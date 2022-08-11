// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// instgen is used to generate the arm64 instruction table for encoding and decoding.
// usages:
//
//	instgen -i=inputDir [-o=outputDir]
//	or
//	instgen -url=linkAddress [-o=outputDir]
//
// Users can download the arm64 xml instruction manual from
// https://developer.arm.com/downloads/-/exploration-tools,
// decompress it, and pass the directory containing the latest XML
// instruction files by "-i" to instgen.
// You can also pass the download link to the manual to instgen by "-url",
// the program will download and unpack the file. The link address is not
// fixed, so we can't hard code it in the program.
//
// The program parses and processes all the xml files, and generates four go files:
// instructions.go, elements.go, ops.go and arm64ops.go to the output directory.
// If -o option is not specified, the files will be written to the current directory.
//
// Since the format of the Arm64 instruction specification document may update,
// this parser may not work for some versions of the XML document. The current
// implementation is based on version:
// https://developer.arm.com/-/media/developer/products/architecture/armv9-a-architecture/2023-03/ISA_A64_xml_A_profile-2023-03.tar.gz

package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
)

var url = flag.String("url", "", "the link address of the xml files")
var input = flag.String("i", "", "the input directory of the xml files")
var output = flag.String("o", "", "the output directory of the generated files")

func usage() {
	fmt.Fprintf(os.Stderr, "usage: instgen -i=\"inputDir\" [-o=\"outputDir\"] \n")
	fmt.Fprintf(os.Stderr, "       instgen -url=\"linkAddress\" [-o=\"outputDir\"]\n")
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() > 2 || *input != "" && *url != "" || *input == "" && *url == "" {
		usage()
	}

	// Step1: download and unpack the arm64 ISA manual if necessary.
	if *input == "" {
		log.Printf("input directory is empty\n")
		// Download the xml files to a temp directory under the current directory.
		curDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Getwd() failed: %v\n", err)
		}
		workDir, err := os.MkdirTemp(curDir, "arm64XMLParser")
		if err != nil {
			log.Fatalf("MkdirTemp() failed: %v\n", err)
		}
		//	defer os.RemoveAll(workDir)
		*input = download(workDir)
	}
	if *output == "" {
		curDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Getwd() failed: %v\n", err)
		}
		*output = curDir
	}

	// Step2: parse each xml file to insts.
	parseXMLFiles(*input)

	// Step3: processing the parsing results.
	processXMLFiles()

	fmt.Printf("len(insts) = %v\n", len(insts))
	fmt.Printf("len(rules) = %v\n", len(rules))

	// Step4: generate instructions.go, elements.go, ops.go and arm64ops.go
	//
	// Sort insts, use the stable sort algorithm to reduce meaningless code changes.
	sort.SliceStable(insts, func(i, j int) bool {
		enci, encj := insts[i].Classes.Iclass[0].Encodings[0], insts[j].Classes.Iclass[0].Encodings[0]
		if enci.arm64Op != encj.arm64Op {
			return enci.arm64Op < encj.arm64Op
		}
		if enci.class != encj.class {
			return enci.class < encj.class
		}
		if enci.goOp != encj.goOp {
			return enci.goOp < encj.goOp
		}
		if len(enci.operands) != len(encj.operands) {
			return len(enci.operands) < len(encj.operands)
		}
		if enci.binary != encj.binary {
			return enci.binary < encj.binary
		}
		if enci.mask != encj.mask {
			return enci.mask < encj.mask
		}
		return enci.asm < encj.asm
	})
	generate(*output)
}

// download downloads the arm64 instruction manual to dir and returns the directory of the xml files.
func download(dir string) string {
	log.Printf("try to download the xml files to %s\n", dir)
	resp, err := http.Get(*url)
	if err != nil {
		log.Fatalf("Get() failed: %v\n", err)
	}
	if resp.StatusCode != 200 {
		log.Fatal(resp.Status)
	}

	fileName := strings.Split(resp.Header["Content-Disposition"][0], "\"")[1]
	compressedFile := path.Join(dir, fileName)
	f, err := os.Create(compressedFile)
	if err != nil {
		log.Fatalf("Create() failed: %v\n", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		log.Fatalf("Copy() failed: %v\n", err)
	}

	// Seek to the beginning of the file.
	f.Seek(0, 0)

	// Extract the compressed file
	log.Printf("try to decompress %s\n", compressedFile)
	extractTarGz(f, dir)

	// Return the directory of the xml files.
	return compressedFile[:strings.Index(compressedFile, ".")]
}

// extractTarGz unpacks a file f in the dir directory to dir.
func extractTarGz(f *os.File, dir string) {
	gr, err := gzip.NewReader(f)
	if err != nil {
		log.Fatalf("gzip.NewReader() failed: %v\n", err)
	}

	// tar read
	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatalf("Next() failed: %v\n", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path.Join(dir, header.Name), 0700); err != nil {
				log.Fatalf("Mkdir() failed: %v\n", err)
			}
		case tar.TypeReg:
			xmlFile, err := os.Create(path.Join(dir, header.Name))
			if err != nil {
				log.Fatalf("Create() failed: %v\n", err)
			}
			if _, err := io.Copy(xmlFile, tr); err != nil {
				log.Fatalf("Copy() failed: %v\n", err)
			}
			if err := xmlFile.Close(); err != nil {
				log.Fatalf("Close() failed: %v\n", err)
			}
		default:
			log.Fatalf("uknown type: %v in %v\n", header.Typeflag, header.Name)
		}
	}
	log.Printf("Decompression completed\n")
}
