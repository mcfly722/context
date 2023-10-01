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

	SetDefer(func(interface{}))
}

type contextState int

const (
	working   contextState = 0
	freezed   contextState = 1
	disposing contextState = 2
)

type context struct {
	parent       *context
	childs       map[*context]*context
	instance     ContextedInstance
	state        contextState
	isOpened     chan struct{}
	root         *root
	deferHandler *func(interface{})
}

type root struct {
	ready sync.Mutex
}

func newEmptyContext() *context {

	return &context{
		parent:       &context{},
		childs:       map[*context]*context{},
		instance:     nil,
		state:        working,
		isOpened:     make(chan struct{}),
		root:         &root{},
		deferHandler: nil,
	}
}

// NewContextFor ...
func (parent *context) NewContextFor(instance ContextedInstance) (Context, error) {

	parent.root.ready.Lock()
	defer parent.root.ready.Unlock()

	switch parent.state {
	case freezed:
		return nil, &CancelInProcessForFreezeError{}
	case disposing:
		return nil, &CancelInProcessForDisposingError{}
	}

	return newContextFor(parent, instance)
}

func newContextFor(parent *context, instance ContextedInstance) (Context, error) {

	newContext := &context{
		parent:       parent,
		childs:       map[*context]*context{},
		instance:     instance,
		state:        working,
		isOpened:     make(chan struct{}),
		root:         parent.root,
		deferHandler: nil,
	}

	parent.childs[newContext] = newContext

	// Start new Context
	go func(current *context) {

		// set defer handler
		defer func() {
			current.root.ready.Lock()
			defer current.root.ready.Unlock()
			if current.deferHandler != nil {
				deferHandler := *(current.deferHandler)
				deferHandler(recover()) // unfortunatelly recover() don't catch panic if you try to call it from function handler, so you need catch it here
			}
		}()

		// execure user context select {...}
		current.instance.Go(current)

		if current.state != disposing {
			panic(ExitFromContextWithoutCancelPanic)
		}

		// Remove node from parent childs and if parent is freezed and empty, initiate it disposing
		current.root.ready.Lock()
		if current.parent != nil {
			delete(current.parent.childs, current)
			if current.parent.state == freezed && len(current.parent.childs) == 0 {
				current.parent.state = disposing
				close(current.parent.isOpened)
			}
		}
		current.root.ready.Unlock()

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

func (current *context) SetDefer(deferFunction func(interface{})) {
	current.root.ready.Lock()
	current.deferHandler = &deferFunction
	current.root.ready.Unlock()
}
