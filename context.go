package context

import (
	"sync"
)

type Context interface {
	// create new child context for instance what implements Instance interface
	NewContextFor(instance ContextedInstance) (Context, error)

	// channel what closes when all childs are closed and you can exit from your current context. Dispose would be called.
	IsOpen() chan struct{}

	// finish all sub*childs, childs and current context
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
	root     *root
}

type root struct {
	ready sync.Mutex
}

func newEmptyContext() *context {
	return &context{
		childs:   map[*context]*context{},
		state:    working,
		isOpened: make(chan struct{}),
		root:     &root{},
	}
}

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

	return newContextFor(parent, instance)
}

func newContextFor(parent *context, instance ContextedInstance) (Context, error) {

	newContext := &context{
		parent:   parent,
		childs:   map[*context]*context{},
		instance: instance,
		state:    working,
		isOpened: make(chan struct{}),
		root:     parent.root,
	}

	parent.childs[newContext] = newContext

	// Start new Context
	go func(current *context) {

		// execure user context select {...}
		current.instance.Go(current)

		if current.state != disposing {
			panic(ExitFromContextWithoutCancelPanic)
		}

		{ // Remove node from parent childs and if parent is freezed and empty, initiate it disposing

			current.root.ready.Lock()
			defer current.root.ready.Unlock()

			if current.parent != nil {
				delete(current.parent.childs, current)
				if current.parent.state == freezed && len(current.parent.childs) == 0 {
					current.parent.state = disposing
					close(current.parent.isOpened)
				}
			}
		}

	}(newContext)

	return newContext, nil
}

// IsOpen ...
func (context *context) IsOpen() chan struct{} {
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

	if current.state == freezed && len(current.childs) == 0 {
		current.state = disposing
		close(current.isOpened)
	}
}
