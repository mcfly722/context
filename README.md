# context
Unfortunately the standard golang [context package](https://github.com/golang/go/tree/master/src/context) does not control closing order of child contexts. ([issue #51075](https://github.com/golang/go/issues/51075))<br>
(parent context could exit earlier than his child, and in this case you could get unpredicted execution behaviour when try to use some parent resources which is already disposed)

To resolve this issue, this is another implementation of context package, and it waits till all child contexts will correctly disposes their resources, only after what parent context also would be disposed and closed.

![alt tag](https://raw.githubusercontent.com/mcfly722/goPackages/main/context/schema.svg)

### How to use it:

Full example you can find in ![context_test.go](https://github.com/mcfly722/goPackages/blob/main/context/context_test.go)

#### 1. Implement context.ContextedInstance interface
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
  # this disposer calls when all child contexts would be closed. You can release your handlers/memory for your node here...
}
```
#### 2. Create your node instance(s)
```
func newNode() *node {
	return &node{
		close: make(chan bool),
	}
}

node1 := newNode()
node2 := newNode()
node3 := newNode()

```


#### 3. Create root context
```
newCtx1 := context.NewContextFor(node1)

```
#### 4. Now you can inherit from root context and create childs and subchilds:
```
newCtx2 := newCtx1.NewContextFor(node2)
newCtx3 := newCtx2.NewContextFor(node3)
...
```
#### 4. Closing
When you send <b>close</b> signal to your node, your node exits from Go() loop and goroutine closes. It means, that there are no more child contexts for this node would be created. After that, for all subchilds and childs this library sends OnDone() signal and waits till they closes. Disposer for any node will fires only after all it childs an subchilds would be disposed.

#### Recomendations and limitations
 1. you have always use <b>OnDone()</b> signal check and exits from your select loop when it comes, otherwise your parent goroutine hangs on closing
 2. create new child context only after your context resources initializations and checks occurs.<br> After creating new child context through <b>NewContextFor(...)</b> method, it starts context goroutine, so, all context resources should be already initialized. Do not initialize any resources in your <b>Go(...)</b> method, initialize them before context creating.
 3. there are no any <b>Schedulers</b>, <b>Values</b> and <b>Close()</b> methods like in original library. All this stuff you can easily to do by your self. My purpose here is to create this library as lightweight and as possible. Testing tree parallelism for race conditions are really really hard thing. I totally rewrites and refactored this small piece of code at least <b>8</b> times, until it becomes as simple and predictable as possible.
<br><br>
If you have any suggestions or recommendations, please, use issue tracker, I would be glad to help.
