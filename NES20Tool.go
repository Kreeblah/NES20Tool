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
   along with NES20Tool.  If not, see <https://www.gnu.org/licenses/>.
*/
package main

import (
	"NES20Tool/FDSTool"
	"NES20Tool/FileTools"
	"NES20Tool/NES20Tool"
	"NES20Tool/ProcessingTools"
	"flag"
	"io/ioutil"
	"os"
)

func main() {
	romSetEnableFDS := flag.Bool("enable-fds", false, "Enable FDS support.  Default: false {true|false}")
	romSetEnableFDSHeaders := flag.Bool("enable-fds-headers", false, "Enable writing FDS headers for organization.  Default: false {true|false}")
	romSetEnableV1 := flag.Bool("enable-ines", false, "Enable iNES header support.  iNES headers will always be lower priority for operations than NES 2.0 headers.  Default: false {true|false}")
	romSetGenerateFDSCRCs := flag.Bool("generate-fds-crcs", false, "Generate FDS CRCs for data chunks.  Few, if any, emulators use these.  Default: false {true|false}")
	romSetCommand := flag.String("operation", "", "Operation to perform on the ROM set.  Default: empty {read|write|convertxml}")
	romSetOrganization := flag.Bool("organization", false, "Read/write relative file location information for automatic organization.  Default: false {true|false}")
	romSetTruncateRoms := flag.Bool("truncate-roms", false, "Truncate PRGROM and CHRROM to the sizes specified in the header.  Default: false {true|false}")
	romSetPreserveTrainers := flag.Bool("preserve-trainers", false, "Preserve trainers in read/write process.  Default: false {true|false}")
	romOutputBasePath := flag.String("rom-output-base-path", "", "The path to use for writing organized roms.  Default: empty")
	romSetSourceDirectory := flag.String("rom-source-path", "", "The path to a directory with NES ROMs to use for the operation.  Default: empty")
	romSetXmlFile := flag.String("xml-file", "", "The path to an XML file to use for the operation.  Default: empty")
	xmlImportFormat := flag.String("xml-import-format", "default", "The format of the imported XML file.  Default: default {default|nes20db}")
	xmlExportFormat := flag.String("xml-export-format", "default", "The format of the exported XML file.  Default: default {default|nes20db}")

	flag.Parse()

	if *romSetCommand != "read" && *romSetCommand != "write" && *romSetCommand != "convertxml" {
		printUsage()
		os.Exit(1)
	}

	if *romSetOrganization && *romSetCommand == "read" && *romSetSourceDirectory == "" {
		printUsage()
		os.Exit(1)
	}

	if *romSetOrganization && *romSetCommand == "write" && *romOutputBasePath == "" {
		printUsage()
		os.Exit(1)
	}

	if *xmlImportFormat != "default" && *xmlImportFormat != "nes20db" {
		printUsage()
		os.Exit(1)
	}

	if *xmlExportFormat != "default" && *xmlExportFormat != "nes20db" {
		printUsage()
		os.Exit(1)
	}

	if *xmlImportFormat == "nes20db" || *xmlExportFormat == "nes20db" {
		*romSetEnableV1 = false
		*romSetEnableFDS = false
		*romSetOrganization = false
	}

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

		if *xmlExportFormat == "default" {
			xmlPayload, err = FileTools.MarshalXMLFromROMMap(romMap, archiveMap, *romSetEnableV1, *romSetPreserveTrainers, *romSetOrganization)
			if err != nil {
				panic(err)
			}
		} else if *xmlExportFormat == "nes20db" {
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
	} else if *romSetCommand == "write" {
		println("Loading XML file from: " + *romSetXmlFile)
		xmlPayload, err := ioutil.ReadFile(*romSetXmlFile)
		if err != nil {
			panic(err)
		}

		println("Reading XML file")
		var romData map[string]*NES20Tool.NESROM
		var archiveData map[string]*FDSTool.FDSArchiveFile
		var hashTypeMatch uint64

		if *xmlImportFormat == "default" {
			romData, archiveData, err = FileTools.UnmarshalXMLToROMMap(string(xmlPayload), *romSetEnableV1, *romSetPreserveTrainers, *romOutputBasePath != "")
			if err != nil {
				panic(err)
			}

			hashTypeMatch = ProcessingTools.HASH_TYPE_SHA256
		} else if *xmlImportFormat == "nes20db" {
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

		matchedRoms := ProcessingTools.ProcessNESROMs(rawRoms, romData, hashTypeMatch, *romSetTruncateRoms, *romSetEnableV1)

		rawArchives := make([]*FDSTool.FDSArchiveFile, 0)
		matchedArchives := make([]*FDSTool.FDSArchiveFile, 0)

		if *romSetEnableFDS {
			println("Processing FDS archives in: " + *romSetSourceDirectory)
			rawArchives, err = FileTools.LoadFDSArchiveRecursive(*romSetSourceDirectory, false)
			if err != nil {
				panic(err)
			}

			matchedArchives = ProcessingTools.ProcessFDSROMs(rawArchives, archiveData, ProcessingTools.HASH_TYPE_SHA256)
		}

		tempBasePath := *romOutputBasePath
		if tempBasePath[len(tempBasePath)-1] != os.PathSeparator {
			tempBasePath = tempBasePath + string(os.PathSeparator)
		}

		for index := range matchedRoms {
			if *romOutputBasePath == "" {
				println("Writing NES ROM: " + matchedRoms[index].Filename)
			} else {
				println("Writing NES ROM: " + tempBasePath + matchedRoms[index].RelativePath)
			}

			if matchedRoms[index].Header20 != nil || (*romSetEnableV1 && matchedRoms[index].Header10 != nil) {
				err = FileTools.WriteROM(matchedRoms[index], *romSetEnableV1, *romSetTruncateRoms, *romSetPreserveTrainers, *romOutputBasePath)
				if err != nil {
					if *romOutputBasePath == "" {
						println("Error writing ROM: " + matchedRoms[index].Filename)
					} else {
						println("Error writing ROM: " + *romOutputBasePath + string(os.PathSeparator) + matchedRoms[index].RelativePath)
					}
					println(err.Error())
				}
			}
		}

		for index := range matchedArchives {
			if *romOutputBasePath == "" {
				println("Writing FDS archive: " + matchedArchives[index].Filename)
			} else {
				println("Writing FDS archive: " + tempBasePath + matchedArchives[index].Filename)
			}

			err = FileTools.WriteFDSArchive(matchedArchives[index], *romSetEnableFDSHeaders, *romOutputBasePath)
			if err != nil {
				if *romOutputBasePath == "" {
					println("Error writing FDS archive: " + matchedArchives[index].Filename)
				} else {
					println("Error writing FDS archive: " + tempBasePath + matchedArchives[index].RelativePath)
				}
				println(err.Error())
			}
		}

		os.Exit(0)
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
