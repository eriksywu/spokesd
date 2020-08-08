package k8s

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const KUBECONFIG = "KUBECONFIG"

// GetClient returns a k8s clientset to the request from inside of cluster
func GetClientFromEnv() (kubernetes.Interface, error) {
	kubeconfig := os.Getenv(KUBECONFIG)
	fmt.Printf("kubeconfig=%s \n", kubeconfig)
	if kubeconfig == "" {
		return nil, fmt.Errorf("no kubeconfig found")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	fmt.Printf("constructing a client for %s \n", config.Host)
	if err != nil {
		return nil, err
	}
	clientset, _ := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Can not create kubernetes client: %v", err)
	}
	return clientset, nil
}
