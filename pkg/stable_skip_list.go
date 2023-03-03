package pkg

import (
	"fmt"
	"golang.org/x/exp/rand"
)

// Cmp must -1 when a < b, 0 when a == b, 1 when a > b
type Cmp[T any] func(a, b T) int

// StableSkipList implements a skip list that allows for duplicate entries.
// Insertion order is retained.
type StableSkipList[T any] interface {
	// Insert always inserts a new value, even if the value is equal to one already in the list.
	// In the case of duplicates, the new value will be inserted _after_ the existing one(s).
	Insert(value T)
	// FindFirst's bool indicates whether the value was actually found
	FindFirst(value T) (T, bool)
	DeleteFirst(value T)
	// First's bool indicates whether the value was actually found
	First() (T, bool)
	// Last's bool indicates whether the value was actually found
	Last() (T, bool)
	fmt.Stringer
}

type stableSkipList[T any] struct {
	forward []*stableSkipListNode[T]
	rand    *rand.Rand
	cmp     Cmp[T]
}

type stableSkipListNode[T any] struct {
	value   T
	forward []*stableSkipListNode[T]
}

func New[T any](cmp Cmp[T]) StableSkipList[T] {
	return &stableSkipList[T]{
		forward: []*stableSkipListNode[T]{},
		rand:    rand.New(rand.NewSource(0)),
		cmp:     cmp,
	}
}

func (ssl *stableSkipList[T]) Insert(value T) {
	traversedNodesByLevel := make([]*stableSkipListNode[T], len(ssl.forward), len(ssl.forward))
	var smallerOrEqualNode *stableSkipListNode[T]

	// we starts at the highest layer and try to make our way towards the insertion point
	// when we can't go any further to the right, we go down to the next layer
	// see figure 3 in the original paper: https://www.cl.cam.ac.uk/teaching/2006/AlgorithI/skiplists.pdf
	// Skip Lists: A Probabilistic Alternative to Balanced Trees (W. Pugh)
	//
	// the FindFirst/DeleteFirst logic follows roughly the same logic, main difference is that Insert
	// aims to go to the right of duplicates, while the other two methods aim to the left
	for level := len(ssl.forward) - 1; level >= 0; level-- {
		if smallerOrEqualNode == nil {
			if ssl.cmp(ssl.forward[level].value, value) == 1 {
				continue
			}

			smallerOrEqualNode = ssl.forward[level]
		}

		for smallerOrEqualNode.forward[level] != nil && ssl.cmp(smallerOrEqualNode.forward[level].value, value) <= 0 {
			smallerOrEqualNode = smallerOrEqualNode.forward[level]
		}

		traversedNodesByLevel[level] = smallerOrEqualNode
	}

	nodeToInsert := &stableSkipListNode[T]{
		value: value,
	}

	attemptToGrow := true
	for level := 0; level < len(ssl.forward); level++ {
		// we always need to insert the node if we are on the bottom layer, otherwise we randomly attempt to grow
		// iff we've grown at the previous levels
		if level == 0 || (attemptToGrow && ssl.rand.Intn(2) == 0) {
			if traversedNodesByLevel[level] != nil {
				nodeToInsert.forward = append(nodeToInsert.forward, traversedNodesByLevel[level].forward[level])
				traversedNodesByLevel[level].forward[level] = nodeToInsert
			} else {
				nodeToInsert.forward = append(nodeToInsert.forward, ssl.forward[level])
				ssl.forward[level] = nodeToInsert
			}

		} else {
			attemptToGrow = false
			// we can't grow anymore there's no need to keep looping
			break
		}
	}

	// handle the case where the list is empty and/or we want to grow to a whole new level
	if len(ssl.forward) == 0 || (attemptToGrow && ssl.rand.Intn(2) == 0) {
		nodeToInsert.forward = append(nodeToInsert.forward, nil)
		ssl.forward = append(ssl.forward, nodeToInsert)
	}
}

func (ssl *stableSkipList[T]) FindFirst(value T) (T, bool) {
	var smallerNode *stableSkipListNode[T]

	for level := len(ssl.forward) - 1; level >= 0; level-- {
		if smallerNode == nil {
			switch ssl.cmp(ssl.forward[level].value, value) {
			case -1:
				smallerNode = ssl.forward[level]
			case 0:
				return ssl.forward[level].value, true
			case 1:
				continue
			default:
				panic("cmp function is not implemented correctly")
			}
		}

	loop:
		for smallerNode.forward[level] != nil {
			switch ssl.cmp(smallerNode.forward[level].value, value) {
			case -1:
				smallerNode = smallerNode.forward[level]
			case 0:
				return smallerNode.forward[level].value, true
			case 1:
				break loop
			default:
				panic("cmp function is not implemented correctly")
			}
		}
	}

	return *new(T), false
}

func (ssl *stableSkipList[T]) DeleteFirst(value T) {
	traversedNodesByLevel := make([]*stableSkipListNode[T], len(ssl.forward), len(ssl.forward))
	var smallerNode *stableSkipListNode[T]

	for level := len(ssl.forward) - 1; level >= 0; level-- {
		if smallerNode == nil {
			if ssl.cmp(ssl.forward[level].value, value) >= 0 {
				continue
			}

			smallerNode = ssl.forward[level]
		}

		for smallerNode.forward[level] != nil && ssl.cmp(smallerNode.forward[level].value, value) == -1 {
			smallerNode = smallerNode.forward[level]
		}

		traversedNodesByLevel[level] = smallerNode
	}

	for level := 0; level < len(ssl.forward); level++ {
		if traversedNodesByLevel[level] != nil &&
			traversedNodesByLevel[level].forward[level] != nil &&
			ssl.cmp(traversedNodesByLevel[level].forward[level].value, value) == 0 {
			traversedNodesByLevel[level].forward[level] = traversedNodesByLevel[level].forward[level].forward[level]
		} else if traversedNodesByLevel[level] == nil && ssl.cmp(ssl.forward[level].value, value) == 0 {
			if ssl.forward[level].forward[level] == nil {
				// no more nodes at this level, so we just shrinks the forward pointer slice
				ssl.forward = ssl.forward[:level]
				return
			} else {
				// we skip over the node to delete
				ssl.forward[level] = ssl.forward[level].forward[level]
			}
		}
	}
}

func (ssl *stableSkipList[T]) First() (T, bool) {
	if len(ssl.forward) == 0 {
		return *new(T), false
	}

	return ssl.forward[0].value, true
}

func (ssl *stableSkipList[T]) Last() (T, bool) {
	if len(ssl.forward) == 0 {
		return *new(T), false
	}

	node := ssl.forward[len(ssl.forward)-1]

	for level := len(ssl.forward) - 1; level >= 0; level-- {
		for node.forward[level] != nil {
			node = node.forward[level]
		}
	}

	return node.value, true
}

func (ssl *stableSkipList[T]) String() string {
	s := ""

	for level := len(ssl.forward) - 1; level >= 0; level-- {
		s += fmt.Sprintf("(%d): -> ", level)

		node := ssl.forward[level]

		for node != nil {
			s += fmt.Sprintf("%+v -> ", node.value)
			node = node.forward[level]
		}

		s += "nil\n"
	}

	return s
}
