package context

// ChildContext obtained from [NewContextFor] function.
//
// Any child could have several subchilds.
//
// During closing, this child would be a parent for all its sub childs.
type ChildContext interface {

	// create new child context for instance what implements Instance interface
	NewContextFor(instance ContextedInstance) (ChildContext, error)

	// cancel current context
	Cancel()
}
