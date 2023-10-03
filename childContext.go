package context

// ChildContext ...
type ChildContext interface {
	// create new child context for instance what implements Instance interface
	NewContextFor(instance ContextedInstance) (ChildContext, error)

	// cancel current context
	Cancel()
}
