// Copyright (c) 2022 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/moutend/go-wav"
)

// demuxFile will demux the given file and write the demuxed data streams into a subfolder with the name of the file.
func demuxFile(filename string) error {

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	chunk, err := ParseChunk64(file)
	if err != nil {
		return fmt.Errorf("failed to parse root chunk: %w", err)
	}

	chunkMXRIFF64, ok := chunk.(*ChunkMXRIFF64)
	if !ok {
		return fmt.Errorf("file doesn't contain a MXRIFF64 as root chunk")
	}

	// Get wave format.
	var audioSampleRate, audioBitDepth, audioChannels int
	var wavObject *wav.File
	for _, childChunk := range chunkMXRIFF64.Chunks {
		if wfmtChunk, ok := childChunk.(*ChunkMXWFMT64); ok {
			audioSampleRate = int(wfmtChunk.ByteRate) / int(wfmtChunk.BytesPerSample) // Seems to be more reliable than just using the SampleRate field.
			audioBitDepth = int(wfmtChunk.ChannelBitDepth)
			audioChannels = int(wfmtChunk.Channels)
			log.Printf("Found waveform table chunk: %v", wfmtChunk)
			log.Printf("Set up audio: SampleRate = %d, BitDepth = %d, Channels = %d", audioSampleRate, audioBitDepth, audioChannels)

			// Set up empty wav object.
			if wavObject, err = wav.New(audioSampleRate, audioBitDepth, audioChannels); err != nil {
				return fmt.Errorf("failed to create wav object: %w", err)
			}

			break
		}
	}

	// Create output directory.
	outputDir := filepath.Base(filename) + "-demuxed"
	outputPath := filepath.Join(filepath.Dir(filename), outputDir)
	if err := os.MkdirAll(outputPath, 0777); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get MXLIST64 with audio and video frames (MXJVFL64).
	var frameCounter, audioSampleCounter int
	for _, childChunk := range chunkMXRIFF64.Chunks {
		if listChunk, ok := childChunk.(*ChunkMXLIST64); ok && listChunk.ContentType == MXJVFL64 {

			// Go through list of video and audio frames.
			for _, frameChunk := range listChunk.Chunks {
				switch frameChunk := frameChunk.(type) {
				case *ChunkMXJVVF64:
					// Create file and write image data into it.
					videoFilename := filepath.Join(outputPath, fmt.Sprintf("video-%06d.jpeg", frameCounter))
					frameCounter++

					if err := os.WriteFile(videoFilename, frameChunk.JPEGData, 0666); err != nil {
						return err
					}

				case *ChunkMXJVAF64:
					// Write audio data into wave object.
					if wavObject != nil {
						if _, err := wavObject.Write(frameChunk.Data); err != nil {
							return err
						}
						audioSampleCounter += int(frameChunk.Samples)
					} else {
						log.Printf("Found audio chunk, but there was no waveform table chunk.")
					}

				}
			}

			break
		}
	}

	// Write wav data into file.
	if wavObject != nil {
		wavTemp, err := wav.Marshal(wavObject)
		if err != nil {
			return fmt.Errorf("failed to encode wave data: %w", err)
		}
		audioFilename := filepath.Join(outputPath, "audio.wav")
		if err := os.WriteFile(audioFilename, wavTemp, 0666); err != nil {
			return fmt.Errorf("failed to write audio file: %w", err)
		}
	}

	log.Printf("Completely demuxed %q: %d video frames, %d audio samples", filename, frameCounter, audioSampleCounter)

	return nil
}
