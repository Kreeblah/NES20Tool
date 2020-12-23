/*
   Copyright 2020, Christopher Gelatt

   This file is part of NESTool.

   NESTool is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   NESTool is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with NESTool.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"NES20Tool/FDSTool"
	"NES20Tool/FileTools"
	"NES20Tool/NESTool"
	"NES20Tool/ProcessingTools"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {
	// Parse the CLI options
	romSetEnableFDS := flag.Bool("enable-fds", false, "Enable FDS support.")
	romSetEnableFDSHeaders := flag.Bool("enable-fds-headers", false, "Enable writing FDS headers for organization.")
	romSetEnableV1 := flag.Bool("enable-ines", false, "Enable iNES header support.  iNES headers will always be lower priority for operations than NES 2.0 headers.")
	romSetGenerateFDSCRCs := flag.Bool("generate-fds-crcs", false, "Generate FDS CRCs for data chunks.  Few, if any, emulators use these.")
	romSetCommand := flag.String("operation", "", "Required.  Operation to perform on the ROM set. {read|write}")
	romSetOrganization := flag.Bool("organization", false, "Read/write relative file location information for automatic organization.")
	romSetTruncateRoms := flag.Bool("truncate-roms", false, "Truncate PRGROM and CHRROM to the sizes specified in the header.")
	romSetPreserveTrainers := flag.Bool("preserve-trainers", false, "Preserve trainers in read/write process.")
	romOutputBasePath := flag.String("rom-output-base-path", "", "The path to use for writing organized NES and/or FDS ROMs.")
	romSetSourceDirectory := flag.String("rom-source-path", "", "Required.  The path to a directory with NES and/or FDS ROMs to use for the operation.")
	romSetXmlFile := flag.String("xml-file", "", "The path to an XML file to use for the operation.")
	xmlFormat := flag.String("xml-format", "default", "The format of the imported or exported XML file. {default|nes20db}")

	flag.Parse()

	// Options validation
	if *romSetCommand != "read" && *romSetCommand != "write" {
		printUsage()
		os.Exit(1)
	}

	if *romSetSourceDirectory == "" {
		printUsage()
		os.Exit(1)
	}

	if *romSetCommand == "write" && *romOutputBasePath == "" {
		if *romSetOrganization {
			printUsage()
			os.Exit(1)
		} else {
			*romOutputBasePath = *romSetSourceDirectory
		}
	}

	if *xmlFormat != "default" && *xmlFormat != "nes20db" {
		printUsage()
		os.Exit(1)
	}

	// nes20db functionality is only for NES 2.0 ROMs
	if *xmlFormat == "nes20db" {
		*romSetEnableV1 = false
		*romSetEnableFDS = false
		*romSetOrganization = false
	}

	// Get the aboslute path for easier calculation of relative ROM paths
	if *romOutputBasePath != "" {
		tempOutputPath, err := filepath.Abs(*romOutputBasePath)
		if err != nil {
			panic(err)
		}

		*romOutputBasePath = tempOutputPath
	}

	if *romSetSourceDirectory != "" {
		tempSourceDirectory, err := filepath.Abs(*romSetSourceDirectory)
		if err != nil {
			panic(err)
		}

		*romSetSourceDirectory = tempSourceDirectory
	}

	// Read a directory structure and generate an XML file to represent it
	if *romSetCommand == "read" {
		println("Loading NES 2.0 ROMs from: " + *romSetSourceDirectory)
		romMap, err := FileTools.LoadROMRecursiveMap(*romSetSourceDirectory, *romSetEnableV1, *romSetPreserveTrainers, ProcessingTools.HASH_TYPE_SHA256)
		if err != nil {
			panic(err)
		}

		archiveMap := make(map[string]*FDSTool.FDSArchiveFile, 0)

		if *romSetEnableFDS {
			println("Loading FDS archives from: " + *romSetSourceDirectory)
			archiveMap, err = FileTools.LoadFDSArchiveRecursiveMap(*romSetSourceDirectory, *romSetGenerateFDSCRCs, ProcessingTools.HASH_TYPE_SHA256)
			if err != nil {
				panic(err)
			}
		}

		println("Generating XML")
		var xmlPayload string

		if *xmlFormat == "default" {
			xmlPayload, err = FileTools.MarshalXMLFromROMMap(romMap, archiveMap, *romSetEnableV1, *romSetPreserveTrainers, *romSetOrganization)
			if err != nil {
				panic(err)
			}
		} else if *xmlFormat == "nes20db" {
			xmlPayload, err = FileTools.MarshalNES20DBXMLFromROMMap(romMap)
			if err != nil {
				panic(err)
			}
		}

		println("Writing XML to: " + *romSetXmlFile)
		err = FileTools.WriteStringToFile(xmlPayload, *romSetXmlFile)
		if err != nil {
			panic(err)
		}

		os.Exit(0)

		// Read an XML file and a source ROM set, match the ROMs in it, and
		// write out a ROM set in a destination location.
	} else if *romSetCommand == "write" {
		println("Loading XML file from: " + *romSetXmlFile)
		xmlPayload, err := ioutil.ReadFile(*romSetXmlFile)
		if err != nil {
			panic(err)
		}

		println("Reading XML file")
		var romData map[string]*NESTool.NESROM
		var archiveData map[string]*FDSTool.FDSArchiveFile
		var hashTypeMatch uint64

		if *xmlFormat == "default" {
			romData, archiveData, err = FileTools.UnmarshalXMLToROMMap(string(xmlPayload), *romSetEnableV1, *romSetPreserveTrainers, *romOutputBasePath != "")
			if err != nil {
				panic(err)
			}

			hashTypeMatch = ProcessingTools.HASH_TYPE_SHA256
		} else if *xmlFormat == "nes20db" {
			romData, err = FileTools.UnmarshalNES20DBXMLToROMMap(string(xmlPayload))
			if err != nil {
				panic(err)
			}

			hashTypeMatch = ProcessingTools.HASH_TYPE_SHA1
		}

		println("Processing NES ROMs in: " + *romSetSourceDirectory)
		rawRoms, err := FileTools.LoadROMRecursive(*romSetSourceDirectory, *romSetEnableV1, *romSetPreserveTrainers)
		if err != nil {
			panic(err)
		}

		matchedRoms := ProcessingTools.ProcessNESROMs(rawRoms, romData, hashTypeMatch, *romSetTruncateRoms, *romSetOrganization, *romSetEnableV1)

		println("Processing UNIF ROMs in: " + *romSetSourceDirectory)
		rawUnifs, err := FileTools.LoadUNIFRecursive(*romSetSourceDirectory)
		if err != nil {
			panic(err)
		}

		matchedUnifs := ProcessingTools.ProcessNESROMs(rawUnifs, romData, hashTypeMatch, *romSetTruncateRoms, *romSetOrganization, *romSetEnableV1)

		matchedRoms = append(matchedRoms, matchedUnifs...)

		rawArchives := make([]*FDSTool.FDSArchiveFile, 0)
		matchedArchives := make([]*FDSTool.FDSArchiveFile, 0)

		if *romSetEnableFDS {
			println("Processing FDS archives in: " + *romSetSourceDirectory)
			rawArchives, err = FileTools.LoadFDSArchiveRecursive(*romSetSourceDirectory, false)
			if err != nil {
				panic(err)
			}

			matchedArchives = ProcessingTools.ProcessFDSROMs(rawArchives, archiveData, ProcessingTools.HASH_TYPE_SHA256, *romSetOrganization)
		}

		tempBasePath := *romOutputBasePath
		if tempBasePath[len(tempBasePath)-1] != os.PathSeparator {
			tempBasePath = tempBasePath + string(os.PathSeparator)
		}

		for index := range matchedRoms {
			tempFilename := matchedRoms[index].Filename
			if tempFilename == "" {
				tempFilename = matchedRoms[index].Name + ".nes"
			}

			tempRelativePath := matchedRoms[index].RelativePath
			if tempRelativePath == "" {
				tempRelativePath = matchedRoms[index].Name + ".nes"
			}

			if *romOutputBasePath == "" {
				println("Writing NES ROM: " + tempFilename)
			} else {
				println("Writing NES ROM: " + tempBasePath + tempRelativePath)
			}

			if matchedRoms[index].Header20 != nil || (*romSetEnableV1 && matchedRoms[index].Header10 != nil) {
				err = FileTools.WriteROM(matchedRoms[index], *romSetEnableV1, *romSetTruncateRoms, *romSetPreserveTrainers, *romOutputBasePath)
				if err != nil {
					if *romOutputBasePath == "" {
						println("Error writing ROM: " + tempFilename)
					} else {
						println("Error writing ROM: " + *romOutputBasePath + string(os.PathSeparator) + tempRelativePath)
					}
					println(err.Error())
				}
			}
		}

		for index := range matchedArchives {
			tempFilename := matchedArchives[index].Filename
			if tempFilename == "" {
				tempFilename = matchedArchives[index].Name + ".nes"
			}

			tempRelativePath := matchedArchives[index].RelativePath
			if tempRelativePath == "" {
				tempRelativePath = matchedArchives[index].Name + ".nes"
			}

			if *romOutputBasePath == "" {
				println("Writing FDS archive: " + tempFilename)
			} else {
				println("Writing FDS archive: " + tempBasePath + tempRelativePath)
			}

			err = FileTools.WriteFDSArchive(matchedArchives[index], *romSetEnableFDSHeaders, *romOutputBasePath)
			if err != nil {
				if *romOutputBasePath == "" {
					println("Error writing FDS archive: " + tempFilename)
				} else {
					println("Error writing FDS archive: " + tempBasePath + tempRelativePath)
				}
				println(err.Error())
			}
		}

		os.Exit(0)
	}
}

// Show the usage options.
func printUsage() {
	println("This utility reads a ROM set which has NES 2.0 headers and")
	println("generates an XML file to describe them, or reads an XML file")
	println("describing NES 2.0 headers and applies them to a ROM set in")
	println("a given directory.  The \"read\" operation generates an XML")
	println("file, and the \"write\" operation is used to update a ROM set.")
	flag.PrintDefaults()
}
