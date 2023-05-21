package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Site describes a website with vinceanalytics configured to send analytics
// stats.
type Site struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SiteSpec `json:"spec"`
	// +optional
	Status SiteStatus `json:"status,omitempty"`
}

type SiteSpec struct {
	//+kubebuilder:validation:Pattern=`(?P<domain>(?:[a-z0-9]+(?:-[a-z0-9]+)*\.)+[a-z]{2,})`
	Domain string `json:"domain"`
	Public bool   `json:"public"`
	Target Target `json:"target"`
}

// Target is a reference to Vince crd resource that this site is attached to. We avoid
// using selectors because there will always be 1:1 mapping between the sites and
// vince instance.
type Target struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type SiteStatus struct {
	Status string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Site `json:"items"`
}
