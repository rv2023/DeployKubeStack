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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1alpha1 "github.com/rv2023/DeployKubeStack/api/v1alpha1"
)

const (
	timeout  = time.Second * 10
	interval = time.Millisecond * 250
)

var _ = Describe("Application Controller", Ordered, func() {
	Context("when creating an Application with minimal spec", func() {
		It("should create Deployment and Service with defaults", func() {
			ctx := context.Background()
			appName := "test-app-minimal"
			app := &appsv1alpha1.Application{
				TypeMeta: metav1.TypeMeta{
					APIVersion: appsv1alpha1.GroupVersion.String(),
					Kind:       "Application",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: "default",
				},
				Spec: appsv1alpha1.ApplicationSpec{
					Image: "nginx:latest",
					Port:  8080,
				},
			}

			Expect(k8sClient.Create(ctx, app)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: appName, Namespace: "default"}

			// Check Application status
			Eventually(func() string {
				createdApp := &appsv1alpha1.Application{}
				err := k8sClient.Get(ctx, lookupKey, createdApp)
				if err != nil {
					return ""
				}
				return createdApp.Status.Phase
			}, timeout, interval).Should(Equal("Ready"))

			// Check Deployment was created
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, lookupKey, deployment)
			}, timeout, interval).Should(Succeed())

			// Verify Deployment spec
			Expect(*deployment.Spec.Replicas).Should(Equal(int32(1)))
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).Should(Equal("nginx:latest"))
			Expect(deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(int32(8080)))
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String()).Should(Equal("100m"))
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()).Should(Equal("100Mi"))

			// Check Service was created
			service := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, lookupKey, service)
			}, timeout, interval).Should(Succeed())

			// Verify Service spec
			Expect(service.Spec.Type).Should(Equal(corev1.ServiceTypeClusterIP))
			Expect(service.Spec.Ports[0].Port).Should(Equal(int32(8080)))
			Expect(service.Spec.Ports[0].TargetPort.IntVal).Should(Equal(int32(8080)))
		})
	})

	Context("when creating an Application with custom replicas and resources", func() {
		It("should create Deployment with custom values", func() {
			ctx := context.Background()
			appName := "test-app-custom"
			replicas := int32(3)
			app := &appsv1alpha1.Application{
				TypeMeta: metav1.TypeMeta{
					APIVersion: appsv1alpha1.GroupVersion.String(),
					Kind:       "Application",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: "default",
				},
				Spec: appsv1alpha1.ApplicationSpec{
					Image:    "nginx:1.21",
					Port:     8080,
					Replicas: &replicas,
					Resources: &corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, app)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: appName, Namespace: "default"}

			// Check Application status
			Eventually(func() string {
				createdApp := &appsv1alpha1.Application{}
				err := k8sClient.Get(ctx, lookupKey, createdApp)
				if err != nil {
					return ""
				}
				return createdApp.Status.Phase
			}, timeout, interval).Should(Equal("Ready"))

			// Check Deployment was created with custom replicas
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, lookupKey, deployment)
			}, timeout, interval).Should(Succeed())

			Expect(*deployment.Spec.Replicas).Should(Equal(int32(3)))
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).Should(Equal("nginx:1.21"))
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String()).Should(Equal("200m"))
			Expect(deployment.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String()).Should(Equal("500m"))
		})
	})

	Context("when Application is deleted", func() {
		It("should set owner references for garbage collection", func() {
			ctx := context.Background()
			appName := "test-app-owner-ref"
			app := &appsv1alpha1.Application{
				TypeMeta: metav1.TypeMeta{
					APIVersion: appsv1alpha1.GroupVersion.String(),
					Kind:       "Application",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: "default",
				},
				Spec: appsv1alpha1.ApplicationSpec{
					Image: "nginx:latest",
					Port:  8080,
				},
			}

			Expect(k8sClient.Create(ctx, app)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: appName, Namespace: "default"}

			// Verify Deployment exists and has owner reference
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, lookupKey, deployment)
			}, timeout, interval).Should(Succeed())

			// Check that deployment has owner reference to Application
			Eventually(func() bool {
				Expect(k8sClient.Get(ctx, lookupKey, deployment)).Should(Succeed())
				return len(deployment.OwnerReferences) > 0 &&
					deployment.OwnerReferences[0].Kind == "Application" &&
					deployment.OwnerReferences[0].Controller != nil &&
					*deployment.OwnerReferences[0].Controller == true
			}, timeout, interval).Should(BeTrue())

			// Verify Service exists and has owner reference
			service := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, lookupKey, service)
			}, timeout, interval).Should(Succeed())

			// Check that service has owner reference to Application
			Eventually(func() bool {
				Expect(k8sClient.Get(ctx, lookupKey, service)).Should(Succeed())
				return len(service.OwnerReferences) > 0 &&
					service.OwnerReferences[0].Kind == "Application" &&
					service.OwnerReferences[0].Controller != nil &&
					*service.OwnerReferences[0].Controller == true
			}, timeout, interval).Should(BeTrue())
		})
	})
})
