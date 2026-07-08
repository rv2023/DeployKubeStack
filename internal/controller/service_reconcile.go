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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/rv2023/DeployKubeStack/api/v1alpha1"
)

// ReconcileService creates or updates the Service for an Application.
func ReconcileService(ctx context.Context, app *appsv1alpha1.Application, r *ApplicationReconciler) error {
	log := logf.FromContext(ctx)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		return mutateService(service, app)
	})

	if err != nil {
		log.Error(err, "failed to reconcile Service", "op", op)
		return err
	}

	log.V(1).Info("reconciled Service", "operation", op)
	return nil
}

// mutateService sets the desired state of the Service.
func mutateService(service *corev1.Service, app *appsv1alpha1.Application) error {
	// Set owner reference for garbage collection
	if err := controllerutil.SetControllerReference(app, service, AppScheme); err != nil {
		return err
	}

	// Set labels
	if service.Labels == nil {
		service.Labels = make(map[string]string)
	}
	service.Labels["app"] = app.Name
	service.Labels["managed-by"] = "deploykubestack"

	// Set spec
	service.Spec = corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP,
		Selector: map[string]string{
			"app": app.Name,
		},
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Port:       app.Spec.Port,
				TargetPort: intstr.FromInt32(app.Spec.Port),
				Protocol:   corev1.ProtocolTCP,
			},
		},
	}

	return nil
}
