package hub

import (
	"strings"
	"testing"
)

func TestNewTopic(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "valid topic",
			args:    []string{"type=alert", "priority=high"},
			want:    "priority=high type=alert",
			wantErr: false,
		},
		{
			name:    "invalid format",
			args:    []string{"type=alert", "priority"},
			wantErr: true,
		},
		{
			name:    "empty topic",
			args:    []string{},
			want:    "",
			wantErr: false,
		},
		{
			name:    "with wildcard",
			args:    []string{"type=*", "priority=high"},
			want:    "priority=high type=*",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTopic(tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTopic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.want {
				t.Errorf("NewTopic() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestT(t *testing.T) {
	t.Run("valid topic", func(t *testing.T) {
		got := T("type=alert", "priority=high")
		want := "priority=high type=alert"
		if got.String() != want {
			t.Errorf("T() = %v, want %v", got.String(), want)
		}
	})

	t.Run("panics on invalid input", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected T() to panic on invalid input")
			}
		}()
		T("type=alert", "priority")
	})
}

func TestTopic_With(t *testing.T) {
	tests := []struct {
		name   string
		base   []string
		args   []string
		expect string
	}{
		{
			name:   "add new attributes",
			base:   []string{"type=alert"},
			args:   []string{"priority=high"},
			expect: "priority=high type=alert",
		},
		{
			name:   "override existing",
			base:   []string{"type=alert", "priority=low"},
			args:   []string{"priority=high"},
			expect: "priority=high type=alert",
		},
		{
			name:   "mixed new and override",
			base:   []string{"type=alert", "priority=low"},
			args:   []string{"priority=high", "source=server"},
			expect: "priority=high source=server type=alert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := T(tt.base...)
			got := base.With(tt.args...)
			if got.String() != tt.expect {
				t.Errorf("With() = %v, want %v", got.String(), tt.expect)
			}
			// Verify original wasn't modified
			if base.String() != T(tt.base...).String() {
				t.Error("Original topic was modified")
			}
		})
	}

	t.Run("panics on invalid input", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected With() to panic on invalid input")
			}
		}()
		T("type=alert").With("priority")
	})
}

func TestTopic_Get(t *testing.T) {
	topic := T("type=alert", "priority=high")

	tests := []struct {
		key  string
		want string
	}{
		{"type", "alert"},
		{"priority", "high"},
		{"missing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := topic.Get(tt.key); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTopic_Each(t *testing.T) {
	topic := T("b=2", "a=1")

	var got []string
	topic.Each(func(k, v string) {
		got = append(got, k+"="+v)
	})

	want := []string{"a=1", "b=2"}
	if len(got) != len(want) {
		t.Fatalf("Each() processed %d items, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Each() item %d = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestTopic_Match(t *testing.T) {
	tests := []struct {
		name   string
		a      string
		b      string
		expect bool
	}{
		{
			name:   "exact match",
			a:      "type=alert priority=high",
			b:      "type=alert priority=high",
			expect: true,
		},
		{
			name:   "wildcard in pattern",
			a:      "type=* priority=high",
			b:      "type=alert priority=high",
			expect: true,
		},
		{
			name:   "wildcard in target",
			a:      "type=alert priority=high",
			b:      "type=* priority=high",
			expect: true,
		},
		{
			name:   "partial match",
			a:      "type=alert",
			b:      "type=alert priority=high",
			expect: true,
		},
		{
			name:   "mismatch",
			a:      "type=alert priority=high",
			b:      "type=alert priority=low",
			expect: false,
		},
		{
			name:   "empty pattern",
			a:      "",
			b:      "type=alert",
			expect: true,
		},
		{
			name:   "empty target",
			a:      "type=alert",
			b:      "",
			expect: false,
		},
		{
			name:   "multiple wildcards",
			a:      "type=* priority=*",
			b:      "type=alert priority=high",
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aa := strings.Split(tt.a, " ")
			bb := strings.Split(tt.b, " ")
			if tt.a == "" {
				aa = []string{}
			}
			if tt.b == "" {
				bb = []string{}
			}
			a := T(aa...)
			b := T(bb...)

			if got := a.Match(b); got != tt.expect {
				t.Errorf("Match() = %v, want %v\na: %v\nb: %v", got, tt.expect, a, b)
			}
		})
	}
}

// Helper method for string representation
func (t *Topic) String() string {
	var s []string
	t.Each(func(k, v string) {
		s = append(s, k+"="+v)
	})
	return strings.Join(s, " ")
}
