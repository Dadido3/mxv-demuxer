// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64_test

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Dadido3/mxv-demuxer/mxriff64"
)

func TestNewFromReader(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "example-files", "Vergleich2.mxv"))
	if err != nil {
		t.Fatalf("Failed to open file: %v.", err)
	}

	riff := mxriff64.NewFromReader(f)

	c, err := riff.ReadChunk64()
	if err != nil {
		t.Fatalf("Failed to read Chunk64: %v.", err)
	}

	var printChunkTree func(chunk any, level int)
	printChunkTree = func(chunk any, level int) {
		switch chunk := chunk.(type) {
		case *mxriff64.Chunk64MXRIFF64:
			log.Printf("%sChunk %s | Total length: %d bytes | %s.", strings.Repeat("\t", level), chunk.Identifier(), chunk.Length(), chunk.Header.FormType)
			log.Printf("%s  Sub-Chunks:", strings.Repeat("\t", level))
			for sc, err := range chunk.Chunks() {
				if err != nil {
					t.Fatalf("Failed to read sub-chunk: %v.", err)
				}
				printChunkTree(sc, level+1)
			}
		case *mxriff64.Chunk64MXLIST64:
			log.Printf("%sChunk %s | Total length: %d bytes | %s.", strings.Repeat("\t", level), chunk.Identifier(), chunk.Length(), chunk.Header.ContentType)
			log.Printf("%s  Sub-Chunks:", strings.Repeat("\t", level))
			for sc, err := range chunk.Chunks() {
				if err != nil {
					t.Fatalf("Failed to read sub-chunk: %v.", err)
				}
				printChunkTree(sc, level+1)
			}
		case *mxriff64.Chunk64MXLIST32:
			log.Printf("%sChunk %s | Total length: %d bytes | %s.", strings.Repeat("\t", level), chunk.Identifier(), chunk.Length(), chunk.Header.ContentType)
			log.Printf("%s  Sub-Chunks:", strings.Repeat("\t", level))
			for sc, err := range chunk.Chunks() {
				if err != nil {
					t.Fatalf("Failed to read sub-chunk: %v.", err)
				}
				printChunkTree(sc, level+1)
			}
		case mxriff64.Chunk32:
			log.Printf("%sChunk %s | Total length: %d bytes.", strings.Repeat("\t", level), chunk.Identifier(), chunk.Length())
		case mxriff64.Chunk64:
			log.Printf("%sChunk %s | Total length: %d bytes.", strings.Repeat("\t", level), chunk.Identifier(), chunk.Length())
		default:
			t.Fatalf("Invalid object %T passed to printChunkTree.", chunk)
		}
	}

	printChunkTree(c, 0)
}
