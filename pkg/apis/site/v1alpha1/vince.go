package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Vince struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VinceSpec `json:"spec"`
	// +optional
	Status VinceStatus `json:"status,omitempty"`
}

type VinceSpec struct {
	Volume             `json:"volume,omitempty"`
	*Resources         `json:"resources,omitempty"`
	NodeAffinity       *v1.NodeAffinity  `json:"nodeAffinity,omitempty"`
	Tolerations        []v1.Toleration   `json:"tolerations,omitempty"`
	InitContainers     []v1.Container    `json:"initContainers,omitempty"`
	PodAnnotations     map[string]string `json:"podAnnotations,omitempty"`
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`
	Env                []v1.EnvVar       `json:"env,omitempty"`
}

type ResourceDescription struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type Resources struct {
	ResourceRequests ResourceDescription `json:"requests,omitempty"`
	ResourceLimits   ResourceDescription `json:"limits,omitempty"`
}
type Volume struct {
	Selector     *metav1.LabelSelector `json:"selector,omitempty"`
	Size         string                `json:"size"`
	StorageClass string                `json:"storageClass,omitempty"`
	SubPath      string                `json:"subPath,omitempty"`
	Iops         *int64                `json:"iops,omitempty"`
	Throughput   *int64                `json:"throughput,omitempty"`
	VolumeType   string                `json:"type,omitempty"`
}

type VinceStatus struct {
	Status string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VinceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Vince `json:"items"`
}
