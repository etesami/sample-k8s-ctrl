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

	"github.com/go-logr/logr"
	kapps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields" // Required for Watching
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"       // Required for Watching
	ctrl "sigs.k8s.io/controller-runtime" // Required for Watching
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil" // Required for Watching
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log" // Required for Watching
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile" // Required for Watching

	test1v1alpha1 "github.com/etesami/sample-k8s-ctrl/api/test1/v1alpha1"
)

const (
	configMapField = ".spec.configMap"
)

// ConfigDeploymentReconciler reconciles a ConfigDeployment object
type ConfigDeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=test1.samples-k8s-ctrl.github.com,resources=configdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=test1.samples-k8s-ctrl.github.com,resources=configdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=test1.samples-k8s-ctrl.github.com,resources=configdeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *ConfigDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// log := r.Log.WithValues("configDeployment", req.NamespacedName)

	log.Info("---------- ---------- ---------- ---------- ---------- ---------- ---------- Reconciling", "[OBJECT]", req)
	var configDeployment test1v1alpha1.ConfigDeployment
	if err := r.Get(ctx, req.NamespacedName, &configDeployment); err != nil {
		log.Error(err, "Unable to fetch ConfigDeployment")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, nil
	}

	var configMapVersion string
	if configDeployment.Spec.ConfigMap != "" {
		foundConfigMap := &corev1.ConfigMap{}
		err := r.Get(ctx, types.NamespacedName{
			Name: configDeployment.Spec.ConfigMap, Namespace: configDeployment.Namespace}, foundConfigMap)
		if err != nil {
			// If a configMap name is provided, then it must exist
			// You will likely want to create an Event for the user to understand why their reconcile is failing.
			log.Error(err, "ConfigMap not found. It is required for a ConfigDeployment object and it does not exist.")
			return ctrl.Result{}, nil
		}

		// Hash the data in some way, or just use the version of the Object
		configMapVersion = foundConfigMap.ResourceVersion
		log.Info("ConfigMap New: ", "[Version]", configMapVersion)
	}

	// Set the information you care about
	deployment := &kapps.Deployment{}
	deployment.ObjectMeta.Namespace = configDeployment.Namespace
	deployment.ObjectMeta.Name = configDeployment.Name + "-deploymentonly"
	deployment.Spec.Template = *configDeployment.Spec.Template
	deployment.Spec.Selector = configDeployment.Spec.Selector
	deployment.Spec.Replicas = configDeployment.Spec.Replicas
	deployment.Spec.Template.ObjectMeta.Labels = configDeployment.Spec.MyMetadata.Labels
	deployment.ObjectMeta.Annotations = map[string]string{"configMapVersion": configMapVersion}

	if err := controllerutil.SetControllerReference(&configDeployment, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	foundDeployment := &kapps.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Deployment", "deployment", deployment.Name)
		err = r.Create(ctx, deployment)
	} else if err == nil {
		// We are about to update the found deployment (foundDeployment)
		// with the new information (deployment)
		foundDeployment.Spec.Replicas = deployment.Spec.Replicas
		foundDeployment.Spec.Template = deployment.Spec.Template
		foundDeployment.Spec.Selector = deployment.Spec.Selector
		foundDeployment.Spec.Template.ObjectMeta.Labels = deployment.Spec.Template.ObjectMeta.Labels

		var needUpdatePods bool = false
		// Check to see if the reconcile was triggered by a change in the ConfigMap
		if foundDeployment.ObjectMeta.Annotations["configMapVersion"] != configMapVersion {
			log.Info("--- --- ---> ConfigMap has changed",
				"[foundDeployment]", foundDeployment.ObjectMeta.Annotations["configMapVersion"],
				"[configMapVersion]", configMapVersion)
			foundDeployment.ObjectMeta.Annotations["configMapVersion"] = configMapVersion
			// If the ConfigMap has changed, then we need to update the pods
			needUpdatePods = true
		}

		log.Info("Start updating Deployment")
		// Update the deployment with the new information
		// Note: This does not force the pods to be re-created
		// If the template has not changed this will not do anything
		if err := r.Update(ctx, foundDeployment); err != nil {
			log.Error(err, "Failed to update Deployment")
			return ctrl.Result{}, err
		} else {
			log.Info("Resource updated.")
		}

		if needUpdatePods {
			// Force to re-create pods
			podList := &corev1.PodList{}
			err = r.List(ctx, podList,
				client.InNamespace(deployment.Namespace),
				client.MatchingLabels{"app": deployment.Spec.Template.ObjectMeta.Labels["app"]})
			if err != nil {
				log.Error(err, "Failed to list pods")
				return ctrl.Result{}, err
			}

			for _, pod := range podList.Items {
				if err := r.Delete(ctx, &pod); err != nil {
					log.Error(err, "failed to delete pod", "podName", pod.Name)
					return ctrl.Result{}, err
				}
				log.Info("Delete the exisiting pod", "[NAME]", pod.Name)
			}
		}
	} else {
		log.Error(err, "Failed. Some other error.")
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// configMapField = ".spec.configMap"
	// IndexField(ctx context.Context, obj Object, field string, extractValue IndexerFunc)
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(), &test1v1alpha1.ConfigDeployment{}, configMapField,
		func(rawObj client.Object) []string {
			// Extract the ConfigMap name from the ConfigDeployment Spec, if one is provided
			configDeployment := rawObj.(*test1v1alpha1.ConfigDeployment)
			if configDeployment.Spec.ConfigMap == "" {
				return nil
			}
			return []string{configDeployment.Spec.ConfigMap}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&test1v1alpha1.ConfigDeployment{}).
		Owns(&kapps.Deployment{}).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForConfigMap),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *ConfigDeploymentReconciler) findObjectsForConfigMap(ctx context.Context, configMap client.Object) []reconcile.Request {
	attachedConfigDeployments := &test1v1alpha1.ConfigDeploymentList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(configMapField, configMap.GetName()),
		Namespace:     configMap.GetNamespace(),
	}
	err := r.List(context.TODO(), attachedConfigDeployments, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedConfigDeployments.Items))
	for i, item := range attachedConfigDeployments.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}
