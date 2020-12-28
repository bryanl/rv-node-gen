package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/bryanl/rv-node-gen/pkg/rvnodegen"
)

func main() {
	var kubeConfigPath string
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeConfigPath, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeConfigPath, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	if err := run(kubeConfigPath); err != nil {
		log.Printf(err.Error())
		os.Exit(1)
	}
}

func run(kubeConfigPath string) error {
	restConfig, err := initRestConfig(kubeConfigPath)
	if err != nil {
		return fmt.Errorf("initialize REST config: %w", err)
	}

	restConfig.QPS = 200
	restConfig.Burst = 400

	client, err := rvnodegen.NewClient(restConfig, rvnodegen.DiscoveryCacheDir("/Users/bryanl/.kube/cache/discovery"))
	if err != nil {
		return fmt.Errorf("initialize cluster client: %w", err)
	}

	informerManager, err := rvnodegen.NewInformerManager(client)
	if err != nil {
		return fmt.Errorf("create informer factory: %w", err)
	}

	gvk := schema.GroupVersionKind{
		Version: "v1",
		Kind:    "Pod",
	}

	gvr, err := informerManager.Resource(gvk)
	if err != nil {
		return fmt.Errorf("get resource for gvk (%s): %w", gvk, err)
	}

	objects, err := informerManager.Lister(gvr).ByNamespace("default").List(labels.Everything())
	if err != nil {
		return fmt.Errorf("list pods: %w", err)
	}

	list, err := toUnstructuredSlice(objects)
	if err != nil {
		return err
	}

	emitter := rvnodegen.NewNodeEmitter()
	visitor := rvnodegen.NewVisitor(emitter, informerManager)

	if err := visitor.Visit(list...); err != nil {
		return fmt.Errorf("visit objects: %w", err)
	}

	spew.Dump(emitter.Nodes())

	return nil
}

func initRestConfig(kubeConfigPath string) (*restclient.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
}

func toUnstructuredSlice(in []runtime.Object) ([]*unstructured.Unstructured, error) {
	var out []*unstructured.Unstructured

	for i := range in {
		object, ok := in[i].(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("object is a %T", in[i])
		}

		out = append(out, object)
	}

	return out, nil
}
