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
	"github.com/go-logr/logr"
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

	simpleGoAppDeployment := &simplegoappk8soperatorv1.SimpleGoAppDeployment{}
	if err := r.Get(ctx, req.NamespacedName, simpleGoAppDeployment); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		logger.Info("SimpleGoAppDeployment was deleted")
		if err := r.reconcileDelete(ctx, req); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if err := r.creatOrUpdateBackendDeployment(ctx, simpleGoAppDeployment, logger); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SimpleGoAppDeploymentReconciler) reconcileDelete(ctx context.Context, req ctrl.Request) error {
	if err := r.deleteBackendDeployment(ctx, req.Namespace); err != nil {
		return err
	}
	return nil
}

func (r *SimpleGoAppDeploymentReconciler) deleteBackendDeployment(ctx context.Context, namespace string) error {
	key := client.ObjectKey{Namespace: namespace, Name: "simple-go-app-backend"}
	deployment := &appsv1.Deployment{}

	if err := r.Get(ctx, key, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	controllerutil.RemoveFinalizer(deployment, simplegoappk8soperatorv1.SimpleGoAppFinalizer+"/backend")
	if err := r.Update(ctx, deployment); err != nil {
		return err
	}

	err := r.Delete(ctx, deployment)
	if err != nil {
		return err
	}
	return nil
}

func (r *SimpleGoAppDeploymentReconciler) creatOrUpdateBackendDeployment(ctx context.Context,
	simpleGoAppDeployment *simplegoappk8soperatorv1.SimpleGoAppDeployment,
	logger logr.Logger) error {
	repl := int32(simpleGoAppDeployment.Spec.Replicas)

	logger.Info("Creating or updating backend deployment")
	logger.Info("Replicas: ", "replicas", repl)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "simple-go-app-backend",
			Namespace: simpleGoAppDeployment.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		controllerutil.AddFinalizer(deployment, simplegoappk8soperatorv1.SimpleGoAppFinalizer+"/backend")
		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: &repl,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "simple-go-app-backend",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "simple-go-app-backend",
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

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleGoAppDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&simplegoappk8soperatorv1.SimpleGoAppDeployment{}).
		Complete(r)
}
