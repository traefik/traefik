package sarama

type Resource struct {
	ResourceType AclResourceType
	ResourceName string
}

func (r *Resource) encode(pe packetEncoder) error {
	pe.putInt8(int8(r.ResourceType))

	if err := pe.putString(r.ResourceName); err != nil {
		return err
	}

	return nil
}

func (r *Resource) decode(pd packetDecoder, version int16) (err error) {
	resourceType, err := pd.getInt8()
	if err != nil {
		return err
	}
	r.ResourceType = AclResourceType(resourceType)

	if r.ResourceName, err = pd.getString(); err != nil {
		return err
	}

	return nil
}

type Acl struct {
	Principal      string
	Host           string
	Operation      AclOperation
	PermissionType AclPermissionType
}

func (a *Acl) encode(pe packetEncoder) error {
	if err := pe.putString(a.Principal); err != nil {
		return err
	}

	if err := pe.putString(a.Host); err != nil {
		return err
	}

	pe.putInt8(int8(a.Operation))
	pe.putInt8(int8(a.PermissionType))

	return nil
}

func (a *Acl) decode(pd packetDecoder, version int16) (err error) {
	if a.Principal, err = pd.getString(); err != nil {
		return err
	}

	if a.Host, err = pd.getString(); err != nil {
		return err
	}

	operation, err := pd.getInt8()
	if err != nil {
		return err
	}
	a.Operation = AclOperation(operation)

	permissionType, err := pd.getInt8()
	if err != nil {
		return err
	}
	a.PermissionType = AclPermissionType(permissionType)

	return nil
}

type ResourceAcls struct {
	Resource
	Acls []*Acl
}

func (r *ResourceAcls) encode(pe packetEncoder) error {
	if err := r.Resource.encode(pe); err != nil {
		return err
	}

	if err := pe.putArrayLength(len(r.Acls)); err != nil {
		return err
	}
	for _, acl := range r.Acls {
		if err := acl.encode(pe); err != nil {
			return err
		}
	}

	return nil
}

func (r *ResourceAcls) decode(pd packetDecoder, version int16) error {
	if err := r.Resource.decode(pd, version); err != nil {
		return err
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	r.Acls = make([]*Acl, n)
	for i := 0; i < n; i++ {
		r.Acls[i] = new(Acl)
		if err := r.Acls[i].decode(pd, version); err != nil {
			return err
		}
	}

	return nil
}
