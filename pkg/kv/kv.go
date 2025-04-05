package kv

import (
	"sort"
	"strings"
)

// KV represents a key-value pair with private fields
type KV struct {
	key   string
	value string
}

// Key returns the key of the key-value pair
func (kv KV) Key() string {
	return kv.key
}

// Value returns the value of the key-value pair
func (kv KV) Value() string {
	return kv.value
}

// Map stores collection of key-value pairs
type Map struct {
	data []KV
}

// Parse processes key-value pairs from input strings and returns Map or error
// Supports two formats:
//  1. "key=value" (single string with separator)
//  2. "key", "value" (two separate strings)
//
// Handles escaped '=' characters (like "\=" in keys/values)
// Returns error if input format is invalid
func Parse(d ...string) (Map, error) {
	var ret Map
	for i := 0; i < len(d); {
		// Find first unescaped '=' position
		p := findUnescapedEquals(d[i])
		if p < 0 {
			// Format: "key", "value" (separate strings)
			if i+1 >= len(d) {
				return Map{}, &ParseError{
					Msg:  "missing value for key",
					Key:  d[i],
					Pos:  i,
					Args: d,
				}
			}
			ret.data = append(ret.data, KV{
				key:   d[i],
				value: d[i+1],
			})
			i += 2
			continue
		}

		ret.data = append(ret.data, KV{
			key:   unescape(d[i][:p]),
			value: unescape(d[i][p+1:]),
		})
		i += 1
	}

	ret.sortKeys()
	return ret, nil
}

// ParseError represents parsing error details
type ParseError struct {
	Msg  string
	Key  string
	Pos  int
	Args []string
}

func (e *ParseError) Error() string {
	return e.Msg + " '" + e.Key + "' at position " + string(rune(e.Pos))
}

// Get returns value by key (empty string if not found)
func (m Map) Get(key string) string {
	for _, kv := range m.data {
		if kv.key == key {
			return kv.value
		}
	}
	return ""
}

// Keys returns all keys in sorted order
func (m Map) Keys() []string {
	keys := make([]string, len(m.data))
	for i, kv := range m.data {
		keys[i] = kv.key
	}
	return keys
}

// Len returns the number of key-value pairs
func (m Map) Len() int {
	return len(m.data)
}

// Each iterates over all key-value pairs in sorted order
func (m Map) Each(fn func(key, value string)) {
	for _, kv := range m.data {
		fn(kv.key, kv.value)
	}
}

// ToMap converts to standard map[string]string
func (m Map) ToMap() map[string]string {
	result := make(map[string]string, len(m.data))
	for _, kv := range m.data {
		result[kv.key] = kv.value
	}
	return result
}

// Match returns true if for all keys in current map:
// - the key exists in other map
// - values are equal OR one of the values is "*"
// Uses the fact that both maps are sorted for O(n+m) comparison
func (m Map) Match(other Map) bool {
	i, j := 0, 0
	lenA, lenB := len(m.data), len(other.data)

	for i < lenA && j < lenB {
		aKey, bKey := m.data[i].key, other.data[j].key

		switch {
		case aKey < bKey:
			// Key exists in A but not in B
			return false
		case aKey > bKey:
			// Key exists in B but not in A - skip
			j++
		default:
			// Keys match - compare values
			aVal, bVal := m.data[i].value, other.data[j].value
			if aVal != "*" && bVal != "*" && aVal != bVal {
				return false
			}
			i++
			j++
		}
	}

	// Check if we processed all keys from A
	return i == lenA
}

// Merge creates new Map with keys from both maps
// Keys from the argument map override keys from the original map
func (m Map) Merge(other Map) Map {
	result := Map{
		data: make([]KV, 0, len(m.data)+len(other.data)),
	}

	i, j := 0, 0
	for i < len(m.data) && j < len(other.data) {
		switch {
		case m.data[i].key < other.data[j].key:
			result.data = append(result.data, m.data[i])
			i++
		case m.data[i].key > other.data[j].key:
			result.data = append(result.data, other.data[j])
			j++
		default:
			// Key exists in both - take value from other map
			result.data = append(result.data, other.data[j])
			i++
			j++
		}
	}

	// Append remaining elements
	result.data = append(result.data, m.data[i:]...)
	result.data = append(result.data, other.data[j:]...)

	return result
}

// sortKeys sorts the key-value pairs by key
func (m *Map) sortKeys() {
	sort.Slice(m.data, func(i, j int) bool {
		return m.data[i].key < m.data[j].key
	})
}

// findUnescapedEquals locates the first '=' not preceded by backslash
func findUnescapedEquals(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			i++ // Skip escaped character
			continue
		}
		if s[i] == '=' {
			return i
		}
	}
	return -1
}

// unescape removes backslash from escaped characters
func unescape(s string) string {
	var buf strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			buf.WriteByte(s[i+1]) // Write escaped char
			i++
		} else {
			buf.WriteByte(s[i])
		}
	}
	return buf.String()
}
