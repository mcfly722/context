package context

// The RootContext interface is returned by the [NewRootContext] function.
type RootContext[M any] interface {

	// Creates new Context from your instance what implements [ContextedInstance] interface.
	// If current root context is already in closing state it returns [ClosingIsInProcessForFreezeError] or [ClosingIsInProcessForDisposingError]
	NewContextFor(instance ContextedInstance[M]) (ChildContext[M], error)

	// Waits till current root context would be Closeed.
	Wait()

	// Close current root context and all childs according reverse order.
	Close()

	// Send control message
	Send(message M) error
}

type rootContext[M any] struct {
	instance   ContextedInstance[M]
	context    Context[M]
	controller chan M
}

// NewRootContext function generates and starts new root context
func NewRootContext[M any](instance ContextedInstance[M]) RootContext[M] {

	root := &rootContext[M]{
		instance:   instance,
		controller: make(chan M),
	}

	emptyContext := newEmptyContext[M]()

	rootContext, _ := newContextFor[M](emptyContext, root)

	root.context = rootContext

	return root
}

// Wait ...
func (root *rootContext[M]) Wait() {
	<-root.controller
}

// Close ...
func (root *rootContext[M]) Close() {
	root.context.Close()
}

func (root *rootContext[M]) Go(current Context[M]) {
	root.instance.Go(current)
	close(root.controller)
}

// This function uses to generate new child context from root or other child context
func (root *rootContext[M]) NewContextFor(instance ContextedInstance[M]) (ChildContext[M], error) {
	return root.context.NewContextFor(instance)
}

// Send Controller message to root context
func (root *rootContext[M]) Send(message M) (err error) {
	return root.context.Send(message)
}
