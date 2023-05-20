package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	Status *VinceStatus `json:"status,omitempty"`
}

type VinceSpec struct {
	Volume    `json:"volume"`
	Container `json:"container"`
}

type Volume struct {
	Selector     *metav1.LabelSelector `json:"selector,omitempty"`
	Size         resource.Quantity     `json:"size"`
	StorageClass string                `json:"storageClass,omitempty"`
	SubPath      string                `json:"subPath,omitempty"`
}

type Container struct {
	//+optional
	Image string `json:"image,omitempty"`

	//+Optional
	Env []v1.EnvVar `json:"env,omitempty"`

	Resources v1.ResourceRequirements `json:"resources"`
}

// VinceStatus tracks status of resources that are created from Vince.
type VinceStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VinceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Vince `json:"items"`
}
