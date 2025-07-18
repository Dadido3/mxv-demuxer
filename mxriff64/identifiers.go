// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64

type FormType [8]byte

func (f FormType) String() string {
	return "FormType:" + string(f[:])
}

var (
	FormTypeMXJVID64 = FormType{'M', 'X', 'J', 'V', 'I', 'D', '6', '4'} // MAGIX JPEG Video. Or just "MXV".
)

type Identifier32 [4]byte

func (i Identifier32) String() string {
	return "Identifier32:" + string(i[:])
}

type Identifier64 [8]byte

func (i Identifier64) String() string {
	return "Identifier64:" + string(i[:])
}

type ContentType [8]byte

func (i ContentType) String() string {
	return "ContentType:" + string(i[:])
}

var (
	ContentTypeMXJVFL64 = ContentType{'M', 'X', 'J', 'V', 'F', 'L', '6', '4'} // MAGIX JPEG Video Frame List: List containing all video and audio frames.
	ContentTypeMXJVTL32 = ContentType{'M', 'X', 'J', 'V', 'T', 'L', '3', '2'} // MAGIX JPEG Video T? L?: List of file offsets to the respective video and audio chunks.
)
