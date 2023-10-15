package context

// This interface should be implemented by your nodes.
//
// The module automatically starts Go(...) method with the current Context and automatically waits until it ends.
// You could not exit from this method without context closing (otherwise [ExitFromContextWithoutClosePanic] occurs).
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
