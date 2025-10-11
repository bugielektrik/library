package helpers

import (
	"reflect"
	"strings"
	"testing"
)

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("\nExpected: %v\nActual:   %v", expected, actual)
	}
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected values to be different, but both were: %v", actual)
	}
}

// AssertNoError checks that no error occurred
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// AssertError checks that an error occurred
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
}

// AssertErrorContains checks that an error contains a specific substring
func AssertErrorContains(t *testing.T, err error, substring string) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
	if !strings.Contains(err.Error(), substring) {
		t.Errorf("Error %q does not contain %q", err.Error(), substring)
	}
}

// AssertTrue checks that a condition is true
func AssertTrue(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if !condition {
		message := "Expected true but got false"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Error(message)
	}
}

// AssertFalse checks that a condition is false
func AssertFalse(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if condition {
		message := "Expected false but got true"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Error(message)
	}
}

// AssertNil checks that a value is nil
func AssertNil(t *testing.T, value interface{}) {
	t.Helper()
	if value != nil && !reflect.ValueOf(value).IsNil() {
		t.Errorf("Expected nil but got: %v", value)
	}
}

// AssertNotNil checks that a value is not nil
func AssertNotNil(t *testing.T, value interface{}) {
	t.Helper()
	if value == nil || reflect.ValueOf(value).IsNil() {
		t.Error("Expected non-nil value but got nil")
	}
}

// AssertLen checks the length of a slice, map, or string
func AssertLen(t *testing.T, collection interface{}, expectedLen int) {
	t.Helper()
	v := reflect.ValueOf(collection)
	actualLen := v.Len()
	if actualLen != expectedLen {
		t.Errorf("Expected length %d but got %d", expectedLen, actualLen)
	}
}

// AssertContains checks if a slice or string contains an element/substring
func AssertContains(t *testing.T, collection, item interface{}) {
	t.Helper()

	// Handle string contains
	if str, ok := collection.(string); ok {
		if substr, ok := item.(string); ok {
			if !strings.Contains(str, substr) {
				t.Errorf("String %q does not contain %q", str, substr)
			}
			return
		}
	}

	// Handle slice contains
	v := reflect.ValueOf(collection)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		t.Fatalf("AssertContains requires a slice, array, or string, got %T", collection)
	}

	for i := 0; i < v.Len(); i++ {
		if reflect.DeepEqual(v.Index(i).Interface(), item) {
			return
		}
	}

	t.Errorf("Collection does not contain %v", item)
}

// AssertPanic checks that a function panics
func AssertPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("Expected panic but function completed normally")
		}
	}()
	fn()
}
