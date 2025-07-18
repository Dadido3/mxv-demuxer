// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64

type ColorFormat [4]byte

var (
	ColorFormatZero  = ColorFormat{0, 0, 0, 0}
	ColorFormatThree = ColorFormat{3, 0, 0, 0}
	ColorFormatI420  = ColorFormat{'I', '4', '2', '0'}
	ColorFormatIYUV  = ColorFormat{'I', 'Y', 'U', 'V'}
	ColorFormatY411  = ColorFormat{'Y', '4', '1', '1'}
	ColorFormatY422  = ColorFormat{'Y', '4', '2', '2'}
	ColorFormatYUNV  = ColorFormat{'Y', 'U', 'N', 'V'}
	ColorFormatYUY2  = ColorFormat{'Y', 'U', 'Y', '2'}
	ColorFormatYUYV  = ColorFormat{'Y', 'U', 'Y', 'V'}
	ColorFormatYV12  = ColorFormat{'Y', 'V', '1', '2'}
)

func (c ColorFormat) String() string {
	return "ColorFormat:" + string(c[:])
}
