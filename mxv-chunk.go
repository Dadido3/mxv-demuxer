// Copyright (c) 2022 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Identifier64 [8]byte

var (
	MXLIST32 = Identifier64{'M', 'X', 'L', 'I', 'S', 'T', '3', '2'}
	MXLIST64 = Identifier64{'M', 'X', 'L', 'I', 'S', 'T', '6', '4'}
	MXRIFF64 = Identifier64{'M', 'X', 'R', 'I', 'F', 'F', '6', '4'}

	MXJVAF64 = Identifier64{'M', 'X', 'J', 'V', 'A', 'F', '6', '4'}
	MXJVCO64 = Identifier64{'M', 'X', 'J', 'V', 'C', 'O', '6', '4'}
	MXJVFL64 = Identifier64{'M', 'X', 'J', 'V', 'F', 'L', '6', '4'}
	MXJVFT64 = Identifier64{'M', 'X', 'J', 'V', 'F', 'T', '6', '4'}
	MXJVH264 = Identifier64{'M', 'X', 'J', 'V', 'H', '2', '6', '4'}
	MXJVHD64 = Identifier64{'M', 'X', 'J', 'V', 'H', 'D', '6', '4'}
	MXJVPD64 = Identifier64{'M', 'X', 'J', 'V', 'P', 'D', '6', '4'}
	MXJVVF64 = Identifier64{'M', 'X', 'J', 'V', 'V', 'F', '6', '4'}
	MXWFMT64 = Identifier64{'M', 'X', 'W', 'F', 'M', 'T', '6', '4'}
)

type Chunk64Header struct {
	Identifier Identifier64 // Type of chunk.
	Length     uint64       // Length of content in bytes.
}

type Chunk64 interface {
	UnmarshalChunk(reader io.Reader, length uint64) error
}

// ParseChunk64 reads and parses a single chunk from the reader.
func ParseChunk64(reader io.Reader) (Chunk64, error) {
	var chunkHeader Chunk64Header
	if err := binary.Read(reader, binary.LittleEndian, &chunkHeader); err != nil {
		return nil, fmt.Errorf("failed to read chunk header: %w", err)
	}

	var chunk Chunk64

	switch chunkHeader.Identifier {
	case MXLIST32, MXLIST64, MXRIFF64:
		var ContentType Identifier64
		if err := binary.Read(reader, binary.LittleEndian, &ContentType); err != nil {
			return nil, fmt.Errorf("failed to read content type: %w", err)
		}

		switch chunkHeader.Identifier {
		case MXLIST32:
			chunk = &ChunkMXLIST32{ContentType: ContentType}
		case MXLIST64:
			chunk = &ChunkMXLIST64{ContentType: ContentType}
		case MXRIFF64:
			chunk = &ChunkMXRIFF64{ContentType: ContentType}
		default:
			return nil, fmt.Errorf("unknown chunk identifier %q", chunkHeader.Identifier)
		}

	default:

		switch chunkHeader.Identifier {
		case MXJVAF64:
			chunk = &ChunkMXJVAF64{}
		case MXJVCO64:
			chunk = &ChunkMXJVCO64{}
		case MXJVFT64:
			chunk = &ChunkMXJVFT64{}
		case MXJVH264:
			chunk = &ChunkMXJVH264{}
		case MXJVHD64:
			chunk = &ChunkMXJVHD64{}
		case MXJVPD64:
			chunk = &ChunkMXJVPD64{}
		case MXJVVF64:
			chunk = &ChunkMXJVVF64{}
		case MXWFMT64:
			chunk = &ChunkMXWFMT64{}
		default:
			return nil, fmt.Errorf("unknown chunk identifier %q", chunkHeader.Identifier)
		}
	}

	// Unmarshal chunk, this will be recursive for lists and riff chunks.
	if err := chunk.UnmarshalChunk(io.LimitReader(reader, int64(chunkHeader.Length)), chunkHeader.Length); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into chunk of type %T: %w", chunk, err)
	}

	return chunk, nil
}

// ParseChunk64All reads and parses all chunks from the reader.
func ParseChunk64All(reader io.Reader) ([]Chunk64, error) {
	var chunks []Chunk64

	for {
		chunk, err := ParseChunk64(reader)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

type ChunkMXLIST32 struct {
	ContentType Identifier64 // Defines the type of the container content. Only used when the Identifier is "MXRIFF64", "MXLIST64" or "MXLIST32". will be ignored in the total length.
	//Chunks      []Chunk32
}

func (c *ChunkMXLIST32) UnmarshalChunk(reader io.Reader, length uint64) error {
	// Read the data, don't do anything with it, yet.
	buf := make([]byte, length)
	_, err := reader.Read(buf)
	return err
}

type ChunkMXLIST64 struct {
	ContentType Identifier64 // Defines the type of the container content. Only used when the Identifier is "MXRIFF64", "MXLIST64" or "MXLIST32". will be ignored in the total length.
	Chunks      []Chunk64
}

func (c *ChunkMXLIST64) UnmarshalChunk(reader io.Reader, length uint64) error {
	chunks, err := ParseChunk64All(reader)
	if err != nil {
		return err
	}

	c.Chunks = chunks
	return nil
}

type ChunkMXRIFF64 struct {
	ContentType Identifier64 // Defines the type of the container content. Only used when the Identifier is "MXRIFF64", "MXLIST64" or "MXLIST32". will be ignored in the total length.
	Chunks      []Chunk64
}

func (c *ChunkMXRIFF64) UnmarshalChunk(reader io.Reader, length uint64) error {
	chunks, err := ParseChunk64All(reader)
	if err != nil {
		return err
	}

	c.Chunks = chunks
	return nil
}

// ChunkMXJVAF64 is an audio frame containing raw audio data.
type ChunkMXJVAF64 struct {
	ChannelBitDepth uint32
	StartSample     uint64
	Samples         uint32
	Data            []byte
}

func (c *ChunkMXJVAF64) UnmarshalChunk(reader io.Reader, length uint64) error {
	if err := binary.Read(reader, binary.LittleEndian, &c.ChannelBitDepth); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.LittleEndian, &c.StartSample); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.LittleEndian, &c.Samples); err != nil {
		return err
	}

	c.Data = make([]byte, length-4-8-4)
	return binary.Read(reader, binary.LittleEndian, &c.Data)
}

// ChunkMXJVCO64 contains unknown data.
type ChunkMXJVCO64 struct {
	Data [24]byte
}

func (c *ChunkMXJVCO64) UnmarshalChunk(reader io.Reader, length uint64) error {
	return binary.Read(reader, binary.LittleEndian, c)
}

// ChunkMXJVFT64 contains a list of references to audio/video frames.
// This maps a frame number to a file offset.
type ChunkMXJVFT64 struct {
	FileOffsets []uint64
}

func (c *ChunkMXJVFT64) UnmarshalChunk(reader io.Reader, length uint64) error {
	c.FileOffsets = make([]uint64, length/8)

	return binary.Read(reader, binary.LittleEndian, &c.FileOffsets)
}

type ChunkMXJVH264 struct {
	Unknown1             uint32
	Unknown2             uint32
	SeekTableFileOffset  uint64 // File offset of ChunkMXJVFT64.
	Frames               uint64
	SeekTableMaxReadSize uint32 // If you read this much data of an offset (that you got from the seek table), you should be able to get at least one full audio and/or video frame. Not really needed, only if you really want to reduce the amount of disk seeks.
	Unknown3             uint32
	Unknown4             uint64
	Unknown5             uint64
	FrameWidth           uint32
	FrameHeight          uint32
	FrameWidth2          uint32 // Maybe second width for anamorphic video?
	FrameHeight2         uint32 // Maybe second height for anamorphic video?
	Unknown6             uint32
	MaxJPEGDataSize      uint32 // Not really known.
	Unknown7             uint64 // Has something to do with audio.
	MaxAudioChunkSize    uint64 // Probably?
	Unknown8             [16]byte
	AudioSampleCounter   uint64 // Contains a bit more samples than there are. (VideoFrames+1) * SampleRate.
}

func (c *ChunkMXJVH264) UnmarshalChunk(reader io.Reader, length uint64) error {
	err := binary.Read(reader, binary.LittleEndian, c)

	// Ignore unexpected end of (chunk). Older formats may have fewer fields.
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return nil
	}
	return err
}

type ChunkMXJVHD64 struct {
	Unknown1             uint32
	Unknown2             uint32
	SeekTableFileOffset  uint64 // File offset of ChunkMXJVFT64.
	Frames               uint64
	SeekTableMaxReadSize uint32 // If you read this much data of an offset (that you got from the seek table), you should be able to get at least one full audio and/or video frame. Not really needed, only if you really want to reduce the amount of disk seeks.
	Unknown3             uint32
	Unknown4             uint64
	Unknown5             uint64
	FrameWidth           uint32
	FrameHeight          uint32
	FrameWidth2          uint32
	FrameHeight2         uint32
	Unknown6             uint32
	MaxJPEGDataSize      uint32 // Not really known.
}

func (c *ChunkMXJVHD64) UnmarshalChunk(reader io.Reader, length uint64) error {
	err := binary.Read(reader, binary.LittleEndian, c)

	// Ignore unexpected end of (chunk). Older formats may have fewer fields.
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return nil
	}
	return err
}

type ChunkMXJVPD64 struct {
	Data [20]byte
}

func (c *ChunkMXJVPD64) UnmarshalChunk(reader io.Reader, length uint64) error {
	return binary.Read(reader, binary.LittleEndian, c)
}

type ChunkMXJVVF64 struct {
	JPEGData []byte
}

func (c *ChunkMXJVVF64) UnmarshalChunk(reader io.Reader, length uint64) error {
	c.JPEGData = make([]byte, length)
	return binary.Read(reader, binary.LittleEndian, &c.JPEGData)
}

type ChunkMXWFMT64 struct {
	Tracks          uint16
	Channels        uint16
	SampleRate      uint32 // May contains really strange sample rates, like 47996.
	ByteRate        uint32
	BytesPerSample  uint16
	ChannelBitDepth uint32
}

func (c *ChunkMXWFMT64) UnmarshalChunk(reader io.Reader, length uint64) error {
	return binary.Read(reader, binary.LittleEndian, c)
}
