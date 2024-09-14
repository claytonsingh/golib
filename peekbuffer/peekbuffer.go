// Package peekbuffer provides a reader with peeking capabilities.
package peekbuffer

import "io"

const FillPeekBufferSize = 4096

// PeekBuffer is a custom reader that wraps an existing io.Reader and provides peeking capability.
// It allows looking ahead in the input stream without consuming the data. Key features:
//
// 1. Implements io.Reader and io.ByteReader interfaces for compatibility with standard readers.
// 2. Provides a Peek and PeekByte method to inspect upcoming data without advancing the read position.
// 3. Prioritizes returning peeked data before reading from the underlying reader.
// 4. Efficiently manages an internal buffer for storing peeked data, growing as needed.
// 5. Handles cases where less data is available than requested during Peek operations.
//
// This structure is useful for scenarios requiring examination of upcoming data to make
// processing decisions, such as detecting file types or parsing structured data streams.
type PeekBuffer struct {
	io.Reader
	io.ByteReader
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
// This method may return fewer bytes than requested, even if the end of the stream hasn't been reached.
//
// Parameters:
//   - p []byte: The slice to read data into.
//
// Returns:
//   - n int: The number of bytes read. This may be less than len(p).
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

// ReadByte implements the io.ByteReader interface.
// It reads and returns a single byte from the buffer if available, or from the underlying reader if the buffer is empty.
//
// Returns:
//   - byte: The byte read.
//   - error: Any error encountered during reading, or io.EOF if the end of the stream is reached.
func (this *PeekBuffer) ReadByte() (byte, error) {
	if len(this.buffer) > 0 {
		b := this.buffer[0]
		this.buffer = this.buffer[1:]
		return b, nil
	} else {
		// Fill the buffer up to MinPeekBufferSize if it's empty
		buf := make([]byte, FillPeekBufferSize)
		n, err := io.ReadAtLeast(this.reader, buf, 1)
		if n > 0 {
			this.buffer = append(this.buffer, buf[:n]...)
			b := this.buffer[0]
			this.buffer = this.buffer[1:]
			return b, nil
		}
		return 0, err
	}
}

// Peek allows looking ahead in the stream without consuming the data.
// It attempts to return up to 'size' bytes from the stream, buffering them if necessary.
// If less than 'size' bytes are available, it returns as much as possible.
// The returned slice is only valid until the next Read operation.
// Note: Modifications to the returned slice will affect subsequent Read operations.
//
// Parameters:
//   - size int: The number of bytes to peek ahead.
//
// Returns:
//   - []byte: A slice containing the peeked data. May be shorter than 'size' if the wrapped stream has less data than requested.
//             Modifying this slice will modify the internal buffer and affect subsequent Read operations.
//   - error: Any error encountered during peeking, or nil if successful.
func (this *PeekBuffer) Peek(size int) ([]byte, error) {
	var err error
	need := size - len(this.buffer)
	if need > 0 {
		// Round up to the next multiple of FillPeekBufferSize
		roundedNeed := ((need + FillPeekBufferSize - 1) / FillPeekBufferSize) * FillPeekBufferSize
		buf := make([]byte, roundedNeed)
		var n int
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

// PeekByte allows looking ahead in the stream at a specific offset without consuming the data.
// It returns the byte at the specified offset if available.
//
// Parameters:
//   - offset int: The offset from the current position to peek at.
//
// Returns:
//   - byte: The byte at the specified offset.
//   - error: Any error encountered during peeking, or io.EOF if the end of the stream is reached.
func (this *PeekBuffer) PeekByte(offset int) (byte, error) {
	peeked, err := this.Peek(offset + 1)
	if err != nil {
		return 0, err
	}
	if offset >= len(peeked) {
		return 0, io.EOF
	}
	return peeked[offset], nil
}
