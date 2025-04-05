package hub

import (
	"testing"
)

func Test_sublist_add(t *testing.T) {
	sl := &sublist{}

	t.Run("add to empty list", func(t *testing.T) {
		sl.add(&sub{id: 2})
		if len(sl.lst) != 1 || sl.lst[0].id != 2 {
			t.Errorf("add() failed, got %v, want [2]", sl.lst)
		}
	})

	t.Run("add in order", func(t *testing.T) {
		sl.add(&sub{id: 1})
		sl.add(&sub{id: 3})
		if len(sl.lst) != 3 || sl.lst[0].id != 1 || sl.lst[1].id != 2 || sl.lst[2].id != 3 {
			t.Errorf("add() ordering failed, got %v, want [1 2 3]", sl.lst)
		}
	})

	t.Run("add duplicate", func(t *testing.T) {
		prevLen := len(sl.lst)
		sl.add(&sub{id: 2}) // Дубликат
		if len(sl.lst) != prevLen+1 {
			t.Error("add() should allow duplicates")
		}
	})
}

func Test_sublist_remove(t *testing.T) {
	sl := &sublist{}

	// Setup initial state
	sl.add(&sub{id: 1})
	sl.add(&sub{id: 2})
	sl.add(&sub{id: 3})

	t.Run("remove middle", func(t *testing.T) {
		sl.remove(2)
		if len(sl.lst) != 2 || sl.lst[0].id != 1 || sl.lst[1].id != 3 {
			t.Errorf("remove() failed, got %v, want [1 3]", sl.lst)
		}
	})

	t.Run("remove first", func(t *testing.T) {
		sl.remove(1)
		if len(sl.lst) != 1 || sl.lst[0].id != 3 {
			t.Errorf("remove() failed, got %v, want [3]", sl.lst)
		}
	})

	t.Run("remove last", func(t *testing.T) {
		sl.remove(3)
		if len(sl.lst) != 0 {
			t.Errorf("remove() failed, got %v, want empty", sl.lst)
		}
	})

	t.Run("remove non-existent", func(t *testing.T) {
		sl.remove(99) // Не должен паниковать
	})
}

func Test_sublist_find(t *testing.T) {
	sl := &sublist{}

	// Setup initial state
	sl.add(&sub{id: 1})
	sl.add(&sub{id: 3})
	sl.add(&sub{id: 5})

	tests := []struct {
		name string
		id   SubID
		want int
	}{
		{"find first", 1, 0},
		{"find middle", 3, 1},
		{"find last", 5, 2},
		{"not found", 2, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sl.find(tt.id); got != tt.want {
				t.Errorf("find() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sublist_maintains_order(t *testing.T) {
	sl := &sublist{}

	// Добавляем в разном порядке
	sl.add(&sub{id: 3})
	sl.add(&sub{id: 1})
	sl.add(&sub{id: 2})

	// Проверяем порядок
	for i, id := range []SubID{1, 2, 3} {
		if sl.lst[i].id != id {
			t.Errorf("Order violation at index %d, got %d, want %d", i, sl.lst[i].id, id)
		}
	}

	// Удаляем и проверяем порядок
	sl.remove(2)
	if sl.lst[0].id != 1 || sl.lst[1].id != 3 {
		t.Errorf("Order after remove failed, got %v, want [1 3]", sl.lst)
	}
}

func Test_mergeSubLists(t *testing.T) {
	// Helper to create test sublists
	makeSubList := func(ids ...SubID) *sublist {
		sl := &sublist{}
		for _, id := range ids {
			sl.lst = append(sl.lst, &sub{id: id})
		}
		return sl
	}

	tests := []struct {
		name  string
		lists []*sublist
		want  []SubID
	}{
		{
			name:  "empty lists",
			lists: []*sublist{},
			want:  []SubID{},
		},
		{
			name:  "single list",
			lists: []*sublist{makeSubList(1, 2, 3)},
			want:  []SubID{1, 2, 3},
		},
		{
			name:  "multiple lists no duplicates",
			lists: []*sublist{makeSubList(1, 3), makeSubList(2, 4)},
			want:  []SubID{1, 2, 3, 4},
		},
		{
			name:  "with duplicates",
			lists: []*sublist{makeSubList(1, 2, 2, 3), makeSubList(2, 4)},
			want:  []SubID{1, 2, 3, 4},
		},
		{
			name:  "all duplicates",
			lists: []*sublist{makeSubList(1, 1), makeSubList(1, 1)},
			want:  []SubID{1},
		},
		{
			name:  "one empty list",
			lists: []*sublist{makeSubList(), makeSubList(1, 2)},
			want:  []SubID{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make([]SubID, 0)
			for s := range mergeSubLists(tt.lists...) {
				got = append(got, s.id)
			}

			if len(got) != len(tt.want) {
				t.Errorf("mergeSubLists() length = %v, want %v", got, tt.want)
				return
			}

			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("mergeSubLists() at index %d = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func Test_mergeSubLists_early_stop(t *testing.T) {
	sl1 := &sublist{
		lst: []*sub{{id: 1}, {id: 2}, {id: 3}},
	}
	sl2 := &sublist{
		lst: []*sub{{id: 4}, {id: 5}},
	}

	count := 0
	for range mergeSubLists(sl1, sl2) {
		count++
		if count == 2 {
			break
		}
	}

	if count != 2 {
		t.Errorf("Expected to process 2 elements, got %d", count)
	}
}

func Test_sublist_len(t *testing.T) {
	t.Run("nil sublist", func(t *testing.T) {
		var sl *sublist
		if got := sl.len(); got != 0 {
			t.Errorf("len() = %v, want 0 for nil sublist", got)
		}
	})

	t.Run("empty sublist", func(t *testing.T) {
		sl := &sublist{}
		if got := sl.len(); got != 0 {
			t.Errorf("len() = %v, want 0 for empty sublist", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		sl := &sublist{}
		sl.add(&sub{id: 1})
		if got := sl.len(); got != 1 {
			t.Errorf("len() = %v, want 1 for single element", got)
		}
	})

	t.Run("multiple elements", func(t *testing.T) {
		sl := &sublist{}
		sl.add(&sub{id: 1})
		sl.add(&sub{id: 2})
		if got := sl.len(); got != 2 {
			t.Errorf("len() = %v, want 2 for two elements", got)
		}
	})

	t.Run("after removal", func(t *testing.T) {
		sl := &sublist{}
		sl.add(&sub{id: 1})
		sl.add(&sub{id: 2})
		sl.remove(1)
		if got := sl.len(); got != 1 {
			t.Errorf("len() = %v, want 1 after removal", got)
		}
	})
}
