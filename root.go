package context

// RootContext ...
type RootContext interface {
	NewContextFor(instance ContextedInstance, componentName string, componentType string) Context
	Terminate()
	Wait()
}

// Root ...
type Root struct {
	ctx       Context
	terminate chan bool
}

// NewRootContext ...
func NewRootContext(debugger Debugger) RootContext {
	root := &Root{
		terminate: make(chan bool),
	}

	root.ctx = newContextFor(root, debugger)

	return root
}

// Go ...
func (root *Root) Go(current Context) {
loop:
	for {
		select {
		case <-root.terminate:
			break loop
		}
	}
}

// Dispose ...
func (root *Root) Dispose(current Context) {}

// NewContextFor ...
func (root *Root) NewContextFor(instance ContextedInstance, componentName string, componentType string) Context {
	return root.ctx.NewContextFor(instance, componentName, componentType)
}

// Terminate ...
func (root *Root) Terminate() {
	go func() {
		root.terminate <- true
	}()
	root.ctx.Wait()
}

// Wait ...
func (root *Root) Wait() {
	root.ctx.Wait()
}
