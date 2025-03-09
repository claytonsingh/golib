# dotaccess
A type-safe library for accessing and modifying deeply nested fields in Go structures using dot notation.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Basic Usage](#basic-usage)
  - [Path Notation](#path-notation)
  - [Pointer Handling](#pointer-handling)
  - [Accessing Unexported Fields](#accessing-unexported-fields)
  - [Error Handling](#error-handling)
- [API Reference](#api-reference)
- [Contributing](#contributing)

## Overview
The dotaccess library works with:
- Nested structs
- Maps
- Slices and arrays
- Pointers (including multi-level pointers)
- Unexported fields (via unsafe mode)

## Features
- Type-safe access and modification using Go's generics
- Dot notation for easy nested field access
- Support for complex nested data structures
- Pointer handling with automatic dereferencing
- Safe handling of nil pointers
- Optional unsafe mode for accessing unexported fields

## Installation
```
go get github.com/claytonsingh/golib/dotaccess
```

## Usage

### Path Notation

Dot notation provides a simple string-based syntax for accessing nested fields. The path is expressed as field names separated by dots:

- **Simple field access**: `"Name"` accesses the top-level field named "Name"
- **Nested struct fields**: `"Address.Street"` accesses the "Street" field within the "Address" struct
- **Map values**: `"Tags.role"` accesses the value with key "role" in the "Tags" map. Only string keys are supported.
- **Slice/array elements**: `"Scores.0"` accesses the first element (index 0) of the "Scores" slice
- **Combined paths**: `"Friends.0.Address.City"` can navigate through slices and nested structs
- **Pointer dereferencing**: `"Contact*"` explicitly dereferences a pointer field. Multiple asterisks can be used for multi-level pointers (e.g., `"Manager**"`)


### Basic Usage

The library handles type conversion and validation automatically based on the generic type parameter used when creating an accessor.

**Note**: The field is bound to the accessor during the call to GetAccessor / GetAccessorDot. If you need to access the same field on different objects, create a multiple accessors.

```go
type Address struct {
	Street string
	City   string
}

type Person struct {
	Name    string
	Age     int
	Address *Address
	Tags    map[string]string
	Scores  []int
}

func main() {
	// Create a sample structure
	p := &Person{
		Name: "John Doe",
		Age:  30,
		Address: &Address{
			Street: "123 Main St",
			City:   "Anytown",
		},
		Tags:   map[string]string{"role": "developer"},
		Scores: []int{95, 80, 90},
	}

	// Access a simple field
	nameAccessor, _ := dotaccess.GetAccessorDot[string](p, "Name")
	fmt.Println("Name:", nameAccessor.Get())

	// Access a nested field
	cityAccessor, _ := dotaccess.GetAccessorDot[string](p, "Address.City")
	fmt.Println("City:", cityAccessor.Get())

	// Access a map value
	roleAccessor, _ := dotaccess.GetAccessorDot[string](p, "Tags.role")
	fmt.Println("Role:", roleAccessor.Get())

	// Access a slice element
	scoreAccessor, _ := dotaccess.GetAccessorDot[int](p, "Scores.0")
	fmt.Println("First Score:", scoreAccessor.Get())

	// Modify a field
	cityAccessor.Set("New City")
	fmt.Println("Updated City:", p.Address.City)
}
```

### Pointer Handling

The library handles pointers intelligently, allowing you to work with pointer fields in a type-safe manner. When accessing fields that are pointers, you have several options:

1. **Automatic Dereferencing**: By default, when you access a pointer field using `GetAccessorDot`, the library will automatically dereference the pointer when you call Get and Set to provide the underlying value.

2. **Pointer Notation**: You can use the `*` suffix at on the last element of the path to explicitly dereference a pointer. For example, `"Name*"` will dereference the pointer at the `Name` field. If you have double pointers or more, you can use multiple `*` suffixes. For example, `"Name**"` will dereference the pointer at the `Name` field twice. The pointer is bound to the accessor instead of the field.

3. **Working with Pointers Directly**: If you need to work with the pointers directly (rather than its value), you can specify the pointer type in the generic parameter. This will return a pointer to the value rather than the value itself. Note that you can not get a pointer to a map value.

```go
type User struct {
	Name *string
}

name := "John Doe"
user := &User{
	Name: &name,
}

// Name: 0xc00008c290 John Doe
fmt.Println("Name:", user.Name, *user.Name)

// Access through pointer is automatically handled
nameAccessor1, _ := dotaccess.GetAccessorDot[string](user, "Name")
fmt.Println("Name:", nameAccessor1.Get())

// This is equivalent to
// newName := "new@example.com"
// user.Name = &newName
nameAccessor1.Set("new1@example.com")

// Name: 0xc00008c2f0 new1@example.com
// pointer and name are both updated
fmt.Println("Name:", user.Name, *user.Name)

// Use pointer notation to get pointer to name field on user
nameAccessor2, _ := dotaccess.GetAccessorDot[*string](user, "Name")
namePtr := nameAccessor2.Get()
*namePtr = "new2@example.com"

// Name: 0xc00008c2f0 new2@example.com
// pointer is the same, but the value is updated
fmt.Println("Name:", user.Name, *user.Name)

// Access the pointer, not the field on user
nameAccessor3, _ := dotaccess.GetAccessorDot[string](user, "Name*")
fmt.Println("Name:", nameAccessor3.Get())

// This is equivalent to *user.Name = "new@example.com"
nameAccessor3.Set("new3@example.com")

// Name: 0xc00008c2f0 new3@example.com
// pointer is the same, but the value is updated
fmt.Println("Name:", user.Name, *user.Name)
```

### Accessing Unexported Fields

The library provides "unsafe" versions of accessor methods to access and modify unexported fields:

**Warning**: Accessing unexported fields bypasses Go's visibility rules and can lead to unexpected behavior. Make sure to wear a hardhat and use this feature with caution.

```go
// Access an unexported field in bytes.Buffer
buf := bytes.NewBuffer([]byte("hello"))

// Access the unexported 'buf' field
bufAccessor, _ := dotaccess.UnsafeGetAccessorDot[[]byte](buf, "buf")

// Get the current buffer content
content := bufAccessor.Get()
fmt.Println(string(content))  // Outputs: "hello"

// Modify the unexported field
bufAccessor.Set([]byte("world"))
fmt.Println(buf.String())  // Outputs: "world"
```

### Error Handling

The accessor functions return an error when they encounter problems:

```go
person := &Person{
	Name: "John Doe",
}

// Access a non-existent field
accessor, err := dotaccess.GetAccessorDot[string](person, "NonExistentField")
if err != nil {
	fmt.Println("Error:", err)  // no field named 'NonExistentField'
	return
}

// Access out-of-bounds slice index
scoresAccessor, err := dotaccess.GetAccessorDot[int](person, "Scores.10")
if err != nil {
	fmt.Println("Error:", err)  // invalid slice/array index '10'
	return
}

// Type mismatch
nameAsInt, err := dotaccess.GetAccessorDot[int](person, "Name")
if err != nil {
	fmt.Println("Error:", err)  // type mismatch: found string but requested int
	return
}
```

Common errors include:
- Field not found in struct
- Invalid array/slice index
- Map key not found
- Type mismatch
- Nil pointer dereference
- Unexported field access (when not using Unsafe methods)

## API Reference

*`GetAccessorDot[T, U](obj *U, path string) (*FieldAccessor[T], error)`*

Gets an accessor for a field using dot notation.

*`UnsafeGetAccessorDot[T, U](obj *U, path string) (*FieldAccessor[T], error)`*

Gets an accessor for a field using dot notation. Unsafe accessor can get and set unexported fields.

*`GetAccessor[T, U](obj *U, path []string, finalDereference int) (*FieldAccessor[T], error)`*

Gets an accessor using a path of strings and a final dereference count.

*`UnsafeGetAccessor[T, U](obj *U, path []string, finalDereference int) (*FieldAccessor[T], error)`*

Gets an accessor using a path of strings and a final dereference count. Unsafe accessor can get and set unexported fields.

*`FieldAccessor[T].Get() T`*

Retrieves the current value of the field.

*`FieldAccessor[T].Set(value T) error`*

Sets the value of the field.

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.
