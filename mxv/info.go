package mxv

// Info contains information about the video and audio data of a MXV file.
type Info struct {
	FrameWidth  uint32
	FrameHeight uint32
	Framerate   float64 // Rate of frame/s.
	VideoFrames uint64  // Total amount of video frames.
	AspectRatio float64 // Output aspect ratio. The final video needs to be stretched to this ratio.

	HasAudio             bool
	AudioChannels        uint16
	AudioSampleRate      uint32
	AudioByteRate        uint32
	AudioBytesPerSample  uint16
	AudioChannelBitDepth uint32
	AudioFrames          uint64
	AudioSamples         uint64
}
