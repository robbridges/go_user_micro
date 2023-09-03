package validator

import (
	"regexp"
	"testing"
)

func TestValidator_Valid(t *testing.T) {
	// Create a new validator
	v := New()

	// Initially, there should be no errors
	if !v.Valid() {
		t.Error("Expected validator to be valid, but it is not")
	}

	// Add an error
	v.AddError("field", "Invalid value")

	// Validator should be invalid now
	if v.Valid() {
		t.Error("Expected validator to be invalid, but it is valid")
	}
}

func TestValidator_AddError(t *testing.T) {
	// Create a new validator
	v := New()

	// Add an error
	v.AddError("field", "Invalid value")

	// Check if the error is added correctly
	if len(v.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(v.Errors))
	}

	// Check the error message
	if v.Errors["field"] != "Invalid value" {
		t.Errorf("Unexpected error message. Expected: %s, Got: %s", "Invalid value", v.Errors["field"])
	}

	// Add another error with the same key
	v.AddError("field", "Another error")

	// Check if the error is overwritten
	if len(v.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(v.Errors))
	}

	// Check the updated error message
	if v.Errors["field"] != "Invalid value" {
		t.Errorf("Unexpected error message. Expected: %s, Got: %s", "Invalid value", v.Errors["field"])
	}
}

func TestPermittedValue(t *testing.T) {
	// Test with integers
	ok := PermittedValue(2, 1, 2, 3)
	if !ok {
		t.Error("Expected permitted value, but got false")
	}

	ok = PermittedValue(4, 1, 2, 3)
	if ok {
		t.Error("Unexpected permitted value, but got true")
	}

	// Test with strings
	ok = PermittedValue("b", "a", "b", "c")
	if !ok {
		t.Error("Expected permitted value, but got false")
	}

	ok = PermittedValue("d", "a", "b", "c")
	if ok {
		t.Error("Unexpected permitted value, but got true")
	}
}

func TestMatches(t *testing.T) {
	rx := regexp.MustCompile(`^\d{3}-\d{3}-\d{4}$`)

	// Test matching value
	ok := Matches("123-456-7890", rx)
	if !ok {
		t.Error("Expected match, but got false")
	}

	// Test non-matching value
	ok = Matches("abc-123", rx)
	if ok {
		t.Error("Unexpected match, but got true")
	}
}

func TestUnique(t *testing.T) {
	// Test with integers
	values := []int{1, 2, 3, 4}
	ok := Unique(values)
	if !ok {
		t.Error("Expected unique values, but got false")
	}

	values = []int{1, 2, 3, 2}
	ok = Unique(values)
	if ok {
		t.Error("Unexpected non-unique values, but got true")
	}

	// Test with strings
	strings := []string{"apple", "banana", "cherry", "banana"}
	ok = Unique(strings)
	if ok {
		t.Error("Unexpected non-unique values, but got true")
	}
}

func TestValidator_Check(t *testing.T) {
	// Create a new validator
	v := New()

	// Call the Check function with a true condition
	v.Check(true, "field", "Valid condition")

	// The validator should not have any errors
	if len(v.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(v.Errors))
	}

	// Call the Check function with a false condition
	v.Check(false, "field", "Invalid condition")

	// The validator should have an error
	if len(v.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(v.Errors))
	}

	// Check the error message
	if v.Errors["field"] != "Invalid condition" {
		t.Errorf("Unexpected error message. Expected: %s, Got: %s", "Invalid condition", v.Errors["field"])
	}
}
