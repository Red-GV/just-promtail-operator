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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loggingv1 "github.com/huikang/promtail-operator/apis/logging/v1"
)

var (
	defautlPromtailImage = "docker.io/grafana/promtail:2.1.0"
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

	// Fetch the promtail instance
	promtail := &loggingv1.Promtail{}
	err = r.Get(ctx, req.NamespacedName, promtail)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			r.Log.Info("Promtail resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		r.Log.Error(err, "Failed to get Promtail")
		return ctrl.Result{}, err
	}

	// Check if the daemonset already exists, if not create a new one
	found := &appsv1.DaemonSet{}
	err = r.Get(ctx, types.NamespacedName{Name: promtail.Name, Namespace: promtail.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new deployment
		de := r.daemonsetForMemcached(promtail)
		r.Log.Info("Creating a new DaemonSet", "Daemonset.Namespace", de.Namespace, "Deployment.Name", de.Name)
		err = r.Create(ctx, de)
		if err != nil {
			r.Log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", de.Namespace, "Deployment.Name", de.Name)
			return ctrl.Result{}, err
		}
		// Daemon created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PromtailReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loggingv1.Promtail{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *PromtailReconciler) daemonsetForMemcached(p *loggingv1.Promtail) *appsv1.DaemonSet {
	ls := labelsForMemcached(p.Name)
	var imageName string
	if p.Spec.Image == "" {
		imageName = defautlPromtailImage
	}
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: imageName,
						Name:  "promtail",
						// Command: []string{"promtail", "-m=64", "-o", "modern", "-v"},
						Args: []string{"-config.file=/etc/promtail/promtail.yaml"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 3101,
							Name:          "http-metrics",
						}},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(p, ds, r.Scheme)
	return ds
}

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForMemcached(name string) map[string]string {
	return map[string]string{"app": "promtail", "memcached_cr": name}
}
