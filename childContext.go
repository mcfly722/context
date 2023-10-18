package context

// ChildContext obtained from the [NewContextFor] function.
//
// Any child could have several subchilds.
//
// During closing, this child would be a parent for all its sub-childs.
type ChildContext interface {

	// create a new child context, for instance, what implements the instance interface
	NewContextFor(instance ContextedInstance) (ChildContext, error)

	// Close current context
	Close()
}
