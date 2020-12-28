package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/bryanl/rv-node-gen/pkg/rvnodegen"
)

type options struct {
	kubeConfigPath string
	httpAddr       string
}

func main() {
	o := options{}

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&o.kubeConfigPath, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&o.kubeConfigPath, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.StringVar(&o.httpAddr, "addr", ":8181", "HTTP listen address")
	flag.Parse()

	if err := run(o); err != nil {
		log.Printf(err.Error())
		os.Exit(1)
	}
}

func run(o options) error {
	restConfig, err := initRestConfig(o.kubeConfigPath)
	if err != nil {
		return fmt.Errorf("initialize REST config: %w", err)
	}

	restConfig.QPS = 200
	restConfig.Burst = 400

	client, err := rvnodegen.NewClient(restConfig, rvnodegen.DiscoveryCacheDir("/Users/bryanl/.kube/cache/discovery"))
	if err != nil {
		return fmt.Errorf("initialize cluster client: %w", err)
	}

	log.Print("Initializing informer manager")
	informerManager, err := rvnodegen.NewInformerManager(client)
	if err != nil {
		return fmt.Errorf("create informer factory: %w", err)
	}
	log.Println("Informer initialized")

	r := mux.NewRouter()
	r.Use(configureCORS)

	r.Handle("/v1/nodes", rvnodegen.NewNodeHandler(informerManager)).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:    o.httpAddr,
		Handler: r,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listener for HTTP server failed: %s", err)
			os.Exit(1)
		}
	}()

	log.Printf("HTTP server listening at %s", o.httpAddr)

	<-done
	log.Print("HTTP server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http: %w", err)
	}

	return nil
}

func configureCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers:", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
		return
	})
}

func initRestConfig(kubeConfigPath string) (*restclient.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
}
