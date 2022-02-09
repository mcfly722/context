package context

import "errors"

// ErrCanceled ...
var ErrCanceled = errors.New("context canceled")

// Context ...
type Context struct {
	id          int64
	parent      *Context
	childs      map[int64]*Context
	nextChildID int64
	constructor chan *Context         // used to attach new Context to Context tree
	destructor  chan *Context         // used to deattach old Context from Context tree
	cancel      chan *cancelOperation // used to deattach Context and all child subcontexts
	cancelDone  chan bool             // send signal when cancel done
	onCancel    chan error            // used to send cancel error
	desposed    chan bool             // used to submit all context tree changes before release runtime
}

type cancelOperation struct {
	context *Context
	error   error
}

func (context *Context) cancelator(err error) {

	for _, ctx := range context.childs {
		// send Cancel to all upper levels
		ctx.cancelator(err)
	}

	for _, ctx := range context.childs {

		// for current level send cancelation to every child
		ctx.onCancel <- err
		<-ctx.desposed

		delete(context.childs, ctx.id)
	}

}

// Background ...
func Background() *Context {
	root := newEmptyContext(nil)

	// here we use one goroutine for root for attach/deattach operations on context tree to do not use mutex'es (they are blocks all tree nodes and really slow).
	go func() {
		for {
			select {
			case newContext := <-root.constructor:
				newContext.id = newContext.parent.nextChildID
				newContext.parent.childs[newContext.id] = newContext
				newContext.parent.nextChildID++

			case cancelOperation := <-root.cancel:
				cancelOperation.context.cancelator(cancelOperation.error)

				cancelOperation.context.cancelDone <- true
			}
		}
	}()

	return root
}

// NewChildContext ...
func (context *Context) NewChildContext() *Context {
	newContext := newEmptyContext(context)
	context.constructor <- newContext
	return newContext
}

func newEmptyContext(parent *Context) *Context {

	newContext := &Context{
		parent:      parent,
		childs:      make(map[int64]*Context),
		nextChildID: 0,
		onCancel:    make(chan error),
		desposed:    make(chan bool),
		cancelDone:  make(chan bool),
	}

	if parent == nil {
		newContext.constructor = make(chan *Context)
		newContext.destructor = make(chan *Context)
		newContext.cancel = make(chan *cancelOperation)
	} else {
		newContext.constructor = parent.constructor
		newContext.destructor = parent.destructor
		newContext.cancel = parent.cancel
	}

	return newContext
}

// Cancel ...
func (context *Context) Cancel(err error) {

	operation := &cancelOperation{
		context: context,
		error:   err,
	}
	context.cancel <- operation

	// waiting till all childs and subchilds would be canceled
	<-context.cancelDone
}

// OnCancel ...
func (context *Context) OnCancel() chan error {
	return context.onCancel
}

// Disposed ...
func (context *Context) Disposed() {
	//context.destructor <- context
	context.desposed <- true
}
