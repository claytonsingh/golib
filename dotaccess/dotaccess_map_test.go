package dotaccess

import (
	"testing"
)

func TestMap(t *testing.T) {

	// Test map access with all supported key types
	// String, Bool, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr
	t.Run("TestMapAccess", func(t *testing.T) {
		// Test map with string keys
		t.Run("StringKeys", func(t *testing.T) {
			data := struct {
				StrMap map[string]int
			}{
				StrMap: map[string]int{
					"one": 1,
					"two": 2,
				},
			}

			accessor, err := GetAccessorDot[int](&data, "StrMap.one")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != 1 {
				t.Errorf("Expected 1, got %d", value)
			}

			// Test setting a value
			err = accessor.Set(100)
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.StrMap["one"] != 100 {
				t.Errorf("Expected 100, got %d", data.StrMap["one"])
			}
		})

		// Test map with bool keys
		t.Run("BoolKeys", func(t *testing.T) {
			data := struct {
				BoolMap map[bool]string
			}{
				BoolMap: map[bool]string{
					true:  "yes",
					false: "no",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "BoolMap.true")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "yes" {
				t.Errorf("Expected 'yes', got %s", value)
			}

			// Also test "false" key
			accessor, err = GetAccessorDot[string](&data, "BoolMap.false")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value = accessor.Get()
			if value != "no" {
				t.Errorf("Expected 'no', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.BoolMap[false] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.BoolMap[false])
			}
		})

		// Test map with int keys
		t.Run("IntKeys", func(t *testing.T) {
			data := struct {
				IntMap map[int]string
			}{
				IntMap: map[int]string{
					1: "one",
					2: "two",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "IntMap.1")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "one" {
				t.Errorf("Expected 'one', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.IntMap[1] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.IntMap[1])
			}
		})

		// Test map with int8 keys
		t.Run("Int8Keys", func(t *testing.T) {
			data := struct {
				Int8Map map[int8]string
			}{
				Int8Map: map[int8]string{
					1: "one",
					2: "two",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "Int8Map.1")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "one" {
				t.Errorf("Expected 'one', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.Int8Map[1] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.Int8Map[1])
			}
		})

		// Test map with int16 keys
		t.Run("Int16Keys", func(t *testing.T) {
			data := struct {
				Int16Map map[int16]string
			}{
				Int16Map: map[int16]string{
					1: "one",
					2: "two",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "Int16Map.1")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "one" {
				t.Errorf("Expected 'one', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.Int16Map[1] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.Int16Map[1])
			}
		})

		// Test map with int32 keys
		t.Run("Int32Keys", func(t *testing.T) {
			data := struct {
				Int32Map map[int32]string
			}{
				Int32Map: map[int32]string{
					1: "one",
					2: "two",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "Int32Map.1")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "one" {
				t.Errorf("Expected 'one', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.Int32Map[1] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.Int32Map[1])
			}
		})

		// Test map with int64 keys
		t.Run("Int64Keys", func(t *testing.T) {
			data := struct {
				Int64Map map[int64]string
			}{
				Int64Map: map[int64]string{
					1: "one",
					2: "two",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "Int64Map.1")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "one" {
				t.Errorf("Expected 'one', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.Int64Map[1] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.Int64Map[1])
			}
		})

		// Test map with uint keys
		t.Run("UintKeys", func(t *testing.T) {
			data := struct {
				UintMap map[uint]string
			}{
				UintMap: map[uint]string{
					1: "one",
					2: "two",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "UintMap.1")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "one" {
				t.Errorf("Expected 'one', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.UintMap[1] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.UintMap[1])
			}
		})

		// Test map with uint8 keys
		t.Run("Uint8Keys", func(t *testing.T) {
			data := struct {
				Uint8Map map[uint8]string
			}{
				Uint8Map: map[uint8]string{
					1: "one",
					2: "two",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "Uint8Map.1")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "one" {
				t.Errorf("Expected 'one', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.Uint8Map[1] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.Uint8Map[1])
			}
		})

		// Test map with uint16 keys
		t.Run("Uint16Keys", func(t *testing.T) {
			data := struct {
				Uint16Map map[uint16]string
			}{
				Uint16Map: map[uint16]string{
					1: "one",
					2: "two",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "Uint16Map.1")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "one" {
				t.Errorf("Expected 'one', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.Uint16Map[1] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.Uint16Map[1])
			}
		})

		// Test map with uint32 keys
		t.Run("Uint32Keys", func(t *testing.T) {
			data := struct {
				Uint32Map map[uint32]string
			}{
				Uint32Map: map[uint32]string{
					1: "one",
					2: "two",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "Uint32Map.1")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "one" {
				t.Errorf("Expected 'one', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.Uint32Map[1] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.Uint32Map[1])
			}
		})

		// Test map with uint64 keys
		t.Run("Uint64Keys", func(t *testing.T) {
			data := struct {
				Uint64Map map[uint64]string
			}{
				Uint64Map: map[uint64]string{
					1: "one",
					2: "two",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "Uint64Map.1")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "one" {
				t.Errorf("Expected 'one', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.Uint64Map[1] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.Uint64Map[1])
			}
		})

		// Test map with uintptr keys
		t.Run("UintptrKeys", func(t *testing.T) {
			data := struct {
				UintptrMap map[uintptr]string
			}{
				UintptrMap: map[uintptr]string{
					1: "one",
					2: "two",
				},
			}

			accessor, err := GetAccessorDot[string](&data, "UintptrMap.1")
			if err != nil {
				t.Fatalf("Failed to get accessor: %v", err)
			}

			value := accessor.Get()
			if value != "one" {
				t.Errorf("Expected 'one', got %s", value)
			}

			// Test setting a value
			err = accessor.Set("updated")
			if err != nil {
				t.Fatalf("Failed to set value: %v", err)
			}

			if data.UintptrMap[1] != "updated" {
				t.Errorf("Expected 'updated', got %s", data.UintptrMap[1])
			}
		})

		// Test invalid conversions
		t.Run("InvalidConversions", func(t *testing.T) {
			data := struct {
				IntMap map[int]string
			}{
				IntMap: map[int]string{
					1: "one",
					2: "two",
				},
			}

			// Try to access with non-numeric string
			_, err := GetAccessorDot[string](&data, "IntMap.abc")
			if err == nil {
				t.Error("Expected error for non-numeric key, got nil")
			}

			// Try to access with invalid bool string
			_, err = GetAccessorDot[string](&data, "IntMap.maybe")
			if err == nil {
				t.Error("Expected error for invalid key, got nil")
			}
		})
	})

	// Test invalid map access
	t.Run("TestInvalidMapAccess", func(t *testing.T) {

		t.Run("TestInvalidMapBoolAccess", func(t *testing.T) {
			data := struct {
				BoolMap map[bool]string
			}{
				BoolMap: map[bool]string{
					true:  "one",
					false: "two",
				},
			}

			// Test accessing non-existent key
			_, err := GetAccessorDot[string](&data, "BoolMap.maybe")
			if err == nil {
				t.Error("Expected error for non-existent key, got nil")
			}
		})

		t.Run("TestInvalidMapIntAccess", func(t *testing.T) {
			data := struct {
				IntMap map[int]string
			}{
				IntMap: map[int]string{
					1: "one",
				},
			}

			// Test accessing non-existent key
			_, err := GetAccessorDot[string](&data, "IntMap.2")
			if err == nil {
				t.Error("Expected error for non-existent key, got nil")
			}

			// Test accessing non-map field
			_, err = GetAccessorDot[string](&data, "IntMap.abc")
			if err == nil {
				t.Error("Expected error for non-map field, got nil")
			}
		})

		t.Run("TestInvalidMapUIntAccess", func(t *testing.T) {
			data := struct {
				UintMap map[uint]string
			}{
				UintMap: map[uint]string{
					1: "one",
				},
			}

			// Test accessing non-existent key
			_, err := GetAccessorDot[string](&data, "UintMap.2")
			if err == nil {
				t.Error("Expected error for non-existent key, got nil")
			}

			// Test accessing non-map field
			_, err = GetAccessorDot[string](&data, "UintMap.abc")
			if err == nil {
				t.Error("Expected error for non-map field, got nil")
			}
		})

		t.Run("TestInvalidMapObjectAccess", func(t *testing.T) {
			type Key struct {
				Alpha int
				Beta  int
			}
			data := struct {
				ObjMap map[Key]string
			}{
				ObjMap: map[Key]string{
					{Alpha: 1, Beta: 2}: "one",
				},
			}

			// Test accessing unsupported key type
			_, err := GetAccessorDot[string](&data, "ObjMap.maybe")
			if err == nil {
				t.Error("Expected error for unsupported key type, got nil")
			}
		})
	})

	// Test pointer to map of pointers
	t.Run("TestPointerToMapOfPointers", func(t *testing.T) {
		one := 1
		two := 2
		data := struct {
			PtrMap *map[string]*int
		}{
			PtrMap: &map[string]*int{
				"one": &one,
				"two": &two,
			},
		}

		accessor, err := GetAccessorDot[int](&data, "PtrMap.one")
		if err != nil {
			t.Fatalf("Failed to get accessor: %v", err)
		}

		value := accessor.Get()
		if value != 1 {
			t.Errorf("Expected 1, got %d", value)
		}

		err = accessor.Set(2)
		if err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}

		if *(*data.PtrMap)["one"] != 2 {
			t.Errorf("Expected 2, got %d", *(*data.PtrMap)["one"])
		}
	})
}
