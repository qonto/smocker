package types

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/sbabiv/xml2map"
	log "github.com/sirupsen/logrus"
	"github.com/smartystreets/assertions"
	"github.com/stretchr/objx"
)

const (
	DefaultMatcher = "ShouldEqual"
)

type Assertion func(actual interface{}, expected ...interface{}) string

var asserts = map[string]Assertion{
	"ShouldResemble":         assertions.ShouldResemble,
	"ShouldAlmostEqual":      assertions.ShouldAlmostEqual,
	"ShouldContainSubstring": assertions.ShouldContainSubstring,
	"ShouldEndWith":          assertions.ShouldEndWith,
	"ShouldEqual":            assertions.ShouldEqual,
	"ShouldEqualJSON":        assertions.ShouldEqualJSON,
	"ShouldEqualXML":         ShouldEqualXML,
	"ShouldStartWith":        assertions.ShouldStartWith,
	"ShouldBeEmpty":          ShouldBeEmpty,
	"ShouldMatch":            ShouldMatch,

	"ShouldNotResemble":         assertions.ShouldNotResemble,
	"ShouldNotAlmostEqual":      assertions.ShouldNotAlmostEqual,
	"ShouldNotContainSubstring": assertions.ShouldNotContainSubstring,
	"ShouldNotEndWith":          assertions.ShouldNotEndWith,
	"ShouldNotEqual":            assertions.ShouldNotEqual,
	"ShouldNotStartWith":        assertions.ShouldNotStartWith,
	"ShouldNotBeEmpty":          ShouldNotBeEmpty,
	"ShouldNotMatch":            ShouldNotMatch,
}

func ShouldMatch(value interface{}, patterns ...interface{}) string {
	valueString, ok := value.(string)
	if !ok {
		return "ShouldMatch works only with strings"
	}

	for _, pattern := range patterns {
		patternString, ok := pattern.(string)
		if !ok {
			return "ShouldMatch works only with strings"
		}

		if match, err := regexp.MatchString(patternString, valueString); !match || err != nil {
			return fmt.Sprintf("Expected %q to match %q (but it didn't)!", valueString, patternString)
		}
	}

	return ""
}

func ShouldEqualXML(actual interface{}, expected ...interface{}) string {
	//Transform string into map
	decoder := xml2map.NewDecoder(strings.NewReader(expected[0].(string)))
	exp, err := decoder.Decode()
	if err != nil {
		panic(err)
	}

	decoder = xml2map.NewDecoder(strings.NewReader(actual.(string)))
	act, err := decoder.Decode()
	if err != nil {
		return fmt.Sprintf("error decoding `%s`, err is: %s", actual.(string), err)
	}

	res := walk(
		reflect.ValueOf(exp),
		reflect.ValueOf(act),
	)

	return res
}

func formatDiff(v1, v2 reflect.Value) string {
	return fmt.Sprintf("%v is different that %v", v1, v2)
}

func walk(v1, v2 reflect.Value) string {
	fmt.Printf("Visiting %v\n", v1)
	// Indirect through pointers and interfaces
	for v1.Kind() == reflect.Ptr || v1.Kind() == reflect.Interface {
		v1 = v1.Elem()
		v2 = v2.Elem()
	}

	switch v1.Kind() {
	case reflect.Array, reflect.Slice:
		if v1.Kind() != v2.Kind() {
			return formatDiff(v1, v2)
		}
		if v1.Len() != v2.Len() {
			return formatDiff(v1, v2)
		}

		for i := 0; i < v1.Len(); i++ {
			ret := walk(v1.Index(i), v2.Index(i))
			if ret != "" {
				return formatDiff(v1, v2)
			}
		}
	case reflect.Map:
		if v1.Kind() != v2.Kind() {
			return formatDiff(v1, v2)
		}
		if v1.Len() != v2.Len() {
			return formatDiff(v1, v2)
		}

		for _, k := range v1.MapKeys() {
			v := v2.MapIndex(k)
			if !v.IsValid() {
				return formatDiff(v1, v2)
			}

			ret := walk(v1.MapIndex(k), v2.MapIndex(k))
			if ret != "" {
				return formatDiff(v1, v2)
			}
		}
	case reflect.String:
		if v1.String() == "[[IGNORE]]" {
			return ""
		}
		if v1.Kind() != v2.Kind() {
			return formatDiff(v1, v2)
		}
		if v1.String() != v2.String() {
			return formatDiff(v1, v2)
		}

	default:
		if v1.Kind() != v2.Kind() {
			return formatDiff(v1, v2)
		}
		if v1.String() != v2.String() {
			return formatDiff(v1, v2)
		}
	}

	return ""
}

func ShouldBeEmpty(value interface{}, patterns ...interface{}) string {
	return assertions.ShouldBeEmpty(value)
}

func ShouldNotBeEmpty(value interface{}, patterns ...interface{}) string {
	return assertions.ShouldNotBeEmpty(value)
}

func ShouldNotMatch(value interface{}, patterns ...interface{}) string {
	valueString, ok := value.(string)
	if !ok {
		return "ShouldNotMatch works only with strings"
	}

	for _, pattern := range patterns {
		patternString, ok := pattern.(string)
		if !ok {
			return "ShouldNotMatch works only with strings"
		}

		if match, err := regexp.MatchString(patternString, valueString); match && err == nil {
			return fmt.Sprintf("Expected %q to not match %q (but it did)!", valueString, patternString)
		}
	}

	return ""
}

type StringMatcher struct {
	Matcher string `json:"matcher" yaml:"matcher,flow"`
	Value   string `json:"value" yaml:"value,flow"`
}

func (sm StringMatcher) Validate() error {
	if _, ok := asserts[sm.Matcher]; !ok {
		return fmt.Errorf("invalid matcher %q", sm.Matcher)
	}

	// Try to compile ShouldMatch regular expressions
	if sm.Matcher == "ShouldMatch" || sm.Matcher == "ShouldNotMatch" {
		if _, err := regexp.Compile(sm.Value); err != nil {
			return fmt.Errorf("invalid regular expression provided to %q operator: %v", sm.Matcher, sm.Value)
		}
	}
	return nil
}

func (sm StringMatcher) Match(value string) bool {
	matcher := asserts[sm.Matcher]
	if matcher == nil {
		log.WithField("matcher", sm.Matcher).Error("Invalid matcher")
		return false
	}

	if res := matcher(value, sm.Value); res != "" {
		log.Tracef("Value doesn't match:\n%s", res)
		return false
	}

	return true
}

func (sm *StringMatcher) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		sm.Matcher = DefaultMatcher
		sm.Value = s
		return nil
	}

	var res struct {
		Matcher string `json:"matcher"`
		Value   string `json:"value"`
	}

	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	sm.Matcher = res.Matcher
	sm.Value = res.Value
	return sm.Validate()
}

func (sm *StringMatcher) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		sm.Matcher = DefaultMatcher
		sm.Value = s
		return nil
	}

	var res struct {
		Matcher string `yaml:"matcher,flow"`
		Value   string `yaml:"value,flow"`
	}

	if err := unmarshal(&res); err != nil {
		return err
	}

	sm.Matcher = res.Matcher
	sm.Value = res.Value
	return sm.Validate()
}

type StringMatcherSlice []StringMatcher

func (sms StringMatcherSlice) Match(values []string) bool {
	if len(sms) > len(values) {
		return false
	}
	for _, matcher := range sms {
		matched := false
		for _, v := range values {
			if matcher.Match(v) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func (sms *StringMatcherSlice) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*sms = []StringMatcher{{
			Matcher: DefaultMatcher,
			Value:   s,
		}}
		return nil
	}

	var sm StringMatcher
	if err := json.Unmarshal(data, &sm); err == nil {
		*sms = []StringMatcher{sm}
		return nil
	}

	var res []StringMatcher
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	*sms = res
	return nil
}

func (sms *StringMatcherSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		*sms = []StringMatcher{{
			Matcher: DefaultMatcher,
			Value:   s,
		}}
		return nil
	}

	var sm StringMatcher
	if err := unmarshal(&sm); err == nil {
		*sms = []StringMatcher{sm}
		return nil
	}

	var res []StringMatcher
	if err := unmarshal(&res); err != nil {
		return err
	}
	*sms = res
	return nil
}

type MultiMapMatcher map[string]StringMatcherSlice

func (mmm MultiMapMatcher) Match(values map[string][]string) bool {
	if len(mmm) > len(values) {
		return false
	}
	for key, matcherValue := range mmm {
		value, ok := values[key]
		if !ok || !matcherValue.Match(value) {
			return false
		}
	}
	return true
}

type BodyMatcher struct {
	bodyString *StringMatcher
	bodyJson   map[string]StringMatcher
}

func (bm BodyMatcher) Match(headers http.Header, value string) bool {
	if bm.bodyString != nil {
		return bm.bodyString.Match(value)
	}

	if headers.Get("Content-Type") == "application/x-www-form-urlencoded" {
		m, err := url.ParseQuery(value)
		if err != nil {
			log.WithError(err).Error("Failed to read request body as encoded form")
		} else if b, err := json.Marshal(m); err != nil {
			log.WithError(err).Error("Failed to serialize form body as JSON")
		} else {
			value = string(b)
		}
	}

	j, err := objx.FromJSON(value)
	if err != nil {
		return false
	}
	for path, matcher := range bm.bodyJson {
		value := j.Get(path)
		if value == nil {
			return false
		}
		if ok := matcher.Match(value.String()); !ok {
			return false
		}
	}
	return true
}

func (bm BodyMatcher) MarshalJSON() ([]byte, error) {
	if bm.bodyString != nil {
		return json.Marshal(bm.bodyString)
	}
	return json.Marshal(bm.bodyJson)
}

func (bm *BodyMatcher) UnmarshalJSON(data []byte) error {
	var s StringMatcher
	if err := json.Unmarshal(data, &s); err == nil {
		if _, ok := asserts[s.Matcher]; ok {
			bm.bodyString = &s
			return nil
		}
	}

	var res map[string]StringMatcher
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}
	bm.bodyJson = res
	return nil
}

func (bm BodyMatcher) MarshalYAML() (interface{}, error) {
	if bm.bodyString != nil {
		return bm.bodyString, nil
	}
	return bm.bodyJson, nil
}

func (bm *BodyMatcher) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s StringMatcher
	if err := unmarshal(&s); err == nil {
		if _, ok := asserts[s.Matcher]; ok {
			bm.bodyString = &s
			return nil
		}
	}

	var res map[string]StringMatcher
	if err := unmarshal(&res); err != nil {
		return err
	}
	bm.bodyJson = res
	return nil
}
