package context

import "sync"

// Context ...
type Context interface {
	// create new child context for instance what implements Instance interface
	NewContextFor(instance ContextedInstance) (Context, error)

	// channel what closes when all childs are closed and you can exit from your current context. Dispose would be called.
	IsOpen() chan struct{}

	// finish all sub*childs, childs and current context
	Cancel()
}

type contextState int64

const (
	working contextState = iota
	freezed
	disposing
)

type context struct {
	parent   *context
	childs   map[*context]empty
	instance ContextedInstance
	state    contextState
	isOpened chan struct{}
	root     *root
}

type root struct {
	ready sync.RWMutex
}

type empty struct{}

// NewContextFor ...
func (parent *context) NewContextFor(instance ContextedInstance) (Context, error) {

	parent.root.ready.Lock()
	defer parent.root.ready.Unlock()

	switch parent.state {
	case freezed:
		return nil, &CancelInProcessError{}
	case disposing:
		return nil, &CancelInProcessError{}
	}

	// context_state_working
	return newContextFor(parent, instance)
}

func newContextFor(parent *context, instance ContextedInstance) (Context, error) {
	// goroutines unsafe
	root := &root{}

	if parent != nil {
		root = parent.root
	}

	newContext := &context{
		parent:   parent,
		childs:   map[*context]empty{},
		instance: instance,
		state:    working,
		isOpened: make(chan struct{}),
		root:     root,
	}

	parent.childs[newContext] = empty{}

	newContext.start()

	return newContext, nil
}

func (current *context) contextRemoveFromParentChildsList() {
	current.root.ready.Lock()
	defer current.root.ready.Unlock()

	if current.parent != nil {
		delete(current.parent.childs, current)
	}
}

// IsOpen ...
func (context *context) IsOpen() chan struct{} {
	context.root.ready.RLock()
	defer context.root.ready.RUnlock()

	if context.state == freezed {
		if len(context.childs) == 0 {
			close(context.isOpened)
			context.state = disposing
		}
	}

	return context.isOpened
}

// Cancel ...
func (current *context) Cancel() {
	current.root.ready.Lock()
	defer current.root.ready.Unlock()

	if current.state == disposing {
		panic(CancelFromDisposeStatePanic)
	}

	current.freezeAllChildsAndSubchilds()
}

func (current *context) freezeAllChildsAndSubchilds() {
	if current.state == working {
		current.state = freezed
		for child := range current.childs {
			child.freezeAllChildsAndSubchilds()
		}
	}
}

func (current *context) start() {
	go func(current *context) {

		current.instance.Go(current)

		if current.state != disposing {
			panic(ExitFromContextWithoutCancelPanic)
		}

		current.contextRemoveFromParentChildsList()
	}(current)
}
