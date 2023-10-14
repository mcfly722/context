package context

// This interface should be implemented by your nodes.
//
// Module automatically starts Go(...) method with current Context and automatically waits till it ends.
// You could not exit from this method without context Finishling (otherwise [ExitFromContextWithoutFinishPanic] occurs).
//
// Example:
//
//	type node struct {}
//	func (node *node) Go(current context.Context) {
//		loop:
//		for {
//			select {
//			case _, isOpened := <-current.Controller():
//				if !isOpened {
//					break loop
//				}
//			}
//		}
//	}
type ContextedInstance[M any] interface {
	Go(current Context[M])
}
