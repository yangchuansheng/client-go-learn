package controller

import (
	"flag"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
)

type ConfigController struct{}

var (
	kubeconfig *string
	config     *rest.Config
)

func (receiver *ConfigController) Initkubeconfig() (*rest.Config, error) {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	var err error
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	return config, err
}

/*
func (receiver *ConfigController) GetClientset() (*kubernetes.Clientset, error) {
	config, err := Initkubeconfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)

	return clientset, err
}

func (receiver *ConfigController) GetDiscoveryClient() (*discovery.DiscoveryClient, error) {
	config, err := Initkubeconfig()
	if err != nil {
		panic(err)
	}
	dc, err := discovery.NewDiscoveryClientForConfig(config)

	return dc, err
}

func (receiver *ConfigController) GetDynamicClient() (dynamic.Interface, error) {
	config, err := Initkubeconfig()
	if err != nil {
		panic(err)
	}
	dyn, err := dynamic.NewForConfig(config)

	return dyn, err
}
*/
