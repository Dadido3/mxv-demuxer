// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64

import (
	"encoding/binary"
	"fmt"
)

type Chunk64 interface {
	Identifier() Identifier64 // Returns the identifier of the chunk.
	Length() int64            // Returns the total length of the chunk, including headers and such.
}

// ReadChunk64 parses the chunk from "a" and returns a Chunk64 that can be used to further inspect the chunk content.
func (a *Accessor) ReadChunk64() (Chunk64, error) {
	var id Identifier64

	if err := binary.Read(a, binary.LittleEndian, &id); err != nil {
		return nil, fmt.Errorf("failed to read identifier: %w", err)
	}

	if chunk, ok := chunk64Registry[id]; ok {
		return chunk.BuildChunk(a)
	}

	// Fall back to dummy chunk, as we want to support reading containers with unknown chunk identifiers.
	dummyChunk := &Chunk64Dummy{}
	return dummyChunk.BuildChunk(a, id)
}

type Chunk64Builder interface {
	Chunk64

	// Parses the data from "a" and returns a Chunk64 that can be used to further inspect the chunk content.
	// The seek position of "a" needs to be at the length field, as the identifier is already read and parsed.
	//
	// This function doesn't need to parse anything beside the chunk header.
	// Which enables quick iteration over chunks without storing or parsing any unnecessary data.
	//
	// Internal: This will be called by ReadChunk64 and should only be used to create new instances of chunk objects.
	BuildChunk(a *Accessor) (Chunk64, error)
}

var chunk64Registry = map[Identifier64]Chunk64Builder{}

// RegisterChunk64 adds the given Chunk64 to the registry.
func RegisterChunk64(c Chunk64Builder) error {
	id := c.Identifier()

	if _, found := chunk64Registry[id]; found {
		return fmt.Errorf("Chunk64 with %s already exists", id)
	}

	chunk64Registry[id] = c

	return nil
}

// MustRegisterChunk64 is similar to RegisterChunk64, but it panics on any error.
func MustRegisterChunk64(c Chunk64Builder) {
	if err := RegisterChunk64(c); err != nil {
		panic(err)
	}
}
