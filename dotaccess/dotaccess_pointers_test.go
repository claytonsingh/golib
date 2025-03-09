package dotaccess

import "testing"

func TestPointers(t *testing.T) {
	t.Run("TestMultiLevelPointerIndirection", func(t *testing.T) {
		// Test structure with fields having multiple levels of indirection
		type NestedStruct struct {
			Value    string
			ValuePtr *string
		}

		type TestStruct struct {
			Name      **string          // Double pointer
			Value     ***int            // Triple pointer
			Nested    **NestedStruct    // Double pointer to struct
			PtrArray  *[]*string        // Pointer to array of pointers
			MapOfPtrs *map[string]**int // Pointer to map of double pointers
		}

		// Setup test data
		name := "John"
		namePtr := &name

		value := 42
		valuePtr := &value
		valuePtrPtr := &valuePtr

		nestedValue := "nested"
		nested := NestedStruct{
			Value:    "direct",
			ValuePtr: &nestedValue,
		}
		nestedPtr := &nested

		strings := []string{"one", "two", "three"}
		stringPtrs := []*string{&strings[0], &strings[1], &strings[2]}

		mapValue1 := 100
		mapValue1Ptr := &mapValue1
		mapValue2 := 200
		mapValue2Ptr := &mapValue2
		mapOfPtrs := map[string]**int{
			"first":  &mapValue1Ptr,
			"second": &mapValue2Ptr,
		}

		// Create the test struct
		testStruct := TestStruct{
			Name:      &namePtr,
			Value:     &valuePtrPtr,
			Nested:    &nestedPtr,
			PtrArray:  &stringPtrs,
			MapOfPtrs: &mapOfPtrs,
		}

		// Test 1: Get and set double pointer string field
		t.Run("double pointer string field", func(t *testing.T) {
			accessor, err := GetAccessorDot[string](&testStruct, "Name")
			if err != nil {
				t.Fatalf("Failed to get accessor for Name: %v", err)
			}

			// Verify initial value
			if accessor.Get() != "John" {
				t.Errorf("Expected Name to be 'John', got '%v'", accessor.Get())
			}

			// Set new value
			err = accessor.Set("Jane")
			if err != nil {
				t.Fatalf("Failed to set Name: %v", err)
			}

			// Verify value changed
			if accessor.Get() != "Jane" {
				t.Errorf("Expected Name to be 'Jane', got '%v'", accessor.Get())
			}

			// Verify original variable changed
			if **testStruct.Name != "Jane" {
				t.Errorf("Expected **Name to be 'Jane', got '%v'", **testStruct.Name)
			}
		})

		// Test 2: Get and set triple pointer int field
		t.Run("triple pointer int field", func(t *testing.T) {
			accessor, err := GetAccessorDot[int](&testStruct, "Value")
			if err != nil {
				t.Fatalf("Failed to get accessor for Value: %v", err)
			}

			// Verify initial value
			if accessor.Get() != 42 {
				t.Errorf("Expected Value to be 42, got '%v'", accessor.Get())
			}

			// Set new value
			err = accessor.Set(99)
			if err != nil {
				t.Fatalf("Failed to set Value: %v", err)
			}

			// Verify value changed
			if accessor.Get() != 99 {
				t.Errorf("Expected Value to be 99, got '%v'", accessor.Get())
			}

			// Verify original variable changed
			if ***testStruct.Value != 99 {
				t.Errorf("Expected ***Value to be 99, got '%v'", ***testStruct.Value)
			}
		})

		// Test 3: Access nested struct through double pointer
		t.Run("double pointer to struct", func(t *testing.T) {
			accessor, err := GetAccessorDot[string](&testStruct, "Nested.Value")
			if err != nil {
				t.Fatalf("Failed to get accessor for Nested.Value: %v", err)
			}

			// Verify initial value
			if accessor.Get() != "direct" {
				t.Errorf("Expected Nested.Value to be 'direct', got '%v'", accessor.Get())
			}

			// Set new value
			err = accessor.Set("changed")
			if err != nil {
				t.Fatalf("Failed to set Nested.Value: %v", err)
			}

			// Verify value changed
			if accessor.Get() != "changed" {
				t.Errorf("Expected Nested.Value to be 'changed', got '%v'", accessor.Get())
			}

			// Test nested pointer field
			ptrAccessor, err := GetAccessorDot[string](&testStruct, "Nested.ValuePtr")
			if err != nil {
				t.Fatalf("Failed to get accessor for Nested.ValuePtr: %v", err)
			}

			// Verify initial value
			if ptrAccessor.Get() != "nested" {
				t.Errorf("Expected Nested.ValuePtr to be 'nested', got '%v'", ptrAccessor.Get())
			}

			// Set new value
			err = ptrAccessor.Set("updated nested")
			if err != nil {
				t.Fatalf("Failed to set Nested.ValuePtr: %v", err)
			}

			// Verify value changed
			if ptrAccessor.Get() != "updated nested" {
				t.Errorf("Expected Nested.ValuePtr to be 'updated nested', got '%v'", ptrAccessor.Get())
			}
		})

		// Test 4: Access element in pointer to array of pointers
		t.Run("pointer to array of pointers", func(t *testing.T) {
			accessor, err := GetAccessorDot[string](&testStruct, "PtrArray.1")
			if err != nil {
				t.Fatalf("Failed to get accessor for PtrArray.1: %v", err)
			}

			// Verify initial value
			if accessor.Get() != "two" {
				t.Errorf("Expected PtrArray.1 to be 'two', got '%v'", accessor.Get())
			}

			// Set new value
			err = accessor.Set("TWO")
			if err != nil {
				t.Fatalf("Failed to set PtrArray.1: %v", err)
			}

			// Verify value changed
			if accessor.Get() != "TWO" {
				t.Errorf("Expected PtrArray.1 to be 'TWO', got '%v'", accessor.Get())
			}
		})

		// Test 5: Access value in pointer to map of double pointers
		t.Run("pointer to map of double pointers", func(t *testing.T) {
			accessor, err := GetAccessorDot[int](&testStruct, "MapOfPtrs.first")
			if err != nil {
				t.Fatalf("Failed to get accessor for MapOfPtrs.first: %v", err)
			}

			// Verify initial value
			if accessor.Get() != 100 {
				t.Errorf("Expected MapOfPtrs.first to be 100, got '%v'", accessor.Get())
			}

			// Set new value
			err = accessor.Set(150)
			if err != nil {
				t.Fatalf("Failed to set MapOfPtrs.first: %v", err)
			}

			// Verify value changed
			if accessor.Get() != 150 {
				t.Errorf("Expected MapOfPtrs.first to be 150, got '%v'", accessor.Get())
			}

			// Verify original value changed
			if **(*testStruct.MapOfPtrs)["first"] != 150 {
				t.Errorf("Expected map value to be 150, got '%v'", **(*testStruct.MapOfPtrs)["first"])
			}
		})
	})

	t.Run("pointer indirection syntax", func(t *testing.T) {
		// Create a multi-level pointer scenario
		str := "original"
		strPtr := &str
		strPtrPtr := &strPtr
		strPtrPtrPtr := &strPtrPtr

		// Create a test struct with multi-level pointer fields
		type TestPointers struct {
			Single *string   // *string
			Double **string  // **string
			Triple ***string // ***string
		}

		testStruct := TestPointers{
			Single: strPtr,       // *string
			Double: strPtrPtr,    // **string
			Triple: strPtrPtrPtr, // ***string
		}

		// 1. Test single level indirection with * syntax
		t.Run("single * indirection with value accessor", func(t *testing.T) {
			origPtr := testStruct.Single

			accessor, err := GetAccessorDot[string](&testStruct, "Single*")
			if err != nil {
				t.Fatalf("Failed to get accessor for Single: %v", err)
			}
			if accessor.Get() != "original" {
				t.Errorf("Expected to get 'original', got '%v'", accessor.Get())
			}

			err = accessor.Set("modified")
			if err != nil {
				t.Fatalf("Failed to set Single: %v", err)
			}
			if *testStruct.Single != "modified" {
				t.Errorf("Expected Single value to be 'modified', got '%v'", *testStruct.Single)
			}
			if origPtr != testStruct.Single {
				t.Errorf("Expected Single pointer to be unchanged")
			}
		})

		// 2. Test single level indirection with * syntax
		t.Run("single * indirection with pointer single accessor", func(t *testing.T) {
			// Reset test value
			*testStruct.Single = "original"
			origPtr := testStruct.Single

			// Access the value via a pointer
			accessor, err := GetAccessorDot[*string](&testStruct, "Single*")
			if err != nil {
				t.Fatalf("Failed to get accessor for Single*: %v", err)
			}

			// Verify we got the pointer
			ptr := accessor.Get()
			if *ptr != "original" {
				t.Errorf("Expected pointed value to be 'original', got '%v'", *ptr)
			}

			// Create new string and pointer
			newValue := "new value"
			newPtr := &newValue

			// Set the value via a pointer
			err = accessor.Set(newPtr)
			if err != nil {
				t.Fatalf("Failed to set Single*: %v", err)
			}

			if *testStruct.Single != "new value" {
				t.Errorf("Expected new pointed value to be 'new value', got '%v'", *testStruct.Single)
			}
			if str != "new value" {
				t.Errorf("Expected string value to be 'new value', got '%v'", str)
			}
			if origPtr != testStruct.Single {
				t.Errorf("Expected Single pointer to be unchanged")
			}
		})

		// 3. Test double pointer indirection with * syntax
		t.Run("double pointer single * indirection with single pointer accessor", func(t *testing.T) {
			// Reset test values
			oldStr := "original"
			oldPtr := &oldStr
			oldPtrPtr := &oldPtr
			testStruct.Double = oldPtrPtr

			// Access the *string pointer, not the **string field or string value
			accessor, err := GetAccessorDot[*string](&testStruct, "Double*")
			if err != nil {
				t.Fatalf("Failed to get accessor for Double**: %v", err)
			}

			// Verify initial value
			if **testStruct.Double != "original" {
				t.Errorf("Initial value mismatch, got '%v'", **testStruct.Double)
			}

			// Create new string and pointer
			newValue := "double new value"
			newPtr := &newValue

			// Set new *string pointer (middle of the chain)
			err = accessor.Set(newPtr)
			if err != nil {
				t.Fatalf("Failed to set Double**: %v", err)
			}

			// Verify that the middle pointer was changed
			if **testStruct.Double != "double new value" {
				t.Errorf("Expected pointed value to be 'double new value', got '%v'", **testStruct.Double)
			}

			// Verify original string is unchanged
			if oldStr != "original" {
				t.Errorf("Original string should be unchanged, got '%v'", oldStr)
			}

			// Old pointer pointer should be unchanged
			if testStruct.Double != oldPtrPtr {
				t.Errorf("Expected root level pointer to be unchanged")
			}

			// Old pointer should be updated
			if *testStruct.Double != newPtr {
				t.Errorf("Expected second level pointer to be updated")
			}
		})

		// 4. Test triple pointer indirection with * syntax
		t.Run("triple *** indirection with value accessor", func(t *testing.T) {
			// Reset test values
			oldVal := "original triple"
			oldPtr := &oldVal
			oldPtrPtr := &oldPtr
			oldPtrPtrPtr := &oldPtrPtr
			testStruct.Triple = oldPtrPtrPtr

			// Access the **string pointer (deep in the chain)
			accessor, err := GetAccessorDot[string](&testStruct, "Triple***")
			if err != nil {
				t.Fatalf("Failed to get accessor for Triple***: %v", err)
			}

			// Create new values for the chain
			newValue := "new triple value"

			// Set new pointer
			err = accessor.Set(newValue)
			if err != nil {
				t.Fatalf("Failed to set Triple***: %v", err)
			}

			// Verify middle of the chain was modified
			if ***testStruct.Triple != "new triple value" {
				t.Errorf("Expected value to be 'new triple value', got '%v'", ***testStruct.Triple)
			}

			// Original shouldn't change
			if oldPtrPtrPtr != testStruct.Triple {
				t.Errorf("Original root level pointer shouldn't change")
			}

			// Original shouldn't change
			if oldPtrPtr != *testStruct.Triple {
				t.Errorf("Original second level pointer shouldn't change")
			}

			// Original shouldn't change
			if oldPtr != **testStruct.Triple {
				t.Errorf("Original third level pointer shouldn't change")
			}
		})
	})
}
