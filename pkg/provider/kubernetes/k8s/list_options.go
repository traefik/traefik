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

// TweakListMergeOptions merges the provided overrides into the
// options. The LabelSelector and FieldSelectors are concatenated together
// using AND logic.
func TweakListMergeOptions(overrideOptions *metav1.ListOptions) internalinterfaces.TweakListOptionsFunc {
	return func(options *metav1.ListOptions) {
		mergeListOptions(options, overrideOptions)
	}
}

// TweakListOptions returns a list option that executes each of the provided
// functions in order.
func TweakListOptions(optionFuncs ...internalinterfaces.TweakListOptionsFunc) internalinterfaces.TweakListOptionsFunc {
	return func(options *metav1.ListOptions) {
		for _, opt := range optionFuncs {
			opt(options)
		}
	}
}

// mergeListOptions merges the provided overrides into the in options using the
// DeepCopyInto method but preserving:
//
//   - LabelSelector by concatenating the existing LabelSelector with the new
//     LabelSelector.
//   - FieldSelector by concatenating the existing FieldSelector with the new
//     FieldSelector.
func mergeListOptions(in, overrides *metav1.ListOptions) {
	if overrides == nil {
		return
	}

	ls := in.LabelSelector
	fs := in.FieldSelector

	overrides.DeepCopyInto(in)
	in.LabelSelector = concatenateSelector(ls, in.LabelSelector)
	in.FieldSelector = concatenateSelector(fs, in.FieldSelector)
}

func concatenateSelector(a, b string) string {
	switch {
	case a == "" && b == "":
		return ""
	case a == "" && b != "":
		return b
	case a != "" && b == "":
		return a
	default:
		return fmt.Sprintf("%s,%s", a, b)
	}
}
