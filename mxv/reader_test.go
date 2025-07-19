// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxv_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Dadido3/mxv-demuxer/mxriff64"
	"github.com/Dadido3/mxv-demuxer/mxv"
	"github.com/google/go-cmp/cmp"
)

func TestNewReader(t *testing.T) {
	tests := []struct {
		filepath string   // The path to the MXV file to be tested.
		mxvInfo  mxv.Info // Expected video/audio information.
	}{
		{
			filepath: filepath.Join("..", "example-files", "Vergleich2.mxv"),
			mxvInfo: mxv.Info{
				ColorFormat: mxriff64.ColorFormatYUY2, FrameWidth: 720, FrameHeight: 576, Framerate: 25, VideoFrames: 349, AspectRatio: 1.3333332999999998,
				HasAudio: true, AudioFormat: mxriff64.AudioFormatPCM, AudioChannels: 2, AudioSampleRate: 48000, AudioByteRate: 192000,
				AudioBytesPerSample: 4, AudioChannelBitDepth: 16, AudioFrames: 28, AudioSamples: 672000,
			},
		},
		{
			filepath: filepath.Join("..", "example-files", "23.976p.mxv"),
			mxvInfo: mxv.Info{
				ColorFormat: mxriff64.ColorFormatYV12, FrameWidth: 1920, FrameHeight: 1080, Framerate: 23.976, VideoFrames: 48, AspectRatio: 1.7777777777777777,
				HasAudio: true, AudioFormat: mxriff64.AudioFormatPCM, AudioChannels: 2, AudioSampleRate: 48000, AudioByteRate: 192000,
				AudioBytesPerSample: 4, AudioChannelBitDepth: 16, AudioFrames: 48, AudioSamples: 96096,
			},
		},
		{
			filepath: filepath.Join("..", "example-files", "24p.mxv"),
			mxvInfo: mxv.Info{
				ColorFormat: mxriff64.ColorFormatYV12, FrameWidth: 1920, FrameHeight: 1080, Framerate: 24, VideoFrames: 48, AspectRatio: 1.7777777777777777,
				HasAudio: true, AudioFormat: mxriff64.AudioFormatPCM, AudioChannels: 2, AudioSampleRate: 48000, AudioByteRate: 192000,
				AudioBytesPerSample: 4, AudioChannelBitDepth: 16, AudioFrames: 48, AudioSamples: 96000,
			},
		},
		{
			filepath: filepath.Join("..", "example-files", "25i.mxv"),
			mxvInfo: mxv.Info{
				ColorFormat: mxriff64.ColorFormatYV12, FrameWidth: 1440, FrameHeight: 1080, Framerate: 25, VideoFrames: 50, AspectRatio: 1.7777777777777777,
				HasAudio: true, AudioFormat: mxriff64.AudioFormatPCM, AudioChannels: 2, AudioSampleRate: 48000, AudioByteRate: 192000,
				AudioBytesPerSample: 4, AudioChannelBitDepth: 16, AudioFrames: 50, AudioSamples: 96000,
			},
		},
		{
			filepath: filepath.Join("..", "example-files", "50p.mxv"),
			mxvInfo: mxv.Info{
				ColorFormat: mxriff64.ColorFormatYV12, FrameWidth: 1920, FrameHeight: 1080, Framerate: 50, VideoFrames: 100, AspectRatio: 1.7777777777777777,
				HasAudio: true, AudioFormat: mxriff64.AudioFormatPCM, AudioChannels: 2, AudioSampleRate: 48000, AudioByteRate: 192000,
				AudioBytesPerSample: 4, AudioChannelBitDepth: 16, AudioFrames: 100, AudioSamples: 96000,
			},
		},
		{
			filepath: filepath.Join("..", "example-files", "60p.mxv"),
			mxvInfo: mxv.Info{
				ColorFormat: mxriff64.ColorFormatYV12, FrameWidth: 1920, FrameHeight: 1080, Framerate: 60, VideoFrames: 120, AspectRatio: 1.7777777777777777,
				HasAudio: true, AudioFormat: mxriff64.AudioFormatPCM, AudioChannels: 2, AudioSampleRate: 48000, AudioByteRate: 192000,
				AudioBytesPerSample: 4, AudioChannelBitDepth: 16, AudioFrames: 120, AudioSamples: 96000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.filepath, func(t *testing.T) {
			f, err := os.Open(tt.filepath)
			if err != nil {
				t.Fatalf("Failed to open file: %v.", err)
			}
			defer f.Close()

			mxvReader, err := mxv.NewReader(f)
			if err != nil {
				t.Fatalf("Failed to read MXV file: %v.", err)
			}

			if !cmp.Equal(tt.mxvInfo, mxvReader.Info) {
				t.Errorf("Parsed video info differs from expected result:\n%s", cmp.Diff(tt.mxvInfo, mxvReader.Info))
			}
		})
	}
}
