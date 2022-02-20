# context
Unfortunately the standard golang [context package](https://github.com/golang/go/tree/master/src/context) does not control order of closing all child contexts. ([issue #51075](https://github.com/golang/go/issues/51075))<br>
(parent context could exit earlier than his child and its child could get unpredicted execution behaviour when try to use some parent resources which is already disposed)

To resolve this issue, this is separate implementation of context package, and it waits till all child contexts will correctly closes before calling close for their parent.


### How to use it:

#### 1. Create root context:
```
context.NewContextTree()
```

#### 2. Create childs and subchilds:
```
context1 := context0.NewChildContext()
context2 := context1.NewChildContext()
context3 := context2.NewChildContext()
```

#### 3. Specify disposer function (if required)
This function would be called on cancel event for current context when parent cancel it before OnDone event.
```
context.SetDisposer(func(err error) {
  fmt.Println(fmt.Sprintf("disposed with cause: %v", err))
})
```

#### 4. Specify deadline for context (if required)
```
context.SetDeadline(time.Now().Add(time.Second * 1))
```

#### 5. Specify OnDone() handler to exit from goroutine (always required!)
```
go func() {
  for {
    select {
    case err := <-ctx.OnDone():  // OnDone signal tells that context goroutine could be closed
                                 // immediately (all context resources already have been disposed)
      return

    default:
      {
        // some work...Do Not stay with empty Default!
      }
    }
  }
}()
```

#### 6. Call context cancelling with some reason
```
context0.Cancel(context.ErrCanceled)
```


### Limitations and specific
* Always use OnDone() handler. Context always tries to send this event on cancel, so, if would not any reader obtain this event from this channel, program just hanged.
* Create new child context only after successfull start new connection/execution/other your job directly before select loop (otherwise it exits on error and OnDone() handler would be missed).
