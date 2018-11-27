package internal

const (
	// TagLabel allow to apply a custom behavior.
	// - "allowEmpty": allow to create an empty struct.
	// - "-": ignore the field.
	TagLabel = "label"

	// TagLabelSliceAsStruct allow to use a slice of struct by creating one entry into the slice.
	// The value is the substitution name use in the label to access the slice.
	TagLabelSliceAsStruct = "label-slice-as-struct"
)
