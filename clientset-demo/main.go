package main

import (
	"clientset-demo/controller"
	"context"
	"fmt"
	"github.com/modood/table"
	"io"
	"k8s.io/client-go/kubernetes"
	"log"
	"strconv"
)

type Result struct {
	No                string
	Name              string
	Namespace         string
	Replicas          string
	Image             string
	CreationTimestamp string
}

var (
	deploymentController = &controller.DeploymentController{}
	customController     = &controller.CustomController{}
)

func main() {
	configController := controller.ConfigController{}
	config, err := configController.Initkubeconfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentList, err := deploymentController.ListDeployments(clientset, "default")
	if err != nil {
		panic(err)
	}

	res := make([]Result, 0)
	for i := 0; i < len(deploymentList.Items); i++ {
		deployment := deploymentList.Items[i]
		res = append(res, Result{
			No:                strconv.Itoa(i),
			Name:              deployment.Name,
			Namespace:         deployment.Namespace,
			Replicas:          strconv.Itoa(int(*deployment.Spec.Replicas)),
			Image:             deployment.Spec.Template.Spec.Containers[0].Image,
			CreationTimestamp: deployment.CreationTimestamp.String(),
		})
	}
	table.Output(res)

	/*
		deployment, err := deploymentController.CreateDeployment(clientset, "default")
		if err != nil {
			panic(err)
		}
		log.Printf("Created Deployment %q.\n", deployment.GetObjectMeta().GetName())

		err = deploymentController.UpdateDeployment(clientset, "default", "nginx1")
		if err != nil {
			panic(fmt.Errorf("update failed: %v", err))
		}
	*/
	results, err := customController.CreateResources(context.Background(), config)
	if err != io.EOF {
		log.Fatal("eof ", err)
	}
	fmt.Printf("Created resource\n %v\n", results)
}
