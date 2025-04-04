package kv

import (
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
