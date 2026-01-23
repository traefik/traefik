package ext

// HTTP is a dynamic.HTTP extension.
type HTTP struct{}

// Router is a dynamic.Router extension.
type Router struct{}

// RouterIng is a RouterIng extension for ingress provider annotations.
type RouterIng struct{}

// ToRouter converts RouterIng to a Router.
func (r RouterIng) ToRouter() Router {
	return Router{}
}
