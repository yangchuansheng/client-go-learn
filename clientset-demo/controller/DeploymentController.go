package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yaml2 "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"log"
)

type DeploymentController struct{}

func (receiver *DeploymentController) ListDeployments(clientset *kubernetes.Clientset, namespace string) (*appsv1.DeploymentList, error) {
	log.Printf("Listing Deployments in namespace %q:\n", namespace)
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	list, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})

	return list, err
}

func (receiver *DeploymentController) ApplyDeployment(clientset *kubernetes.Clientset, namespace string) (*appsv1.Deployment, error) {
	log.Println("Creating Deployment...")
	var (
		data []byte
		err  error
	)
	if data, err = ioutil.ReadFile("manifests/nginx.yaml"); err != nil {
		log.Println(err)
	}
	if data, err = yaml2.ToJSON(data); err != nil {
		log.Println(err)
	}
	deployment := &appsv1.Deployment{}
	if err := json.Unmarshal(data, deployment); err != nil {
		log.Println(err)
	}
	result, err := clientset.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})

	return result, err
}

func (receiver *DeploymentController) UpdateDeployment(clientset *kubernetes.Clientset, namespace, name string) error {
	log.Println("Updating Deployment...")
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, getErr := deploymentsClient.Get(context.TODO(), name, metav1.GetOptions{})
		if errors.IsNotFound(getErr) {
			log.Printf("Deployment %v not found.", name)
		} else if statusError, isError := getErr.(*errors.StatusError); isError == true {
			log.Printf("Error geting deployment status of %v: %v", name, statusError.ErrStatus.Message)
		} else if getErr != nil {
			panic(fmt.Errorf("failed to get the latest version of Deployment(nginx): %v", getErr))
		}
		/*
			replicas := int32(1)
			result.Spec.Replicas = &replicas

			result.Spec.Replicas = pointer.Int32Ptr(1)
		*/
		*result.Spec.Replicas = 1
		result.Spec.Template.Spec.Containers[0].Image = "nginx:alpine"
		_, updateErr := deploymentsClient.Update(context.TODO(), result, metav1.UpdateOptions{})

		return updateErr
	})

	if retryErr != nil {
		return retryErr
	}

	log.Println("Updated deployment...")
	return nil
}
