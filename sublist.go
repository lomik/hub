package hub

import (
	"iter"
	"sort"
)

type sublist struct {
	lst []*sub
}

// add inserts new subscription while maintaining sort order by SubID
func (sl *sublist) add(s *sub) {
	// Find insertion point using binary search
	idx := sort.Search(len(sl.lst), func(i int) bool {
		return sl.lst[i].id >= s.id
	})

	// Grow slice
	sl.lst = append(sl.lst, nil)

	// Shift elements if needed
	if idx < len(sl.lst)-1 {
		copy(sl.lst[idx+1:], sl.lst[idx:])
	}

	// Insert new sub
	sl.lst[idx] = s
}

// remove deletes subscription by ID while maintaining order
func (sl *sublist) remove(id SubID) {
	// Find index using binary search
	idx := sort.Search(len(sl.lst), func(i int) bool {
		return sl.lst[i].id >= id
	})

	// Check if found
	if idx < len(sl.lst) && sl.lst[idx].id == id {
		// Remove by shifting
		copy(sl.lst[idx:], sl.lst[idx+1:])
		sl.lst = sl.lst[:len(sl.lst)-1]
	}
}

// find returns subscription index or -1 if not found
func (sl *sublist) find(id SubID) int {
	idx := sort.Search(len(sl.lst), func(i int) bool {
		return sl.lst[i].id >= id
	})

	if idx < len(sl.lst) && sl.lst[idx].id == id {
		return idx
	}
	return -1
}

// len returns count of subscriptions
func (sl *sublist) len() int {
	if sl == nil {
		return 0
	}
	return len(sl.lst)
}

// mergeSubLists returns an iterator over all subscriptions from given sublists.
// Lists must be sorted by SubID. Duplicates are automatically skipped.
func mergeSubLists(lists ...*sublist) iter.Seq[*sub] {
	return func(yield func(*sub) bool) {
		// Create pointers array for each list
		indices := make([]int, len(lists))
		prevID := SubID(0) // Tracks last emitted ID

		for {
			// Find the smallest ID among all lists
			var smallest *sub
			smallestListIdx := -1

			for i, lst := range lists {
				if indices[i] >= len(lst.lst) {
					continue // This list is exhausted
				}

				current := lst.lst[indices[i]]
				if smallest == nil || current.id < smallest.id {
					smallest = current
					smallestListIdx = i
				}
			}

			if smallest == nil {
				return // All lists exhausted
			}

			// Skip duplicates
			if smallest.id != prevID {
				prevID = smallest.id
				if !yield(smallest) {
					return // Caller requested stop
				}
			}

			// Advance the pointer for the list we took the element from
			indices[smallestListIdx]++
		}
	}
}
