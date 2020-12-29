package rvnodegen

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/bryanl/rv-node-gen/internal/log"
)

// Server is the node gen server.
type Server struct {
	addr           string
	kubeConfigPath string
}

// NewServer creates an instance of Server.
func NewServer(kubeConfigPath, addr string) *Server {
	s := &Server{
		addr:           addr,
		kubeConfigPath: kubeConfigPath,
	}
	return s
}

// Run runs the server.
func (s *Server) Run(ctx context.Context) error {
	logger := log.From(ctx)

	restConfig, err := initRestConfig(s.kubeConfigPath)
	if err != nil {
		return fmt.Errorf("initialize REST config: %w", err)
	}

	restConfig.QPS = 200
	restConfig.Burst = 400

	client, err := NewClient(restConfig, DiscoveryCacheDir("/Users/bryanl/.kube/cache/discovery"))
	if err != nil {
		return fmt.Errorf("initialize cluster client: %w", err)
	}

	logger.Info("Initializing informer manager")
	informerManager, err := NewInformerManager(client)
	if err != nil {
		return fmt.Errorf("create informer factory: %w", err)
	}
	logger.Info("Informer initialized")

	api := NewAPI(informerManager.Lister())

	srv := &http.Server{
		Addr:    s.addr,
		Handler: api.Handler(ctx),
		BaseContext: func(listener net.Listener) context.Context {
			return log.With(ctx, logger)
		},
		ErrorLog:     logger.StdLogger(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "HTTP listener failed")
			os.Exit(1)
		}
	}()

	logger.Info("HTTP server starting", "addr", s.addr)

	<-done
	logger.Info("HTTP server stopping")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http: %w", err)
	}
	logger.Info("HTTP server stopped")

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
