// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64

type AudioFormat uint16

const (
	AudioFormatPCM       AudioFormat = 1
	AudioFormatMSADPCM   AudioFormat = 2
	AudioFormatIEEEFloat AudioFormat = 3
	AudioFormatIBMCVSD   AudioFormat = 5
	AudioFormatALAW      AudioFormat = 6
	AudioFormatMULAW     AudioFormat = 7
)

func (c AudioFormat) String() string {
	switch c {
	case AudioFormatPCM:
		return "AudioFormat:PCM"
	case AudioFormatMSADPCM:
		return "AudioFormat:MS ADPCM"
	case AudioFormatIEEEFloat:
		return "AudioFormat:IEEE FLOAT"
	case AudioFormatIBMCVSD:
		return "AudioFormat:IBM CVSD"
	case AudioFormatALAW:
		return "AudioFormat:ALAW"
	case AudioFormatMULAW:
		return "AudioFormat:MULAW"
	}

	return "AudioFormat:Unknown"
}
