package dotaccess

import (
	"testing"
)

func TestSlice(t *testing.T) {

	t.Run("TestSliceAccess", func(t *testing.T) {
		data := struct {
			IntSlice []int
		}{
			IntSlice: []int{1, 2, 3},
		}

		accessor, err := NewAccessorDot[int](&data, "IntSlice.1")
		if err != nil {
			t.Fatalf("Failed to get accessor for IntSlice.1: %v", err)
		}

		value := accessor.Get()
		if value != 2 {
			t.Errorf("Expected value 2, got %v", value)
		}

		err = accessor.Set(4)
		if err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}
	})

	t.Run("TestSliceAccessErrors", func(t *testing.T) {
		data := struct {
			IntSlice []int
		}{
			IntSlice: []int{1, 2, 3},
		}

		// Test accessing non-existent index
		_, err := NewAccessorDot[int](&data, "IntSlice.4")
		if err == nil {
			t.Error("Expected error for non-existent index, got nil")
		}

		// Test accessing non-slice field
		_, err = NewAccessorDot[int](&data, "IntSlice.abc")
		if err == nil {
			t.Error("Expected error for non-slice field, got nil")
		}
	})

	// test pointer to slice of pointers
	t.Run("TestPointerToSliceOfPointers", func(t *testing.T) {
		data := struct {
			IntSlice *[]*int
		}{
			IntSlice: &[]*int{},
		}
		for i := range 3 {
			*data.IntSlice = append(*data.IntSlice, &i)
		}

		accessor, err := NewAccessorDot[int](&data, "IntSlice.1")
		if err != nil {
			t.Fatalf("Failed to get accessor for IntSlice.1: %v", err)
		}
		value := accessor.Get()
		if value != 1 {
			t.Errorf("Expected value 1, got %v", value)
		}

		err = accessor.Set(4)
		if err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}

		value = accessor.Get()
		if value != 4 {
			t.Errorf("Expected value 4, got %v", value)
		}
	})
}
