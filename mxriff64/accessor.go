// Copyright (c) 2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package mxriff64

import (
	"fmt"
	"io"
)

// An accessor wraps any io.Reader, io.Seeker and/or io.Writer.
// The io interfaces are optional and can be nil.
type Accessor struct {
	io.Reader
	io.Seeker
	//io.Writer // TODO: This could be easily used to extend the MXRIFF64 lib with write support

	Pos int64 // The current file offset.
}

// Starting point for reading a MXRIFF64 container based on an io.Reader.
func NewFromReader(b io.Reader) *Accessor {
	return &Accessor{
		Reader: b,
	}
}

// Starting point for reading a MXRIFF64 container based on an io.ReadSeeker.
func NewFromReadSeeker(b io.ReadSeeker) *Accessor {
	return &Accessor{
		Reader: b,
		Seeker: b,
	}
}

func (a *Accessor) Read(p []byte) (n int, err error) {
	if a.Reader != nil {
		n, err = a.Reader.Read(p)
		a.Pos += int64(n)
		return
	}

	return 0, fmt.Errorf("the accessor doesn't support reading")
}

func (a *Accessor) Seek(offset int64, whence int) (o int64, err error) {
	// Ignore any seek operation to the current position.
	switch whence {
	case io.SeekStart:
		if offset == a.Pos {
			return a.Pos, nil
		}
	case io.SeekCurrent:
		if offset == 0 {
			return a.Pos, nil
		}
	}

	if a.Seeker != nil {
		o, err = a.Seeker.Seek(offset, whence)
		a.Pos = o
		return
	}

	// Support for virtual seeking by reading and discarding data.
	// Works only in the forward direction.
	if a.Reader != nil {
		var diff int64
		switch whence {
		case io.SeekStart:
			diff = offset - a.Pos
		case io.SeekCurrent:
			diff = offset
		case io.SeekEnd:
			return 0, fmt.Errorf("io.SeekEnd is not supported")
		}

		if diff < 0 {
			return 0, fmt.Errorf("the accessor doesn't support negative seeking")
		}
		if n, err := io.CopyN(io.Discard, a.Reader, diff); err != nil {
			a.Pos += n
			return 0, fmt.Errorf("failed to discard %d bytes from the reader: %w", diff, err)
		}
		a.Pos += diff
		return a.Pos, nil
	}

	return a.Pos, fmt.Errorf("the accessor doesn't support seeking")
}

/*func (a *Accessor) Write(p []byte) (n int, err error) {
	if a.Writer != nil {
		n, err = a.Writer.Write(p)
		a.Pos += int64(n)
		return
	}

	return 0, fmt.Errorf("the accessor doesn't support writing")
}*/
