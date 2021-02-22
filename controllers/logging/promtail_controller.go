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
	"io/ioutil"
	"path"
	"time"

	"github.com/go-logr/logr"
	// "github.com/grafana/loki/pkg/promtail/config"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// "gopkg.in/yaml.v2"

	loggingv1 "github.com/huikang/promtail-operator/apis/logging/v1"
)

var (
	defautlPromtailImage = "docker.io/grafana/promtail:2.1.0"
	PromtailScrapeConfig = "scrape_config.yaml"
)

// PromtailReconciler reconciles a Promtail object
type PromtailReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	AssetsPath string
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

	// Ensure existence of secret for config, if not create a default one
	if updated, err := r.createOrUpdateScrapeConfigSecret(ctx, promtail); err != nil {
		r.Log.Error(err, "Failed to reconcile secret for scrape configuration")
		return ctrl.Result{}, err
	} else if updated {
		r.Log.Info("secret updated")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Second,
		}, nil
	}
	r.Log.Info("secret reconciled")

	// Check if the daemonset already exists, if not create a new one
	found := &appsv1.DaemonSet{}
	err = r.Get(ctx, types.NamespacedName{Name: promtail.Name, Namespace: promtail.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new deployment
		de := r.daemonsetForPromtail(promtail)
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

func (r *PromtailReconciler) createOrUpdateScrapeConfigSecret(ctx context.Context, promtail *loggingv1.Promtail) (bool, error) {
	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: promtail.Name, Namespace: promtail.Namespace}, secret)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new secret
		se, err := r.secretForPromtail(promtail)
		if se == nil {
			r.Log.Error(err, "Failed to create new secret")
			return false, err
		}
		r.Log.Info("Creating a new secret", "Secret.Namespace", se.Namespace, "Secret.Name", se.Name)
		err = r.Create(ctx, se)
		if err != nil {
			r.Log.Error(err, "Failed to create new secret", "Secret.Namespace", se.Namespace, "Secret.Name", se.Name)
			return false, err
		}

		// secret created successfully - return and requeue with some delay
		return true, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Secret")
		return false, err
	}
	return false, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PromtailReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&loggingv1.Promtail{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

func (r *PromtailReconciler) secretForPromtail(p *loggingv1.Promtail) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
		},
		StringData: map[string]string{},
	}
	filepath := path.Join(r.AssetsPath, PromtailScrapeConfig)
	r.Log.Info("load", "scrape config", filepath)
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		r.Log.Error(err, "failed to read asset", filepath)
		return nil, err
	}

	/*
		TODO (huikang): why the following generated config does not work?
		var config config.Config
		err = yaml.Unmarshal(f, &config)
		if err != nil {
			r.Log.Error(err, "error parsing yaml")
			return nil, err
		}
		strConfig, err := yaml.Marshal(config)
		if err != nil {
			r.Log.Error(err, "error convert promtail config to byte data")
			return nil, err
		}
	*/
	secret.StringData["promtail.yaml"] = string(f)

	ctrl.SetControllerReference(p, secret, r.Scheme)
	return secret, nil
}

func (r *PromtailReconciler) daemonsetForPromtail(p *loggingv1.Promtail) *appsv1.DaemonSet {
	ls := labelsForMemcached(p.Name)
	var imageName string
	if p.Spec.Image == "" {
		imageName = defautlPromtailImage
	}

	// prepare volumnes
	volumes := []corev1.Volume{
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: p.Name,
				},
			},
		},
	}
	VolumeMounts := []corev1.VolumeMount{
		{
			Name:      "config",
			MountPath: "/etc/promtail",
		},
	}

	var appendVolume func(name string, hostPath string, mountPath string)
	appendVolume = func(name string, hostPath string, mountPath string) {
		mountType := corev1.HostPathDirectory
		volume := corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: hostPath,
					Type: &mountType,
				},
			},
		}
		volumes = append(volumes, volume)

		vm := corev1.VolumeMount{
			Name:      name,
			MountPath: mountPath,
			ReadOnly:  true,
		}
		VolumeMounts = append(VolumeMounts, vm)
	}
	// appendVolume("run", "/run/promtail", "/run/promtail")
	appendVolume("pods", "/var/log/pods", "/var/log/pods")
	appendVolume("containers", "/var/lib/docker/containers", "/var/lib/docker/containers")

	var userid int64
	var groupid int64
	userid, groupid = 0, 0
	var allowPrivilegeEscalation bool
	readOnlyRootFilesystem := true
	capabilities := corev1.Capabilities{
		Drop: []corev1.Capability{"ALL"},
	}
	tolerations := []corev1.Toleration{
		{
			Effect:   corev1.TaintEffectNoSchedule,
			Key:      "node-role.kubernetes.io/master",
			Operator: corev1.TolerationOpExists,
		},
	}
	objectFieldSelector := corev1.ObjectFieldSelector{
		FieldPath: "spec.nodeName",
	}
	envVarSource := corev1.EnvVarSource{
		FieldRef: &objectFieldSelector,
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
					ServiceAccountName: "promtail",
					SecurityContext: &corev1.PodSecurityContext{
						RunAsGroup: &userid,
						RunAsUser:  &groupid,
					},
					Containers: []corev1.Container{
						{
							Image: imageName,
							Name:  "promtail",
							Args:  []string{"-config.file=/etc/promtail/promtail.yaml"},
							Ports: []corev1.ContainerPort{{
								ContainerPort: 3101,
								Name:          "http-metrics",
							}},
							VolumeMounts: VolumeMounts,
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
								Capabilities:             &capabilities,
								ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							},
							Env: []corev1.EnvVar{
								{
									Name:      "HOSTNAME",
									ValueFrom: &envVarSource,
								},
							},
						},
					},
					Tolerations: tolerations,
					Volumes:     volumes,
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
