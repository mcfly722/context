package context

// ChildContext obtained from [NewContextFor] function.
//
// Any child could have several subchilds.
//
// During closing, this child would be a parent for all its sub childs.
type ChildContext[M any] interface {

	// create new child context for instance what implements Instance interface
	NewContextFor(instance ContextedInstance[M]) (ChildContext[M], error)

	// cancel current context
	Cancel()
}
