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

package test1

import (
	"context"

	kapps "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	test1v1alpha1 "github.com/etesami/sample-k8s-ctrl/api/test1/v1alpha1"
	"github.com/go-logr/logr"
)

// SimpleDeploymentReconciler reconciles a SimpleDeployment object
type SimpleDeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=test1.samples-k8s-ctrl.github.com,resources=simpledeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=test1.samples-k8s-ctrl.github.com,resources=simpledeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=test1.samples-k8s-ctrl.github.com,resources=simpledeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SimpleDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *SimpleDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	log := r.Log.WithValues("simpleDeployment", req.NamespacedName)

	var simpleDeployment test1v1alpha1.SimpleDeployment
	klog.Infof("Fetching the simpledeployment")
	if err := r.Get(ctx, req.NamespacedName, &simpleDeployment); err != nil {
		log.Error(err, "unable to fetch SimpleDeployment")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	deployment := &kapps.Deployment{}
	// Set the information you care about
	deployment.ObjectMeta.Namespace = simpleDeployment.Namespace
	deployment.ObjectMeta.Name = simpleDeployment.Name
	deployment.Spec.Replicas = simpleDeployment.Spec.Replicas
	deployment.Spec.Template = *simpleDeployment.Spec.Template
	deployment.Spec.Selector = simpleDeployment.Spec.Selector
	deployment.Spec.Template.ObjectMeta.Labels = simpleDeployment.Spec.MyMetadata.Labels

	// var labelSelector v1.LabelSelector
	// mapString := map[string]interface{}{
	// 	"matchLabels": map[string]string{
	// 		"app": "nginx",
	// 	},
	// }
	// // Convert the nested map to a map[string]string
	// matchLabels, ok := mapString["matchLabels"].(map[string]string)
	// if !ok {
	// 	fmt.Println("Error: Invalid structure1")
	// 	return ctrl.Result{}, nil
	// }
	// if err := v1.Convert_Map_string_To_string_To_v1_LabelSelector(&matchLabels, &labelSelector, nil); err != nil {
	// 	log.Error(err, "unable to convert jsonString to labelSelector: matchLabels")
	// 	return ctrl.Result{}, err
	// } else {
	// 	deployment.Spec.Selector = &labelSelector
	// }

	klog.Infof("Calling SetControllerReference")
	if err := controllerutil.SetControllerReference(&simpleDeployment, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	foundDeployment := &kapps.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		klog.Infof("Creating Deployment", "deployment", deployment.Name)
		err = r.Create(ctx, deployment)
	} else if err == nil {
		if foundDeployment.Spec.Replicas != deployment.Spec.Replicas {
			foundDeployment.Spec.Replicas = deployment.Spec.Replicas
			klog.Info("Updating Deployment", "deployment", deployment.Name)
			err = r.Update(ctx, foundDeployment)
		}
	} else {
		klog.Error("88888888888888 Something else happened")
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&test1v1alpha1.SimpleDeployment{}).
		// Owns(&kapps.Deployment{}).
		Complete(r)
}
