// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64

import (
	"encoding/binary"
	"fmt"
)

// MAGIX JPEG Video Header version 2.
type Chunk64MXJVH264 struct {
	*Accessor

	Header struct {
		DataLength int64
	}

	Data Chunk64MXJVH264Data
}

type Chunk64MXJVH264Data struct {
	Chunk64MXJVHD64Data // This is an extension of Chunk64MXJVHD64.

	AudioFrames       uint64      // Number of AFTE entries (These point to MXJVAF64 chunks that contain waveform data)
	MaxAudioChunkSize uint64      // The size of the largest MXJVAF64 chunk. (The size includes the chunk identifier and length field and all its data)
	AspectRatio       float64     // Final image aspect ratio. If this ratio != FrameWidth / FrameHeight the video doesn't have square pixels.
	ColorFormat       ColorFormat // Color format.
	Unknown4          uint32
	AudioSamples      uint64 // Total number of audio samples.
}

// Returns the identifier of the chunk.
func (c *Chunk64MXJVH264) Identifier() Identifier64 {
	return Identifier64{'M', 'X', 'J', 'V', 'H', '2', '6', '4'}
}

// Returns the total length of the chunk, including headers and such.
func (c *Chunk64MXJVH264) Length() int64 {
	return 8 + 8 + c.Header.DataLength
}

// Parses the data from "a" and returns a Chunk64 that can be used to further inspect the chunk content.
// The seek position of "a" needs to be at the length field, as the identifier is already read and parsed.
//
// This function doesn't need to parse anything beside the chunk header.
// Which enables quick iteration over chunks without storing or parsing any unnecessary data.
//
// Internal: This will be called by ReadChunk64 and should only be used to create new instances of chunk objects.
func (*Chunk64MXJVH264) BuildChunk(a *Accessor) (Chunk64, error) {
	if a == nil {
		return nil, fmt.Errorf("accessor is nil")
	}

	c := &Chunk64MXJVH264{Accessor: a}

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
	MustRegisterChunk64(&Chunk64MXJVH264{})
}
