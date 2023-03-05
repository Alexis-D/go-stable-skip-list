package pkg

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
	"sort"
	"testing"
)

type item struct {
	insertedAt int
	value      int
}

func cmp() func(a item, b item) int {
	return func(a, b item) int {
		if a.value < b.value {
			return -1
		} else if a.value == b.value {
			return 0
		}

		return 1
	}
}

func fixedRand(n uint32) randUint32Fn {
	return func() uint32 {
		return n
	}
}

// checkInvariants attempts to check invariants such as:
// * if a layer is above another one, it should have fewer nodes
// * if a value is left of another in a layer, then it should be smaller or equal
// * no node is taller than the heads
// * insertion order within one level is respected, and the earliest insertAfter always shows up first when its value appears on a level
// * no nil head
func checkInvariants(t *testing.T, sl *stableSkipList[item]) {
	assert.NotNil(t, sl.heads)

	prevLevelSize := -1
	valueToFirstInsertion := make(map[int]int)
	headHeight := len(sl.heads)
	for level := 0; level < headHeight; level++ {
		head := sl.heads[level]
		levelSize := 0
		valuesSeenInThisLayer := make(map[int]bool)
		assert.NotNil(t, head)

		for head != nil {
			if existing, ok := valueToFirstInsertion[head.value.value]; ok {
				if !valuesSeenInThisLayer[head.value.value] {
					// if a value appear on a layer, the leftmost dupe should be the first to have been inserted
					assert.Equal(t, existing, head.value.insertedAt)
				} else {
					assert.Greater(t, head.value.insertedAt, existing)
				}
			} else {
				valueToFirstInsertion[head.value.value] = head.value.insertedAt
			}
			valuesSeenInThisLayer[head.value.value] = true

			if head.forward[level] != nil {
				assert.GreaterOrEqual(t, head.forward[level].value.value, head.value.value)
				if head.value == head.forward[level].value {
					assert.Greater(t, head.forward[level].value.insertedAt, head.value.insertedAt)
				}
				assert.GreaterOrEqual(t, headHeight, len(head.forward[level].forward))
			}

			levelSize++
			head = head.forward[level]
		}

		if level > 0 {
			assert.GreaterOrEqual(t, prevLevelSize, levelSize)
		}

		prevLevelSize = levelSize
	}
}

func insertedAtFn() func() int {
	insertedAt := 0
	return func() int {
		x := insertedAt
		insertedAt++
		return x
	}
}

func TestInsert(t *testing.T) {
	i := insertedAtFn()
	// exercise all code paths / edge cases
	// use debugger + fmt.Println(sl) to understand failing tests (if any)
	item0 := item{
		insertedAt: i(),
		value:      0,
	}
	item1 := item{
		insertedAt: i(),
		value:      1,
	}
	item2 := item{
		insertedAt: i(),
		value:      2,
	}

	sl := New(cmp()).(*stableSkipList[item])
	sl.randUint32 = fixedRand(1 << 0)
	sl.Insert(item0)
	checkInvariants(t, sl)
	assert.Len(t, sl.heads, 1)
	assert.Equal(t, item0, sl.heads[0].value)
	assert.Len(t, sl.heads[0].forward, 1)
	assert.Nil(t, sl.heads[0].forward[0])

	sl.randUint32 = fixedRand(1 << 0)
	sl.Insert(item1)
	checkInvariants(t, sl)
	assert.Len(t, sl.heads, 1)
	assert.Len(t, sl.heads[0].forward, 1)
	assert.Equal(t, item1, sl.heads[0].forward[0].value)
	assert.Len(t, sl.heads[0].forward[0].forward, 1)
	assert.Nil(t, sl.heads[0].forward[0].forward[0])

	sl.randUint32 = fixedRand(1 << 2)
	sl.Insert(item2)
	checkInvariants(t, sl)
	assert.Len(t, sl.heads, 2)
	assert.Len(t, sl.heads[1].forward, 2)
	assert.Equal(t, item2, sl.heads[1].value)
	assert.Len(t, sl.heads[1].forward, 2)
	assert.Nil(t, sl.heads[1].forward[1])

	item01stdupe := item{
		insertedAt: i(),
		value:      item0.value,
	}
	sl.randUint32 = fixedRand(1 << 3)
	sl.Insert(item01stdupe)
	checkInvariants(t, sl)
	assert.Len(t, sl.heads, 3)
	assert.Len(t, sl.heads[2].forward, 3)
	assert.Equal(t, item0, sl.heads[2].value)
	assert.Equal(t, item0, sl.heads[0].value)
	assert.Equal(t, item01stdupe, sl.heads[0].forward[0].value)

	item02nddupe := item{
		insertedAt: i(),
		value:      item0.value,
	}
	sl.randUint32 = fixedRand(1 << 1)
	sl.Insert(item02nddupe)
	checkInvariants(t, sl)
	assert.Equal(t, item0, sl.heads[0].value)
	assert.Equal(t, item01stdupe, sl.heads[0].forward[0].value)
	assert.Equal(t, item02nddupe, sl.heads[0].forward[0].forward[0].value)

	item03rddupe := item{
		insertedAt: i(),
		value:      item0.value,
	}
	sl.randUint32 = fixedRand(1 << 2)
	sl.Insert(item03rddupe)
	checkInvariants(t, sl)
	item04thdupe := item{
		insertedAt: i(),
		value:      item0.value,
	}
	sl.randUint32 = fixedRand(1 << 1)
	sl.Insert(item04thdupe)
	checkInvariants(t, sl)
	assert.Equal(t, item0, sl.heads[0].value)
	assert.Equal(t, item03rddupe, sl.heads[0].forward[0].forward[0].forward[0].value)
	assert.Equal(t, item04thdupe, sl.heads[0].forward[0].forward[0].forward[0].forward[0].value)

	smallestItem := item{
		insertedAt: i(),
		value:      -1,
	}
	sl.randUint32 = fixedRand(1 << 0)
	sl.Insert(smallestItem)
	checkInvariants(t, sl)
	assert.Equal(t, smallestItem, sl.heads[0].value)

	item2dupe := item{
		insertedAt: i(),
		value:      item2.value,
	}
	sl.randUint32 = fixedRand(1 << 2)
	sl.Insert(item2dupe)
	checkInvariants(t, sl)
	assert.Equal(t, item2dupe, sl.heads[1].forward[1].forward[1].forward[1].value)

	item1dupe := item{
		insertedAt: i(),
		value:      item1.value,
	}
	sl.randUint32 = fixedRand(1 << 2)
	sl.Insert(item1dupe)
	checkInvariants(t, sl)
	assert.Equal(t, item1, sl.heads[1].forward[1].forward[1].value)
	assert.Equal(t, item1, sl.heads[0].forward[0].forward[0].forward[0].forward[0].forward[0].forward[0].value)
	assert.Equal(t, item1dupe, sl.heads[0].forward[0].forward[0].forward[0].forward[0].forward[0].forward[0].forward[0].value)
}

func TestDeleteFirst(t *testing.T) {
	i := insertedAtFn()
	item0 := item{
		insertedAt: i(),
		value:      0,
	}
	item1 := item{
		insertedAt: i(),
		value:      1,
	}
	item2 := item{
		insertedAt: i(),
		value:      2,
	}
	sl := New(cmp()).(*stableSkipList[item])
	sl.randUint32 = fixedRand(1 << 0)
	sl.Insert(item0)
	checkInvariants(t, sl)
	sl.Insert(item1)
	checkInvariants(t, sl)
	sl.Insert(item2)
	checkInvariants(t, sl)

	sl.DeleteFirst(item1)
	checkInvariants(t, sl)
	_, found := sl.FindFirst(item1)
	assert.False(t, found)

	sl.DeleteFirst(item0)
	checkInvariants(t, sl)
	_, found = sl.FindFirst(item0)
	assert.False(t, found)

	sl.randUint32 = fixedRand(1 << 2)
	item21stdupe := item{
		insertedAt: i(),
		value:      item2.value,
	}
	sl.Insert(item21stdupe)
	checkInvariants(t, sl)
	sl.DeleteFirst(item2)
	checkInvariants(t, sl)
	_, found = sl.FindFirst(item21stdupe)
	assert.True(t, found)

	for j := 0; j < 9; j++ {
		sl.randUint32 = fixedRand(1 << (j % 3))
		sl.Insert(item{
			insertedAt: i(),
			value:      item2.value,
		})
		checkInvariants(t, sl)
	}

	sl.DeleteFirst(item2)
	checkInvariants(t, sl)

	sl.randUint32 = fixedRand(1 << 1)
	smallestItem := item{
		insertedAt: i(),
		value:      -1,
	}
	sl.Insert(smallestItem)
	checkInvariants(t, sl)
	sl.DeleteFirst(smallestItem)
	checkInvariants(t, sl)

	item3 := item{
		insertedAt: i(),
		value:      3,
	}
	sl.Insert(item3)
	checkInvariants(t, sl)
	sl.DeleteFirst(item3)
	checkInvariants(t, sl)

	for j := 0; j < 9; j++ {
		sl.randUint32 = fixedRand(1 << (j % 4))
		sl.Insert(item{
			insertedAt: i(),
			value:      item3.value,
		})
		checkInvariants(t, sl)
	}

	sl.DeleteFirst(item3)
	checkInvariants(t, sl)
}

func TestFuzz(t *testing.T) {
	// we test the StableSkipList implementation by implementing something that support a similar behavior (minus complexity)
	// using a slice that is sorted/stable, we mimic insertAfter/deletes by finding the right index in the slice via binary search
	// and modify the slice accordingly.
	//
	// after doing lots of random insertAfter/deletes, we do check the stable skip list invariants + check that the bottom
	// layer matches the order of the slice

	for seed := 0; seed < 128; seed++ {
		t.Run(fmt.Sprintf("Fuzzing, iteration #%d", seed), func(t *testing.T) {
			slice := make([]item, 0, 0)
			sl := New(cmp())
			rnd := rand.New(rand.NewSource(uint64(seed)))

			for i := 0; i < (1 << (seed % 16)); i++ {
				valueForThisIteration := i % (seed + 13)
				itemToHandle := item{
					insertedAt: i,
					value:      valueForThisIteration, // we want dupes!
				}

				if rnd.Intn(3) > 0 {
					sl.Insert(itemToHandle)
					pos := sort.Search(len(slice), func(idx int) bool {
						// we search +1 because we want to find where to insertAfter our item
						return slice[idx].value >= valueForThisIteration+1
					})

					if pos == len(slice) {
						slice = append(slice, itemToHandle)
					} else {
						slice = append(slice[:pos+1], slice[pos:]...)
						slice[pos] = itemToHandle
					}
				} else {
					sl.DeleteFirst(itemToHandle)
					pos := sort.Search(len(slice), func(idx int) bool {
						return slice[idx].value >= itemToHandle.value
					})

					if pos < len(slice) && slice[pos].value == itemToHandle.value {
						slice = append(slice[:pos], slice[pos+1:]...)
					}
				}
			}

			slImpl := sl.(*stableSkipList[item])
			checkInvariants(t, slImpl)
			if len(slice) == 0 {
				assert.Empty(t, slImpl.heads)
			} else {
				head := slImpl.heads[0]
				for i := 0; i < len(slice); i++ {
					assert.NotNil(t, head)
					assert.Equal(t, slice[i], head.value)
					head = head.forward[0]
				}
			}
		})
	}
}
