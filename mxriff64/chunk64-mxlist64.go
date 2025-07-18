// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64

import (
	"encoding/binary"
	"fmt"
	"io"
	"iter"
)

type Chunk64MXLIST64 struct {
	*Accessor

	Header struct {
		DataLength  int64
		ContentType ContentType // Type of data that is stored in this list container.
	}

	dataStartOffset int64 // File offset where the chunk data starts. This is the beginning of the sub-chunk list.
}

// Returns the identifier of the chunk.
func (c *Chunk64MXLIST64) Identifier() Identifier64 {
	return Identifier64{'M', 'X', 'L', 'I', 'S', 'T', '6', '4'}
}

// Returns the total length of the chunk, including headers and such.
func (c *Chunk64MXLIST64) Length() int64 {
	return 8 + 8 + 8 + c.Header.DataLength
}

// Chunks returns an iterator listing all sub-chunks.
//
// Any error is returned as the second value.
// In case there is an error, the iteration will stop.
func (c *Chunk64MXLIST64) Chunks() iter.Seq2[Chunk64, error] {
	return func(yield func(Chunk64, error) bool) {
		chunkPos := c.dataStartOffset // The file offset where the current chunk starts.
		for chunkPos < c.dataStartOffset+int64(c.Header.DataLength) {
			if _, err := c.Accessor.Seek(chunkPos, io.SeekStart); err != nil {
				yield(nil, fmt.Errorf("failed to seek to the start of the sub-chunk list: %w", err))
				return
			}

			sc, err := c.Accessor.ReadChunk64()
			if err != nil {
				yield(nil, fmt.Errorf("failed to read chunk: %w", err))
				return
			}

			if chunkPos+sc.Length() > c.dataStartOffset+int64(c.Header.DataLength) {
				yield(nil, fmt.Errorf("the sub-chunk goes beyond the parent chunk"))
				return
			}

			if !yield(sc, nil) {
				return
			}

			chunkPos += sc.Length()
		}
	}
}

// Parses the data from "a" and returns a Chunk64 that can be used to further inspect the chunk content.
// The seek position of "a" needs to be at the length field, as the identifier is already read and parsed.
//
// This function doesn't need to parse anything beside the chunk header.
// Which enables quick iteration over chunks without storing or parsing any unnecessary data.
//
// Internal: This will be called by ReadChunk64 and should only be used to create new instances of chunk objects.
func (*Chunk64MXLIST64) BuildChunk(a *Accessor) (Chunk64, error) {
	if a == nil {
		return nil, fmt.Errorf("accessor is nil")
	}

	c := &Chunk64MXLIST64{Accessor: a}

	if err := binary.Read(c, binary.LittleEndian, &c.Header); err != nil {
		return nil, fmt.Errorf("failed to read header of %q chunk: %w", c.Identifier(), err)
	}

	c.dataStartOffset = c.Accessor.Pos

	return c, nil
}

func init() {
	MustRegisterChunk64(&Chunk64MXLIST64{})
}
