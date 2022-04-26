package main

import (
	"clientset-demo/controller"
	"clientset-demo/util"
	"context"
	"fmt"
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

	// Apply Deployment from YAML
	util.Prompt()
	_, err = deploymentController.ApplyDeployment(clientset, "default")
	if err != nil {
		panic(err)
	}

	// Update Deployment
	util.Prompt()
	err = deploymentController.UpdateDeployment(clientset, "default", "nginx")
	if err != nil {
		panic(fmt.Errorf("Update deployment nginx failed: %v", err))
	}

	// Delete Deployment
	util.Prompt()
	err = deploymentController.DeleteDeployments(clientset, "default", "nginx")
	if err != nil {
		panic(fmt.Errorf("Delete deployment nginx failed: %v", err))
	}

	// Create or Update any resources from YAML
	util.Prompt()
	err = customController.ApplyResources(context.Background(), config)
	if err != io.EOF {
		log.Fatal("eof ", err)
	}

	// Apply any resources from YAML
	util.Prompt()
	err = customController.AdvancedApplyResources(config)
	if err != nil {
		panic(err)
	}
}
