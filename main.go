// Copyright (c) 2022-2025 David Vogel
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/earthboundkid/versioninfo/v2"
)

func main() {
	flag.Parse()

	filenames := flag.Args()

	log.Printf("Started mxv-demuxer %v.", versioninfo.Version)

	if len(filenames) == 0 {
		var err error
		if filenames, err = findFiles("."); err != nil {
			log.Panicf("Failed to find files: %v", err)
		}
	}

	for _, filename := range filenames {
		log.Printf("Starting to demux %q...", filename)
		if err := demuxFile(filename); err != nil {
			log.Printf("Failed to demux %q: %v", filename, err)
		}
	}
}

func findFiles(root string) ([]string, error) {
	fileInfos, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && strings.ToLower(filepath.Ext(fileInfo.Name())) == ".mxv" {
			files = append(files, filepath.Join(root, fileInfo.Name()))
		}
	}

	return files, nil
}
