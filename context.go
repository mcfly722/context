package context

import (
	"sync"
)

// Context ...
type Context struct {
	id          int64
	parent      *Context
	childs      map[int64]*Context
	nextChildID int64
	onCancel    chan *error
	waitGroup   *sync.WaitGroup

	ready sync.Mutex
}

// Background ...
func Background() *Context {
	return newEmptyContext(0, nil)
}

// NewChildContext ...
func (context *Context) NewChildContext() *Context {
	context.ready.Lock()
	newContext := newEmptyContext(context.nextChildID, context)
	context.childs[context.nextChildID] = newContext
	context.nextChildID++

	context.waitGroup.Add(1)

	context.ready.Unlock()
	return newContext
}

func newEmptyContext(id int64, parent *Context) *Context {
	return &Context{
		id:          id,
		parent:      parent,
		childs:      make(map[int64]*Context),
		nextChildID: 0,
		onCancel:    make(chan *error),
		waitGroup:   &sync.WaitGroup{},
	}
}

// Cancel ...
func (context *Context) Cancel(err *error) {

	context.ready.Lock()
	for _, ctx := range context.childs {
		// send Cancel to all upper levels
		ctx.Cancel(err)
	}

	for _, ctx := range context.childs {
		// for current level send cancelation to every child
		ctx.onCancel <- err
	}
	context.ready.Unlock()

	// wait until all childs have been disposed
	context.waitGroup.Wait()

}

// OnCancel ...
func (context *Context) OnCancel() chan *error {
	return context.onCancel
}

// Disposed ...
func (context *Context) Disposed() {

	if context.parent != nil {

		parent := context.parent

		parent.ready.Lock()
		if _, ok := parent.childs[context.id]; ok {
			delete(parent.childs, context.id)
			parent.waitGroup.Done()
		}
		parent.ready.Unlock()

	}

}
