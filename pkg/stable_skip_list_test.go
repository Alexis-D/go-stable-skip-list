package pkg

import (
	"github.com/stretchr/testify/assert"
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

func TestStableSkipList(t *testing.T) {
	testSlice := []int{12, 13, 17, 12, 10, 9, 8, 11, 18, 19, 21, 23, 22, 1}

	for _, test := range []struct {
		name        string
		insertItems []int
		fn          func(t *testing.T, sl StableSkipList[item])
	}{
		{
			name: "empty skip list behave as expected",
			fn: func(t *testing.T, sl StableSkipList[item]) {
				_, found := sl.First()
				assert.False(t, found)
				_, found = sl.Last()
				assert.False(t, found)
			},
		}, {
			name:        "first and last return the correct values",
			insertItems: testSlice,
			fn: func(t *testing.T, sl StableSkipList[item]) {
				first, found := sl.First()
				assert.True(t, found)
				assert.Equal(t, 1, first.value)
				last, found := sl.Last()
				assert.True(t, found)
				assert.Equal(t, 23, last.value)
			},
		}, {
			name:        "FindFirst works, even with duplicates",
			insertItems: testSlice,
			fn: func(t *testing.T, sl StableSkipList[item]) {
				value, found := sl.FindFirst(item{value: 12})
				assert.True(t, found)
				assert.Equal(t, 0, value.insertedAt)
				assert.Equal(t, 12, value.value)

				value, found = sl.FindFirst(item{value: 19})
				assert.True(t, found)
				assert.Equal(t, 19, value.value)

				_, found = sl.FindFirst(item{value: -1})
				assert.False(t, found)
			},
		}, {
			name:        "DeleteFirst works, even with duplicates",
			insertItems: testSlice,
			fn: func(t *testing.T, sl StableSkipList[item]) {
				sl.DeleteFirst(item{value: 12})
				value, found := sl.FindFirst(item{value: 12})
				assert.True(t, found)
				assert.Equal(t, 3, value.insertedAt)
				assert.Equal(t, 12, value.value)

				sl.DeleteFirst(item{value: 19})
				_, found = sl.FindFirst(item{value: 19})
				assert.False(t, found)
				sl.DeleteFirst(item{value: 21})
				_, found = sl.FindFirst(item{value: 21})
				assert.False(t, found)
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			sl := New(cmp())
			if test.insertItems != nil {
				for i, x := range test.insertItems {
					sl.Insert(item{
						insertedAt: i,
						value:      x,
					})
				}
			}
			test.fn(t, sl)
		})
	}
}
