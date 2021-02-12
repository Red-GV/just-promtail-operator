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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PromtailSpec defines the desired state of Promtail
type PromtailSpec struct {
	// Strategy can be either 'DaemonSet' (default) or 'Sidecar'
	// +optional
	Strategy string `json:"strategy,omitempty"`

	// +optional
	// +listType=atomic
	Volumes []v1.Volume `json:"volumes,omitempty"`

	// +optional
	// +listType=atomic
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`

	// +optional
	Image string `json:"image,omitempty"`

	// +nullable
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Specification of the promtail deployment
	//
	// +optional
	Config PromtailConfig `json:"promtailConfig"`
}

// PromtailConfig defines the configuration of the promtail deployment
type PromtailConfig struct {
	// Specification of the promtail deployment
	//
	// +optional
	Client ClientConfig `json:"client"`

	// Specification of ScrapeConfigs
	//
	// +optional
	ScrapeConfigs ScrapeConfig `json:"scrape_configs"`
}

// ClientConfig defines the configuration of the promtail deployment
type ClientConfig struct {
	// The url of loki backend
	//
	// +optional
	URL string `json:"url,omitempty"`
	// Maximum amount of time to wait before sending a batch
	//
	// +optional
	Batchwait string `json:"batchwait,omitempty"`
}

// ScrapeConfig configures how Promtail can scrape logs
type ScrapeConfig struct {
	// The url of loki backend
	//
	// +optional
	Jobs []ScrapeJob `json:"scrape_jobs"`
}

// ElasticsearchNode struct represents individual node in Elasticsearch cluster
type ScrapeJob struct {
	// name of the scrape job
	//
	// +optional
	JobName string `json:"job_name,omitempty"`
}

// PromtailStatus defines the observed state of Promtail
type PromtailStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Promtail is the Schema for the promtails API
type Promtail struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PromtailSpec   `json:"spec,omitempty"`
	Status PromtailStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PromtailList contains a list of Promtail
type PromtailList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Promtail `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Promtail{}, &PromtailList{})
}
