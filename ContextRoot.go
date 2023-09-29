package context

import "sync"

// RootContext ...
type RootContext interface {

	// create new child context for instance what implements Instance interface
	NewContextFor(instance ContextedInstance) (Context, error)

	// TBD
	Wait()

	// TBD
	Cancel()
}

type rootContext struct {
	ready sync.RWMutex
}

// NewRootContext ...
func NewRootContext() RootContext {
	return &rootContext{}
}

func (root *rootContext) NewContextFor(instance ContextedInstance) (Context, error) {
	return newContextFor(nil, &rootContextInstance{})
}

func (root *rootContext) Wait() {
	//TBD
}

func (root *rootContext) Cancel() {
	//TBD
}

type rootContextInstance struct{}

func (instance *rootContextInstance) Go(current Context) {
	//TBD
}

/*
func (root *Root) Go(current Context) {
	<-current.IsOpened()
	/*
	   loop:
	   	for {
	   		select {
	   		case _, opened := <-current.IsOpened():
	   			if !opened {
	   				break loop
	   			}
	   		}
	   	}
}
*/
