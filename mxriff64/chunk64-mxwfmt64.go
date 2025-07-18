// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64

import (
	"encoding/binary"
	"fmt"
)

// MAGIX JPEG Video Header.
type Chunk64MXWFMT64 struct {
	*Accessor

	Header struct {
		DataLength int64
	}

	Data Chunk64MXWFMT64Data
}

type Chunk64MXWFMT64Data struct {
	Tracks          uint16 // Guess: The number of tracks if there are multiple. But it could also encode the byte format used.
	Channels        uint16 // The number of channels.
	SampleRate      uint32 // Samples per second.
	ByteRate        uint32 // Bytes per second.
	BytesPerSample  uint16 // The number of bytes per sample. (Channels * ChannelBitDepth / 8) or (ByteRate / SampleRate)
	ChannelBitDepth uint32 // Bits per channel per sample.
}

// Returns the identifier of the chunk.
func (c *Chunk64MXWFMT64) Identifier() Identifier64 {
	return Identifier64{'M', 'X', 'W', 'F', 'M', 'T', '6', '4'}
}

// Returns the total length of the chunk, including headers and such.
func (c *Chunk64MXWFMT64) Length() int64 {
	return 8 + 8 + c.Header.DataLength
}

// Parses the data from "a" and returns a Chunk64 that can be used to further inspect the chunk content.
// The seek position of "a" needs to be at the length field, as the identifier is already read and parsed.
//
// This function doesn't need to parse anything beside the chunk header.
// Which enables quick iteration over chunks without storing or parsing any unnecessary data.
//
// Internal: This will be called by ReadChunk64 and should only be used to create new instances of chunk objects.
func (*Chunk64MXWFMT64) BuildChunk(a *Accessor) (Chunk64, error) {
	if a == nil {
		return nil, fmt.Errorf("accessor is nil")
	}

	c := &Chunk64MXWFMT64{Accessor: a}

	if err := binary.Read(c, binary.LittleEndian, &c.Header); err != nil {
		return nil, fmt.Errorf("failed to read header of %q chunk: %w", c.Identifier(), err)
	}

	dataStartOffset := c.Accessor.Pos

	if err := binary.Read(c, binary.LittleEndian, &c.Data); err != nil {
		return nil, fmt.Errorf("failed to read data of %q chunk: %w", c.Identifier(), err)
	}

	readBytes := c.Accessor.Pos - dataStartOffset
	if readBytes != c.Header.DataLength {
		return nil, fmt.Errorf("unexpected data length in header of %T. Got %d bytes, but expect %d bytes", c, c.Header.DataLength, readBytes)
	}

	return c, nil
}

func init() {
	MustRegisterChunk64(&Chunk64MXWFMT64{})
}
