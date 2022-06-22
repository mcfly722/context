package context

// RootContext ...
type RootContext interface {
	NewContextFor(instance ContextedInstance, componentName string, componentType string) Context // create new child context
	Cancel()                                                                                      // cancel root context with all childs
	Wait()                                                                                        // waits till root context would be closed
	Log(eventType int, msg string)                                                                // log context event
}

// Root ...
type Root struct {
	ctx Context
}

// Go ...
func (root *Root) Go(current Context) {
loop:
	for {
		select {
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}
	}
}

// NewRootContext ...
func NewRootContext(debugger Debugger) RootContext {
	root := &Root{}

	root.ctx, _ = newContextFor(root, debugger)

	return root
}

// Cancel ...
func (root *Root) Cancel() {
	root.ctx.Cancel()
}

// Wait ...
func (root *Root) Wait() {
	root.ctx.wait()
}

// Log ...
func (root *Root) Log(eventType int, msg string) {
	root.ctx.Log(eventType, msg)
}

// NewContextFor ...
func (root *Root) NewContextFor(instance ContextedInstance, componentName string, componentType string) Context {
	child, _ := root.ctx.NewContextFor(instance, componentName, componentType)
	return child
}
