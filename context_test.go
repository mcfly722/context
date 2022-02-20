package context_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mcfly722/goPackages/context"
)

func buildTree(context context.Context, currentPath string, width int, depth int) context.Context {

	if depth > 0 {

		context := context.NewChildContext()

		context.SetDisposer(func(err error) {
			fmt.Println(fmt.Sprintf("%v disposed", currentPath))
		})

		for i := 0; i < width; i++ {
			buildTree(context, fmt.Sprintf("%v->%v", currentPath, i), width, depth-1)
		}

		fmt.Println(fmt.Sprintf("created %v", currentPath))

		go func() {

			for {
				select {
				case <-context.OnDone():
					fmt.Println(fmt.Sprintf("%v closed", currentPath))
					return
				}
			}

		}()

		return context
	}
	return nil
}

func Test_CancelHive(t *testing.T) {
	rootContext := context.NewContextTree()

	ctx := buildTree(rootContext, "0", 3, 5)

	//time.Sleep(1 * time.Millisecond)
	fmt.Println("\n\ncanceling 0->2->1->...")
	ctx.GetChild(2).GetChild(1).Cancel(context.ErrCanceled)

	//time.Sleep(2 * time.Millisecond)

	fmt.Println("finished")
}

func Test_CancelHiveByTimeout(t *testing.T) {
	rootContext := context.NewContextTree()

	ctx := buildTree(rootContext, "0", 3, 5)

	//time.Sleep(1 * time.Millisecond)
	fmt.Println("\n\ncanceling 0->2->1->... by timeout")
	ctx.GetChild(2).GetChild(1).SetDeadline(time.Now().Add(1 * time.Millisecond))
	//time.Sleep(2 * time.Millisecond)

	fmt.Println("finished")
}

func Test_CancelRoot(t *testing.T) {
	rootContext := context.NewContextTree()

	ctx := buildTree(rootContext, "0", 3, 5)

	//time.Sleep(1 * time.Millisecond)
	fmt.Println("\n\ncanceling 0")
	ctx.Cancel(context.ErrCanceled)

	//time.Sleep(2 * time.Millisecond)

	fmt.Println("finished")
}

func Test_CancelOnlyRoot(t *testing.T) {
	rootContext := context.NewContextTree()

	go func() {
		for {
			select {
			case <-rootContext.OnDone():
				return
			}
		}
	}()

	fmt.Println("\n\ncanceling 0")
	rootContext.Cancel(context.ErrCanceled)
}

/*
func Test_RaceTree(t *testing.T) {
		rootContext := context.NewContextTree()
		ctx := buildTree(rootContext, "0", 3, 5)
}
*/
