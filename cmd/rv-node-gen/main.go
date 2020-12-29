package main

import (
	"context"
	"flag"
	golog "log"
	"os"
	"path/filepath"

	"k8s.io/client-go/util/homedir"

	"github.com/bryanl/rv-node-gen/internal/log"
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
		golog.Printf(err.Error())
		os.Exit(1)
	}
}

func run(o options) error {
	logger := log.New()
	ctx := log.With(context.Background(), logger)

	server := rvnodegen.NewServer(o.kubeConfigPath, o.httpAddr)
	return server.Run(ctx)
}
