package context_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mcfly722/goPackages/context"
)

func buildTree(context *context.Context, currentPath string, width int, depth int) {

	if depth > 0 {

		ctx := context.NewChildContext()
		fmt.Println(fmt.Sprintf("creating %v", currentPath))

		for i := 1; i <= width; i++ {
			buildTree(ctx, fmt.Sprintf("%v->%v", currentPath, i), width, depth-1)
		}

		go func() {
			for {
				select {
				case <-ctx.OnCancel():
					fmt.Println(fmt.Sprintf("canceling %v", currentPath))

					ctx.Disposed()
					return
				default:
				}
			}

		}()

	}

}

func Test_TreeOrder(t *testing.T) {

	ctx := (context.Background()).NewChildContext()

	buildTree(ctx, "0", 3, 4)

	fmt.Println("canceling main context")

	err := errors.New("test")

	ctx.Cancel(&err)

	fmt.Println("finished")
}
