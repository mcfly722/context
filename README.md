# context
![Version: version](https://img.shields.io/badge/version-v1.0.3-success.svg)
![Tests: tests](https://img.shields.io/badge/tests-✔6|✘0-success.svg)
[![License: GPL3.0](https://img.shields.io/badge/License-GPL3.0-blue.svg)](https://www.gnu.org/licenses/gpl-3.0.html)
<br>
Unfortunately the standard golang [context package](https://github.com/golang/go/tree/master/src/context) does not control closing order of child contexts ([issue #51075](https://github.com/golang/go/issues/51075)).<br>
(parent context could exit earlier than his child, and in this case you could get unpredicted execution behaviour when you try to use some parent resources which is already closed)

To resolve this issue, here is another implementation of this context pattern.<br>
It waits till all child contexts will correctly closes (parent event loop would be available for servicing it childs). Only when all childs would be closed, then parent would exit to.

### Documentation: [GoDoc](https://pkg.go.dev/github.com/mcfly722/context)

### How to use it:

Full example you can find here: [examples_test.go](https://github.com/mcfly722/goPackages/blob/main/context/examples_test.go)


#### 1. Add new context import to your project
```
import (
	context "github.com/mcfly722/context"
)
```
#### 2. Implement context.ContextedInstance interface with your node
Your node should contains <b>Go(..)</b> method:
```
type node struct {
	name: string,
}

func (node *node) Go(current context.Context) {
loop:
	for {
		select {
		case _, isOpened := <-current.Context(): // this method returns context channel. If it closes, it means that we need to finish select loop
			if !isOpened {
				break loop
			}
		default: // you can use default or not, it works in both cases
			{
			}
		}
	}
	fmt.Printf("context %v closed\n", node.name)
}
```
#### 3. Create your node instance(s)
```
node0 := &node{name : "root"}
node1 := &node{name : "1"}
node2 := &node{name : "2"}
node3 := &node{name : "3"}
```
#### 3. Create root context
```
ctx0 := context.NewRootContext(node0)
```
#### 4. Now you can inherit from root context and create childs and subchilds contexts:
```
ctx1, err := ctx0.NewContextFor(node1)
if err != nil {
	// child context does not created successfully, possibly parent is in closing state, you need just exit
} else {
	// child context created successfully
}
```
```
ctx2, err := ctx1.NewContextFor(node2)
...
```
```
ctx3, err := ctx2.NewContextFor(node3)
...
```
#### 5. Closing
```
ctx0.Close()
```
It would close all contexts in reverse order 3->2->1->root.

### Restrictions
 1. Do not exit from your context goroutine without checking that *current.Context()* channel is closed. It is potential lock or race, and this library restricts it (panic occurs especially to exclude this code mistake).<br>
 2. Always check NewContextFor(...) error. Parent could be in closing state, it this case child would not be created.<br>

### Common questions
 1. I want wait till child context will be closed. Where is <b>context.Wait()</b>?<br>
 <b>context.Wait()</b> is a race condition potential mistake. You send close to child and wait in parent, but childs at this moment do not know anything about closing. It continues to send data to parent through channels. Parent blocked, it waits with <b>contenxt.Wait()</b>. Child also blocked on channel send. It is full dead block.
 2. Why <b>rootContext.Wait()</b> exists?<br>
 <b>rootContext</b> has its own empty goroutine loop without any send/receive, so, deadblock from scenario 3 is not possible.
 3. Where is '<b>Deadlines</b>','<b>Timeouts</b>','<b>Values</b>' like in original context?<br>
<b>It's all sugar.</b><br>
This timeouts/deadlines you can implement in your select loop (see: [<-time.After(...)](https://pkg.go.dev/time#After) or [<-time.Tick(...)](https://pkg.go.dev/time#Tick))<br>
For values use constructor function with parameters.<br>

### By the way
My goal here is to simplify this not trivial hierarchical multi thread pattern as much as possible without any compromises.<br>
It main responsibility - provide safety sequenced closing through parent/child hierarchy (graceful shutdown). and nothing more unnecessary!<br>
I hope my efforts are not in vain, and I would be glad to any recommendations and suggestions.<br>
<br>
McFly.
