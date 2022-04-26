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
	"strconv"
)

type DeploymentController struct{}

type Result struct {
	No                string
	Name              string
	Namespace         string
	Replicas          string
	Image             string
	CreationTimestamp string
}

func (receiver *DeploymentController) ListDeployments(clientset *kubernetes.Clientset, namespace string) (*appsv1.DeploymentList, []Result, error) {
	log.Printf("Listing Deployments in namespace %q:\n", namespace)
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	deploymentsList, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	res := make([]Result, 0)
	for i := 0; i < len(deploymentsList.Items); i++ {
		deployment := deploymentsList.Items[i]
		res = append(res, Result{
			No:                strconv.Itoa(i),
			Name:              deployment.Name,
			Namespace:         deployment.Namespace,
			Replicas:          strconv.Itoa(int(*deployment.Spec.Replicas)),
			Image:             deployment.Spec.Template.Spec.Containers[0].Image,
			CreationTimestamp: deployment.CreationTimestamp.String(),
		})
	}

	return deploymentsList, res, err
}

func (receiver *DeploymentController) ApplyDeployment(clientset *kubernetes.Clientset, namespace string) (*appsv1.Deployment, error) {
	log.Println("Creating or Updating Deployment...")
	var (
		data   []byte
		err    error
		result *appsv1.Deployment
	)
	// 读取 YAML 文件
	if data, err = ioutil.ReadFile("manifests/deployment.yaml"); err != nil {
		panic(err)
	}
	// YAML 转 JSON
	if data, err = yaml2.ToJSON(data); err != nil {
		panic(err)
	}
	// JSON 转 struct
	deployment := &appsv1.Deployment{}
	if err := json.Unmarshal(data, deployment); err != nil {
		panic(err)
	}

	deployments := clientset.AppsV1().Deployments(namespace)
	if _, err = deployments.Get(context.TODO(), deployment.Name, metav1.GetOptions{}); err != nil {
		if result, err = deployments.Create(context.Background(), deployment, metav1.CreateOptions{}); err != nil {
			log.Fatal("err:\t", err)
		}
		fmt.Printf("Created Deployment %q.\n", deployment.GetObjectMeta().GetName())
	} else {
		if result, err = deployments.Update(context.Background(), deployment, metav1.UpdateOptions{}); err != nil {
			log.Fatal("err:\t", err)
		}
		fmt.Printf("Updated Deployment %q.\n", deployment.GetObjectMeta().GetName())
	}

	return result, err
}

func (receiver *DeploymentController) UpdateDeployment(clientset *kubernetes.Clientset, namespace, name string) error {
	log.Println("Updating Deployment...")
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	//    You have two options to Update() this Deployment:
	//
	//    1. Modify the "deployment" variable and call: Update(deployment).
	//       This works like the "kubectl replace" command and it overwrites/loses changes
	//       made by other clients between you Create() and Update() the object.
	//    2. Modify the "result" returned by Get() and retry Update(result) until
	//       you no longer get a conflict error. This way, you can preserve changes made
	//       by other clients between Create() and Update(). This is implemented below
	//			 using the retry utility package included with client-go. (RECOMMENDED)
	//
	// More Info:
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
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
