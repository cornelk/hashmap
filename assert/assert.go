// Package assert contains test assertion helpers.
package assert

import (
	"fmt"
	"reflect"
	"testing"
)

func equal(expected, actual any) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	if reflect.DeepEqual(expected, actual) {
		return true
	}

	actualType := reflect.TypeOf(actual)
	if actualType == nil {
		return false
	}
	expectedValue := reflect.ValueOf(expected)
	if expectedValue.IsValid() && expectedValue.Type().ConvertibleTo(actualType) {
		return reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual)
	}

	return false
}

func fail(t *testing.T, message string, errorMessage ...string) {
	t.Helper()
	if len(errorMessage) != 0 {
		message = fmt.Sprintf("%s\n%s", message, errorMessage)
	}
	t.Errorf(message)
	t.FailNow()
}

// Equal asserts that two objects are equal.
func Equal(t *testing.T, expected, actual any, errorMessage ...string) {
	t.Helper()
	if equal(expected, actual) {
		return
	}

	msg := fmt.Sprintf("Not equal: \nexpected: %v\nactual  : %v", expected, actual)
	fail(t, msg, errorMessage...)
}

// True asserts that the specified value is true.
func True(t *testing.T, value bool, errorMessage ...string) {
	t.Helper()
	if value {
		return
	}
	fail(t, "Unexpected false", errorMessage...)
}

// False asserts that the specified value is false.
func False(t *testing.T, value bool, errorMessage ...string) {
	t.Helper()
	if !value {
		return
	}
	fail(t, "Unexpected true", errorMessage...)
}
