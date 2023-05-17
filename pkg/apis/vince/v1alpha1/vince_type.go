package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
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
	Volume v1.PersistentVolumeClaimSpec `json:"volume"`
	//+optional
	VolumeSubPath string       `json:"volume_subpath,omitempty"`
	Container     v1.Container `json:"container"`
}

// VinceStatus tracks status of resources that are created from Vince.
type VinceStatus struct {
	// The state of the Secret resource used to configure the vince resource.
	// +kubebuilder:validation:Enum=Created;Resolved
	//+optional
	Secret string `json:"secret,omitempty"`
	//The state of the Service resource exposing the Vince instance.
	//+optional
	Service *v1.ServiceStatus `json:"service,omitempty"`
	// We track the status of the pod linked to this vince resource deployment.
	// Calls for site management are made directly on this pod.
	// +optional
	Pod *v1.PodStatus `json:"pod,omitempty"`
	// We track the status of the stateful set linked to this vince resource.
	// +optional
	StatefulSet *appsv1.StatefulSetStatus `json:"stateful_set,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VinceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Vince `json:"items"`
}
