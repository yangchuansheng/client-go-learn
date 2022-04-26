package main

import (
	"clientset-demo/controller"
	"context"
	"github.com/modood/table"
	"io"
	"k8s.io/client-go/kubernetes"
	"log"
)

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

	// List Deployments
	_, res, err := deploymentController.ListDeployments(clientset, "default")
	if err != nil {
		panic(err)
	}
	table.Output(res)

	/*
		_, err = deploymentController.ApplyDeployment(clientset, "default")
		if err != nil {
			panic(err)
		}

		err = deploymentController.UpdateDeployment(clientset, "default", "nginx1")
		if err != nil {
			panic(fmt.Errorf("update failed: %v", err))
		}

		err = customController.AdvancedApplyResources(config)
		if err != nil {
			panic(err)
		}
	*/
	err = customController.ApplyResources(context.Background(), config)
	if err != io.EOF {
		log.Fatal("eof ", err)
	}
}
