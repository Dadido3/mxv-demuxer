// Copyright (c) 2022 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"os"
	"testing"
)

func TestParseChunk64(t *testing.T) {
	filename := "mxv-examples/Vergleich2.mxv"

	file, err := os.Open(filename)
	if err != nil {
		t.Errorf("Failed to open file: %v", err)
		return
	}

	chunk, err := ParseChunk64(file)
	if err != nil {
		t.Errorf("Failed to parse root chunk: %v", err)
		return
	}

	chunkMXRIFF64, ok := chunk.(*ChunkMXRIFF64)
	if !ok {
		t.Errorf("File doesn't contain a MXRIFF64 as root chunk")
		return
	}

	if len(chunkMXRIFF64.Chunks) != 8 {
		t.Errorf("File contains wrong amount of chunks. Got %d, expected %d.", len(chunkMXRIFF64.Chunks), 8)
	}
}
