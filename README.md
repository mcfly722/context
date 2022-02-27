# context
Unfortunately the standard golang [context package](https://github.com/golang/go/tree/master/src/context) does not control order of closing all child contexts. ([issue #51075](https://github.com/golang/go/issues/51075))<br>
(parent context could exit earlier than his child and its child could get unpredicted execution behaviour when try to use some parent resources which is already disposed)

To resolve this issue, this is another implementation of context package, and it waits till all child contexts will correctly disposes their resources, only after what parent context also would be disposed and closed.


### How to use it:

#### 1. Create root context:
```
rootContext := context.NewContextTree(
  func(reason context.Reason) {           // desposer function
    fmt.Println(fmt.Sprintf("root disposed. Reason=%v", reason))
  }, func(reason context.Reason) {        // finalizer function
    fmt.Println(fmt.Sprintf("root finished. Reason=%v", reason))
    onDone <- true
  })
```
<b>Disposer</b> - this function release context resources (connections, caches, etc..)<br>
<b>Finalizer</b> - this function calls only after canceled and disposed context and it main purpose is to tell to goroutine to finish it waiting select loop.<br>
Both this functions are separated, because there are scenario 1,2 when you try to create child for already closed context. In this case disposer would be called directly from <b>NewChildContext</b> method, but finalizer - not. This is required to release context resources before exiting (goroutine with waiting loop should not be created).

#### 2. Create childs and subchilds:
```
onDone := make(chan bool)

ctx, err := rootContext.NewChildContext(
  func(reason context.Reason) { // Disposer
    fmt.Println(fmt.Sprintf("%v disposed", currentContextName))
  }, func(reason context.Reason) { // Finalizer
    onDone <- true
  })

if err != nil { # parent context already canceled (disposer called automatically, finalizer not)
  fmt.Printf(fmt.Sprintf("error: %v", err))
  return nil
}


```

#### 3. Wait till context would be canceled for OnDone signal from finalizer
This function would be called on cancel event for current context when parent cancel it before OnDone event.
```
go func() {
  for {
    select {
    case <-onDone:
      return
    }
  }
}()

```
#### 4. Specify deadline for context (if required)
```
ctx.SetDeadline(time.Now().Add(time.Second * 1))
```


#### 5. Call context cancelling with some reason
```
ctx.Cancel(context.ReasonCanceled)
```
<br><br>

### Controversial Scenarios

##### 1. Creating child from canceled context
In this scenario disposer calling immediately directly from <b>NewChildContext</b> and <b>NewChildContext</b> returns error. You should exit from function with this error without starting any waiting goroutine.

##### 2. Creating and Canceling context on same Tree from different goroutines
It is supported scenario and it resolved with <b>tree.changesAllowed</b> mutex. There are one mutex for whole tree for all structure changes.

##### 3. Calling cancel on already canceled context
This scenario possible only if there are some code error or race condition between cancelling same context from different goroutines. Panic generates and this code should be rewrited.

##### 4. Empty Disposer and Finalizer
Supported scenario
