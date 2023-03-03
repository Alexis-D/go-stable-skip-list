package pkg

import (
	"fmt"
	"golang.org/x/exp/rand"
)

// Cmp must -1 when a < b, 0 when a == b, 1 when a > b
type Cmp[T any] func(a, b T) int

// StableSkipList implements a skip list that allows for duplicate entries
// insertion order is retained.
type StableSkipList[T any] interface {
	Insert(value T)
	FindFirst(value T) (T, bool)
	DeleteFirst(value T)
	First() (T, bool)
	Last() (T, bool)
	fmt.Stringer
}

type skipList[T any] struct {
	forward []*skipListNode[T]
	rand    *rand.Rand
	cmp     Cmp[T]
}

type skipListNode[T any] struct {
	value   T
	forward []*skipListNode[T]
}

func New[T any](cmp Cmp[T]) StableSkipList[T] {
	return &skipList[T]{
		forward: []*skipListNode[T]{},
		rand:    rand.New(rand.NewSource(0)),
		cmp:     cmp,
	}
}

func (sl *skipList[T]) Insert(value T) {
	path := make([]*skipListNode[T], len(sl.forward), len(sl.forward))
	var node *skipListNode[T]

	for level := len(sl.forward) - 1; level >= 0; level-- {
		if node == nil {
			if sl.cmp(sl.forward[level].value, value) == 1 {
				continue
			}

			node = sl.forward[level]
		}

		for node.forward[level] != nil && sl.cmp(node.forward[level].value, value) <= 0 {
			node = node.forward[level]
		}

		path[level] = node
	}

	nodeToInsert := &skipListNode[T]{
		value: value,
	}

	attemptToGrow := true
	for level := 0; level < len(sl.forward); level++ {
		if level == 0 || (attemptToGrow && sl.rand.Intn(2) == 0) {
			if path[level] != nil {
				nodeToInsert.forward = append(nodeToInsert.forward, path[level].forward[level])
				path[level].forward[level] = nodeToInsert
			} else {
				nodeToInsert.forward = append(nodeToInsert.forward, sl.forward[level])
				sl.forward[level] = nodeToInsert
			}

		} else {
			attemptToGrow = false
		}
	}

	if len(sl.forward) == 0 || (attemptToGrow && sl.rand.Intn(2) == 0) {
		nodeToInsert.forward = append(nodeToInsert.forward, nil)
		sl.forward = append(sl.forward, nodeToInsert)
	}
}

func (sl *skipList[T]) FindFirst(value T) (T, bool) {
	var node *skipListNode[T]

	for level := len(sl.forward) - 1; level >= 0; level-- {
		if node == nil {
			switch sl.cmp(sl.forward[level].value, value) {
			case -1:
				node = sl.forward[level]
			case 0:
				return sl.forward[level].value, true
			case 1:
				continue
			default:
				panic("cmp function is not implemented correctly")
			}
		}

	loop:
		for node.forward[level] != nil {
			switch sl.cmp(node.forward[level].value, value) {
			case -1:
				node = node.forward[level]
			case 0:
				return node.forward[level].value, true
			case 1:
				break loop
			default:
				panic("cmp function is not implemented correctly")
			}
		}
	}

	return *new(T), false
}

func (sl *skipList[T]) DeleteFirst(value T) {
	path := make([]*skipListNode[T], len(sl.forward), len(sl.forward))
	var node *skipListNode[T]

	for level := len(sl.forward) - 1; level >= 0; level-- {
		if node == nil {
			if sl.cmp(sl.forward[level].value, value) >= 0 {
				continue
			}

			node = sl.forward[level]
		}

		for node.forward[level] != nil && sl.cmp(node.forward[level].value, value) == -1 {
			node = node.forward[level]
		}

		path[level] = node
	}

	for level := 0; level < len(sl.forward); level++ {
		if path[level] != nil && path[level].forward[level] != nil && sl.cmp(path[level].forward[level].value, value) == 0 {
			path[level].forward[level] = path[level].forward[level].forward[level]
		} else if path[level] == nil && sl.cmp(sl.forward[level].value, value) == 0 {
			if sl.forward[level].forward[level] == nil {
				sl.forward = sl.forward[:level]
			} else {
				sl.forward[level] = sl.forward[level].forward[level]
			}
		}
	}
}
func (sl *skipList[T]) First() (T, bool) {
	if len(sl.forward) == 0 {
		return *new(T), false
	}

	return sl.forward[0].value, true
}

func (sl *skipList[T]) Last() (T, bool) {
	if len(sl.forward) == 0 {
		return *new(T), false
	}

	node := sl.forward[len(sl.forward)-1]

	for level := len(sl.forward) - 1; level >= 0; level-- {
		for node.forward[level] != nil {
			node = node.forward[level]
		}
	}

	return node.value, true
}

func (sl *skipList[T]) String() string {
	s := ""

	for level := len(sl.forward) - 1; level >= 0; level-- {
		s += fmt.Sprintf("(%d): -> ", level)

		node := sl.forward[level]

		for node != nil {
			s += fmt.Sprintf("%+v -> ", node.value)
			node = node.forward[level]
		}

		s += "nil\n"
	}

	return s
}
