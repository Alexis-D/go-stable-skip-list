package pkg

import (
	"fmt"
	"golang.org/x/exp/rand"
	"math/bits"
	"strings"
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
	// FindFirstGreaterEq(value T) (T, bool)
	DeleteFirst(value T)
	// First's bool indicates whether the value was actually found
	First() (T, bool)
	// Last's bool indicates whether the value was actually found
	Last() (T, bool)
	fmt.Stringer
}

type randUint32Fn func() uint32

type stableSkipList[T any] struct {
	heads      []*stableSkipListNode[T]
	randUint32 randUint32Fn
	cmp        Cmp[T]
}

type stableSkipListNode[T any] struct {
	value   T
	forward []*stableSkipListNode[T]
}

func (ssl *stableSkipList[T]) insertHead(level int, newNode *stableSkipListNode[T]) {
	newNode.forward[level] = ssl.heads[level]
	ssl.heads[level] = newNode
}

func (ssln *stableSkipListNode[T]) insertAfter(level int, newNode *stableSkipListNode[T]) {
	newNode.forward[level] = ssln.forward[level]
	ssln.forward[level] = newNode
}

func New[T any](cmp Cmp[T]) StableSkipList[T] {
	r := rand.New(rand.NewSource(0))
	return &stableSkipList[T]{
		heads: []*stableSkipListNode[T]{},
		randUint32: func() uint32 {
			return r.Uint32()
		},
		cmp: cmp,
	}
}

// Insert inserts the value in the list.
//
// see figure 3 in the original paper: https://www.cl.cam.ac.uk/teaching/2006/AlgorithI/skiplists.pdf
// Skip Lists: A Probabilistic Alternative to Balanced Trees (W. Pugh)
//
// This implementation differs from the paper as it needs to account for duplicates, to do that it needs to follow
// a couple rules:
// * at level i, the first node with value v _is guaranteed_ to be the first v value inserted into the list
// * when inserting v, it will be inserted to the right of all equal nodes at that level
func (ssl *stableSkipList[T]) Insert(value T) {
	nodeToInsert := &stableSkipListNode[T]{
		value: value,
	}
	newHeight := ssl.newHeight()
	existing := ssl.findFirstNode(value)

	if existing != nil && newHeight > len(existing.forward) {
		// we will need to grow the _existing_ node
		// the new node will use the height of the existing node to avoid growing the list unnecessarily
		nodeToInsert.forward = make([]*stableSkipListNode[T], len(existing.forward), len(existing.forward))
		existing.forward = append(existing.forward, make(
			[]*stableSkipListNode[T],
			newHeight-len(existing.forward),
			newHeight-len(existing.forward),
		)...)
	} else {
		nodeToInsert.forward = make([]*stableSkipListNode[T], newHeight, newHeight)
	}

	originalHeadHeight := len(ssl.heads)
	if newHeight > originalHeadHeight {
		// we need to grow the whole list
		if existing != nil {
			ssl.heads = append(ssl.heads, existing)
		} else {
			ssl.heads = append(ssl.heads, nodeToInsert)
		}
	}

	// largestSmaller is the node that will come _right before_ a node with the value we're looking for on a given level
	// largestEq if the last node on a level to come with the value we're looking for
	var largestSmaller, largestEq *stableSkipListNode[T]
	for level := originalHeadHeight - 1; level >= 0; level-- {
		if largestSmaller == nil || largestEq == nil {
			switch ssl.cmp(ssl.heads[level].value, value) {
			case -1: // head is smaller than our target
				if largestSmaller == nil {
					largestSmaller = ssl.heads[level]
				}
			case 0: // head is equal to our target
				if largestEq == nil {
					largestEq = ssl.heads[level]
				}
			case 1: // head is larger than our target
				if level < len(nodeToInsert.forward) {
					// we make the list head point to our new node, and our new node to the old head
					ssl.insertHead(level, nodeToInsert)
				}

				if existing != nil {
					if level < len(existing.forward) {
						// we insertAfter our existing node at the list' head, and point the existing node to the old head
						ssl.insertHead(level, existing)
					}
				}

				continue
			}
		}

		for largestSmaller != nil && largestSmaller.forward[level] != nil && ssl.cmp(largestSmaller.forward[level].value, value) == -1 {
			largestSmaller = largestSmaller.forward[level]
		}

		if largestEq == nil && largestSmaller.forward[level] != nil && ssl.cmp(largestSmaller.forward[level].value, value) == 0 {
			largestEq = largestSmaller.forward[level]
		}

		for largestEq != nil && largestEq.forward[level] != nil && ssl.cmp(largestEq.forward[level].value, value) == 0 {
			largestEq = largestEq.forward[level]
		}

		if largestSmaller != nil {
			if existing != nil {
				if largestSmaller.forward[level] != existing && level < len(existing.forward) {
					// we need to grow our existing node
					largestSmaller.insertAfter(level, existing)
				}
			}
		}

		if largestEq != nil {
			if level < len(nodeToInsert.forward) {
				// we insertAfter to the right
				largestEq.insertAfter(level, nodeToInsert)
			}
		}

		if largestSmaller != nil && largestEq == nil {
			if level < len(nodeToInsert.forward) {
				// we're inserting a new non-dupe value into the tree
				largestSmaller.insertAfter(level, nodeToInsert)
			}
		}
	}
}

// newHeight returns an integer in the range [1, min(33, len(ssl.heads)+1)]
func (ssl *stableSkipList[T]) newHeight() int {
	height := bits.TrailingZeros32(ssl.randUint32())

	if height == 0 {
		return 1
	} else if height <= len(ssl.heads) {
		return height
	}

	return len(ssl.heads) + 1
}

func (ssl *stableSkipList[T]) FindFirst(value T) (T, bool) {
	maybeNode := ssl.findFirstNode(value)
	if maybeNode != nil {
		return maybeNode.value, true
	}
	return *new(T), false
}

func (ssl *stableSkipList[T]) findFirstNode(value T) *stableSkipListNode[T] {
	var smallerNode *stableSkipListNode[T]

	for level := len(ssl.heads) - 1; level >= 0; level-- {
		if smallerNode == nil {
			switch ssl.cmp(ssl.heads[level].value, value) {
			case -1:
				smallerNode = ssl.heads[level]
			case 0:
				return ssl.heads[level]
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
				return smallerNode.forward[level]
			case 1:
				break loop
			default:
				panic("cmp function is not implemented correctly")
			}
		}
	}

	return nil
}

func (ssl *stableSkipList[T]) DeleteFirst(value T) {
	var smallerNode *stableSkipListNode[T]
	for level := len(ssl.heads) - 1; level >= 0; level-- {
		if smallerNode == nil {
			switch ssl.cmp(ssl.heads[level].value, value) {
			case -1:
				smallerNode = ssl.heads[level]
			case 0: // we found a head to delete
				nodeToDelete := ssl.heads[level]
				if nodeToDelete.forward[level] != nil && ssl.cmp(nodeToDelete.forward[level].value, value) == 0 {
					// we have a dupe on this level
					nextLogicalDupe := nodeToDelete.forward[0]
					if nodeToDelete.forward[level] != nextLogicalDupe && len(nodeToDelete.forward) > len(nextLogicalDupe.forward) {
						// and the dupe was not the next logical dupe (we know that from looking at the bottom layer)
						// we make the next logic dupe copy the tree of the node we are deleting
						nextLogicalDupe.forward = append(
							nextLogicalDupe.forward,
							nodeToDelete.forward[len(nextLogicalDupe.forward):len(nodeToDelete.forward)]...)
					}

					// point the head to the next logical dupe
					ssl.heads[level] = nextLogicalDupe
				} else {
					// no dupe on this level, we can shrink our node and remove it from this level
					next := nodeToDelete.forward[level]
					nodeToDelete.forward = nodeToDelete.forward[:len(nodeToDelete.forward)-1]
					ssl.heads[level] = next

					// ensure invariant: no heads point to nil
					if ssl.heads[level] == nil {
						ssl.heads = ssl.heads[:len(ssl.heads)-1]
					}
				}
			case 1:
				continue
			}
		}

		for smallerNode != nil && smallerNode.forward[level] != nil && ssl.cmp(smallerNode.forward[level].value, value) == -1 {
			smallerNode = smallerNode.forward[level]
		}

		if smallerNode != nil && smallerNode.forward[level] != nil && ssl.cmp(smallerNode.forward[level].value, value) == 0 {
			nodeToDelete := smallerNode.forward[level]
			if nodeToDelete.forward[level] != nil && ssl.cmp(nodeToDelete.value, nodeToDelete.forward[level].value) == 0 {
				nextLogicalDupe := nodeToDelete.forward[0]
				if nodeToDelete.forward[level] != nextLogicalDupe && len(nodeToDelete.forward) > len(nextLogicalDupe.forward) {
					nextLogicalDupe.forward = append(
						nextLogicalDupe.forward,
						nodeToDelete.forward[len(nextLogicalDupe.forward):len(nodeToDelete.forward)]...)
				}

				smallerNode.forward[level] = nextLogicalDupe
			} else {
				next := nodeToDelete.forward[level]
				nodeToDelete.forward = nodeToDelete.forward[:len(nodeToDelete.forward)-1]
				smallerNode.forward[level] = next
			}
		}
	}
}

func (ssl *stableSkipList[T]) First() (T, bool) {
	if len(ssl.heads) == 0 {
		return *new(T), false
	}

	return ssl.heads[0].value, true
}

func (ssl *stableSkipList[T]) Last() (T, bool) {
	if len(ssl.heads) == 0 {
		return *new(T), false
	}

	node := ssl.heads[len(ssl.heads)-1]
	for level := len(ssl.heads) - 1; level >= 0; level-- {
		for node.forward[level] != nil {
			node = node.forward[level]
		}
	}
	return node.value, true
}

func (ssl *stableSkipList[T]) String() string {
	var sb strings.Builder
	for level := len(ssl.heads) - 1; level >= 0; level-- {
		sb.WriteString(fmt.Sprintf("(%d): -> ", level))
		node := ssl.heads[level]
		for node != nil {
			sb.WriteString(fmt.Sprintf("%+v -> ", node.value))
			node = node.forward[level]
		}
		sb.WriteString("nil\n")
	}
	return sb.String()
}
