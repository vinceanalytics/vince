/*
Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE Version 3
*/
// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	resource "k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeApplyConfiguration represents an declarative configuration of the Volume type for use
// with apply.
type VolumeApplyConfiguration struct {
	Selector     *v1.LabelSelector  `json:"selector,omitempty"`
	Size         *resource.Quantity `json:"size,omitempty"`
	StorageClass *string            `json:"storageClass,omitempty"`
	SubPath      *string            `json:"subPath,omitempty"`
}

// VolumeApplyConfiguration constructs an declarative configuration of the Volume type for use with
// apply.
func Volume() *VolumeApplyConfiguration {
	return &VolumeApplyConfiguration{}
}

// WithSelector sets the Selector field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Selector field is set to the value of the last call.
func (b *VolumeApplyConfiguration) WithSelector(value v1.LabelSelector) *VolumeApplyConfiguration {
	b.Selector = &value
	return b
}

// WithSize sets the Size field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Size field is set to the value of the last call.
func (b *VolumeApplyConfiguration) WithSize(value resource.Quantity) *VolumeApplyConfiguration {
	b.Size = &value
	return b
}

// WithStorageClass sets the StorageClass field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the StorageClass field is set to the value of the last call.
func (b *VolumeApplyConfiguration) WithStorageClass(value string) *VolumeApplyConfiguration {
	b.StorageClass = &value
	return b
}

// WithSubPath sets the SubPath field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SubPath field is set to the value of the last call.
func (b *VolumeApplyConfiguration) WithSubPath(value string) *VolumeApplyConfiguration {
	b.SubPath = &value
	return b
}