// Copyright (c) 2022 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

func main() {
	if err := demuxFile("D:\\Eigenes\\Magix Daten\\__Demo\\02 Walker.mxv"); err != nil {
		panic(err)
	}

}
