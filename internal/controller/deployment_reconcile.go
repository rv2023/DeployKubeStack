/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/rv2023/DeployKubeStack/api/v1alpha1"
)

const deploymentFinalizerName = "deploykubestack.com/deployment-finalizer"

// ReconcileDeployment creates or updates the Deployment for an Application.
func ReconcileDeployment(ctx context.Context, app *appsv1alpha1.Application, r *ApplicationReconciler) error {
	log := logf.FromContext(ctx)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		return mutateDeployment(deployment, app)
	})

	if err != nil {
		log.Error(err, "failed to reconcile Deployment", "op", op)
		return err
	}

	log.V(1).Info("reconciled Deployment", "operation", op)
	return nil
}

// mutateDeployment sets the desired state of the Deployment.
func mutateDeployment(deployment *appsv1.Deployment, app *appsv1alpha1.Application) error {
	// Set owner reference for garbage collection
	if err := controllerutil.SetControllerReference(app, deployment, AppScheme); err != nil {
		return err
	}

	// Set labels
	if deployment.Labels == nil {
		deployment.Labels = make(map[string]string)
	}
	deployment.Labels["app"] = app.Name
	deployment.Labels["managed-by"] = "deploykubestack"

	// Get replicas, use default if not specified
	replicas := int32(1)
	if app.Spec.Replicas != nil {
		replicas = *app.Spec.Replicas
	}

	// Get resource requirements with defaults
	resources := getResourceRequirements(app.Spec.Resources)

	// Set spec
	deployment.Spec = appsv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": app.Name,
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": app.Name,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "app",
						Image: app.Spec.Image,
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: app.Spec.Port,
								Protocol:      corev1.ProtocolTCP,
							},
						},
						Resources: resources,
					},
				},
			},
		},
	}

	return nil
}

// getResourceRequirements returns resource requirements with defaults if not specified.
// Default: 100m CPU, 100Mi memory for both requests and limits.
func getResourceRequirements(resources *corev1.ResourceRequirements) corev1.ResourceRequirements {
	if resources != nil && (len(resources.Requests) > 0 || len(resources.Limits) > 0) {
		return *resources
	}

	// Set defaults
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("100Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("100Mi"),
		},
	}
}
