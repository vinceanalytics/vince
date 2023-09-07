package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Config holds configuration for a vince instance
type Config struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ConfigSpec   `json:"spec"`
	Status            ConfigStatus `json:"status,omitempty"`
}

type ConfigSpec struct {
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

type ConfigStatus struct {
	// A list of sites attached to this Vince instance.
	//+optional
	Sites []string `json:"sites,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Config `json:"items"`
}
