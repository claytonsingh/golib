package peekbuffer

import "io"

// PeekBuffer is a custom reader that wraps an existing io.Reader and provides peeking capability.
// It allows looking ahead in the input stream without consuming the data. Key features:
//
// 1. Implements io.Reader interface for compatibility with standard readers.
// 2. Provides a Peek method to inspect upcoming data without advancing the read position.
// 3. Prioritizes returning peeked data before reading from the underlying reader.
// 4. Efficiently manages an internal buffer for storing peeked data, growing as needed.
// 5. Handles cases where less data is available than requested during Peek operations.
//
// This structure is useful for scenarios requiring examination of upcoming data to make
// processing decisions, such as detecting file types or parsing structured data streams.
type PeekBuffer struct {
	io.Reader
	reader io.Reader
	buffer []byte
}

// NewPeekBuffer creates and returns a new PeekBuffer instance that wraps the provided reader.
// It initializes the PeekBuffer with an empty buffer.
//
// Parameters:
//   - reader io.Reader: The underlying reader to wrap.
//
// Returns:
//   - *PeekBuffer: A new PeekBuffer instance.
func NewPeekBuffer(reader io.Reader) *PeekBuffer {
	return &PeekBuffer{
		reader: reader,
	}
}

// Read implements the io.Reader interface.
// It first returns any data in the buffer before reading from the wrapped reader.
//
// Parameters:
//   - p []byte: The slice to read data into.
//
// Returns:
//   - n int: The number of bytes read.
//   - err error: Any error encountered during reading, or io.EOF if the end of the stream is reached.
func (this *PeekBuffer) Read(p []byte) (n int, err error) {
	if len(this.buffer) > 0 {
		n := copy(p, this.buffer)
		this.buffer = this.buffer[n:]
		return n, nil
	} else {
		return this.reader.Read(p)
	}
}

// Peek allows looking ahead in the stream without consuming the data.
// It attempts to return up to 'size' bytes from the stream, buffering them if necessary.
// If less than 'size' bytes are available, it returns as much as possible.
// The returned slice is only valid until the next Read operation.
//
// Parameters:
//   - size int: The number of bytes to peek ahead.
//
// Returns:
//   - []byte: A slice containing the peeked data. May be shorter than 'size' if the wapped stream has less data then requested.
//   - error: Any error encountered during peeking, or nil if successful.
func (this *PeekBuffer) Peek(size int) ([]byte, error) {
	var err error
	need := size - len(this.buffer)
	if need > 0 {
		var n int
		buf := make([]byte, need)
		n, err = io.ReadFull(this.reader, buf)
		if n > 0 {
			this.buffer = append(this.buffer, buf[:n]...)
		}
	}

	have := len(this.buffer)
	if size < have {
		have = size
	}

	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return this.buffer[:have], err
	}
	return this.buffer[:have], nil
}
