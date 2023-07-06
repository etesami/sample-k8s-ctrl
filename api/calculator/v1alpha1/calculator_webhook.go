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

package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var calculatorlog = logf.Log.WithName("calculator-resource")

func (r *Calculator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-calculator-samples-k8s-ctrl-github-com-v1alpha1-calculator,mutating=true,failurePolicy=fail,sideEffects=None,groups=calculator.samples-k8s-ctrl.github.com,resources=calculators,verbs=create;update,versions=v1alpha1,name=mcalculator.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Calculator{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Calculator) Default() {
	calculatorlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-calculator-samples-k8s-ctrl-github-com-v1alpha1-calculator,mutating=false,failurePolicy=fail,sideEffects=None,groups=calculator.samples-k8s-ctrl.github.com,resources=calculators,verbs=create;update,versions=v1alpha1,name=vcalculator.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Calculator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Calculator) ValidateCreate() (admission.Warnings, error) {
	calculatorlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	klog.Infof("Validate create", "name", r.Name)
	if !isInList([]string{"add", "subtract", "multiply", "divide"}, r.Spec.Operation) {
		return nil, fmt.Errorf("Operation %s is not supported", r.Spec.Operation)
	}
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Calculator) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	calculatorlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Calculator) ValidateDelete() (admission.Warnings, error) {
	calculatorlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func isInList(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}