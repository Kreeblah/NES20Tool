/*
   Copyright 2020, Christopher Gelatt

   This file is part of NES20Tool.

   NES20Tool is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   NES20Tool is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with Foobar.  If not, see <https://www.gnu.org/licenses/>.
*/
package main

import (
	"NES20Tool/FileTools"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"io/ioutil"
	"os"
	"reflect"
)

func main() {
	romSetCommand := flag.String("operation", "", "Operation to perform on the ROM set.  {read|write}")
	romSetDirectory := flag.String("rom-path", "", "The path to a directory with NES ROMs to use for the operation.")
	romSetXmlFile := flag.String("xml-file", "", "The path to an XML file to use for the operation.")

	flag.Parse()

	if *romSetCommand == "read" {
		if romSetXmlFile != nil && romSetDirectory != nil {
			println("Loading NES 2.0 ROMs from: " + *romSetDirectory)
			romMap, err := FileTools.LoadROMRecursiveMap(*romSetDirectory)
			if err != nil {
				panic(err)
			}

			println("Generating XML")
			xmlPayload, err := FileTools.MarshalXMLFromROMMap(romMap)
			if err != nil {
				panic(err)
			}

			println("Writing XML to: " + *romSetXmlFile)
			err = FileTools.WriteStringToFile(xmlPayload, *romSetXmlFile)
			if err != nil {
				panic(err)
			}

			os.Exit(0)
		} else {
			printUsage()
			os.Exit(1)
		}
	} else if *romSetCommand == "write" {
		if romSetXmlFile != nil && romSetDirectory != nil {

			println("Loading XML file from: " + *romSetXmlFile)
			xmlPayload, err := ioutil.ReadFile(*romSetXmlFile)
			if err != nil {
				panic(err)
			}

			println("Reading XML file")
			romData, err := FileTools.UnmarshalXMLToROMMap(string(xmlPayload))
			if err != nil {
				panic(err)
			}

			println("Processing NES ROMs in: " + *romSetDirectory)
			rawRoms, err := FileTools.LoadROMRecursive(*romSetDirectory)
			if err != nil {
				panic(err)
			}

			for key, _ := range rawRoms {
				println("Checking ROM: " + rawRoms[key].Filename)
				if romData[rawRoms[key].SHA256] != nil {
					println("Matched NES 2.0 ROM: " + romData[rawRoms[key].SHA256].Name)
					if !reflect.DeepEqual(romData[rawRoms[key].SHA256].Header, rawRoms[key].Header) {
						println("Writing NES 2.0 ROM: " + rawRoms[key].Filename)
						rawRoms[key].Header = romData[rawRoms[key].SHA256].Header
						err = FileTools.WriteROM(rawRoms[key])
						if err != nil {
							println("Error writing NES 2.0 ROM: " + rawRoms[key].Filename)
							println(err.Error())
						}
					} else {
						println("Skipping ROM (already up to date): " + rawRoms[key].Filename)
					}
				} else {
					tempCrc32Bytes := make([]byte, 4)
					binary.BigEndian.PutUint32(tempCrc32Bytes, rawRoms[key].CRC32)
					tempCrc32String := hex.EncodeToString(tempCrc32Bytes)
					println("Failed to match ROM: " + rawRoms[key].Filename)
					println("ROM CRC32:  " + tempCrc32String)
					println("ROM MD5:    " + hex.EncodeToString(rawRoms[key].MD5[:]))
					println("ROM SHA1:   " + hex.EncodeToString(rawRoms[key].SHA1[:]))
					println("ROM SHA256: " + hex.EncodeToString(rawRoms[key].SHA256[:]))
				}
			}

			os.Exit(0)
		} else {
			printUsage()
			os.Exit(1)
		}
	} else {
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	println("This utility reads a ROM set which has NES 2.0 headers and")
	println("generates an XML file to describe them, or reads an XML file")
	println("describing NES 2.0 headers and applies them to a ROM set in")
	println("a given directory.  The \"read\" operation generates an XML")
	println("file, and the \"write\" operation is used to update a ROM set.")
	flag.PrintDefaults()
}
