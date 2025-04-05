package hub

import "github.com/lomik/hub/pkg/kv"

// Any is a special value that matches any other value in topic matching
const Any string = "*"

// Topic represents a named channel for event distribution with key-value attributes.
// It's immutable after creation and safe for concurrent use.
type Topic struct {
	mp kv.Map
}

// NewTopic creates a new Topic from key-value pairs.
// Pairs can be provided in two formats:
//   - "key=value" strings
//   - Separate "key", "value" arguments
//
// Returns error if input format is invalid.
//
// Example:
//
//	t, err := NewTopic("type=alert", "severity=high")
func NewTopic(args ...string) (*Topic, error) {
	mp, err := kv.Parse(args...)
	if err != nil {
		return nil, err
	}
	return &Topic{mp: mp}, nil
}

// T creates a new Topic from key-value pairs, panicking on error.
// Simplified version of NewTopic for use in tests and initialization.
//
// Example:
//
//	t := T("type=alert", "severity=high")
func T(args ...string) *Topic {
	mp, err := kv.Parse(args...)
	if err != nil {
		panic(err)
	}
	return &Topic{mp: mp}
}

// With creates a new Topic by merging current attributes with new ones.
// New attributes override existing ones with the same keys.
// Panics if new attributes have invalid format.
//
// Example:
//
//	t1 := T("type=alert", "severity=high")
//	t2 := t1.With("severity=low", "source=server")
//	// t2 now has: type=alert, severity=low, source=server
func (t *Topic) With(args ...string) *Topic {
	other, err := kv.Parse(args...)
	if err != nil {
		panic(err)
	}
	return &Topic{mp: t.mp.Merge(other)}
}

// Get returns the value for the specified key.
// Returns empty string if key doesn't exist.
//
// Example:
//
//	t := T("type=alert")
//	v := t.Get("type") // returns "alert"
func (t *Topic) Get(k string) string {
	return t.mp.Get(k)
}

// Each iterates over all key-value pairs in the Topic.
// Pairs are processed in sorted key order.
//
// Example:
//
//	t := T("b=2", "a=1")
//	t.Each(func(k, v string) {
//	    fmt.Printf("%s=%s\n", k, v)
//	})
//	// Output:
//	// a=1
//	// b=2
func (t *Topic) Each(cb func(k, v string)) {
	t.mp.Each(cb)
}

// Match checks if this Topic matches another Topic.
// A Topic matches if:
//   - All keys in this Topic exist in the other Topic
//   - Corresponding values are equal or one of them is Any ("*")
//
// Does not consider additional keys in the other Topic.
//
// Example:
//
//	t1 := T("type=alert", "severity=high")
//	t2 := T("type=alert", "severity=*", "source=server")
//	t1.Match(t2) // returns true
func (t *Topic) Match(other *Topic) bool {
	return t.mp.Match(other.mp)
}

// Len returns the number of key-value pairs
func (t *Topic) Len() int {
	return t.mp.Len()
}
