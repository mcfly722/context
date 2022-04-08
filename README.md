# context
Unfortunately the standard golang [context package](https://github.com/golang/go/tree/master/src/context) does not control order of closing all child contexts. ([issue #51075](https://github.com/golang/go/issues/51075))<br>
(parent context could exit earlier than his child and its child could get unpredicted execution behaviour when try to use some parent resources which is already disposed)

To resolve this issue, this is another implementation of context package, and it waits till all child contexts will correctly disposes their resources, only after what parent context also would be disposed and closed.

![alt tag](https://raw.githubusercontent.com/mcfly722/goPackages/main/context/schema.svg)

### How to use it:

Full example you can find in </b>context_test.go</b>

#### 1. Implement ContextedInstance interface
Define your instance with <b>Go(..)</b> and <b>Dispose()</b> methods
```
type node struct {
  close chan bool
}

func (node *node) Go(current context.Context) {
  loop:
  	for {
  		select {
  		case <-node.close:
  			break loop
  		case <-current.OnDone():
  			break loop
  		}
  	}
}

func (node *node) Dispose() {
  # this disposer calls when all child contexts would be closed. You can release your handlers/memory here...
}
```


#### 2. Create root context
```
newCtx1 := context.NewContextFor(node1)

```
#### 3. Create childs and subchilds:
```
newCtx2 := newCtx1.NewContextFor(node2)
newCtx3 := newCtx2.NewContextFor(node3)
...
```
#### 4. Closing
When you send <b>close</b> signal, your node context exits from loop and goroutine closes. It means that there are no more child contexts for this node would be created. For all subchilds and childs sends OnDone() signal and they starts to close. Disposer for current node will fires only after all childs contexts would be disposed.
