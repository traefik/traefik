package dynamic

type ModifyHeader struct {
	Set    map[string]string `json:"set,omitempty" toml:"set,omitempty" yaml:"set,omitempty"  export:"true"`
	Append map[string]string `json:"append,omitempty" toml:"set,omitempty" yaml:"append,omitempty" export:"true"`
	Delete []string          `json:"delete,omitempty" toml:"delete,omitempty" yaml:"delete,omitempty" export:"true"`
}

func (mh *ModifyHeader) IsDefined() bool {
	return mh != nil && (len(mh.Set) != 0 || len(mh.Append) != 0 || len(mh.Delete) != 0)
}
