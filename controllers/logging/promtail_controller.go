/*
Copyright 2021.

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

package logging

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loggingv1 "github.com/huikang/promtail-operator/apis/logging/v1"
)

// PromtailReconciler reconciles a Promtail object
type PromtailReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=logging.just-loki.io,resources=promtails,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=logging.just-loki.io,resources=promtails/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=logging.just-loki.io,resources=promtails/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods;nodes;nodes/proxy;services;endpoints,verbs=get;watch;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Promtail object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *PromtailReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("promtail", req.NamespacedName)

	// your logic here
	// Ensure existence of servicesaccount
	sa, err := promtailServiceAccount()
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "initializing node-exporter Service failed")
	}

	if err := createOrUpdateServiceAccount(sa); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "Failed to reconcile ServiceAccount for Elasticsearch cluster")
	}

	// Ensure existence of config maps
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PromtailReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loggingv1.Promtail{}).
		Complete(r)
}
