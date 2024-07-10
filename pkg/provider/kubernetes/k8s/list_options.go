package k8s

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers/internalinterfaces"
)

const (
	labelSelectorNotHelm = "owner!=helm"
)

// TweakListOptionWithLabelSelector returns a list option that filters the list
// using the provided label selector.
func TweakListOptionWithLabelSelector(labelSelector string) internalinterfaces.TweakListOptionsFunc {
	return func(options *metav1.ListOptions) {
		options.LabelSelector = labelSelector
	}
}

// TweakListOptionNotOwnedByHelm returns a list option that excludes objects
// owned by Helm.
func TweakListOptionNotOwnedByHelm() internalinterfaces.TweakListOptionsFunc {
	return TweakListOptionWithLabelSelector(labelSelectorNotHelm)
}

