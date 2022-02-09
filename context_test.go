package context

import (
	"errors"
	"fmt"
	"testing"
)

func buildTree(context *Context, currentPath string, width int, depth int) {

	if depth > 0 {

		ctx := context.NewChildContext()
		fmt.Println(fmt.Sprintf("created %v", currentPath))

		for i := 0; i < width; i++ {
			buildTree(ctx, fmt.Sprintf("%v->%v", currentPath, i), width, depth-1)
		}

		go func() {
			for {
				select {
				case err := <-ctx.OnCancel():
					fmt.Println(fmt.Sprintf("canceled %v (%v)", currentPath, err))
					ctx.Disposed()
					return
				}
			}

		}()

	}

}

func Test_TreeOrder(t *testing.T) {
	ctx := Background()
	buildTree(ctx, "0", 3, 5)

	fmt.Println("\ncanceling hive 0->2->1")
	ctx1 := ctx.childs[0].childs[2].childs[1]
	err1 := errors.New("test")

	fmt.Println(fmt.Sprintf("%v", err1))

	ctx1.Cancel(err1)

	fmt.Println("\ncanceling main context")

	ctx.Cancel(ErrCanceled)

	fmt.Println("finished")
}
