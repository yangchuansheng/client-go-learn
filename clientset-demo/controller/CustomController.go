package controller

import (
	"bytes"
	"context"
	"github.com/pytimer/k8sutil/apply"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"log"
)

type CustomController struct{}

func (receiver *CustomController) CreateResources(ctx context.Context, config *rest.Config) error {
	log.Println("Creating Resources...")
	var (
		data []byte
		err  error
	)
	if data, err = ioutil.ReadFile("manifests/nginx.yaml"); err != nil {
		log.Println(err)
	}

	// Prepare the dynamic client and typed client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(data), 100)
	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}

		if len(rawObj.Raw) == 0 {
			// if the yaml object is empty just continue to the next one
			continue
		}

		// Decode YAML manifest into unstructured.Unstructured
		obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			log.Fatal(err)
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		gr, err := restmapper.GetAPIGroupResources(clientset.Discovery())
		if err != nil {
			panic(err)
		}

		// Find GVR
		mapper := restmapper.NewDiscoveryRESTMapper(gr)
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			panic(err)
		}

		// Obtain REST interface for the GVR
		var dri dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			// namespaced resources should specify the namespace
			if unstructuredObj.GetNamespace() == "" {
				unstructuredObj.SetNamespace("default")
			}
			dri = dyn.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
		} else {
			// for cluster-wide resources
			dri = dyn.Resource(mapping.Resource)
		}

		// Create the object
		if _, err := dri.Create(ctx, unstructuredObj, metav1.CreateOptions{}); err != nil {
			log.Fatal(err)
		}
	}

	return err
}

func (receiver *CustomController) ApplyResources(ctx context.Context, config *rest.Config) error {
	log.Println("Creating or Updating Resources...")
	var (
		data []byte
		err  error
	)
	if data, err = ioutil.ReadFile("manifests/nginx.yaml"); err != nil {
		log.Println(err)
	}

	// Prepare the dynamic client and typed client
	// configController := ConfigController{}
	// config, err := configController.Initkubeconfig()
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err)
	}
	applyOptions := apply.NewApplyOptions(dyn, discoveryClient)
	if err := applyOptions.Apply(context.TODO(), data); err != nil {
		log.Fatalf("apply error: %v", err)
	}

	return err
}