# `go-stable-skip-list`

This is a toy[^1] [skip list (wikipedia)](https://en.wikipedia.org/wiki/Skip_list) implementation. The main difference
compared to various skip lists implementations out there is that it is 'stable', as in, it not only allows for
duplicates but it also retains them in the order they were inserted. As such `Insert(value T)` does always perform an
insert and never updates a value.

Quick example of how this is used:

```go
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

func main() {
	sl := New(cmp())
	for i, x := range []int{12, 13, 17, 12, 10, 9, 8, 11, 18, 19, 21, 23, 22, 1} {
		sl.Insert(item{
			insertedAt: i,
			value:      x,
		})
	}

	sl.FindFirst(item{value: 12})   // finds the first 12
	sl.DeleteFirst(item{value: 12}) // delete the first 12
	sl.FindFirst(item{value: 12})   // finds the second 12

	fmt.Println(sl)
}
```

It will output the list:

```
(3): -> {insertedAt:9 value:19} -> {insertedAt:10 value:21} -> nil
(2): -> {insertedAt:6 value:8} -> {insertedAt:5 value:9} -> {insertedAt:8 value:18} -> {insertedAt:9 value:19} -> {insertedAt:10 value:21} -> {insertedAt:11 value:23} -> nil
(1): -> {insertedAt:6 value:8} -> {insertedAt:5 value:9} -> {insertedAt:7 value:11} -> {insertedAt:2 value:17} -> {insertedAt:8 value:18} -> {insertedAt:9 value:19} -> {insertedAt:10 value:21} -> {insertedAt:11 value:23} -> nil
(0): -> {insertedAt:13 value:1} -> {insertedAt:6 value:8} -> {insertedAt:5 value:9} -> {insertedAt:4 value:10} -> {insertedAt:7 value:11} -> {insertedAt:3 value:12} -> {insertedAt:1 value:13} -> {insertedAt:2 value:17} -> {insertedAt:8 value:18} -> {insertedAt:9 value:19} -> {insertedAt:10 value:21} -> {insertedAt:12 value:22} -> {insertedAt:11 value:23} -> nil
```

How does the algo differ from a textbook skip list? It ensures it builds a 'wall' when duplicates are encountered, such
that the first inserted 'dupe' is always the first one to be hit, probably easier to visualize (nodes being represented
as value (insertion timestamp)):

```
-------------------------------------> 5 (4) -> nil
-> 1 (0) ----------> 3 (2) ----------> 5 (4) -> nil
-> 1 (0) -> 2 (1) -> 3 (2) -> 3 (3) -> 5 (4) -> nil
```

Now inserting `3 (5)` results in (note how we grew `3 (2)` instead of `3 (5)`):

```
-------------------> 3 (2) -------------------> 5 (4) -> nil
-> 1 (0) ----------> 3 (2) ----------> 3 (5) -> 5 (4) -> nil
-> 1 (0) -> 2 (1) -> 3 (2) -> 3 (3) -> 3 (5) -> 5 (4) -> nil
```

Deletion works the same way, e.g. if we remove `3`, we now get:

```
-------------------------------------> 5 (4) -> nil
-> 1 (0) ----------> 3 (3) -> 3 (5) -> 5 (4) -> nil
-> 1 (0) -> 2 (1) -> 3 (3) -> 3 (5) -> 5 (4) -> nil
```


[^1]: it's not threadsafe, it has very little tests, it's not very configurable, etc
