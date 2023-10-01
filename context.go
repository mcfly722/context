package context

type Context interface {
	// create new child context for instance what implements Instance interface
	NewContextFor(instance ContextedInstance) (Context, error)

	// channel what closes when all childs are closed and you can exit from your current context. Dispose would be called.
	IsOpen() chan struct{}

	// finish all sub*childs, hilds and current context
	Cancel()
}

type contextState int

const (
	working   contextState = 0
	freezed   contextState = 1
	disposing contextState = 2
)

type context struct {
	parent   *context
	childs   map[*context]*context
	instance ContextedInstance
	state    contextState
	isOpened chan struct{}
	tree     *tree
}

func newEmptyContext() *context {
	context := &context{
		childs:   map[*context]*context{},
		state:    working,
		isOpened: make(chan struct{}),
	}

	context.tree = newTree(context)

	return context
}

// NewContextFor ...
func (parent *context) NewContextFor(instance ContextedInstance) (Context, error) {
	newContext, err := newContextFor(parent, instance)
	if err != nil {
		return nil, err
	}

	err = sendContextTo(parent.tree.new, newContext)
	if err != nil {
		return nil, err
	}

	return newContext, nil
}

func (current *context) Cancel() {
	sendContextTo(current.parent.tree.close, current)
}

// IsOpen ...
func (context *context) IsOpen() chan struct{} {
	return context.isOpened
}

func newContextFor(parent *context, instance ContextedInstance) (*context, error) {
	// goroutines unsafe

	newContext := &context{
		parent:   parent,
		childs:   map[*context]*context{},
		instance: instance,
		state:    working,
		isOpened: make(chan struct{}),
		tree:     parent.tree,
	}
	return newContext, nil
}
