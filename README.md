# context
![Version: version](https://img.shields.io/badge/version-v1.0.6-success.svg)
![Tests: tests](https://img.shields.io/badge/tests-✔6|✘0-success.svg)
[![License: GPL3.0](https://img.shields.io/badge/License-GPL3.0-blue.svg)](https://www.gnu.org/licenses/gpl-3.0.html)
<br>
Unfortunately, the standard golang [context package](https://github.com/golang/go/tree/master/src/context) does not control the closing order of child contexts ([issue #51075](https://github.com/golang/go/issues/51075)).<br>
(the parent context could exit earlier than his child, and in this case, you could get unpredicted execution behavior when you try to use some parent resources that are already closed)

To resolve this issue, here is another implementation of this context pattern.<br>
It waits until all child contexts correctly close (parent event loop would be available for servicing their children). Only when all children would be closed parents will exits too.

### Documentation: [GoDoc](https://pkg.go.dev/github.com/mcfly722/context)

### How to use it:

A full example can be found here: [examples_test.go](https://github.com/mcfly722/goPackages/blob/main/context/examples_test.go)


#### 1. Add a new context import to your project:
```
import (
	context "github.com/mcfly722/context"
)
```
#### 2. Implement context.ContextedInstance interface with your node
Your node should contain <b>Go(..)</b> method:
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
#### 4. Now you can inherit from the root context and create child and subchild contexts:
```
ctx1, err := ctx0.NewContextFor(node1)
if err != nil {
	// child context is not created successfully, possibly the parent is in a closing state, you just need to exit
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
It would close all contexts in reverse order: 3->2->1->root.

### Restrictions
 1. Do not exit from your context goroutine without checking that *current.Context()* channel is closed. It is a potential lock or race, and this library restricts it (panic occurs especially to exclude this code mistake).<br>
 2. Always check NewContextFor(...) error. A parent could be in a closed state; in this case, a child would not be created.<br>

### Common questions
 1. Is there any send method to send some control messages from parent to child to change their state?<br>
 No. The only possible way to implement this without races, is to use the same channel that currently used for closing. Unfortunately, GoLang has library race between channel.Send and channel.Close methods (see [issue #30372](https://github.com/golang/go/issues/30372)).
 2. I want to wait until child context is closed. Where is <b>context.Wait()</b>?<br>
 <b>context.Wait()</b> is a race condition potential mistake. You send close to the child and wait for the parent, but children at this moment do not know anything about closing. It continues to send data to parents through channels. Parent blocked, it waits with <b>context.Wait()</b>. The child was also blocked on channel sending. It is a full dead block.
 3. Why does <b>rootContext.Wait()</b> exist?<br>
 <b>rootContext</b> has its own empty goroutine loop without any send or receive, so deadblock from scenario 3 is not possible.
 4. Where is '<b>Deadlines</b>','<b>Timeouts</b>','<b>Values</b>' like in the original context?<br>
<b>It's all sugar.</b><br>
This timeout or deadlines you can implement in your select loop (see: [<-time.After(...)](https://pkg.go.dev/time#After) or [<-time.Tick(...)](https://pkg.go.dev/time#Tick))<br>
For values, use the constructor function with parameters.<br>

### By the way
My goal here is to simplify this not-trivial hierarchical multi-threaded pattern as much as possible without any compromises.<br>
Its main responsibility is to provide safety-sequenced closing through parent-child hierarchy (graceful shutdown). and nothing more is unnecessary.<br><br>