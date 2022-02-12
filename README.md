# context
Unfortunately the standard golang [context package](https://github.com/golang/go/tree/master/src/context) does not control order of closing all child contexts. ([issue #51075](https://github.com/golang/go/issues/51075))<br>
(parent context could exit earlier than his child and its child could get unpredicted execution behaviour when try to use some parent resources which is already disposed)

To resolve this issue, this is separate implementation of context package, and it waits till all child contexts will correctly closes before calling close for their parent.


### How to use it:

#### 1. Create root context:
```
context0 := context.Background()
```

#### 2. Create childs and subchilds:
```
context1 := context0.NewChildContext()
context2 := context1.NewChildContext()
context3 := context2.NewChildContext()
```

#### 3. Wait for onCancel event
When all context data have been released after <b>onCancel</b> event, call <b>Disposed()</b> method. It tells to framework that child completely released it resources and it's parent don't need wait this child any more.
```
go func() {
  for {
    select {
    case err := <-ctx.OnCancel():

      fmt.Println(fmt.Sprintf("canceled (%v)", err))

      ctx.Disposed() //                   <--- !!! Here we tell that context completely finished !!!
      return
    }
  }

}()
```

#### 4. Close context and all it's subchilds
```
context0.Cancel(context.Canceled)
```

### Limitations and specific
There are no additional error checks, so there are prohibited scenarios:<br>

a) call <b>Cancel()</b> second time for same <b>context</b><br>
b) cancel context without select and <b>OnCancel</b> handler<br>
c) <b>OnCancel</b> handler without <b>Disposed()</b> method<br>
