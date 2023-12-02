package context

// The RootContext interface is returned by the [NewRootContext] function.
type RootContext interface {

	// Creates new Context from your instance what implements [ContextedInstance] interface.
	// If current root context is already in closing state it returns [ClosingIsInProcessForFreezeError] or [ClosingIsInProcessForDisposingError]
	NewContextFor(instance ContextedInstance) (ChildContext, error)

	// Waits till current root context would be Closeed.
	Wait()

	// Close current root context and all childs according reverse order.
	Close()
}

type rootContext struct {
	instance ContextedInstance
	context  *context
	done     chan struct{}
}

// NewRootContext function generates and starts new root context
func NewRootContext(instance ContextedInstance) RootContext {

	root := &rootContext{
		instance: instance,
		done:     make(chan struct{}),
	}

	emptyContext := newEmptyContext()

	rootContext, _ := newContextFor(emptyContext, root)

	root.context = rootContext

	return root
}

// Wait ...
func (root *rootContext) Wait() {
	<-root.done
}

// Close ...
func (root *rootContext) Close() {
	root.context.Close()
}

func (root *rootContext) Go(current Context) {
	root.instance.Go(current)
	close(root.done)
}

// This function uses to generate new child context from root or other child context
func (root *rootContext) NewContextFor(instance ContextedInstance) (ChildContext, error) {
	return root.context.NewContextFor(instance)
}
