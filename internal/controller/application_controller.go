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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/rv2023/DeployKubeStack/api/v1alpha1"
)

var AppScheme *runtime.Scheme

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.deploykubestack.com,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.deploykubestack.com,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.deploykubestack.com,resources=applications/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=create;get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=create;get;list;watch;update;patch;delete

// Reconcile reconciles an Application by creating and managing its child resources.
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the Application CR
	app := &appsv1alpha1.Application{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		log.Error(err, "unable to fetch Application")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Store scheme for use in reconcilers
	AppScheme = r.Scheme

	// Update status to Provisioning
	app.Status.Phase = "Provisioning"
	if err := r.Status().Update(ctx, app); err != nil {
		log.Error(err, "failed to update Application status to Provisioning")
		return ctrl.Result{}, err
	}

	// Reconcile Deployment
	if err := ReconcileDeployment(ctx, app, r); err != nil {
		log.Error(err, "failed to reconcile Deployment")
		app.Status.Phase = "Degraded"
		app.Status.Message = "Failed to create Deployment"
		r.Status().Update(ctx, app) // nolint:errcheck
		return ctrl.Result{}, err
	}
	app.Status.DeploymentReady = true

	// Reconcile Service
	if err := ReconcileService(ctx, app, r); err != nil {
		log.Error(err, "failed to reconcile Service")
		app.Status.Phase = "Degraded"
		app.Status.Message = "Failed to create Service"
		r.Status().Update(ctx, app) // nolint:errcheck
		return ctrl.Result{}, err
	}
	app.Status.ServiceReady = true

	// Update status to Ready
	app.Status.Phase = "Ready"
	app.Status.Message = "Application deployed successfully"
	if err := r.Status().Update(ctx, app); err != nil {
		log.Error(err, "failed to update Application status to Ready")
		return ctrl.Result{}, err
	}

	log.Info("Application reconciliation completed successfully", "name", app.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Named("application").
		Complete(r)
}
