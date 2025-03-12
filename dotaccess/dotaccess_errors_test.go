package dotaccess

import (
	"sync"
	"testing"
)

func TestGetAccessor_Errors(t *testing.T) {
	person := Person{
		Name: "John Doe",
		Age:  30,
		Tags: []string{"customer"},
	}

	// Test invalid field
	_, err := NewAccessorDot[string](&person, "InvalidField")
	if err == nil {
		t.Error("Expected error for invalid field, got nil")
	}

	// Test invalid path
	_, err = NewAccessorDot[string](&person, "")
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}

	// Test nil object
	var nilPerson *Person
	_, err = NewAccessorDot[string](nilPerson, "something")
	if err == nil {
		t.Error("Expected error for nil object, got nil")
	}

	// Test out of bounds slice access
	_, err = NewAccessorDot[string](&person, "Tags.5")
	if err == nil {
		t.Error("Expected error for out of bounds slice access, got nil")
	}

	// Test invalid slice index
	_, err = NewAccessorDot[string](&person, "Tags.abc")
	if err == nil {
		t.Error("Expected error for invalid slice index, got nil")
	}

	// Test type mismatch on set
	accessor, _ := NewAccessorDot[int](&person, "Age")
	err = accessor.Set(42) // Now type safe, can't pass string
	if err != nil {
		t.Errorf("Unexpected error for type-safe set: %v", err)
	}

	// Test asking for wrong type in accessor
	_, err = NewAccessorDot[string](&person, "Age")
	if err == nil {
		t.Error("Expected error for type mismatch in accessor, got nil")
	}

	// Test with empty path
	_, err = NewAccessorDot[string](&person, "")
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}

	// Test with empty path
	_, err = NewAccessor[string](&person, []string{}, 0)
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}

	// Test pointer to nil
	var nilPtr *int
	_, err = NewAccessorDot[int](nilPtr, "Age")
	if err == nil {
		t.Error("Expected error for pointer to nil, got nil")
	}

}

func Test_Errors(t *testing.T) {
	t.Run("DereferenceDepthLimit", func(t *testing.T) {

		// test dereference depth limit with pointer reference to self
		ref := struct {
			Self *any
		}{}

		var anyPtr any = &ref.Self
		ref.Self = &anyPtr

		_, err := NewAccessorDot[string](&ref, "Self.0")
		if err == nil {
			t.Error("Expected error for dereference depth limit, got nil")
		}
	})

	t.Run("UnexportedFieldError", func(t *testing.T) {
		// Create a sync.Mutex to test accessing its unexported 'state' field
		mutex := &sync.Mutex{}

		_, err := NewAccessorDot[int32](&mutex, "state")
		if err == nil {
			t.Error("Expected error for unexported field, got nil")
		}
	})

	t.Run("UnspportedObjectType", func(t *testing.T) {
		// Try getting a field on a string
		str := "not a pointer"
		_, err := NewAccessorDot[string](&str, "field")
		if err == nil {
			t.Error("Expected error for unsupported object type, got nil")
		}
	})

	t.Run("NilPointerError", func(t *testing.T) {
		// Try getting a field on a struct with a nil pointer
		testStruct := struct {
			Value *int
		}{
			Value: nil,
		}
		_, err := NewAccessorDot[int](&testStruct, "Value.field")
		if err == nil {
			t.Error("Expected error for nil pointer, got nil")
		}
	})

	t.Run("RequestDoublePointerToField", func(t *testing.T) {
		// Try getting a field on a struct with a double pointer
		testStruct := struct {
			Value int
		}{
			Value: 4,
		}
		_, err := NewAccessorDot[**int](&testStruct, "Value")
		if err == nil {
			t.Error("Expected error for double pointer to field, got nil")
		}
	})

	t.Run("RequestPointerToMapElement", func(t *testing.T) {
		// Try getting a field on a map with a pointer to an element
		testMap := map[string]int{
			"key": 4,
		}
		_, err := NewAccessorDot[*int](&testMap, "key")
		if err == nil {
			t.Error("Expected error for pointer to map element, got nil")
		}
	})
}
