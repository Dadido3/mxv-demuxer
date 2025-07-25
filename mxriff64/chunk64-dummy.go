// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64

import (
	"encoding/binary"
	"fmt"
	"io"
)

// This is a placeholder for unknown chunks.
type Chunk64Dummy struct {
	*Accessor

	ID Identifier64

	// Assume all unknown Chunk64 chunks have a header size of 16 byte.
	Header struct {
		DataLength int64
	}

	dataStartOffset int64 // File offset where the chunk data starts. This is the beginning of the sub-chunk list.
}

// Returns the identifier of the chunk.
func (c *Chunk64Dummy) Identifier() Identifier64 {
	return c.ID
}

// Returns the total length of the chunk, including headers and such.
func (c *Chunk64Dummy) Length() int64 {
	return 8 + 8 + c.Header.DataLength
}

// Returns an io.Reader with the chunk data.
func (c *Chunk64Dummy) DataReader() (io.Reader, error) {
	if _, err := c.Accessor.Seek(c.dataStartOffset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to the start of the chunk data: %w", err)
	}

	return io.LimitReader(c.Accessor, c.Header.DataLength), nil
}

// Parses the data from "a" and returns a Chunk64 that can be used to further inspect the chunk content.
// The seek position of "a" needs to be at the length field, as the identifier is already read and parsed.
//
// This function doesn't need to parse anything beside the chunk header.
// Which enables quick iteration over chunks without storing or parsing any unnecessary data.
//
// Internal: This will be called by ReadChunk64 and should only be used to create new instances of chunk objects.
func (*Chunk64Dummy) BuildChunk(a *Accessor, id Identifier64) (Chunk64, error) {
	if a == nil {
		return nil, fmt.Errorf("accessor is nil")
	}

	c := &Chunk64Dummy{Accessor: a, ID: id}

	if err := binary.Read(c, binary.LittleEndian, &c.Header); err != nil {
		return nil, fmt.Errorf("failed to read header of %q chunk: %w", c.Identifier(), err)
	}

	c.dataStartOffset = c.Accessor.Pos

	return c, nil
}
