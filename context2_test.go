package context_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	context "github.com/mcfly722/context"
)

type node2 struct {
	i      int
	path   string
	childs []context.ChildContext[*state2]
	ready  sync.Mutex
}

func newNode2(path string, i int) *node2 {

	return &node2{
		path:   path,
		i:      i,
		childs: []context.ChildContext[*state2]{},
	}
}

func (node *node2) appendChild(child context.ChildContext[*state2]) {
	node.ready.Lock()
	node.childs = append(node.childs, child)
	node.ready.Unlock()
}

func (node *node2) trySendToRandomChildRandomState() {
	node.ready.Lock()
	if len(node.childs) > 0 {
		i := rand.Intn(len(node.childs))
		(node.childs[i]).Send(&state2{name: "random state"})
	}
	node.ready.Unlock()
}

type state2 struct {
	name string
}

var messages2 int = 0
var messagesLock2 sync.Mutex

func (node *node2) Go(current context.Context[*state2]) {
loop:
	for {
		select {
		case <-time.After(time.Duration(rand.Intn(10)) * time.Microsecond):
			node.trySendToRandomChildRandomState()
		case <-time.After(time.Duration(rand.Intn(100)) * time.Microsecond):
			current.Close()
		case _, isOpened := <-current.Controller():
			if !isOpened {
				break loop
			} else {
				messagesLock2.Lock()
				messages2 += 1
				messagesLock2.Unlock()
			}
		default:
			{
			}
		}
	}
}

func (parent *node2) simpleTree2(context context.ChildContext[*state2], width int, height int) {
	if height > 1 {

		for i := 0; i < width; i++ {

			newNode := newNode2(fmt.Sprintf("%v->%v", parent.path, i), i)
			newContext, err := context.NewContextFor(newNode)
			if err == nil {
				//fmt.Printf("%v started\n", newNode.name())
				newNode.simpleTree2(newContext, width, height-1)
				parent.appendChild(newContext)
			}
		}

	}

}

const race_iterations int = 10000

func Test_Race_RandomSimpleTree3x3(t *testing.T) {

	for i := 1; i <= race_iterations; i++ {
		if (i % 100) == 0 {
			messagesLock2.Lock()
			fmt.Printf("Test_Race_RandomSimpleTree3x3 %v/%v messages=%v\n", i, race_iterations, messages2)
			messagesLock2.Unlock()
		}

		rootNode := newNode2("", 0)

		rootContext := context.NewRootContext[*state2](rootNode)

		rootNode.simpleTree2(rootContext, 3, 3)

		go func() {
			rootContext.Close()
		}()

		rootContext.Wait()

	}
}
