package context

import (
	"sync"
)

type tree struct {
	new      chan *treeMessage
	close    chan *treeMessage
	finished chan *treeMessage
	root     *context
	ready    sync.Mutex
}

func newTree(rootContext *context) *tree {

	tree := &tree{
		new:      make(chan *treeMessage),
		close:    make(chan *treeMessage),
		finished: make(chan *treeMessage),
		root:     rootContext,
	}

	go func() {
		select {
		case treeMessage := <-tree.new:
			treeMessage.answer(func() error {
				tree.ready.Lock()
				defer tree.ready.Unlock()
				switch treeMessage.context.parent.state {
				case freezed:
					return &CancelInProcessError{}
				case disposing:
					return &CancelInProcessError{}
				default:
					{
						newContext := treeMessage.context
						newContext.parent.childs[newContext] = newContext

						go func(context *context) {
							context.instance.Go(context)
							if context.state != disposing {
								panic(ExitFromContextWithoutCancelPanic)
							}
							sendContextTo(context.tree.finished, context)
						}(newContext)

					}
				}
				return nil
			})
		case treeMessage := <-tree.close:
			treeMessage.answer(func() error {

				if treeMessage.context.state == disposing {
					panic(CancelFromDisposeStatePanic)
				}

				treeMessage.context.freezeAllChildsAndSubchilds()

				return nil
			})
		case treeMessage := <-tree.finished:
			treeMessage.answer(func() error {

				delete(treeMessage.context.parent.childs, treeMessage.context)

				tree.disposeFreezedEmptyParents()

				return nil
			})
		default:
			{
			}

		}

		close(tree.new)
		close(tree.close)

	}()

	return tree
}

func (current *context) freezeAllChildsAndSubchilds() {
	if current.state == working {
		current.state = freezed
		for child := range current.childs {
			child.freezeAllChildsAndSubchilds()
		}
	}
}

func (tree *tree) disposeFreezedEmptyParents() {
	tree.root.disposeFreezedEmptyParentsRecursively()
}

func (current *context) disposeFreezedEmptyParentsRecursively() {
	for child := range current.childs {
		child.disposeFreezedEmptyParentsRecursively()
	}
	if current.state == freezed && len(current.childs) == 0 {
		current.state = disposing
	}
}
