package context

// Context ...
type Context interface {
	// create new child context for instance what implements Instance interface
	NewContextFor(instance ContextedInstance) (Context, error)

	// channel what closes when all childs are closed and you can exit from your current context. Dispose would be called.
	IsOpened() chan struct{}

	// finish all sub*childs, childs and current context
	Cancel()
}

type contextState int64

const (
	context_state_working contextState = iota
	context_state_freeze
	context_state_dispose
)

type context struct {
	parent   *context
	childs   []*context
	instance ContextedInstance
	state    contextState
	isOpened chan struct{}
	root     *rootContext
}

func newContextFor(parent *context, instance ContextedInstance) (Context, error) {

	root := &rootContext{}

	if parent != nil {
		root = parent.root
	}

	newContext := &context{
		parent:   parent,
		childs:   []*context{},
		instance: instance,
		state:    context_state_working,
		isOpened: make(chan struct{}),
		root:     root,
	}

	parent.childs = append(parent.childs, newContext)

	newContext.start()

	return newContext, nil
}

// NewContextFor ...
func (parent *context) NewContextFor(instance ContextedInstance) (Context, error) {

	parent.root.ready.Lock()
	defer parent.root.ready.Unlock()

	switch parent.state {
	case context_state_freeze:
		return nil, &NewContextInstanceFromFreezedStateError{}
	case context_state_dispose:
		return nil, &NewContextInstanceFromDisposingStateError{}
	}

	// context_state_working
	return newContextFor(parent, instance)
}

// IsOpened ...
func (context *context) IsOpened() chan struct{} {
	context.root.ready.RLock()
	defer context.root.ready.RUnlock()

	if context.state == context_state_freeze {
		if len(context.childs) == 0 {
			close(context.isOpened)
			context.state = context_state_dispose
		}
	}

	return context.isOpened
}

// Cancel ...
func (current *context) Cancel() {
	current.root.ready.Lock()
	defer current.root.ready.Unlock()

	/* TBD...
	if context.state == context_state_dispose {
		return &GracefulFinishFromDisposeStateError{}
	}
	*/

	current.recursiveCancelForAllChildsAndSubchilds()

}

func (current *context) recursiveCancelForAllChildsAndSubchilds() {
	if current.state == context_state_working {
		current.state = context_state_freeze
		for _, child := range current.childs {
			child.recursiveCancelForAllChildsAndSubchilds()
		}
	}
}

func (context *context) start() {
	// TBD
}
