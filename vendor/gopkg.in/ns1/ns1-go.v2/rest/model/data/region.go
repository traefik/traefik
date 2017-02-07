package data

// Region is a metadata table with a name/key.
// Can be thought of as metadata groupings.
type Region struct {
	Meta Meta `json:"meta,omitempty"`
}

// Regions is simply a mapping of Regions inside a record.
type Regions map[string]Region
