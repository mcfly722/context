package context_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mcfly722/goPackages/context"
)

var contextTree map[string]context.Context = make(map[string]context.Context)
var contextTreeReady sync.Mutex

func buildTree(parent context.Context, currentPath string, width int, depth int) context.Context {

	if depth > 0 {

		onDone := make(chan bool)

		ctx, err := parent.NewChildContext(
			func(reason context.Reason) { // Disposer
				fmt.Println(fmt.Sprintf("%v disposed", currentPath))
			}, func(reason context.Reason) { // Finalizer
				onDone <- true
			})

		if err != nil {
			fmt.Printf(fmt.Sprintf("error: %v", err))
			return nil
		}

		for i := 0; i < width; i++ {
			buildTree(ctx, fmt.Sprintf("%v->%v", currentPath, i), width, depth-1)
		}

		contextTreeReady.Lock()
		fmt.Println(fmt.Sprintf("[%v] created %v", len(contextTree), currentPath))
		contextTree[currentPath] = ctx
		contextTreeReady.Unlock()

		go func() {

			for {
				select {
				case <-onDone:
					fmt.Println(fmt.Sprintf("[%v] finished", currentPath))
					return
				}
			}
		}()

		return ctx
	}

	return nil
}

func Test_CancelHive(t *testing.T) {

	onDone := make(chan bool)

	rootContext := context.NewContextTree(
		func(reason context.Reason) {
			fmt.Println(fmt.Sprintf("root disposed. Reason=%v", reason))
		}, func(reason context.Reason) {
			fmt.Println(fmt.Sprintf("root finished. Reason=%v", reason))
			onDone <- true
		})

	buildTree(rootContext, "0", 3, 5)

	fmt.Println("\n\nCancelling 0->1->2")
	contextTree["0->1->2"].Cancel(context.ReasonCanceled)

	fmt.Println("\n\nCancelling 0")
	rootContext.Cancel(context.ReasonCanceled)

	<-onDone

	time.Sleep(time.Second * 1)

	fmt.Println("finished")
}

/*
func Test_CancelRoot(t *testing.T) {

	rootContext := context.NewContextTree(func(reason context.Reason) {
		fmt.Println(fmt.Sprintf("root disposed. Reason=%v", reason))
	}, nil)

	buildTree(rootContext, "0", 5, 7)

	rootContext.Cancel(context.ReasonCanceled)

	time.Sleep(time.Second * 1)

	fmt.Println(fmt.Sprintf("finished. mapSize = %v", len(contextTree)))
}
*/
