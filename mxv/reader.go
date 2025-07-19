// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxv

import (
	"cmp"
	"fmt"
	"io"
	"iter"
	"slices"

	"github.com/Dadido3/mxv-demuxer/mxriff64"
	go_cmp "github.com/google/go-cmp/cmp"
)

type Reader struct {
	accessor *mxriff64.Accessor

	chunkVideoHeader2 *mxriff64.Chunk64MXJVH264
	chunkVideoHeader  *mxriff64.Chunk64MXJVHD64
	chunkWaveFormat   *mxriff64.Chunk64MXWFMT64
	chunkFrameList    *mxriff64.Chunk64MXLIST64
	chunkFrameTable   *mxriff64.Chunk64MXJVFT64
	chunkLookupList   *mxriff64.Chunk64MXLIST32

	// Info is filled by NewReader.
	Info Info

	// Cached list of video frame chunk offsets.
	videoFrameOffsets []mxriff64.Chunk32VFTEData

	// Cached list of audio frame chunk offsets.
	audioFrameOffsets []mxriff64.Chunk32AFTEData
}

// NewReader creates a new reader from the given io.ReadSeeker.
func NewReader(rs io.ReadSeeker) (*Reader, error) {
	r := &Reader{
		accessor: mxriff64.NewFromReadSeeker(rs),
	}

	rootChunk, err := r.accessor.ReadChunk64()
	if err != nil {
		return nil, fmt.Errorf("failed to read root chunk: %w", err)
	}

	mxriffChunk, ok := rootChunk.(*mxriff64.Chunk64MXRIFF64)
	if !ok {
		return nil, fmt.Errorf("invalid root chunk type. Got %T, want %T", rootChunk, mxriffChunk)
	}

	if mxriffChunk.Header.FormType != mxriff64.FormTypeMXJVID64 {
		return nil, fmt.Errorf("unexpected form type. Got %s, want %s", mxriffChunk.Header.FormType, mxriff64.FormTypeMXJVID64)
	}

	for sc, err := range mxriffChunk.Chunks() {
		if err != nil {
			return nil, fmt.Errorf("failed to get sub-chunk from root chunk: %w", err)
		}

		switch sc := sc.(type) {
		case *mxriff64.Chunk64MXJVH264:
			r.chunkVideoHeader2 = sc
		case *mxriff64.Chunk64MXJVHD64:
			r.chunkVideoHeader = sc
		case *mxriff64.Chunk64MXWFMT64:
			r.chunkWaveFormat = sc
		case *mxriff64.Chunk64MXLIST64:
			switch sc.Header.ContentType {
			case mxriff64.ContentTypeMXJVFL64:
				r.chunkFrameList = sc
			}
		case *mxriff64.Chunk64MXJVFT64:
			r.chunkFrameTable = sc
		case *mxriff64.Chunk64MXLIST32:
			switch sc.Header.ContentType {
			case mxriff64.ContentTypeMXJVTL32:
				r.chunkLookupList = sc
			}
		}
	}

	if r.chunkVideoHeader != nil && r.chunkVideoHeader2 != nil {
		if r.chunkVideoHeader.Data != r.chunkVideoHeader2.Data.Chunk64MXJVHD64Data {
			return nil, fmt.Errorf("the two video headers contain contradicting information:\n%s", go_cmp.Diff(r.chunkVideoHeader.Data, r.chunkVideoHeader2.Data.Chunk64MXJVHD64Data))
		}
	}

	// TODO: Fall back to MXJVHD64 if there is no MXJVH264 chunk
	if r.chunkVideoHeader2 != nil {
		r.Info.FrameWidth = r.chunkVideoHeader2.Data.FrameWidth   // Ignore FrameWidth2
		r.Info.FrameHeight = r.chunkVideoHeader2.Data.FrameHeight // Ignore FrameHeight2
		r.Info.Framerate = r.chunkVideoHeader2.Data.Framerate
		r.Info.VideoFrames = r.chunkVideoHeader2.Data.VideoFrames
		r.Info.AspectRatio = r.chunkVideoHeader2.Data.AspectRatio
		r.Info.ColorFormat = r.chunkVideoHeader2.Data.ColorFormat

		r.Info.HasAudio = r.chunkVideoHeader2.Data.Flags&0b00000100 != 0
		r.Info.AudioFrames = r.chunkVideoHeader2.Data.AudioFrames
		r.Info.AudioSamples = r.chunkVideoHeader2.Data.AudioSamples
	} else {
		return nil, fmt.Errorf("couldn't find a MXJVH264 chunk")
	}

	if r.Info.HasAudio {
		if r.chunkWaveFormat != nil {
			if r.chunkWaveFormat.Data.Tracks != 1 {
				return nil, fmt.Errorf("can't handle audio tracks != 1") // It's not even clear if this field stores the number of audio tracks, or some other data.
			}

			r.Info.AudioChannels = r.chunkWaveFormat.Data.Channels
			r.Info.AudioSampleRate = r.chunkWaveFormat.Data.ByteRate / uint32(r.chunkWaveFormat.Data.BytesPerSample) // r.chunkWaveFormat.Data.SampleRate does not seem reliable and can differ slightly.
			r.Info.AudioByteRate = r.chunkWaveFormat.Data.ByteRate
			r.Info.AudioBytesPerSample = r.chunkWaveFormat.Data.BytesPerSample
			r.Info.AudioChannelBitDepth = r.chunkWaveFormat.Data.ChannelBitDepth
		} else {
			return nil, fmt.Errorf("couldn't find MXWFMT64 chunk even though container should have audio data")
		}
	}

	return r, nil
}

// VideoFrames returns an iterator over all video frames.
//
// The frame data can be read by calling VideoFrameData with the frame number returned by this iterator.
//
// To ensure this function succeeds, you have to call `PrepareLookupTable()` first.
func (r *Reader) VideoFrames() iter.Seq2[int, mxriff64.Chunk32VFTEData] {
	return func(yield func(int, mxriff64.Chunk32VFTEData) bool) {
		r.PrepareLookupTable() // Ignore error.

		for frame, vfte := range r.videoFrameOffsets {
			if !yield(frame, vfte) {
				return
			}
		}
	}
}

// VideoFrameData returns a reader to the raw JPEG image data for the given frame.
//
// The range of valid frame numbers is [0...Info.VideoFrames-1].
func (r *Reader) VideoFrameData(frame int) (io.Reader, error) {
	if err := r.PrepareLookupTable(); err != nil {
		return nil, fmt.Errorf("failed to prepare frame chunk lookup table: %w", err)
	}

	if r.videoFrameOffsets == nil {
		return nil, fmt.Errorf("container doesn't contain any video frame chunk lookup entries")
	}

	if frame < 0 || frame >= len(r.videoFrameOffsets) {
		return nil, fmt.Errorf("requested video frame %d is outside of the valid range from %d to %d", frame, 0, len(r.videoFrameOffsets)-1)
	}
	vfte := r.videoFrameOffsets[frame]

	if _, err := r.accessor.Seek(vfte.VideoFrameChunkOffset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to video frame chunk: %w", err)
	}

	chunk, err := r.accessor.ReadChunk64()
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk: %w", err)
	}

	// Check size, but only if the VideoFrameChunkSize field is != 0.
	// VideoFrameChunkSize being zero may be some sort of corruption that occurs in older MXV files.
	if vfte.VideoFrameChunkSize != 0 && chunk.Length() != int64(vfte.VideoFrameChunkSize) {
		return nil, fmt.Errorf("parsed chunk is of wrong size. Got %d bytes, want %d bytes", chunk.Length(), vfte.VideoFrameChunkSize)
	}

	if frameChunk, ok := chunk.(*mxriff64.Chunk64MXJVVF64); !ok {
		return nil, fmt.Errorf("parsed chunk is not a video frame chunk. Got %T, want %T", chunk, frameChunk)
	} else {
		return frameChunk.DataReader()
	}
}

// AudioFrames returns an iterator over all audio frames.
//
// The frame data can be read by calling AudioFrameData with the frame number returned by this iterator.
//
// To ensure this function succeeds, you have to call `PrepareLookupTable()` first.
func (r *Reader) AudioFrames() iter.Seq2[int, mxriff64.Chunk32AFTEData] {
	return func(yield func(int, mxriff64.Chunk32AFTEData) bool) {
		r.PrepareLookupTable() // Ignore error.

		for frame, afte := range r.audioFrameOffsets {
			if !yield(frame, afte) {
				return
			}
		}
	}
}

// AudioFrameFromSample returns the frame number and sample length of the audio frame that contains the given sample.
func (r *Reader) AudioFrameFromSample(sample uint64) (frame int, samples uint32, err error) {
	if err := r.PrepareLookupTable(); err != nil {
		return 0, 0, fmt.Errorf("failed to prepare frame chunk lookup table: %w", err)
	}

	if r.audioFrameOffsets == nil {
		return 0, 0, fmt.Errorf("container doesn't contain any audio frame chunk lookup entries")
	}

	for frame, afte := range r.audioFrameOffsets {
		if sample >= afte.StartSample && sample < afte.StartSample+uint64(afte.Samples) {
			return frame, afte.Samples, nil
		}
	}

	return 0, 0, fmt.Errorf("couldn't find any audio frame that contains sample %d", sample)
}

// AudioFrameData returns a reader to the raw PCM audio data for the given sample range.
//
// The range of valid frame numbers is [0...Info.AudioFrames-1].
func (r *Reader) AudioFrameData(frame int) (reader io.Reader, startSample uint64, samples uint32, err error) {
	if err := r.PrepareLookupTable(); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to prepare frame chunk lookup table: %w", err)
	}

	if r.audioFrameOffsets == nil {
		return nil, 0, 0, fmt.Errorf("container doesn't contain any audio frame chunk lookup entries")
	}

	if frame < 0 || frame >= len(r.audioFrameOffsets) {
		return nil, 0, 0, fmt.Errorf("requested audio frame %d is outside of the valid range from %d to %d", frame, 0, len(r.audioFrameOffsets)-1)
	}
	afte := r.audioFrameOffsets[frame]

	if _, err := r.accessor.Seek(afte.AudioFrameChunkOffset, io.SeekStart); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to seek to audio frame chunk: %w", err)
	}

	chunk, err := r.accessor.ReadChunk64()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read chunk: %w", err)
	}

	// Check size, but only if the AudioFrameChunkSize field is != 0.
	// AudioFrameChunkSize being zero may be some sort of corruption that occurs in older MXV files.
	if afte.AudioFrameChunkSize != 0 && chunk.Length() != int64(afte.AudioFrameChunkSize) {
		return nil, 0, 0, fmt.Errorf("parsed chunk is of wrong size. Got %d bytes, want %d bytes", chunk.Length(), afte.AudioFrameChunkSize)
	}

	if frameChunk, ok := chunk.(*mxriff64.Chunk64MXJVAF64); !ok {
		return nil, 0, 0, fmt.Errorf("parsed chunk is not an audio frame chunk. Got %T, want %T", chunk, frameChunk)
	} else {
		reader, err := frameChunk.DataReader()
		return reader, frameChunk.Data.StartSample, frameChunk.Data.Samples, err
	}
}

// Reads and caches the audio and video frame chunk lookup table.
//
// Keeping this table in RAM uses about 12 bytes per video and 24 bytes per audio frame.
func (r *Reader) PrepareLookupTable() error {
	if r.videoFrameOffsets != nil || r.audioFrameOffsets != nil {
		return nil
	}

	// Cache is empty, rebuild it.

	if r.chunkLookupList == nil {
		return fmt.Errorf("couldn't find MXLIST32 chunk with %s", mxriff64.ContentTypeMXJVTL32)
	}

	// Read frame table from container.
	for chunk, err := range r.chunkLookupList.Chunks() {
		if err != nil {
			return fmt.Errorf("failed to get sub-chunk from audio/video lookup table: %w", err)
		}
		switch chunk := chunk.(type) {
		case *mxriff64.Chunk32VFTE:
			r.videoFrameOffsets = append(r.videoFrameOffsets, chunk.Data)
		case *mxriff64.Chunk32AFTE:
			r.audioFrameOffsets = append(r.audioFrameOffsets, chunk.Data)
		}
	}

	// Ensure that the audio frames are ordered, even though they are most likely already in order.
	slices.SortFunc(r.audioFrameOffsets, func(a, b mxriff64.Chunk32AFTEData) int { return cmp.Compare(a.StartSample, b.StartSample) })

	// Check that we got as many frames as stated in the header.
	if r.Info.VideoFrames != uint64(len(r.videoFrameOffsets)) {
		return fmt.Errorf("actual number of video frames (%d) differs from header value (%d)", len(r.videoFrameOffsets), r.Info.VideoFrames)
	}
	if r.Info.AudioFrames != uint64(len(r.audioFrameOffsets)) {
		return fmt.Errorf("actual number of audio frames (%d) differs from header value (%d)", len(r.audioFrameOffsets), r.Info.AudioFrames)
	}

	// Also check that we got the promised amount of audio samples without gaps or overlaps.
	var audioSamples uint64
	for _, afte := range r.audioFrameOffsets {
		if afte.StartSample != audioSamples {
			return fmt.Errorf("there is a gap or overlap in the audio data")
		}
		audioSamples += uint64(afte.Samples)
	}
	if audioSamples != r.Info.AudioSamples {
		return fmt.Errorf("actual number of audio samples (%d) differs from header value (%d)", audioSamples, r.Info.AudioSamples)
	}

	return nil
}
