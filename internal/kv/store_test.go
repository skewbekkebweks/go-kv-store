package kv

import (
	"math"
	"slices"
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	tests := []struct {
		name  string
		input string
		ret   bool
	}{
		{name: "Add new 1", input: "1", ret: true},
		{name: "Add existing 1", input: "1", ret: false},
	}

	s := NewStore()

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			ret, err := s.Set(test.input, "", 0)
			if err != nil {
				t.Errorf("err != nil")
			}
			if ret != test.ret {
				t.Errorf("%v != %v", ret, test.ret)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
		ret    bool
	}{
		{name: "Get existing 1", input: "1", output: "123", ret: true},
		{name: "Get not-existent 2", input: "2", output: "", ret: false},
	}

	s := NewStore()
	s.Set("1", "123", 0)

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output, ret := s.Get(test.input)
			if output != test.output || ret != test.ret {
				t.Errorf("(%v, %v) != (%v, %v)", output, ret, test.output, test.ret)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name  string
		input string
		ret   bool
	}{
		{name: "Del existing 1", input: "1", ret: true},
		{name: "Del not-existent 2", input: "2", ret: false},
	}

	s := NewStore()
	s.Set("1", "123", 0)

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			ret := s.Delete(test.input)
			if ret != test.ret {
				t.Errorf("%v != %v", ret, test.ret)
			}
			_, ok := s.Get(test.input)
			if ok {
				t.Errorf("%v is existing after Delete", test.input)
			}
		})
	}
}

func TestKeys(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		limit  int
		output []string
	}{
		{name: "Get all sorted", prefix: "", limit: math.MaxInt, output: []string{"aaa", "abcaa", "abcbb", "abcde", "abdef"}},
		{name: "Get 0", prefix: "", limit: 0, output: []string{}},
		{name: "Get with prefix abc", prefix: "abc", limit: 2, output: []string{"abcaa", "abcbb"}},
	}

	s := NewStore()
	s.Set("aaa", "", 0)
	s.Set("abcde", "", 0)
	s.Set("abcaa", "", 0)
	s.Set("abcbb", "", 0)
	s.Set("abdef", "", 0)

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			output := s.Keys(test.prefix, test.limit)
			if !slices.Equal(test.output, output) {
				t.Errorf("expected: %v, received: %v", test.output, output)
			}
		})
	}
}

func TestTTL(t *testing.T) {
	i := 0
	now := func() time.Time {
		return time.Time{}.Add(time.Duration(i) * time.Second)
	}
	s := newStore(now)

	s.Set("a", "123", time.Second)
	_, ok := s.Get("a")
	if !ok {
		t.Error("expected: exist, received: !exist")
	}
	i += 1
	_, ok = s.Get("a")
	if ok {
		t.Error("expected: !exist, received: exist")
	}
}

func TestExpire(t *testing.T) {
	i := 0
	now := func() time.Time {
		return time.Time{}.Add(time.Duration(i) * time.Second)
	}
	s := newStore(now)

	s.Set("a", "123", 2*time.Second)
	_, ok := s.Get("a")
	if !ok {
		t.Error("expected: exist, received: !exist")
	}
	i += 1
	updated, err := s.Expire("a", 2*time.Second)
	if err != nil {
		t.Fatal("expected err == nil")
	}
	if !updated {
		t.Error("expected updated == true")
	}
	i += 2
	_, ok = s.Get("a")
	if ok {
		t.Error("expected: !exist, received: exist")
	}
	updated, err = s.Expire("a", time.Second)
	if err != nil {
		t.Fatal("expected err == nil")
	}
	if updated {
		t.Error("expected updated == false")
	}
}
