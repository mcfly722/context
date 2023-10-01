package context

// RootContext ...
type RootContext interface {
	NewContextFor(instance ContextedInstance) (Context, error)
	Wait()
	Cancel()
}

type rootContext struct {
	instance ContextedInstance
	context  Context
	done     chan struct{}
}

func NewRootContext(instance ContextedInstance) (RootContext, error) {

	root := &rootContext{
		instance: instance,
		done:     make(chan struct{}),
	}

	emptyContext := newEmptyContext()

	rootContext, _ := newContextFor(emptyContext, root)

	root.context = rootContext

	return root, nil
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

func (root *rootContext) NewContextFor(instance ContextedInstance) (Context, error) {
	return root.context.NewContextFor(instance)
}
