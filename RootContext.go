package context

// The RootContext interface is returned by the [NewRootContext] function.
type RootContext interface {

	// Method creates new Context from your instance what implements [ContextedInstance] interface.
	// If current root context is already in closing state it returns [CancelInProcessForFreezeError] or [CancelInProcessForDisposingError]
	NewContextFor(instance ContextedInstance) (ChildContext, error)

	// Method waits till current root context would be canceled.
	Wait()

	// Method cancel current root context and all childs according reverse order.
	Cancel()
}

type rootContext struct {
	instance ContextedInstance
	context  Context
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

// Cancel ...
func (root *rootContext) Cancel() {
	root.context.Cancel()
}

func (root *rootContext) Go(current Context) {
	root.instance.Go(current)
	close(root.done)
}

// This function uses to generate new child context from root or other child context
func (root *rootContext) NewContextFor(instance ContextedInstance) (ChildContext, error) {
	return root.context.NewContextFor(instance)
}
