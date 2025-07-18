// Copyright (c) 2022-2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/Dadido3/mxv-demuxer/mxv"
	"github.com/moutend/go-wav"
)

// demuxFile will demux the given file and write the demuxed data streams into a subfolder with the name of the file.
func demuxFile(filename string) error {

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	mxvReader, err := mxv.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to read MXV file: %w", err)
	}

	if err := mxvReader.PrepareLookupTable(); err != nil {
		return fmt.Errorf("failed to prepare lookup table: %w", err)
	}

	// Create output directory.
	outputDir := filepath.Base(filename) + "-demuxed"
	outputPath := filepath.Join(filepath.Dir(filename), outputDir)
	if err := os.MkdirAll(outputPath, 0777); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	log.Printf("MXV info: %+v.", mxvReader.Info)

	// Video frames.
	log.Printf("Extracting video frames.")
	for frame := range mxvReader.VideoFrames() {
		videoFilename := filepath.Join(outputPath, fmt.Sprintf("video-%06d.jpeg", frame))
		file, err := os.Create(videoFilename)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		frameReader, err := mxvReader.VideoFrameData(frame)
		if err != nil {
			return fmt.Errorf("failed to get video data stream: %w", err)
		}
		if _, err := io.Copy(file, frameReader); err != nil {
			return fmt.Errorf("failed to copy video data stream: %w", err)
		}
	}

	log.Printf("Finished extracting video frames.")

	if mxvReader.Info.HasAudio {
		// TODO: go-wav doesn't support streaming, therefore replace it

		// Set up empty wav object.
		wavObject, err := wav.New(int(mxvReader.Info.AudioSampleRate), int(mxvReader.Info.AudioChannelBitDepth), int(mxvReader.Info.AudioChannels))
		if err != nil {
			return fmt.Errorf("failed to create wav object: %w", err)
		}

		// Append sample data.
		log.Printf("Extracting audio samples.")
		for frame := range mxvReader.AudioFrames() {
			frameReader, _, _, err := mxvReader.AudioFrameData(frame)
			if err != nil {
				return fmt.Errorf("failed to get audio data stream: %w", err)
			}
			frameData, err := io.ReadAll(frameReader)
			if err != nil {
				return fmt.Errorf("failed to read audio data stream: %w", err)
			}
			if _, err := wavObject.Write(frameData); err != nil {
				return fmt.Errorf("failed to append audio data to wave object: %w", err)
			}
		}

		// Write audio file to disk.
		audioFilename := filepath.Join(outputPath, "audio.wav")
		log.Printf("Writing extracted audio data to disk at %q.", audioFilename)
		wavTemp, err := wav.Marshal(wavObject)
		if err != nil {
			return fmt.Errorf("failed to encode wave data: %w", err)
		}
		if err := os.WriteFile(audioFilename, wavTemp, 0666); err != nil {
			return fmt.Errorf("failed to write audio file: %w", err)
		}

		log.Printf("Finished writing audio data.")
	}

	log.Printf("Completely demuxed %q.", filename)

	return nil
}
