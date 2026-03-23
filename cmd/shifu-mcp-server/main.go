package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/edgenesis/shifu/pkg/deviceapi"
	"github.com/edgenesis/shifu/pkg/k8s/api/v1alpha1"
	mcpserver "github.com/edgenesis/shifu/pkg/mcp/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var (
		kubeconfig string
		addr       string
	)
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (uses in-cluster config if empty)")
	flag.StringVar(&addr, "addr", ":8443", "Address to listen on")
	flag.Parse()

	config, err := getRestConfig(kubeconfig)
	if err != nil {
		log.Fatalf("Failed to get Kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes clientset: %v", err)
	}

	edClient, err := newEdgeDeviceRestClient(config)
	if err != nil {
		log.Fatalf("Failed to create EdgeDevice REST client: %v", err)
	}

	edgeLister := func(ctx context.Context) ([]v1alpha1.EdgeDevice, error) {
		edList := &v1alpha1.EdgeDeviceList{}
		err := edClient.Get().
			Resource("edgedevices").
			Do(ctx).
			Into(edList)
		if err != nil {
			return nil, fmt.Errorf("listing EdgeDevices: %w", err)
		}
		return edList.Items, nil
	}

	resolver := deviceapi.NewResolver(clientset, edgeLister)
	apiClient := deviceapi.NewClient(resolver)
	server := mcpserver.New(apiClient)

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)

	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)

	log.Printf("Shifu MCP Server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getRestConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func newEdgeDeviceRestClient(config *rest.Config) (*rest.RESTClient, error) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}

	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{
		Group:   v1alpha1.GroupVersion.Group,
		Version: v1alpha1.GroupVersion.Version,
	}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	return rest.UnversionedRESTClientFor(&crdConfig)
}
