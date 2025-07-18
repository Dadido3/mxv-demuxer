// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64

import (
	"encoding/binary"
	"fmt"
)

// MAGIX Video Frame Table Entry.
type Chunk32VFTE struct {
	*Accessor

	// Assume all unknown Chunk64 chunks have a header size of 8 byte.
	Header struct {
		DataLength int32
	}

	Data Chunk32VFTEData
}

type Chunk32VFTEData struct {
	VideoFrameChunkOffset int64  // The file offset of the MXJVVF64 chunk that contains the video frame.
	VideoFrameChunkSize   uint32 // The size of the MXJVVF64 chunk that contains the video frame.
}

// Returns the identifier of the chunk.
func (c *Chunk32VFTE) Identifier() Identifier32 {
	return Identifier32{'V', 'F', 'T', 'E'}
}

// Returns the total length of the chunk, including headers and such.
func (c *Chunk32VFTE) Length() int32 {
	return 4 + 4 + c.Header.DataLength
}

// Parses the data from "a" and returns a Chunk32 that can be used to further inspect the chunk content.
// The seek position of "a" needs to be at the length field, as the identifier is already read and parsed.
//
// This function doesn't need to parse anything beside the chunk header.
// Which enables quick iteration over chunks without storing or parsing any unnecessary data.
//
// Internal: This will be called by ReadChunk32 and should only be used to create new instances of chunk objects.
func (*Chunk32VFTE) BuildChunk(a *Accessor) (Chunk32, error) {
	if a == nil {
		return nil, fmt.Errorf("accessor is nil")
	}

	c := &Chunk32VFTE{Accessor: a}

	if err := binary.Read(c, binary.LittleEndian, &c.Header); err != nil {
		return nil, fmt.Errorf("failed to read header of %q chunk: %w", c.Identifier(), err)
	}

	dataStartOffset := c.Accessor.Pos

	if err := binary.Read(c, binary.LittleEndian, &c.Data); err != nil {
		return nil, fmt.Errorf("failed to read data of %q chunk: %w", c.Identifier(), err)
	}

	readBytes := c.Accessor.Pos - dataStartOffset
	if readBytes != int64(c.Header.DataLength) {
		return nil, fmt.Errorf("unexpected data length in header of %T. Got %d bytes, but expect %d bytes", c, c.Header.DataLength, readBytes)
	}

	return c, nil
}

func init() {
	MustRegisterChunk32(&Chunk32VFTE{})
}
