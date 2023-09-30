package context

// Context ...
type ContextNode interface {
	// create new child context for instance what implements Instance interface
	NewContextFor(instance ContextedInstance) (Context, error)
}
