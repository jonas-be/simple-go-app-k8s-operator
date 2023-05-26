/*
Copyright 2023.

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

package controllers

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	simplegoappk8soperatorv1 "jonasbe.de/simple-go-app-k8s-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// SimpleGoAppDeploymentReconciler reconciles a SimpleGoAppDeployment object
type SimpleGoAppDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=simple-go-app-k8s-operator.jonasbe.de,resources=simplegoappdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=simple-go-app-k8s-operator.jonasbe.de,resources=simplegoappdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=simple-go-app-k8s-operator.jonasbe.de,resources=simplegoappdeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=deployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SimpleGoAppDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *SimpleGoAppDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling SimpleGoAppDeployment")
	simpleGoAppDeployment := &simplegoappk8soperatorv1.SimpleGoAppDeployment{}
	if err := r.Get(ctx, req.NamespacedName, simpleGoAppDeployment); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if !simpleGoAppDeployment.DeletionTimestamp.IsZero() {
		if err := r.reconcileDelete(ctx, simpleGoAppDeployment); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if err := r.reconcileEnsure(ctx, simpleGoAppDeployment); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *SimpleGoAppDeploymentReconciler) reconcileEnsure(ctx context.Context, simpleGoAppDeployment *simplegoappk8soperatorv1.SimpleGoAppDeployment) error {
	if controllerutil.AddFinalizer(simpleGoAppDeployment, simplegoappk8soperatorv1.SimpleGoAppFinalizer) {
		if err := r.Update(ctx, simpleGoAppDeployment); err != nil {
			return err
		}
	}

	if err := r.creatOrUpdateBackendDeployment(ctx, simpleGoAppDeployment); err != nil {
		return err
	}
	if err := r.createOrUpdateBackendService(ctx, simpleGoAppDeployment); err != nil {
		return err
	}
	if err := r.createOrUpdateFrontendDeployment(ctx, simpleGoAppDeployment); err != nil {
		return err
	}
	if err := r.createOrUpdateFrontendService(ctx, simpleGoAppDeployment); err != nil {
		return err
	}
	return nil
}

func (r *SimpleGoAppDeploymentReconciler) reconcileDelete(ctx context.Context, simpleGoAppDeployment *simplegoappk8soperatorv1.SimpleGoAppDeployment) error {
	if err := r.deleteFrontendService(ctx, simpleGoAppDeployment); err != nil {
		return err
	}
	if err := r.deleteFrontendDeployment(ctx, simpleGoAppDeployment); err != nil {
		return err
	}
	if err := r.deleteBackendService(ctx, simpleGoAppDeployment); err != nil {
		return err
	}
	if err := r.deleteBackendDeployment(ctx, simpleGoAppDeployment); err != nil {
		return err
	}
	if controllerutil.RemoveFinalizer(simpleGoAppDeployment, simplegoappk8soperatorv1.SimpleGoAppFinalizer) {
		if err := r.Update(ctx, simpleGoAppDeployment); err != nil {
			return err
		}
	}
	return nil
}

func (r *SimpleGoAppDeploymentReconciler) deleteBackendDeployment(ctx context.Context, simpleGoAppDeployment *simplegoappk8soperatorv1.SimpleGoAppDeployment) error {
	key := client.ObjectKey{Namespace: simpleGoAppDeployment.Namespace, Name: simpleGoAppDeployment.Name + "-backend"}
	deployment := &appsv1.Deployment{}

	if err := r.Get(ctx, key, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if controllerutil.RemoveFinalizer(deployment, simplegoappk8soperatorv1.SimpleGoAppFinalizer+"/backend") {
		if err := r.Update(ctx, deployment); err != nil {
			return err
		}
	}

	err := r.Delete(ctx, deployment)
	if err != nil {
		return err
	}
	return nil
}

func (r *SimpleGoAppDeploymentReconciler) deleteBackendService(ctx context.Context, simpleGoAppDeployment *simplegoappk8soperatorv1.SimpleGoAppDeployment) error {
	key := client.ObjectKey{Namespace: simpleGoAppDeployment.Namespace, Name: simpleGoAppDeployment.Name + "-backend"}
	service := &v1.Service{}

	if err := r.Get(ctx, key, service); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if controllerutil.RemoveFinalizer(service, simplegoappk8soperatorv1.SimpleGoAppFinalizer+"/backend-service") {
		if err := r.Update(ctx, service); err != nil {
			return err
		}
	}

	err := r.Delete(ctx, service)
	if err != nil {
		return err
	}
	return nil
}

func (r *SimpleGoAppDeploymentReconciler) deleteFrontendDeployment(ctx context.Context, simpleGoAppDeployment *simplegoappk8soperatorv1.SimpleGoAppDeployment) error {
	key := client.ObjectKey{Namespace: simpleGoAppDeployment.Namespace, Name: simpleGoAppDeployment.Name + "-frontend"}
	deployment := &appsv1.Deployment{}

	if err := r.Get(ctx, key, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if controllerutil.RemoveFinalizer(deployment, simplegoappk8soperatorv1.SimpleGoAppFinalizer+"/frontend") {
		if err := r.Update(ctx, deployment); err != nil {
			return err
		}
	}

	err := r.Delete(ctx, deployment)
	if err != nil {
		return err
	}
	return nil
}

func (r SimpleGoAppDeploymentReconciler) deleteFrontendService(ctx context.Context, simpleGoAppDeployment *simplegoappk8soperatorv1.SimpleGoAppDeployment) error {
	key := client.ObjectKey{Namespace: simpleGoAppDeployment.Namespace, Name: simpleGoAppDeployment.Name + "-frontend"}
	service := &v1.Service{}

	if err := r.Get(ctx, key, service); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if controllerutil.RemoveFinalizer(service, simplegoappk8soperatorv1.SimpleGoAppFinalizer+"/frontend-service") {
		if err := r.Update(ctx, service); err != nil {
			return err
		}
	}

	err := r.Delete(ctx, service)
	if err != nil {
		return err
	}
	return nil
}

func (r *SimpleGoAppDeploymentReconciler) creatOrUpdateBackendDeployment(ctx context.Context,
	simpleGoAppDeployment *simplegoappk8soperatorv1.SimpleGoAppDeployment) error {
	repl := int32(simpleGoAppDeployment.Spec.Replicas)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      simpleGoAppDeployment.Name + "-backend",
			Namespace: simpleGoAppDeployment.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		controllerutil.AddFinalizer(deployment, simplegoappk8soperatorv1.SimpleGoAppFinalizer+"/backend")
		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: &repl,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": simpleGoAppDeployment.Name + "-backend",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": simpleGoAppDeployment.Name + "-backend",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "simple-go-app-backend",
							Image: "jonasbe25/simple-go-app-backend:latest",
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 50051,
									Protocol:      v1.ProtocolTCP,
								},
							},
							Env: []v1.EnvVar{
								{Name: "CONTENT", Value: simpleGoAppDeployment.Spec.ReturnValue},
							},
						},
					},
				},
			},
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (r *SimpleGoAppDeploymentReconciler) createOrUpdateBackendService(ctx context.Context, simpleGoAppDeployment *simplegoappk8soperatorv1.SimpleGoAppDeployment) error {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      simpleGoAppDeployment.Name + "-backend",
			Namespace: simpleGoAppDeployment.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		controllerutil.AddFinalizer(service, simplegoappk8soperatorv1.SimpleGoAppFinalizer+"/backend-service")
		service.Spec = v1.ServiceSpec{
			Selector: map[string]string{
				"app": simpleGoAppDeployment.Name + "-backend",
			},
			Ports: []v1.ServicePort{
				{
					Port: 50051,
				},
			},
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (r *SimpleGoAppDeploymentReconciler) createOrUpdateFrontendDeployment(ctx context.Context, simpleGoAppDeployment *simplegoappk8soperatorv1.SimpleGoAppDeployment) error {
	repl := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      simpleGoAppDeployment.Name + "-frontend",
			Namespace: simpleGoAppDeployment.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		controllerutil.AddFinalizer(deployment, simplegoappk8soperatorv1.SimpleGoAppFinalizer+"/frontend")
		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: &repl,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": simpleGoAppDeployment.Name + "-frontend",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": simpleGoAppDeployment.Name + "-frontend",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "simple-go-app-frontend",
							Image: "jonasbe25/simple-go-app-web:latest",
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: []v1.EnvVar{
								{
									Name: "TARGET",
									Value: fmt.Sprintf("%s-backend.%s.svc.cluster.local:50051",
										simpleGoAppDeployment.Name,
										simpleGoAppDeployment.Namespace),
								},
							},
						},
					},
				},
			},
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (r SimpleGoAppDeploymentReconciler) createOrUpdateFrontendService(ctx context.Context, simpleGoAppDeployment *simplegoappk8soperatorv1.SimpleGoAppDeployment) error {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      simpleGoAppDeployment.Name + "-frontend",
			Namespace: simpleGoAppDeployment.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		controllerutil.AddFinalizer(service, simplegoappk8soperatorv1.SimpleGoAppFinalizer+"/frontend-service")
		service.Spec = v1.ServiceSpec{
			Selector: map[string]string{
				"app": simpleGoAppDeployment.Name + "-frontend",
			},
			Type: "NodePort",
			Ports: []v1.ServicePort{
				{
					Port:     8080,
					NodePort: simpleGoAppDeployment.Spec.NodePort,
				},
			},
		}
		return nil
	}); err != nil {
		return err
	}
	return nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleGoAppDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&simplegoappk8soperatorv1.SimpleGoAppDeployment{}).
		Complete(r)
}
