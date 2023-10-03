package context

// This interface should implement by your nodes.
//
// Module automatically starts Go(...) method with current Context and automatically waits till it ends.
// You could not exit from this method without context cancelling (otherwise [ExitFromContextWithoutCancelPanic] occurs).
//
// Example:
//
//	type node struct {}
//	func (node *node) Go(current context.Context) {
//		loop:
//		for {
//			select {
//			case _, isOpened := <-current.Context():
//				if !isOpened {
//					break loop
//				}
//			}
//		}
//	}
type ContextedInstance interface {
	Go(current Context)
}
