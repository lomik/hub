package kv

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    Map
		wantErr bool
	}{
		{
			name:  "empty",
			input: []string{""},
			want:  Map{data: []KV{}},
		},
		{
			name:  "simple key=value pairs",
			input: []string{"a=1", "b=2"},
			want: Map{data: []KV{
				{key: "a", value: "1"},
				{key: "b", value: "2"},
			}},
		},
		{
			name:  "separate key value pairs",
			input: []string{"a", "1", "b", "2"},
			want: Map{data: []KV{
				{key: "a", value: "1"},
				{key: "b", value: "2"},
			}},
		},
		{
			name:  "mixed format",
			input: []string{"a=1", "b", "2", "c=3"},
			want: Map{data: []KV{
				{key: "a", value: "1"},
				{key: "b", value: "2"},
				{key: "c", value: "3"},
			}},
		},
		{
			name:  "escaped equals",
			input: []string{`a\=b=c`, `d=e\=f`},
			want: Map{data: []KV{
				{key: "a=b", value: "c"},
				{key: "d", value: "e=f"},
			}},
		},
		{
			name:    "missing value",
			input:   []string{"a"},
			wantErr: true,
		},
		{
			name:  "empty input",
			input: []string{},
			want:  Map{data: []KV{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareMaps(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareMaps(a, b Map) bool {
	if a.Len() != b.Len() {
		return false
	}
	for i := 0; i < a.Len(); i++ {
		if a.data[i].key != b.data[i].key || a.data[i].value != b.data[i].value {
			return false
		}
	}
	return true
}

func TestGet(t *testing.T) {
	m := Map{data: []KV{
		{key: "a", value: "1"},
		{key: "b", value: "2"},
	}}

	tests := []struct {
		key  string
		want string
	}{
		{"a", "1"},
		{"b", "2"},
		{"c", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := m.Get(tt.key); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeys(t *testing.T) {
	m := Map{data: []KV{
		{key: "b", value: "2"},
		{key: "a", value: "1"},
		{key: "c", value: "3"},
	}}
	m.sortKeys()

	want := []string{"a", "b", "c"}
	got := m.Keys()

	if len(got) != len(want) {
		t.Fatalf("Keys() length = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Keys()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestLen(t *testing.T) {
	m := Map{data: []KV{
		{key: "a", value: "1"},
		{key: "b", value: "2"},
	}}

	if got := m.Len(); got != 2 {
		t.Errorf("Len() = %d, want 2", got)
	}
}

func TestEach(t *testing.T) {
	m := Map{data: []KV{
		{key: "a", value: "1"},
		{key: "b", value: "2"},
	}}

	var gotKeys []string
	var gotValues []string
	m.Each(func(k, v string) {
		gotKeys = append(gotKeys, k)
		gotValues = append(gotValues, v)
	})

	wantKeys := []string{"a", "b"}
	wantValues := []string{"1", "2"}

	for i := range wantKeys {
		if gotKeys[i] != wantKeys[i] {
			t.Errorf("Each() key = %v, want %v", gotKeys[i], wantKeys[i])
		}
		if gotValues[i] != wantValues[i] {
			t.Errorf("Each() value = %v, want %v", gotValues[i], wantValues[i])
		}
	}
}

func TestToMap(t *testing.T) {
	m := Map{data: []KV{
		{key: "a", value: "1"},
		{key: "b", value: "2"},
	}}

	got := m.ToMap()
	want := map[string]string{
		"a": "1",
		"b": "2",
	}

	if len(got) != len(want) {
		t.Fatalf("ToMap() length = %d, want %d", len(got), len(want))
	}

	for k, v := range want {
		if got[k] != v {
			t.Errorf("ToMap()[%s] = %s, want %s", k, got[k], v)
		}
	}
}

func TestKVMethods(t *testing.T) {
	kv := KV{key: "test", value: "123"}

	if kv.Key() != "test" {
		t.Errorf("Key() = %v, want test", kv.Key())
	}

	if kv.Value() != "123" {
		t.Errorf("Value() = %v, want 123", kv.Value())
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		name   string
		a      string
		b      string
		expect bool
	}{
		{
			name:   "exact match",
			a:      "color=red size=large",
			b:      "color=red size=large weight=heavy",
			expect: true,
		},
		{
			name:   "wildcard in a",
			a:      "color=* size=large",
			b:      "color=blue size=large",
			expect: true,
		},
		{
			name:   "wildcard in b",
			a:      "color=red size=large",
			b:      "color=* size=large",
			expect: true,
		},
		{
			name:   "missing key in b",
			a:      "color=red size=large",
			b:      "color=red",
			expect: false,
		},
		{
			name:   "value mismatch",
			a:      "color=red size=large",
			b:      "color=red size=small",
			expect: false,
		},
		{
			name:   "empty a",
			a:      "",
			b:      "any=value",
			expect: true,
		},
		{
			name:   "different order",
			a:      "color=red size=large",
			b:      "size=large color=red",
			expect: true,
		},
		{
			name:   "partial with wildcard",
			a:      "color=*",
			b:      "color=blue size=large",
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := parseSpaceSeparated(tt.a)
			if err != nil {
				t.Fatalf("Failed to parse a: %v", err)
			}

			b, err := parseSpaceSeparated(tt.b)
			if err != nil {
				t.Fatalf("Failed to parse b: %v", err)
			}

			result := a.Match(b)
			if result != tt.expect {
				t.Errorf("Match() = %v, want %v\nA: %v\nB: %v", result, tt.expect, tt.a, tt.b)
			}
		})
	}
}

// parseSpaceSeparated converts space-separated key=value pairs to Map
func parseSpaceSeparated(s string) (Map, error) {
	if s == "" {
		return Map{}, nil
	}
	return Parse(strings.Split(s, " ")...)
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name   string
		a      string
		b      string
		expect string
	}{
		{
			name:   "disjoint keys",
			a:      "a=1 b=2",
			b:      "c=3 d=4",
			expect: "a=1 b=2 c=3 d=4",
		},
		{
			name:   "overlapping keys",
			a:      "a=1 b=2",
			b:      "b=3 c=4",
			expect: "a=1 b=3 c=4",
		},
		{
			name:   "empty first map",
			a:      "",
			b:      "a=1 b=2",
			expect: "a=1 b=2",
		},
		{
			name:   "empty second map",
			a:      "a=1 b=2",
			b:      "",
			expect: "a=1 b=2",
		},
		{
			name:   "multiple overrides",
			a:      "a=1 b=2 c=3",
			b:      "b=4 c=5 d=6",
			expect: "a=1 b=4 c=5 d=6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := mustParse(t, tt.a)
			b := mustParse(t, tt.b)
			expect := mustParse(t, tt.expect)

			result := a.Merge(b)

			if !reflect.DeepEqual(result, expect) {
				t.Errorf("Merge() mismatch:\nGot:    %v\nExpect: %v",
					formatMap(result), formatMap(expect))
			}
		})
	}
}

// mustParse is a helper that parses space-separated key-value pairs or fails the test
func mustParse(t *testing.T, s string) Map {
	if s == "" {
		return Map{}
	}
	m, err := Parse(strings.Split(s, " ")...)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return m
}

// formatMap returns string representation of Map for error messages
func formatMap(m Map) string {
	var sb strings.Builder
	for i, kv := range m.data {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(kv.key)
		sb.WriteString("=")
		sb.WriteString(kv.value)
	}
	return sb.String()
}
