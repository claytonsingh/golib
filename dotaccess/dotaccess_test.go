package dotaccess

import (
	"bytes"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"
)

// Test structures
type Address struct {
	Street  string
	City    string
	ZipCode string
}

type Person struct {
	Name       string
	Age        int
	Address    Address
	Tags       []string
	Scores     map[string]int
	Contact    *Contact
	FriendIDs  *[]int             // Pointer to slice
	Friends    []*Person          // Slice of pointers
	Properties *map[string]string // Pointer to map
	Metadata   map[string]*string // Map of pointers
	Manager    **Person           // Double pointer
}

type Contact struct {
	Email     string
	Phone     string
	Preferred bool
}

func TestGetAccessor_Struct(t *testing.T) {
	person := Person{
		Name: "John Doe",
		Age:  30,
		Address: Address{
			Street:  "123 Main St",
			City:    "Anytown",
			ZipCode: "12345",
		},
		Tags:   []string{"customer", "premium"},
		Scores: map[string]int{"math": 95, "science": 88},
		Contact: &Contact{
			Email:     "john@example.com",
			Phone:     "555-1234",
			Preferred: true,
		},
	}

	// Test simple field access
	accessor1, err := GetAccessorDot[string](&person, "Name")
	if err != nil {
		t.Fatalf("Failed to get accessor for Name: %v", err)
	}
	if accessor1.Get() != "John Doe" {
		t.Errorf("Expected Name to be 'John Doe', got '%v'", accessor1.Get())
	}

	// Test nested struct field
	accessor2, err := GetAccessorDot[string](&person, "Address.Street")
	if err != nil {
		t.Fatalf("Failed to get accessor for Address.Street: %v", err)
	}
	if accessor2.Get() != "123 Main St" {
		t.Errorf("Expected Address.Street to be '123 Main St', got '%v'", accessor2.Get())
	}

	// Test slice access
	accessor3, err := GetAccessorDot[string](&person, "Tags.1")
	if err != nil {
		t.Fatalf("Failed to get accessor for Tags.1: %v", err)
	}
	if accessor3.Get() != "premium" {
		t.Errorf("Expected Tags.1 to be 'premium', got '%v'", accessor3.Get())
	}

	// Test map access
	accessor4, err := GetAccessorDot[int](&person, "Scores.math")
	if err != nil {
		t.Fatalf("Failed to get accessor for Scores.math: %v", err)
	}
	if accessor4.Get() != 95 {
		t.Errorf("Expected Scores.math to be 95, got '%v'", accessor4.Get())
	}

	// Test pointer dereferencing
	accessor5, err := GetAccessorDot[string](&person, "Contact.Email")
	if err != nil {
		t.Fatalf("Failed to get accessor for Contact.Email: %v", err)
	}
	if accessor5.Get() != "john@example.com" {
		t.Errorf("Expected Contact.Email to be 'john@example.com', got '%v'", accessor5.Get())
	}
}

func TestGetAccessor_Set(t *testing.T) {
	// Use a pointer to person since we need to modify it
	person := &Person{
		Name: "John Doe",
		Age:  30,
		Address: Address{
			Street:  "123 Main St",
			City:    "Anytown",
			ZipCode: "12345",
		},
		Tags:   []string{"customer", "premium"},
		Scores: map[string]int{"math": 95, "science": 88},
		Contact: &Contact{
			Email:     "john@example.com",
			Phone:     "555-1234",
			Preferred: true,
		},
	}

	// Test setting a simple field
	accessor1, err := GetAccessorDot[string](person, "Name")
	if err != nil {
		t.Fatalf("Failed to get accessor for Name: %v", err)
	}
	err = accessor1.Set("Jane Doe")
	if err != nil {
		t.Fatalf("Failed to set Name: %v", err)
	}
	if person.Name != "Jane Doe" {
		t.Errorf("Expected Name to be 'Jane Doe', got '%v'", person.Name)
	}

	// Test setting a nested struct field
	accessor2, err := GetAccessorDot[string](person, "Address.City")
	if err != nil {
		t.Fatalf("Failed to get accessor for Address.City: %v", err)
	}
	err = accessor2.Set("New City")
	if err != nil {
		t.Fatalf("Failed to set Address.City: %v", err)
	}
	if person.Address.City != "New City" {
		t.Errorf("Expected Address.City to be 'New City', got '%v'", person.Address.City)
	}

	// Test setting a slice element
	accessor3, err := GetAccessorDot[string](person, "Tags.0")
	if err != nil {
		t.Fatalf("Failed to get accessor for Tags.0: %v", err)
	}
	err = accessor3.Set("vip")
	if err != nil {
		t.Fatalf("Failed to set Tags.0: %v", err)
	}
	if person.Tags[0] != "vip" {
		t.Errorf("Expected Tags.0 to be 'vip', got '%v'", person.Tags[0])
	}

	// Test setting a field through a pointer
	accessor4, err := GetAccessorDot[bool](person, "Contact.Preferred")
	if err != nil {
		t.Fatalf("Failed to get accessor for Contact.Preferred: %v", err)
	}
	err = accessor4.Set(false)
	if err != nil {
		t.Fatalf("Failed to set Contact.Preferred: %v", err)
	}
	if person.Contact.Preferred != false {
		t.Errorf("Expected Contact.Preferred to be false, got '%v'", person.Contact.Preferred)
	}

	// Test getting and setting nil values
	accessor5, err := GetAccessorDot[*Contact](person, "Contact")
	if err != nil {
		t.Fatalf("Failed to get accessor for Contact: %v", err)
	}

	// First verify we can get the non-nil value
	initialContact := accessor5.Get()
	if initialContact == nil {
		t.Errorf("Expected Contact to be non-nil initially")
	}

	// Set the field to nil
	err = accessor5.Set(nil)
	if err != nil {
		t.Fatalf("Failed to set Contact to nil: %v", err)
	}

	// Verify the field is now nil
	if person.Contact != nil {
		t.Errorf("Expected Contact to be nil after Set(nil), got %v", person.Contact)
	}

	// Verify nil is returned from Get
	nilContact := accessor5.Get()
	if nilContact != nil {
		t.Errorf("Expected Get() to return nil for Contact, got %v", nilContact)
	}

	// Set it back to a non-nil value
	newContact := &Contact{
		Email:     "new@example.com",
		Phone:     "555-5678",
		Preferred: true,
	}
	err = accessor5.Set(newContact)
	if err != nil {
		t.Fatalf("Failed to set Contact to new value: %v", err)
	}

	// Verify the field has the new value
	if person.Contact != newContact {
		t.Errorf("Expected Contact to be set to new value")
	}
}

func TestGetAccessor_Errors(t *testing.T) {
	person := Person{
		Name: "John Doe",
		Age:  30,
		Tags: []string{"customer"},
	}

	// Test invalid field
	_, err := GetAccessorDot[string](&person, "InvalidField")
	if err == nil {
		t.Error("Expected error for invalid field, got nil")
	}

	// Test invalid path
	_, err = GetAccessorDot[string](&person, "")
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}

	// Test nil object
	var nilPerson *Person
	_, err = GetAccessorDot[string](nilPerson, "something")
	if err == nil {
		t.Error("Expected error for nil object, got nil")
	}

	// Test out of bounds slice access
	_, err = GetAccessorDot[string](&person, "Tags.5")
	if err == nil {
		t.Error("Expected error for out of bounds slice access, got nil")
	}

	// Test invalid slice index
	_, err = GetAccessorDot[string](&person, "Tags.abc")
	if err == nil {
		t.Error("Expected error for invalid slice index, got nil")
	}

	// Test type mismatch on set
	accessor, _ := GetAccessorDot[int](&person, "Age")
	err = accessor.Set(42) // Now type safe, can't pass string
	if err != nil {
		t.Errorf("Unexpected error for type-safe set: %v", err)
	}

	// Test type mismatch at compile time by intentionally using wrong type
	// This won't compile if uncommented:
	// accessor.Set("not an int")

	// Test asking for wrong type in accessor
	_, err = GetAccessorDot[string](&person, "Age")
	if err == nil {
		t.Error("Expected error for type mismatch in accessor, got nil")
	}
}

func TestGetAccessor_ComplexPath(t *testing.T) {
	// Create a complex nested structure
	data := map[string]any{
		"users": []map[string]any{
			{
				"name": "Alice",
				"accounts": map[string]any{
					"savings": map[string]any{
						"balance": 1000.50,
					},
					"checking": map[string]any{
						"balance": 250.75,
					},
				},
			},
			{
				"name": "Bob",
				"accounts": map[string]any{
					"savings": map[string]any{
						"balance": 2500.00,
					},
				},
			},
		},
	}

	// Test complex path access
	accessor1, err := GetAccessorDot[float64](&data, "users.0.accounts.savings.balance")
	if err != nil {
		t.Fatalf("Failed to get accessor for users.0.accounts.savings.balance: %v", err)
	}

	balance := accessor1.Get()
	if !reflect.DeepEqual(balance, 1000.50) {
		t.Errorf("Expected balance to be 1000.50, got %v", balance)
	}

	// Test setting a value in a complex path
	accessor2, err := GetAccessorDot[string](&data, "users.1.name")
	if err != nil {
		t.Fatalf("Failed to get accessor for users.1.name: %v", err)
	}
	err = accessor2.Set("Robert")
	if err != nil {
		t.Fatalf("Failed to set users.1.name: %v", err)
	}

	users := data["users"].([]map[string]any)
	if users[1]["name"] != "Robert" {
		t.Errorf("Expected users.1.name to be 'Robert', got '%v'", users[1]["name"])
	}
}

func TestGetAccessor_ComplexPointers(t *testing.T) {
	// Setup test data with complex pointer types
	friendIDs := []int{101, 102, 103}

	emailAlice := "alice@example.com"
	emailBob := "bob@example.com"

	properties := map[string]string{
		"hair":   "brown",
		"eyes":   "blue",
		"height": "180cm",
	}

	// Prepare test data
	bob := &Person{
		Name: "Bob Smith",
		Age:  28,
	}

	jane := &Person{
		Name: "Jane Doe",
	}

	manager := &bob // Double pointer scenario

	alice := Person{
		Name:       "Alice Johnson",
		Age:        35,
		FriendIDs:  &friendIDs,           // Pointer to slice
		Friends:    []*Person{bob, jane}, // Slice of pointers
		Properties: &properties,          // Pointer to map
		Metadata: map[string]*string{ // Map of pointers
			"email":     &emailAlice,
			"alt_email": &emailBob,
		},
		Manager: manager, // Double pointer
		Contact: &Contact{
			Email: "alice@example.com",
		},
	}

	// Test pointer to slice
	accessor1, err := GetAccessorDot[int](&alice, "FriendIDs.1")
	if err != nil {
		t.Fatalf("Failed to get accessor for FriendIDs.1: %v", err)
	}
	if accessor1.Get() != 102 {
		t.Errorf("Expected FriendIDs.1 to be 102, got '%v'", accessor1.Get())
	}

	// Test setting value in pointer to slice
	accessor2, err := GetAccessorDot[int](&alice, "FriendIDs.0")
	if err != nil {
		t.Fatalf("Failed to get accessor for FriendIDs.0: %v", err)
	}
	err = accessor2.Set(999)
	if err != nil {
		t.Fatalf("Failed to set FriendIDs.0: %v", err)
	}
	if (*alice.FriendIDs)[0] != 999 {
		t.Errorf("Expected FriendIDs.0 to be 999, got '%v'", (*alice.FriendIDs)[0])
	}

	// Test slice of pointers
	accessor3, err := GetAccessorDot[string](&alice, "Friends.0.Name")
	if err != nil {
		t.Fatalf("Failed to get accessor for Friends.0.Name: %v", err)
	}
	if accessor3.Get() != "Bob Smith" {
		t.Errorf("Expected Friends.0.Name to be 'Bob Smith', got '%v'", accessor3.Get())
	}

	// Test setting value in slice of pointers
	accessor4, err := GetAccessorDot[int](&alice, "Friends.1.Age")
	if err != nil {
		t.Fatalf("Failed to get accessor for Friends.1.Age: %v", err)
	}
	err = accessor4.Set(32)
	if err != nil {
		t.Fatalf("Failed to set Friends.1.Age: %v", err)
	}
	if alice.Friends[1].Age != 32 {
		t.Errorf("Expected Friends.1.Age to be 32, got '%v'", alice.Friends[1].Age)
	}

	// Test pointer to map
	accessor5, err := GetAccessorDot[string](&alice, "Properties.eyes")
	if err != nil {
		t.Fatalf("Failed to get accessor for Properties.eyes: %v", err)
	}
	if accessor5.Get() != "blue" {
		t.Errorf("Expected Properties.eyes to be 'blue', got '%v'", accessor5.Get())
	}

	// Test setting value in pointer to map
	accessor6, err := GetAccessorDot[string](&alice, "Properties.hair")
	if err != nil {
		t.Fatalf("Failed to get accessor for Properties.hair: %v", err)
	}
	err = accessor6.Set("blonde")
	if err != nil {
		t.Fatalf("Failed to set Properties.hair: %v", err)
	}
	if (*alice.Properties)["hair"] != "blonde" {
		t.Errorf("Expected Properties.hair to be 'blonde', got '%v'", (*alice.Properties)["hair"])
	}

	// Test map of pointers
	accessor7, err := GetAccessorDot[*string](&alice, "Metadata.email")
	if err != nil {
		t.Fatalf("Failed to get accessor for Metadata.email: %v", err)
	}
	if *accessor7.Get() != "alice@example.com" {
		t.Errorf("Expected Metadata.email to be 'alice@example.com', got '%v'", *accessor7.Get())
	}

	// Test setting value in map of pointers
	newEmail := "new_bob@example.com"
	accessor8, err := GetAccessorDot[*string](&alice, "Metadata.alt_email")
	if err != nil {
		t.Fatalf("Failed to get accessor for Metadata.alt_email: %v", err)
	}
	err = accessor8.Set(&newEmail)
	if err != nil {
		t.Fatalf("Failed to set Metadata.alt_email: %v", err)
	}
	if *alice.Metadata["alt_email"] != "new_bob@example.com" {
		t.Errorf("Expected Metadata.alt_email to be 'new_bob@example.com', got '%v'", *alice.Metadata["alt_email"])
	}

	// Test double pointer access
	accessor9, err := GetAccessorDot[string](&alice, "Manager.Name")
	if err != nil {
		t.Fatalf("Failed to get accessor for Manager.Name: %v", err)
	}
	if accessor9.Get() != "Bob Smith" {
		t.Errorf("Expected Manager.Name to be 'Bob Smith', got '%v'", accessor9.Get())
	}

	// Test setting through double pointer
	accessor10, err := GetAccessorDot[int](&alice, "Manager.Age")
	if err != nil {
		t.Fatalf("Failed to get accessor for Manager.Age: %v", err)
	}
	err = accessor10.Set(40)
	if err != nil {
		t.Fatalf("Failed to set Manager.Age: %v", err)
	}
	if (*alice.Manager).Age != 40 {
		t.Errorf("Expected Manager.Age to be 40, got '%v'", (*alice.Manager).Age)
	}
	// We should also verify bob.Age changed since it's the same object
	if bob.Age != 40 {
		t.Errorf("Expected bob.Age to also be 40, got '%v'", bob.Age)
	}

	// Test double pointer -> value (multiple levels of dereferencing)
	t.Run("double pointer -> value", func(t *testing.T) {
		accessor, err := GetAccessorDot[int](&alice, "Manager.Age")
		if err != nil {
			t.Fatalf("Failed to get accessor for Manager.Age: %v", err)
		}
		if accessor.Get() != 40 {
			t.Errorf("Expected Manager.Age to be 28, got '%v'", accessor.Get())
		}

		// Test setting through multiple levels of pointers
		err = accessor.Set(45)
		if err != nil {
			t.Fatalf("Failed to set Manager.Age: %v", err)
		}

		// Verify the change propagated through the double pointer
		if (*alice.Manager).Age != 45 {
			t.Errorf("Expected (*Manager).Age to be 45, got '%v'", (*alice.Manager).Age)
		}
		if bob.Age != 45 {
			t.Errorf("Expected bob.Age to be 45, got '%v'", bob.Age)
		}
	})

	// Test getting a pointer to value (value to pointer)
	t.Run("value to pointer", func(t *testing.T) {
		// Get the pointer to the Name field
		ptrAccessor, err := GetAccessorDot[*string](&alice, "Name")
		if err != nil {
			t.Fatalf("Failed to get pointer accessor for &Name: %v", err)
		}

		namePtr := ptrAccessor.Get()
		if *namePtr != "Alice Johnson" {
			t.Errorf("Expected *namePtr to be 'Alice Johnson', got '%v'", *namePtr)
		}

		// Modify through the obtained pointer
		*namePtr = "Alice Smith"

		// Verify the original was modified
		if alice.Name != "Alice Smith" {
			t.Errorf("Expected alice.Name to be 'Alice Smith', got '%v'", alice.Name)
		}
	})

	// Test pointer to double pointer - getting **Contact from *Contact field
	t.Run("pointer to double pointer", func(t *testing.T) {
		// Get a double pointer to the Contact field (which is already *Contact)
		contactPtrAccessor, err := UnsafeGetAccessorDot[**Contact](&alice, "Contact")
		if err != nil {
			t.Fatalf("Failed to get double pointer accessor for Contact: %v", err)
		}

		contactPtr := contactPtrAccessor.Get()

		// Verify we got the right double pointer
		if (**contactPtr).Email != "alice@example.com" {
			t.Errorf("Expected (**contactPtr).Email to be 'alice@example.com', got '%v'", (**contactPtr).Email)
		}

		// Create a new Contact to set
		newContact := &Contact{
			Email:     "alice.new@example.com",
			Phone:     "555-9876",
			Preferred: false,
		}

		// Set through the accessor
		err = contactPtrAccessor.Set(&newContact)
		if err != nil {
			t.Fatalf("Failed to set Contact: %v", err)
		}

		// Verify the contact was changed
		if alice.Contact.Email != "alice.new@example.com" {
			t.Errorf("Expected new Contact.Email to be 'alice.new@example.com', got '%v'", alice.Contact.Email)
		}

		// Verify the original pointer chain was maintained
		if (**contactPtr).Email != "alice.new@example.com" {
			t.Errorf("Expected (**contactPtr).Email to be updated to 'alice.new@example.com', got '%v'", (**contactPtr).Email)
		}
	})

	// The existing "pointer to double pointer" test could be renamed to be clearer
	t.Run("double pointer to double pointer", func(t *testing.T) {
		// Get the pointer to the Manager field (which is already **Person)
		ptrAccessor, err := GetAccessorDot[**Person](&alice, "Manager")
		if err != nil {
			t.Fatalf("Failed to get accessor for Manager: %v", err)
		}

		managerPtr := ptrAccessor.Get()

		// Verify we got the right double pointer
		if (**managerPtr).Name != "Bob Smith" {
			t.Errorf("Expected (**managerPtr).Name to be 'Bob Smith', got '%v'", (**managerPtr).Name)
		}

		// Create a new person to set as manager
		charlie := &Person{
			Name: "Charlie Brown",
			Age:  50,
		}
		newManager := &charlie

		// Set through the accessor
		err = ptrAccessor.Set(newManager)
		if err != nil {
			t.Fatalf("Failed to set Manager: %v", err)
		}

		// Verify the manager was changed
		if (*alice.Manager).Name != "Charlie Brown" {
			t.Errorf("Expected new Manager.Name to be 'Charlie Brown', got '%v'", (*alice.Manager).Name)
		}

		// Verify the original pointer chain was maintained
		if (**managerPtr).Name != "Bob Smith" {
			t.Errorf("Expected (**managerPtr).Name to be 'Bob Smith', got '%v'", (**managerPtr).Name)
		}
	})
}

func TestUnexportedFields(t *testing.T) {
	// Test case 1: bytes.Buffer has unexported 'buf' field
	t.Run("bytes.Buffer", func(t *testing.T) {
		// Create a buffer with some content
		buf := bytes.NewBuffer([]byte("hello"))

		// Access the unexported 'buf' field
		accessor, err := UnsafeGetAccessorDot[[]byte](buf, "buf")
		if err != nil {
			t.Fatalf("Failed to get accessor for bytes.Buffer.buf: %v", err)
		}

		// Get the current value
		bufSlice := accessor.Get()
		if string(bufSlice[:5]) != "hello" {
			t.Errorf("Expected 'hello', got '%s'", string(bufSlice[:5]))
		}

		// Set a new value
		newBuf := []byte("world")
		err = accessor.Set(newBuf)
		if err != nil {
			t.Fatalf("Failed to set bytes.Buffer.buf: %v", err)
		}

		// Verify the change worked
		if buf.String() != "world" {
			t.Errorf("Expected 'world', got '%s'", buf.String())
		}
	})

	// Test case 2: time.Time has unexported fields like 'sec', 'nsec', etc.
	t.Run("time.Time", func(t *testing.T) {
		tm := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		// Access the unexported 'wall' field
		accessor, err := UnsafeGetAccessorDot[uint64](&tm, "wall")
		if err != nil {
			t.Fatalf("Failed to get accessor for time.Time.wall: %v", err)
		}

		// Get the current value
		wallValue := accessor.Get()
		t.Logf("time.Time.wall type: %T, value: %v", wallValue, wallValue)

		// We won't try to modify this as time.Time's internals are complex
		// and highly implementation-dependent
	})

	// Test case 3: sync.Mutex has unexported 'state' field
	t.Run("sync.Mutex", func(t *testing.T) {
		var mu sync.Mutex

		// Lock it so state isn't zero
		mu.Lock()

		// Access the unexported 'state' field
		accessor, err := UnsafeGetAccessorDot[int32](&mu, "state")
		if err != nil {
			t.Fatalf("Failed to get accessor for sync.Mutex.state: %v", err)
		}

		// Get the current value while locked
		lockedState := accessor.Get()
		t.Logf("Locked mutex state type: %T, value: %v", lockedState, lockedState)

		// Verify non-zero while locked
		if lockedState == 0 {
			t.Error("Expected non-zero state value while locked")
		}

		// Unlock and check state changed
		mu.Unlock()

		// Get state after unlock
		unlockedState := accessor.Get()
		t.Logf("Unlocked mutex state type: %T, value: %v", unlockedState, unlockedState)

		// Verify state changed after unlock
		if unlockedState == lockedState {
			t.Error("Expected mutex state to change after unlock")
		}
	})
}

func TestNestedUnexportedFields(t *testing.T) {
	// Testing with http.Transport's dialsInProgress.headPos field
	t.Run("http.Transport.dialsInProgress.headPos", func(t *testing.T) {
		// Create an http.Transport
		transport := &http.Transport{}

		// First, let's verify we can access the dialsInProgress field
		// Using any as the type since the internal structure varies across Go versions
		dialsAccessor, err := UnsafeGetAccessorDot[any](&transport, "dialsInProgress")
		if err != nil {
			t.Fatalf("Failed to get accessor for transport.dialsInProgress: %v", err)
		}

		dialsValue := dialsAccessor.Get()
		t.Logf("Initial dialsInProgress value: %v (type: %T)", dialsValue, dialsValue)

		// Now try to access the nested headPos field
		headPosAccessor, err := UnsafeGetAccessorDot[int](&transport, "dialsInProgress.headPos")
		if err != nil {
			t.Fatalf("Failed to get accessor for transport.dialsInProgress.headPos: %v", err)
		}

		// Get the current headPos value
		initialHeadPos := headPosAccessor.Get()
		t.Logf("Initial headPos value: %v (type: %T)", initialHeadPos, initialHeadPos)

		// Try to modify the headPos value
		err = headPosAccessor.Set(42)
		if err != nil {
			t.Fatalf("Failed to set transport.dialsInProgress.headPos: %v", err)
		}

		// Get the updated value to verify it changed
		updatedHeadPos := headPosAccessor.Get()
		t.Logf("Updated headPos value: %v (type: %T)", updatedHeadPos, updatedHeadPos)

		// Verify the value changed as expected
		if updatedHeadPos != 42 {
			t.Errorf("Expected headPos to be 42 after update, got %v", updatedHeadPos)
		}

		// Reset to original value for cleanliness
		err = headPosAccessor.Set(initialHeadPos)
		if err != nil {
			t.Fatalf("Failed to reset transport.dialsInProgress.headPos: %v", err)
		}
	})
}

func TestMultiLevelPointerIndirection(t *testing.T) {
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
}

func TestPointerIndirectionSyntax(t *testing.T) {
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
}

func TestAccessorAny(t *testing.T) {
	// Test accessing a field as 'any' type
	t.Run("access field as any type", func(t *testing.T) {
		person := Person{
			Name: "John Doe",
			Age:  30,
		}

		var unknown any = &person

		// Get accessor for Name field as any type
		accessor, err := GetAccessorDot[any](&unknown, "Name")
		if err != nil {
			t.Fatalf("Failed to get accessor for Name as any: %v", err)
		}

		// Verify we can get the value
		value := accessor.Get()
		if value != "John Doe" {
			t.Errorf("Expected value to be 'John Doe', got '%v'", value)
		}

		// Test setting a new value
		err = accessor.Set("Jane Smith")
		if err != nil {
			t.Fatalf("Failed to set Name: %v", err)
		}

		// Verify the value was updated
		if person.Name != "Jane Smith" {
			t.Errorf("Expected Name to be 'Jane Smith', got '%s'", person.Name)
		}

		// Verify we can get the updated value through the accessor
		updatedValue := accessor.Get()
		if updatedValue != "Jane Smith" {
			t.Errorf("Expected accessor to return 'Jane Smith', got '%v'", updatedValue)
		}
	})
}
