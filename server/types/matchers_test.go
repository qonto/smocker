package types

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestStringMatcherJSON(t *testing.T) {
	test := `"test"`
	serialized := `{"matcher":"ShouldEqual","value":"test"}`

	var res StringMatcher
	if err := json.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res.Matcher != DefaultMatcher {
		t.Fatalf("matcher %s should be equal to %s", res.Matcher, DefaultMatcher)
	}
	if res.Value != "test" {
		t.Fatalf("value %s should be equal to %s", res.Value, "test")
	}

	b, err := json.Marshal(&res)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != serialized {
		t.Fatalf("serialized value %s should be equal to %s", string(b), test)
	}

	test = `{"matcher":"ShouldEqual","value":"test2"}`
	serialized = test
	res = StringMatcher{}
	if err = json.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res.Matcher != "ShouldEqual" {
		t.Fatalf("matcher %s should be equal to %s", res.Matcher, "ShouldEqual")
	}
	if res.Value != "test2" {
		t.Fatalf("value %s should be equal to %s", res.Value, "test2")
	}

	b, err = json.Marshal(&res)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != serialized {
		t.Fatalf("serialized value %s should be equal to %s", string(b), test)
	}
}

func TestStringMatcherYAML(t *testing.T) {
	test := `test`
	var res StringMatcher
	if err := yaml.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res.Matcher != DefaultMatcher {
		t.Fatalf("matcher %s should be equal to %s", res.Matcher, DefaultMatcher)
	}
	if res.Value != "test" {
		t.Fatalf("value %s should be equal to %s", res.Value, "test")
	}

	if _, err := yaml.Marshal(&res); err != nil {
		t.Fatal(err)
	}

	test = `{"matcher":"ShouldEqual","value":"test2"}`
	res = StringMatcher{}
	if err := yaml.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res.Matcher != "ShouldEqual" {
		t.Fatalf("matcher %s should be equal to %s", res.Matcher, "ShouldEqual")
	}
	if res.Value != "test2" {
		t.Fatalf("value %s should be equal to %s", res.Value, "test2")
	}

	if _, err := yaml.Marshal(&res); err != nil {
		t.Fatal(err)
	}
}

func TestMultiMapMatcherJSON(t *testing.T) {
	test := `{"test":"test"}`
	serialized := `{"test":[{"matcher":"ShouldEqual","value":"test"}]}`
	var res MultiMapMatcher
	if err := json.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res == nil {
		t.Fatal("multimap matcher should not be nil")
	}

	for _, value := range res {
		if value[0].Matcher != DefaultMatcher {
			t.Fatalf("matcher %s should be equal to %s", value[0].Matcher, DefaultMatcher)
		}
	}

	expected := MultiMapMatcher{
		"test": {
			{Matcher: "ShouldEqual", Value: "test"},
		},
	}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("values %v should be equal to %v", res, expected)
	}

	b, err := json.Marshal(&res)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != serialized {
		t.Fatalf("serialized value %s should be equal to %s", string(b), test)
	}

	test = `{"test":{"matcher":"ShouldEqual","value":"test3"}}`
	serialized = `{"test":[{"matcher":"ShouldEqual","value":"test3"}]}`
	res = MultiMapMatcher{}
	if err = json.Unmarshal([]byte(test), &res); err != nil {
		t.Fatal(err)
	}

	if res["test"][0].Matcher != "ShouldEqual" {
		t.Fatalf("matcher %s should be equal to %s", res["test"][0].Matcher, "ShouldEqual")
	}

	expected = MultiMapMatcher{
		"test": {
			{Matcher: "ShouldEqual", Value: "test3"},
		},
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("values %v should be equal to %v", res, expected)
	}

	b, err = json.Marshal(&res)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != serialized {
		t.Fatalf("serialized value %s should be equal to %s", string(b), test)
	}
}

func TestShouldEqualJSON(t *testing.T) {
	provided := `{
		"testkey":"testvalue"
	}`
	expected := `{"testkey" : "testvalue"}`
	diff := assertions.ShouldEqualJSON(provided, expected)
	assert.Empty(t, diff)
}

func Test_walk(t *testing.T) {

	tests := []struct {
		name     string
		expected string
		actual   string
		want     bool
	}{
		{
			name:     "ok 1",
			expected: "<toto>PLOP</toto>",
			actual:   "<toto>PLOP</toto>",
			want:     true,
		},
		{
			name:     "nok 1",
			expected: "<toto>PLOP</toto>",
			actual:   "<toto>PLUP</toto>",
			want:     false,
		},
		{
			name:     "ok 2",
			expected: "<toto><tutu><tata>d</tata></tutu></toto>",
			actual:   "<toto><tutu><tata>d</tata></tutu></toto>",
			want:     true,
		},
		{
			name:     "nok 2",
			expected: "<toto><tutu><tata>d</tata></tutu></toto>",
			actual:   "<toto><tutu><toto>d</toto></tutu></toto>",
			want:     false,
		},
		{
			name:     "ok array",
			expected: "<toto><tutu>1</tutu><tutu>2</tutu></toto>",
			actual:   "<toto><tutu>1</tutu><tutu>2</tutu></toto>",
			want:     true,
		},
		{
			name:     "nok array",
			expected: "<toto><tutu>1</tutu><tutu>2</tutu></toto>",
			actual:   "<toto><tutu>1</tutu><tutu>3</tutu></toto>",
			want:     false,
		},
		{
			name:     "ok ignore",
			expected: "<toto><tutu><tata>${xmlunit.ignore}</tata></tutu></toto>",
			actual:   "<toto><tutu><tata>d</tata></tutu></toto>",
			want:     true,
		},
		{
			name:     "ok ignore a sub map",
			expected: "<toto><tutu><tata>${xmlunit.ignore}</tata></tutu></toto>",
			actual:   "<toto><tutu><tata><a>r</a></tata></tutu></toto>",
			want:     true,
		},
		{
			name:     "ok xml attribute",
			expected: `<toto tt="aa">PLOP</toto>`,
			actual:   `<toto tt="aa">PLOP</toto>`,
			want:     true,
		},
		{
			name:     "nok xml attribute",
			expected: `<toto tt="aa">PLOP</toto>`,
			actual:   `<toto tt="ba">PLOP</toto>`,
			want:     false,
		},
		{
			name:     "nok xml attribute",
			expected: `<toto tt="aa">PLOP</toto>`,
			actual:   `<toto>PLOP</toto>`,
			want:     false,
		},
		{
			name:     "nok xml attribute",
			expected: `<toto>PLOP</toto>`,
			actual:   `<toto tt="aa">PLOP</toto>`,
			want:     false,
		},
		{
			name:     "ok different order",
			expected: "<toto><tutu>a</tutu><tata>b</tata></toto>",
			actual:   "<toto><tata>b</tata><tutu>a</tutu></toto>",
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := ShouldEqualXML(tt.actual, tt.expected)
			got := diff == ""
			assert.Equal(t, tt.want, got)
		})
	}
}
