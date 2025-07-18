// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64

import (
	"encoding/binary"
	"fmt"
)

type Chunk32 interface {
	Identifier() Identifier32 // Returns the identifier of the chunk.
	Length() int32            // Returns the total length of the chunk, including headers and such.
}

// ReadChunk32 parses the chunk from "a" and returns a Chunk32 that can be used to further inspect the chunk content.
func (a *Accessor) ReadChunk32() (Chunk32, error) {
	var id Identifier32

	if err := binary.Read(a, binary.LittleEndian, &id); err != nil {
		return nil, fmt.Errorf("failed to read Chunk32 identifier: %w", err)
	}

	if chunk, ok := chunk32Registry[id]; ok {
		return chunk.BuildChunk(a)
	}

	// Fall back to dummy chunk, as we want to support reading containers with unknown chunk identifiers.
	dummyChunk := &Chunk32Dummy{}
	return dummyChunk.BuildChunk(a, id)
}

type Chunk32Builder interface {
	Chunk32

	// Parses the data from "a" and returns a Chunk32 that can be used to further inspect the chunk content.
	// The seek position of "a" needs to be at the length field, as the identifier is already read and parsed.
	//
	// This function doesn't need to parse anything beside the chunk header.
	// Which enables quick iteration over chunks without storing or parsing any unnecessary data.
	//
	// Internal: This will be called by ReadChunk32 and should only be used to create new instances of chunk objects.
	BuildChunk(a *Accessor) (Chunk32, error)
}

var chunk32Registry = map[Identifier32]Chunk32Builder{}

// RegisterChunk32 adds the given Chunk32 to the registry.
func RegisterChunk32(c Chunk32Builder) error {
	id := c.Identifier()

	if _, found := chunk32Registry[id]; found {
		return fmt.Errorf("Chunk32 with %s already exists", id)
	}

	chunk32Registry[id] = c

	return nil
}

// MustRegisterChunk32 is similar to RegisterChunk32, but it panics on any error.
func MustRegisterChunk32(c Chunk32Builder) {
	if err := RegisterChunk32(c); err != nil {
		panic(err)
	}
}
