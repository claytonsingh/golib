package dotaccess

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// MaxDereferenceDepth limits consecutive pointer/interface dereferences to prevent infinite loops.
// Linked lists (struct->pointer->struct) are unaffected. This global safety limit applies to all
// GetAccessor functions and can be adjusted if needed.
var MaxDereferenceDepth int = 100

// FieldAccessor provides type-safe access to get and set values in nested structures
type FieldAccessor[T any] struct {
	target       reflect.Value
	key          reflect.Value // Key in map or index in slice (as reflect.Value)
	fieldType    FieldType     // Type of the field (regular, map element, slice element, etc.)
	isUnexported bool          // Track if field is unexported
	ptrDepthDiff int           // Difference between target and source pointer depths (target - source)
}

// FieldType represents the type of field being accessed
type FieldType int

const (
	FieldTypeRegular FieldType = iota
	FieldTypeMapElement
	FieldTypeSliceElement
)

// Get retrieves the current value of the field with type safety
func (a *FieldAccessor[T]) Get() T {
	// Get the val of the field
	var val reflect.Value
	switch a.fieldType {
	case FieldTypeMapElement:
		val = a.target.MapIndex(a.key)
	case FieldTypeSliceElement:
		val = a.target.Index(int(a.key.Int()))
	default:
		val = a.target
	}

	// Target has more pointers than source - need to add pointers
	if a.ptrDepthDiff > 0 {
		// Map elements cannot be addressed
		if a.fieldType == FieldTypeMapElement {
			// This should never happen as it is checked in the getAccessor function
			panic("cannot get pointer to map element")
		}

		// Add the required number of pointers
		for i := 0; i < a.ptrDepthDiff; i++ {
			// Create a pointer to the value
			val = reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr()))
		}
	}

	// Source has more pointers than target - need to dereference
	// Dereference until we match the target pointer depth
	for i := 0; i < -a.ptrDepthDiff; i++ {
		if val.IsNil() {
			panic("cannot dereference nil pointer")
		}
		val = val.Elem()
	}

	// Handle unexported fields
	if a.isUnexported {
		return reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem().Interface().(T)
	} else {
		return val.Interface().(T)
	}
}

// Set attempts to set a new value to the field with type safety
func (a *FieldAccessor[T]) Set(value T) error {

	val := reflect.ValueOf(value)

	// Target needs more pointers than source has
	for i := 0; i < -a.ptrDepthDiff; i++ {
		// Create a new value that points to the original
		newValue := reflect.New(val.Type())
		newValue.Elem().Set(val)
		val = newValue
	}

	// Target needs fewer pointers than source has - dereference
	for i := 0; i < a.ptrDepthDiff; i++ {
		if val.IsNil() {
			return errors.New("cannot dereference nil pointer")
		}
		val = val.Elem()
	}

	// Handle different field types
	switch a.fieldType {
	case FieldTypeMapElement:
		// Handle map elements
		a.target.SetMapIndex(a.key, val)

	case FieldTypeSliceElement:
		// Handle slice elements
		index := int(a.key.Int())
		a.target.Index(index).Set(val)

	default:
		// Handle regular fields (exported or unexported)
		if a.isUnexported {
			reflect.NewAt(a.target.Type(), unsafe.Pointer(a.target.UnsafeAddr())).Elem().Set(val)
		} else {
			a.target.Set(val)
		}
	}

	return nil
}

// GetAccessorDot returns a type-safe accessor for the field at the given path
// Path is in dot notation, e.g. "person.address.street"
// Supports navigating through structs, maps, slices, and pointers
func GetAccessorDot[T any, U any](obj *U, path string) (*FieldAccessor[T], error) {
	parts := strings.Split(path, ".")
	// Grab the last part of the path and count the number of pointers in it
	// Remove the pointers from the last part of the path
	lastPart := parts[len(parts)-1]
	finalDereference := strings.Count(lastPart, "*")
	parts[len(parts)-1] = strings.ReplaceAll(lastPart, "*", "")
	return getAccessor[T](obj, parts, false, finalDereference)
}

// UnsafeGetAccessorDot returns a type-safe accessor for the field at the given path
// Path is in dot notation, e.g. "person.address.street"
// Supports navigating through structs, maps, slices, and pointers
// Unsafe accessor will return the value as is, even if the field is unexported
func UnsafeGetAccessorDot[T any, U any](obj *U, path string) (*FieldAccessor[T], error) {
	parts := strings.Split(path, ".")
	// Grab the last part of the path and count the number of pointers in it
	// Remove the pointers from the last part of the path
	lastPart := parts[len(parts)-1]
	finalDereference := strings.Count(lastPart, "*")
	parts[len(parts)-1] = strings.ReplaceAll(lastPart, "*", "")
	return getAccessor[T](obj, parts, true, finalDereference)
}

// GetAccessor returns a type-safe accessor for the field at the given path
// Path is in dot notation, e.g. "person.address.street"
// Supports navigating through structs, maps, slices, and pointers
func GetAccessor[T any, U any](obj *U, path []string, finalDereference int) (*FieldAccessor[T], error) {
	return getAccessor[T](obj, path, false, finalDereference)
}

// UnsafeGetAccessor returns a type-safe accessor for the field at the given path
// Path is in dot notation, e.g. "person.address.street"
// Supports navigating through structs, maps, slices, and pointers
// Unsafe accessor will return the value as is, even if the field is unexported
func UnsafeGetAccessor[T any, U any](obj *U, path []string, finalDereference int) (*FieldAccessor[T], error) {
	return getAccessor[T](obj, path, true, finalDereference)
}

// getAccessor is a helper function that returns a type-safe accessor for the field at the given path
// If unsafe is true, the accessor will return the value as is, even if the field is unexported
func getAccessor[T any, U any](obj *U, path []string, allowUnexported bool, finalDereference int) (*FieldAccessor[T], error) {
	if obj == nil {
		return nil, errors.New("object is nil")
	}

	if len(path) == 0 {
		return nil, errors.New("path is empty")
	}

	// Initialize our fa accessor with the starting value
	var parent reflect.Value
	a := FieldAccessor[T]{
		target:    reflect.ValueOf(obj),
		fieldType: FieldTypeRegular,
	}

	for i, part := range path {

		if part == "" {
			return nil, errors.New("path is empty at index " + strconv.Itoa(i))
		}

		maxDereferences := MaxDereferenceDepth
		for {
			if maxDereferences <= 0 {
				return nil, fmt.Errorf("dereference depth limit of %d reached at '%s'", MaxDereferenceDepth, strings.Join(path[:i+1], "."))
			}
			maxDereferences--

			// Dereference pointers and interfaces
			kind := a.target.Kind()
			if kind == reflect.Ptr || kind == reflect.Interface {
				if a.target.IsNil() {
					return nil, fmt.Errorf("nil pointer at '%s'", strings.Join(path[:i+1], "."))
				}
				a.target = a.target.Elem()
				continue
			}
			break
		}

		switch a.target.Kind() {
		case reflect.Struct:
			fieldValue := a.target.FieldByName(part)
			if !fieldValue.IsValid() {
				return nil, fmt.Errorf("no field named '%s' in struct at '%s'", part, strings.Join(path[:i], "."))
			}

			// Check if the field is unexported
			canAddr := fieldValue.CanAddr()
			canSet := fieldValue.CanSet()
			parent = a.target
			a = FieldAccessor[T]{
				target:       fieldValue,
				isUnexported: !canSet && canAddr,
				fieldType:    FieldTypeRegular,
			}

			if a.isUnexported && !allowUnexported {
				return nil, fmt.Errorf("cannot access field '%s' of %s type at '%s'", part, parent.Type().String(), strings.Join(path[:i], "."))
			}

		// String, Bool, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr
		case reflect.Map:
			keyValue := reflect.ValueOf(part)
			if !keyValue.Type().AssignableTo(a.target.Type().Key()) {
				// Try to convert string key to the map's key type
				keyType := a.target.Type().Key()
				switch keyType.Kind() {
				case reflect.Bool:
					// Convert string to bool
					b, err := strconv.ParseBool(part)
					if err != nil {
						return nil, fmt.Errorf("cannot convert key '%s' to bool for map at '%s'", part, strings.Join(path[:i], "."))
					}
					keyValue = reflect.ValueOf(b)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					// Convert string to any integer type
					index, err := strconv.ParseInt(part, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("cannot convert key '%s' to %s for map at '%s'", part, keyType.Kind(), strings.Join(path[:i], "."))
					}
					// Convert to the specific int type
					keyValue = reflect.ValueOf(index).Convert(keyType)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					// Convert string to any unsigned integer type
					index, err := strconv.ParseUint(part, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("cannot convert key '%s' to %s for map at '%s'", part, keyType.Kind(), strings.Join(path[:i], "."))
					}
					// Convert to the specific uint type
					keyValue = reflect.ValueOf(index).Convert(keyType)
				default:
					return nil, fmt.Errorf("incompatible key type for map at '%s', cannot convert string to %s", strings.Join(path[:i], "."), keyType.Kind())
				}
			}

			mapValue := a.target.MapIndex(keyValue)
			if !mapValue.IsValid() {
				return nil, fmt.Errorf("key '%s' not found in map at '%s'", part, strings.Join(path[:i], "."))
			}

			// Remember the map and key for setting values later
			parent = a.target
			a = FieldAccessor[T]{
				target:    mapValue,
				key:       keyValue,
				fieldType: FieldTypeMapElement,
			}

		case reflect.Slice, reflect.Array:
			index, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid slice/array index '%s' at '%s'", part, strings.Join(path[:i], "."))
			}
			if index < 0 || index >= a.target.Len() {
				return nil, fmt.Errorf("index %d out of range for slice/array at '%s'", index, strings.Join(path[:i], "."))
			}

			// Remember the slice and index for setting values later
			parent = a.target
			a = FieldAccessor[T]{
				target:    a.target.Index(index),
				key:       reflect.ValueOf(index),
				fieldType: FieldTypeSliceElement,
			}

		default:
			return nil, fmt.Errorf("cannot access field '%s' of %s type at '%s'", part, a.target.Type().String(), strings.Join(path[:i], "."))
		}
	}

	// Dereference the value the required number of times
	for i := 0; i < finalDereference; i++ {
		if a.target.IsNil() {
			return nil, fmt.Errorf("nil pointer at '%s'", strings.Join(path, "."))
		}
		parent = a.target
		a = FieldAccessor[T]{
			target:    a.target.Elem(),
			fieldType: FieldTypeRegular,
		}
	}

	// Type validation - ensure the final value can be converted to the requested type T
	if a.target.IsValid() {
		// Get the type of T
		var zero [0]T
		targetType := reflect.TypeOf(zero).Elem()
		sourceType := a.target.Type()

		// Check pointer depths and determine base types (without pointers)
		targetBaseType, targetPtrDepth := getPointerTypeAndDepth(targetType)
		sourceBaseType, sourcePtrDepth := getPointerTypeAndDepth(sourceType)
		a.ptrDepthDiff = targetPtrDepth - sourcePtrDepth

		if sourceBaseType.AssignableTo(targetBaseType) || a.target.Kind() == reflect.Interface {
			if a.ptrDepthDiff > 0 {
				// Target has more pointers than source
				if a.ptrDepthDiff > 1 {
					// Cannot add multiple levels of pointers
					return nil, fmt.Errorf("cannot add multiple levels of pointers: source has %d, target requires %d", sourcePtrDepth, targetPtrDepth)
				}

				// T -> *T conversion - make sure we can make a pointer to it
				if a.fieldType == FieldTypeMapElement {
					return nil, fmt.Errorf("cannot request a pointer to a map element: found %s but requested %s", sourceType, targetType)
				}

				if !a.target.CanAddr() {
					return nil, fmt.Errorf("type mismatch: found %s but requested %s", sourceType, targetType)
				}
			}
			// Other cases (equal pointer depth or target has fewer pointers) are handled fine
		} else {
			return nil, fmt.Errorf("type mismatch: found %s but requested %s", sourceType, targetType)
		}
	}

	// If we are accessing a map or slice element, we access the parent
	if a.fieldType == FieldTypeMapElement || a.fieldType == FieldTypeSliceElement {
		a.target = parent
	}

	return &a, nil
}

// getPointerTypeAndDepth returns the base type (without pointers) and the number of pointer indirections
func getPointerTypeAndDepth(t reflect.Type) (reflect.Type, int) {
	depth := 0
	for t.Kind() == reflect.Ptr {
		depth++
		t = t.Elem()
	}
	return t, depth
}
