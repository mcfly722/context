package context

import (
	"sync"
)

const maxCancelCallsBeforeOnDoneReached int = 3 // you have select {} loop and could be several Close() calls from different events. To do not block execution for unblocking send, used channel with this length. This value should be >= 1. (ideal is to have unlimited lenght nonblocking channel, but it needs additional implementation)

// Context ...
type Context interface {
	NewContextFor(instance ContextedInstance, componentName string, componentType string) Context
	OnDone() chan bool // buffered channel with size=1. It is essential to do not block on a send onDone to do not stuck if GoRun method has no OnDone check.
	Wait()
	Log(eventType int, msg string)
}

// ContextedInstance ...
type ContextedInstance interface {
	Go(current Context)      // here we could put our events loop and wait timeouts/events/onDone signal
	Dispose(current Context) // Dispose fires only when current and all child GoRun's has been finished. It is garantee that there are no any other resources/calls which tries to use current context, this context could be gracefully closed
}

type tree struct {
	changesAllowed sync.Mutex
}

type ctx struct {
	id               int64
	parent           *ctx
	debugger         Debugger
	debuggerNodePath []DebugNode // it is not a pointer, it is full array copy
	childs           map[int64]*ctx
	nextChildID      int64
	instance         ContextedInstance
	waitGroup        sync.WaitGroup
	currentLoop      sync.WaitGroup
	onDone           chan bool
	closed           bool
	tree             *tree
}

func newContextFor(instance ContextedInstance, debugger Debugger) Context {

	newContext := &ctx{
		id:               0,
		parent:           nil,
		debugger:         debugger,
		debuggerNodePath: []DebugNode{DebugNode{ID: 0, ComponentType: "root", ComponentName: "root"}},
		childs:           make(map[int64]*ctx),
		nextChildID:      0,
		instance:         instance,
		onDone:           make(chan bool, maxCancelCallsBeforeOnDoneReached),
		closed:           false,
		tree:             &tree{},
	}

	newContext.start()

	return newContext
}

// StartNewFor ...
func (context *ctx) NewContextFor(instance ContextedInstance, componentName string, componentType string) Context {

	// attach to parent new child
	parent := context

	context.tree.changesAllowed.Lock()

	newContext := &ctx{
		id:               parent.nextChildID,
		parent:           parent,
		debugger:         parent.debugger,
		debuggerNodePath: append(parent.debuggerNodePath, DebugNode{ID: parent.nextChildID, ComponentName: componentName, ComponentType: componentType}),
		childs:           make(map[int64]*ctx),
		nextChildID:      0,
		instance:         instance,
		onDone:           make(chan bool, maxCancelCallsBeforeOnDoneReached),
		closed:           parent.closed,
		tree:             parent.tree,
	}

	parent.childs[parent.nextChildID] = newContext
	parent.nextChildID++

	parent.waitGroup.Add(1)

	parent.tree.changesAllowed.Unlock()

	newContext.start()

	return newContext

}

func (context *ctx) Log(eventType int, msg string) {
	context.debugger.Log(context.debuggerNodePath, eventType, msg)
}

func (context *ctx) start() {

	context.currentLoop.Add(1)

	go func(ctx *ctx) {

		ctx.debugger.Log(ctx.debuggerNodePath, 100, "started")

		{ // wait till context execution would be finished, only after that you can dispose all context resources, otherwise it could try to create new child context on disposed resources
			ctx.instance.Go(ctx)
			ctx.debugger.Log(ctx.debuggerNodePath, 101, "finished")
		}

		{ // stop all childs contexts
			for _, child := range ctx.childs {
				child.onDone <- true
			}
			//  and wait them
			ctx.waitGroup.Wait()
		}

		{ // all childs and subchilds contexts has been stopped and disposed, we can gracefully dispose current context resources
			ctx.debugger.Log(ctx.debuggerNodePath, 101, "disposing")
			ctx.instance.Dispose(ctx)
			ctx.debugger.Log(ctx.debuggerNodePath, 100, "disposed")
			ctx.currentLoop.Done()
		}

		{ // for parent, this context excluded from wait group
			if ctx.parent != nil {
				ctx.parent.waitGroup.Done()
			}
		}

		{ // unbind closed context from tree
			ctx.tree.changesAllowed.Lock()
			if ctx.parent != nil {
				delete(ctx.childs, ctx.id)
			}
			ctx.tree.changesAllowed.Unlock()
		}

	}(context)
}

// OnDone ...
func (context *ctx) OnDone() chan bool {
	return context.onDone
}

// Wait ...
func (context *ctx) Wait() {
	context.waitGroup.Wait()   // wait all childs
	context.currentLoop.Wait() // wait for current context disposing
}
