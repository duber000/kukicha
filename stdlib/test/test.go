package test

import (
	"fmt"
	"reflect"
	"testing"
)

// AssertEqual fails the test if got != want (using reflect.DeepEqual).
func AssertEqual(t *testing.T, got, want any, msgAndArgs ...any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		prefix := ""
		if len(msgAndArgs) > 0 {
			prefix = fmt.Sprintf("%v: ", msgAndArgs[0])
		}
		t.Errorf("%sexpected %v, got %v", prefix, want, got)
	}
}

// AssertTrue fails the test if condition is false.
func AssertTrue(t *testing.T, condition bool, msgAndArgs ...any) {
	t.Helper()
	if !condition {
		prefix := "expected true"
		if len(msgAndArgs) > 0 {
			prefix = fmt.Sprintf("%v", msgAndArgs[0])
		}
		t.Errorf("%s", prefix)
	}
}

// AssertFalse fails the test if condition is true.
func AssertFalse(t *testing.T, condition bool, msgAndArgs ...any) {
	t.Helper()
	if condition {
		prefix := "expected false"
		if len(msgAndArgs) > 0 {
			prefix = fmt.Sprintf("%v", msgAndArgs[0])
		}
		t.Errorf("%s", prefix)
	}
}

// AssertNoError fails the test if err is non-nil.
func AssertNoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err != nil {
		prefix := ""
		if len(msgAndArgs) > 0 {
			prefix = fmt.Sprintf("%v: ", msgAndArgs[0])
		}
		t.Errorf("%sunexpected error: %v", prefix, err)
	}
}

// AssertError fails the test if err is nil.
func AssertError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err == nil {
		prefix := "expected an error, got nil"
		if len(msgAndArgs) > 0 {
			prefix = fmt.Sprintf("%v", msgAndArgs[0])
		}
		t.Errorf("%s", prefix)
	}
}

// AssertNotEmpty fails the test if val is nil or the zero value of its type.
func AssertNotEmpty(t *testing.T, val any, msgAndArgs ...any) {
	t.Helper()
	if val == nil {
		prefix := "expected non-empty value, got nil"
		if len(msgAndArgs) > 0 {
			prefix = fmt.Sprintf("%v", msgAndArgs[0])
		}
		t.Errorf("%s", prefix)
		return
	}
	v := reflect.ValueOf(val)
	if v.IsZero() {
		prefix := "expected non-empty value, got zero value"
		if len(msgAndArgs) > 0 {
			prefix = fmt.Sprintf("%v", msgAndArgs[0])
		}
		t.Errorf("%s", prefix)
	}
}
