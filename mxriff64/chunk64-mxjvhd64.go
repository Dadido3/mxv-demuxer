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
type Chunk64MXJVHD64 struct {
	*Accessor

	Header struct {
		DataLength int64
	}

	Data Chunk64MXJVHD64Data
}

type Chunk64MXJVHD64Data struct {
	StructSize       uint32 // Seems to be always 112, may be the size of this struct.
	Unknown1         uint32 // Guess: May be some sort of version. (File or encoding software)
	FrameTableOffset uint64 // File offset of Chunk64MXJVFT64.
	VideoFrames      uint64 // The total number of VFTE entries (These point to MXJVVF64 chunks that contain images). It's possible that several VFTE entries point to the same MXJVVF64 chunk.
	MaxReadSize      uint32 // Guess: Seems to be the max of all video and audio frame chunk pairs. Perhaps to tell any software what the max. buffer size needs to be.
	Unknown2         uint32 // Guess: May be some sort of version. (File or encoding software)
	Unknown3         uint64
	Framerate        float64 // Number of frames per second. For interlaced video it will store the number of full frames per second. I.e. PAL has 50 fields/s and therefore 25 full frames/s.
	FrameWidth       uint32
	FrameHeight      uint32
	FrameWidth2      uint32 // Maybe needed when the video is anamorphic? It's the same as the above width in all my test files.
	FrameHeight2     uint32 // Maybe needed when the video is anamorphic? It's the same as the above height in all my test files.
	Flags            uint32
	MaxJPEGSize      uint32 // The JPEG data size of the largest MXJVVF64 chunk. (The size only includes the JPEG data)
}

// Returns the identifier of the chunk.
func (c *Chunk64MXJVHD64) Identifier() Identifier64 {
	return Identifier64{'M', 'X', 'J', 'V', 'H', 'D', '6', '4'}
}

// Returns the total length of the chunk, including headers and such.
func (c *Chunk64MXJVHD64) Length() int64 {
	return 8 + 8 + c.Header.DataLength
}

// Parses the data from "a" and returns a Chunk64 that can be used to further inspect the chunk content.
// The seek position of "a" needs to be at the length field, as the identifier is already read and parsed.
//
// This function doesn't need to parse anything beside the chunk header.
// Which enables quick iteration over chunks without storing or parsing any unnecessary data.
//
// Internal: This will be called by ReadChunk64 and should only be used to create new instances of chunk objects.
func (*Chunk64MXJVHD64) BuildChunk(a *Accessor) (Chunk64, error) {
	if a == nil {
		return nil, fmt.Errorf("accessor is nil")
	}

	c := &Chunk64MXJVHD64{Accessor: a}

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
	MustRegisterChunk64(&Chunk64MXJVHD64{})
}
