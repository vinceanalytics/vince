package k8s

import (
	"k8s.io/apimachinery/pkg/api/meta"
)

// ResourceFilter holds resource filtering rules.
type ResourceFilter struct {
	watchedNamespaces []string
	ignoredNamespaces []string
	ignoredApps       []string
}

type namespaceName struct {
	Name      string
	Namespace string
}

// ResourceFilterOption adds a filtering rule to the given ResourceFilter.
type ResourceFilterOption func(filter *ResourceFilter)

// WatchNamespaces add the given namespaces to the list of namespaces to watch.
func WatchNamespaces(namespaces ...string) ResourceFilterOption {
	return func(filter *ResourceFilter) {
		filter.watchedNamespaces = append(filter.watchedNamespaces, namespaces...)
	}
}

// IgnoreNamespaces adds the given namespaces to the list of namespaces to ignore.
func IgnoreNamespaces(namespaces ...string) ResourceFilterOption {
	return func(filter *ResourceFilter) {
		filter.ignoredNamespaces = append(filter.ignoredNamespaces, namespaces...)
	}
}

// IgnoreApps add the given apps to the list of apps to ignore. An app is a Kubernetes object
// with an "app" label, the name of the app being the value of the label.
func IgnoreApps(apps ...string) ResourceFilterOption {
	return func(filter *ResourceFilter) {
		filter.ignoredApps = append(filter.ignoredApps, apps...)
	}
}

// NewResourceFilter creates a new ResourceFilter, configured with the given options.
func NewResourceFilter(opts ...ResourceFilterOption) *ResourceFilter {
	filter := &ResourceFilter{}

	for _, opt := range opts {
		opt(filter)
	}

	return filter
}

// IsIgnored returns true if the resource should be ignored.
func (f *ResourceFilter) IsIgnored(obj interface{}) bool {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return true
	}

	pMeta := meta.AsPartialObjectMetadata(accessor)

	// If we are not watching all namespaces, check if the namespace is in the watch list.
	if len(f.watchedNamespaces) > 0 && !contains(f.watchedNamespaces, pMeta.Namespace) {
		return true
	}

	// Check if the namespace is not explicitly ignored.
	if contains(f.ignoredNamespaces, pMeta.Namespace) {
		return true
	}

	// Check if the "app" label doesn't contain a value which is ignored.
	if contains(f.ignoredApps, pMeta.Labels["app"]) {
		return true
	}
	return false
}

func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}

	return false
}

func containsNamespaceName(slice []namespaceName, nn namespaceName) bool {
	for _, item := range slice {
		if item.Namespace == nn.Namespace && item.Name == nn.Name {
			return true
		}
	}

	return false
}
