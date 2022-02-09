# context
Unfortunately the standard golang [context package](https://github.com/golang/go/tree/master/src/context) does not control order of closing all child contexts. ([issue #51075](https://github.com/golang/go/issues/51075))<br>
(parent context could exit earlier than his child and its child could get unpredicted execution behaviour when try to use some parent resources which is already disposed)

To resolve this issue, this is separate implementation of context package, and it waits till all child contexts will correctly closes before calling close for their parent.


### How to use it:

#### 1. Create root context:
```
rootContext := context.Background()
```

#### 2. Create childs and subchilds:
```
childContext1 := rootContext.NewChildContext()
childContext2 := childContext1.NewChildContext()
childContext3 := childContext2.NewChildContext()
```

#### 3. Use select for unblocking listening onCancel channel.
When all context data have been released, call <b>Disposed()</b> method. It tells to framework that child completely released it resources and it's parent don't need wait it any more.
```
go func() {
  for {
    select {
    case err := <-rootContext.OnCancel():
      fmt.Println(fmt.Sprintf("canceled (%v)", err))

      rootContext.Disposed()
      return
    }
  }

}()
```

### Limitations and specific
1. There are no additional error checks so scenarios:<br>

a) call <b>Cancel()</b> second time for same <b>context</b><br>
b) cancel context without select and <b>OnCancel</b> handler<br>
c) <b>OnCancel</b> handler without <b>Disposed()</b> method<br>

this scenarios are not allowed and you should control it by yourself.

2. There are no mutex'es in this module for any contexts to block operations between delete and add to childs in context tree.<br>
Instead of it, it is implemented through several channels to root goroutine node which orchestrate all this collisions. This code little bit more complex, but stable and fast
(race conditions and collisions tested in additional tests).
