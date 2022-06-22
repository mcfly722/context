# context
Unfortunately the standard golang [context package](https://github.com/golang/go/tree/master/src/context) does not control closing order of child contexts. ([issue #51075](https://github.com/golang/go/issues/51075))<br>
(parent context could exit earlier than his child, and in this case you could get unpredicted execution behaviour when you try to use some parent resources which is already disposed)

To resolve this issue, this is another implementation of context package, and it waits till all child contexts will correctly disposes their resources (parent event loop would be available for servicing it), only after what parent context also would be exited from loop and disposes it resources.

![alt tag](https://raw.githubusercontent.com/mcfly722/goPackages/main/context/schema.svg)

### How to use it:

Full example you can find in [context_test.go](https://github.com/mcfly722/goPackages/blob/main/context/context_test.go)

#### 1. Implement context.ContextedInstance interface
Define your instance with <b>Go(..)</b> method:
```
type node struct {
  close chan bool
}

func (node *node) Go(current context.Context) {
  loop:
  	for {
  		select {
  		  case <-node.close:
  		    context.Cancel()
  		    break                // !!!!! do not exit from loop here! panic will occur if some childs are left unclosed
  		  case _, opened <-current.Opened():
  		    if !opened {
  		      break loop
  		    }
  		    break
  		}
  	}

  # you can dispose your context resources here ...
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
rootCtx := context.NewRootContext(context.NewEmptyDebugger())
```
You also can use Console debugger like this:
```
rootCtx := context.NewRootContext(context.NewConsoleLogDebugger())

```
Or implement your own debugger (<b>Debugger</b> interface)

#### 4. Now you can inherit from root context and create childs and subchilds:
```

newCtx1 := rootCtx.NewContextFor(node1, "1", "node")
newCtx2, err := newCtx1.NewContextFor(node2, "2", "node")
if err!=nil {return err}
newCtx3, err := newCtx2.NewContextFor(node3, "3", "node")
if err!=nil {return err}
...
```
#### 4. Closing
Do not close your events loop using your own chan events. For it, use <b>current.Close()</b> call. This method correctly closes all your childs without blocking, only after that it closes your goroutine through <b>current.Opened()</b> channel.


#### Recomendations and limitations
 1. you have always use <b>current.Close()</b> call to exit from current goroutine, do not exit from your loop on external signals
 2. use <b>NewContextFor()</b> only from started goroutine. Do not call it from constructors or parents.
 3. I want wait till child context will be closed. Where is <b>context.Wait()</b>?<br>
 <b>context.Wait()</b> is a race condition potential mistake. You send close to child and wait in parent, but childs at this moment do not know anything about closing. It continues to send data to parent through channels. Parent blocked, it waits with <b>contenxt.Wait()</b>. Child also blocked on channel send. It is full dead block.
 4. Why <b>rootContext.Wait()</b> exists?<br>
 <b>rootContext</b> has its own empty goroutine loop without any send/receive, so, deadblock from scenario 3 is not possible.
 <br><br>
If you have any suggestions or recommendations, please, use issue tracker, I would be glad to help.
