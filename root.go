package context

// RootContext ...
type RootContext interface {
	NewContextFor(instance ContextedInstance) Context
	Terminate()
	Wait()
}

// Root ...
type Root struct {
	ctx       Context
	terminate chan bool
}

// NewRootContext ...
func NewRootContext() RootContext {
	root := &Root{
		terminate: make(chan bool),
	}
	root.ctx = NewContextFor(root)
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
func (root *Root) Dispose() {}

// NewContextFor ...
func (root *Root) NewContextFor(instance ContextedInstance) Context {
	return root.ctx.NewContextFor(instance)
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
