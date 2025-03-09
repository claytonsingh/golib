package dotaccess

import (
	"reflect"
	"testing"
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

	t.Run("pointer to slice", func(t *testing.T) {

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
	})

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
