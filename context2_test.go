package context_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type node2 struct {
	i    int
	path string
}

func newNode2(path string, i int) *node2 {

	return &node2{
		path: path,
		i:    i,
	}
}

func (node *node2) Go(current context.Context) {
loop:
	for {
		select {
		case <-time.After(time.Duration(rand.Intn(100)) * time.Microsecond):
			current.Cancel()
		case _, opened := <-current.IsOpen():
			if !opened {
				break loop
			}
		default:
			{
			}
		}
	}
}

func (parent *node2) simpleTree2(context context.ContextNode, width int, height int) {
	if height > 1 {

		for i := 0; i < width; i++ {

			newNode := newNode2(fmt.Sprintf("%v->%v", parent.path, i), i)
			newContext, err := context.NewContextFor(newNode)
			if err == nil {
				//fmt.Printf("%v started\n", newNode.name())
				newNode.simpleTree2(newContext, width, height-1)
			}

		}

	}

}

const race_iterations int = 10000

func Test_Race_RandomSimpleTree3x3(t *testing.T) {

	for i := 0; i < race_iterations; i++ {
		fmt.Printf("Test_Race_RandomSimpleTree3x3 %v/%v\n", i+1, race_iterations)

		rootNode := newNode2("", 0)

		rootContext, err := context.NewRootContext(rootNode)
		if err != nil {
			t.Fatal(err)
		}

		rootNode.simpleTree2(rootContext, 3, 3)

		go func() {
			rootContext.Cancel()
		}()

		rootContext.Wait()
	}
}