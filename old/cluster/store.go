package cluster

// Object is the struct to store
type Object interface{}

// Store is a generic interface to represents a storage
type Store interface {
	Load() (Object, error)
	Get() Object
	Begin() (Transaction, Object, error)
}

// Transaction allows to set a struct in the KV store
type Transaction interface {
	Commit(object Object) error
}
