package peekbuffer

import (
	"bytes"
	"io"
	"testing"
)

func TestNewPeekBuffer(t *testing.T) {
	reader := bytes.NewReader([]byte("test"))
	pb := NewPeekBuffer(reader)
	if pb == nil {
		t.Fatal("NewPeekBuffer returned nil")
	}
	if pb.reader != reader {
		t.Error("NewPeekBuffer did not set the reader correctly")
	}
}

func TestPeekBuffer_Read(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		readSize int
		want     string
		wantErr  error
	}{
		{"Read all", "hello", 5, "hello", nil},
		{"Read partial", "world", 3, "wor", nil},
		{"Read empty", "", 5, "", io.EOF},
		{"Read more than available", "test", 5, "test", io.ErrUnexpectedEOF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPeekBuffer(bytes.NewReader([]byte(tt.input)))
			buf := make([]byte, tt.readSize)
			_, err := io.ReadFull(pb, buf)

			if err != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == nil && string(buf) != tt.want {
				t.Errorf("Read() got = %v, want %v", string(buf), tt.want)
			}
		})
	}
}

func TestPeekBuffer_Peek(t *testing.T) {
	const input = "hello world"

	type peekTest struct {
		size int
		want string
	}

	tests := []struct {
		name    string
		peeks   []peekTest
		wantErr bool
	}{
		{"Peek all", []peekTest{{11, "hello world"}}, false},
		{"Peek partial", []peekTest{{5, "hello"}}, false},
		{"Peek more than available", []peekTest{{15, "hello world"}}, false},
		{"Peek empty", []peekTest{{0, ""}}, false},
		{"Multiple peeks", []peekTest{{5, "hello"}, {3, "hel"}, {11, "hello world"}}, false},
		{"Increasing peeks", []peekTest{{2, "he"}, {5, "hello"}, {8, "hello wo"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPeekBuffer(bytes.NewReader([]byte(input)))

			for _, peek := range tt.peeks {
				got, err := pb.Peek(peek.size)

				if (err != nil) != tt.wantErr {
					t.Errorf("Peek(%d) error = %v, wantErr %v", peek.size, err, tt.wantErr)
					return
				}
				if string(got) != peek.want {
					t.Errorf("Peek(%d) got = %v, want %v", peek.size, string(got), peek.want)
				}
			}
		})
	}
}

func TestPeekBuffer_PeekThenRead(t *testing.T) {
	input := "hello world"
	pb := NewPeekBuffer(bytes.NewReader([]byte(input)))

	// Peek first 5 bytes
	peeked, err := pb.Peek(5)
	if err != nil {
		t.Fatalf("Peek() error = %v", err)
	}
	if string(peeked) != "hello" {
		t.Errorf("Peek() got = %v, want %v", string(peeked), "hello")
	}

	// Read 6 bytes (including the peeked data)
	buf := make([]byte, 6)
	n, err := io.ReadFull(pb, buf)
	if err != nil {
		t.Fatalf("ReadFull() error = %v", err)
	}
	if n != 6 || string(buf) != "hello " {
		t.Errorf("ReadFull() got = %v, want %v", string(buf), "hello ")
	}

	// Read remaining data
	remaining, err := io.ReadAll(pb)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(remaining) != "world" {
		t.Errorf("ReadAll() got = %v, want %v", string(remaining), "world")
	}
}

func TestPeekBuffer_LargeInput(t *testing.T) {
	largeInput := make([]byte, 1<<20) // 1 MB of data
	for i := range largeInput {
		largeInput[i] = byte(i % 256)
	}

	pb := NewPeekBuffer(bytes.NewReader(largeInput))

	// Peek at the first 1024 bytes
	peeked, err := pb.Peek(1024)
	if err != nil || len(peeked) != 1024 {
		t.Errorf("Large input Peek failed: %v", err)
	}

	// Read the entire input
	read, err := io.ReadAll(pb)
	if err != nil || len(read) != len(largeInput) {
		t.Errorf("Large input Read failed: %v", err)
	}
}

func TestPeekBuffer_MultipleReads(t *testing.T) {
	input := "multiple read test"
	pb := NewPeekBuffer(bytes.NewReader([]byte(input)))

	reads := []struct {
		size int
		want string
	}{
		{5, "multi"},
		{3, "ple"},
		{5, " read"},
		{5, " test"},
	}

	for _, r := range reads {
		buf := make([]byte, r.size)
		_, err := io.ReadFull(pb, buf)
		if err != nil {
			t.Errorf("Multiple Read failed: got %s, want %s, err: %v", string(buf), r.want, err)
			return
		}
		if string(buf) != r.want {
			t.Errorf("Multiple Read failed: got %s, want %s", string(buf), r.want)
		}
	}
}

func TestPeekBuffer_PeekAfterPartialRead(t *testing.T) {
	input := "peek after read"
	pb := NewPeekBuffer(bytes.NewReader([]byte(input)))

	// Read first 5 bytes
	buf := make([]byte, 5)
	_, err := io.ReadFull(pb, buf)
	if err != nil {
		t.Errorf("Initial Read failed: %v", err)
		return
	}
	if string(buf) != "peek " {
		t.Errorf("Initial Read failed: got %s, want %s", string(buf), "peek ")
	}

	// Peek at next 5 bytes
	peeked, err := pb.Peek(5)
	if err != nil || string(peeked) != "after" {
		t.Errorf("Peek after partial Read failed: %v", err)
	}
}

func TestPeekBuffer_ErrorHandling(t *testing.T) {
	errorReader := &ErrorReader{err: io.ErrUnexpectedEOF}
	pb := NewPeekBuffer(errorReader)

	_, err := pb.Peek(5)
	if err != nil && err != io.ErrUnexpectedEOF {
		t.Errorf("Expected no ErrUnexpectedEOF on Peek, got %v", err)
	}

	_, err = pb.Read(make([]byte, 5))
	if err == nil {
		t.Errorf("Expected error on Read, got nil")
	}
}

// ErrorReader is a mock reader that always returns an error
type ErrorReader struct {
	err error
}

func (e *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func TestPeekBuffer_ModifyPeekedData(t *testing.T) {
	input := "modify peeked data"
	pb := NewPeekBuffer(bytes.NewReader([]byte(input)))

	// Peek first 6 bytes
	peeked, err := pb.Peek(6)
	if err != nil {
		t.Fatalf("Peek() error = %v", err)
	}
	if string(peeked) != "modify" {
		t.Errorf("Peek() got = %v, want %v", string(peeked), "modify")
	}

	// Modify the peeked data
	copy(peeked, "MODIFY")

	// Read remaining data
	remaining, err := io.ReadAll(pb)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(remaining) != "MODIFY peeked data" {
		t.Errorf("ReadAll() got = %v, want %v", string(remaining), "MODIFY peeked data")
	}
}

func TestPeekBuffer_ReadByte(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    byte
		wantErr error
	}{
		{"Read first byte", "hello", 'h', nil},
		{"Read first byte", "world", 'w', nil},
		{"Read empty", "", 0, io.EOF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPeekBuffer(bytes.NewReader([]byte(tt.input)))
			got, err := pb.ReadByte()

			if err != tt.wantErr {
				t.Errorf("readBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == nil && got != tt.want {
				t.Errorf("readBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPeekBuffer_PeekByte(t *testing.T) {
	const input = "hello world"

	type peekTest struct {
		offset int
		want   byte
	}

	tests := []struct {
		name    string
		peeks   []peekTest
		wantErr bool
	}{
		{"Peek first byte", []peekTest{{0, 'h'}}, false},
		{"Peek fifth byte", []peekTest{{4, 'o'}}, false},
		{"Peek last byte", []peekTest{{10, 'd'}}, false},
		{"Peek out of bounds", []peekTest{{11, 0}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPeekBuffer(bytes.NewReader([]byte(input)))

			for _, peek := range tt.peeks {
				got, err := pb.PeekByte(peek.offset)

				if (err != nil) != tt.wantErr {
					t.Errorf("PeekByte(%d) error = %v, wantErr %v", peek.offset, err, tt.wantErr)
					return
				}
				if err == nil && got != peek.want {
					t.Errorf("PeekByte(%d) got = %v, want %v", peek.offset, got, peek.want)
				}
			}
		})
	}
}
