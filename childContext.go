package context

// ChildContext obtained from the [NewContextFor] function.
//
// Any child could have several subchilds.
//
// During closing, this child would be a parent for all its sub-childs.
type ChildContext[M any] interface {

	// create a new child context, for instance, what implements the instance interface
	NewContextFor(instance ContextedInstance[M]) (ChildContext[M], error)

	// Close current context
	Close()

	// Send a control message.
	Send(message M) (err error)
}
