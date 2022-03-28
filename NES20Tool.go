/*
   Copyright 2021-2022, Christopher Gelatt

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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	// Parse the CLI options
	romSetEnableFDS := flag.Bool("enable-fds", false, "Enable FDS support.")
	romSetEnableFDSHeaders := flag.Bool("enable-fds-headers", false, "Enable writing FDS headers for organization.")
	romSetEnableV1 := flag.Bool("enable-ines", false, "Enable iNES header support.  iNES headers will always be lower priority for operations than NES 2.0 headers.")
	romSetGenerateFDSCRCs := flag.Bool("generate-fds-crcs", false, "Generate FDS CRCs for data chunks.  Few, if any, emulators use these.")
	romSetCommand := flag.String("operation", "", "Required.  Operation to perform on the ROM or ROM set. {read|write|transform|rominfo|editheaderfield}")
	romSetOrganization := flag.Bool("organization", false, "Read/write relative file location information for automatic organization.")
	romSetPrintChecksums := flag.Bool("print-checksums", false, "Print checksums as ROMs are loaded or processed.")
	romSetTruncateRoms := flag.Bool("truncate-roms", false, "Truncate PRGROM and CHRROM to the sizes specified in the header.")
	romSetPreserveTrainers := flag.Bool("preserve-trainers", false, "Preserve trainers in read/write process.")
	romOutputBasePath := flag.String("rom-output-base-path", "", "The path to use for writing organized NES and/or FDS ROMs.")
	romSetSourceDirectory := flag.String("rom-source-path", "", "Required.  The path to a directory with NES and/or FDS ROMs to use for the operation.")
	romSetXmlFile := flag.String("xml-file", "", "The path to an XML file to use for the operation.")
	xmlFormat := flag.String("xml-format", "default", "The format of the imported or exported XML file. {default|nes20db}")
	formatTransformDestination := flag.String("format-transform-destination", "", "Destination file for format transform operations.")
	formatTransformType := flag.String("format-transform-type", "", "Format of destination file for transform operations. {default|nes20db|sanni}")
	romToAnalyze := flag.String("rom-file", "", "An NES ROM file to analyze with the rominfo operation.")
	inputRom := flag.String("input-rom", "", "The ROM to edit when editing a header field.")
	outputRom := flag.String("output-rom", "", "The ROM to write when editing a header field.")
	romFieldName := flag.String("rom-field-name", "", "The ROM field to edit when editing a header field.")
	romFieldValue := flag.String("rom-field-value", "", "The data to apply to the specified ROM field when editing a header field.")

	flag.Parse()

	// Options validation
	if *romSetCommand != "read" && *romSetCommand != "write" && *romSetCommand != "transform" && *romSetCommand != "rominfo" && *romSetCommand != "editheaderfield" {
		printUsage()
		os.Exit(1)
	}

	if *romSetSourceDirectory == "" && *romSetCommand != "transform" && *romSetCommand != "rominfo" && *romSetCommand != "editheaderfield" {
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

	if *romSetCommand == "transform" && (*formatTransformDestination == "" || *formatTransformType == "") {
		printUsage()
		os.Exit(1)
	}

	if *romSetCommand == "transform" {
		*romSetOrganization = true
	}

	if *romSetCommand == "rominfo" && *romToAnalyze == "" {
		printUsage()
		os.Exit(1)
	}

	if *romSetCommand == "editheaderfield" && (*romFieldName == "" || *romFieldValue == "" || *inputRom == "" || *outputRom == "") {
		printUsage()
		os.Exit(1)
	}

	// Read a directory structure and generate an XML file to represent it
	if *romSetCommand == "read" {
		println("Loading NES 2.0 ROMs from: " + *romSetSourceDirectory)
		romMap, err := FileTools.LoadROMRecursiveMap(*romSetSourceDirectory, *romSetEnableV1, *romSetPreserveTrainers, ProcessingTools.HASH_TYPE_SHA256, *romSetPrintChecksums)
		if err != nil {
			panic(err)
		}

		archiveMap := make(map[string]*FDSTool.FDSArchiveFile, 0)

		if *romSetEnableFDS {
			println("Loading FDS archives from: " + *romSetSourceDirectory)
			archiveMap, err = FileTools.LoadFDSArchiveRecursiveMap(*romSetSourceDirectory, *romSetGenerateFDSCRCs, ProcessingTools.HASH_TYPE_SHA256, *romSetPrintChecksums)
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
			xmlPayload, err = FileTools.MarshalNES20DBXMLFromROMMap(romMap, *romSetOrganization)
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
			romData, err = FileTools.UnmarshalNES20DBXMLToROMMap(string(xmlPayload), *romSetOrganization)
			if err != nil {
				panic(err)
			}

			hashTypeMatch = ProcessingTools.HASH_TYPE_SHA1
		}

		println("Processing NES ROMs in: " + *romSetSourceDirectory)
		rawRoms, err := FileTools.LoadROMRecursive(*romSetSourceDirectory, *romSetEnableV1, *romSetPreserveTrainers, *romSetPrintChecksums)
		if err != nil {
			panic(err)
		}

		matchedRoms := ProcessingTools.ProcessNESROMs(rawRoms, romData, hashTypeMatch, *romSetTruncateRoms, *romSetOrganization, *romSetEnableV1)

		println("Processing UNIF ROMs in: " + *romSetSourceDirectory)
		rawUnifs, err := FileTools.LoadUNIFRecursive(*romSetSourceDirectory, *romSetPrintChecksums)
		if err != nil {
			panic(err)
		}

		matchedUnifs := ProcessingTools.ProcessNESROMs(rawUnifs, romData, hashTypeMatch, *romSetTruncateRoms, *romSetOrganization, *romSetEnableV1)

		matchedRoms = append(matchedRoms, matchedUnifs...)

		rawArchives := make([]*FDSTool.FDSArchiveFile, 0)
		matchedArchives := make([]*FDSTool.FDSArchiveFile, 0)

		if *romSetEnableFDS {
			println("Processing FDS archives in: " + *romSetSourceDirectory)
			rawArchives, err = FileTools.LoadFDSArchiveRecursive(*romSetSourceDirectory, false, *romSetPrintChecksums)
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
	} else if *romSetCommand == "transform" {
		println("Loading XML file from: " + *romSetXmlFile)
		xmlPayload, err := ioutil.ReadFile(*romSetXmlFile)
		if err != nil {
			panic(err)
		}

		println("Reading source data")
		var romData map[string]*NESTool.NESROM
		var archiveData map[string]*FDSTool.FDSArchiveFile

		if *xmlFormat == "default" {
			romData, archiveData, err = FileTools.UnmarshalXMLToROMMap(string(xmlPayload), *romSetEnableV1, *romSetPreserveTrainers, *romOutputBasePath != "")
			if err != nil {
				panic(err)
			}
		} else if *xmlFormat == "nes20db" {
			romData, err = FileTools.UnmarshalNES20DBXMLToROMMap(string(xmlPayload), *romSetOrganization)
			if err != nil {
				panic(err)
			}
		}

		transformPayloadString := ""
		transformPayloadBytes := make([]byte, 0)

		if *formatTransformType == "default" {
			transformPayloadString, err = FileTools.MarshalXMLFromROMMap(romData, archiveData, *romSetEnableV1, *romSetPreserveTrainers, *romSetOrganization)
			if err != nil {
				panic(err)
			}
		} else if *formatTransformType == "nes20db" {
			transformPayloadString, err = FileTools.MarshalNES20DBXMLFromROMMap(romData, *romSetOrganization)
			if err != nil {
				panic(err)
			}
		} else if *formatTransformType == "sanni" {
			transformPayloadBytes, err = FileTools.MarshalDBFileFromROMMap(romData, *romSetEnableV1)
			if err != nil {
				panic(err)
			}
		}

		println("Writing transformed payload to: " + *formatTransformDestination)
		if len(transformPayloadString) > 0 {
			err = FileTools.WriteStringToFile(transformPayloadString, *formatTransformDestination)
		} else if len(transformPayloadBytes) > 0 {
			err = FileTools.WriteBytesToFile(transformPayloadBytes, *formatTransformDestination)
		} else {
			os.Exit(1)
		}
		if err != nil {
			panic(err)
		}

		os.Exit(0)
	} else if *romSetCommand == "rominfo" {
		rom, err := FileTools.LoadROM(*romToAnalyze, true, true, "", false)
		if err != nil {
			panic(err)
		}

		fmt.Println(rom)

		os.Exit(0)
	} else if *romSetCommand == "editheaderfield" {
		inputFileName := filepath.Base(*inputRom)
		inputFilePath := filepath.Dir(*inputRom)
		outputFileName := filepath.Base(*outputRom)
		outputFilePath := filepath.Dir(*outputRom)

		nesRom, err := FileTools.LoadROM(inputFileName, true, true, inputFilePath, false)
		if err != nil {
			panic(err)
		}

		if nesRom == nil {
			println("Unable to read ROM: " + *inputRom)
			os.Exit(1)
		}

		paramInt, strConvErr := strconv.ParseUint(*romFieldValue, 10, 64)

		switch *romFieldName {
		case "prg-rom-byte-size":
			if strConvErr != nil {
				panic(err)
			}

			if nesRom.Header20 != nil {
				err = NESTool.UpdateSizes(nesRom, paramInt, nesRom.Header20.CHRROMCalculatedSize)
			} else if nesRom.Header10 != nil {
				err = NESTool.UpdateSizes(nesRom, paramInt, nesRom.Header10.CHRROMCalculatedSize)
			} else {
				println("No valid ROM found")
				os.Exit(1)
			}

			if err != nil {
				panic(err)
			}
		case "prg-ram-size":
			if strConvErr != nil {
				panic(err)
			}

			if nesRom.Header20 != nil {
				if paramInt > 15 {
					println("For NES 2.0 ROMs, PRG RAM size is calculated as 64*x^X bytes, where X is 0-15")
					os.Exit(1)
				}

				nesRom.Header20.PRGRAMSize = uint8(paramInt)
			} else if nesRom.Header10 != nil {
				if paramInt > 255 {
					println("For iNES ROMs, PRG RAM can have no more than 255 8KB units")
					os.Exit(1)
				}

				nesRom.Header10.PRGRAMSize = uint8(paramInt)
			} else {
				println("No valid ROM found")
				os.Exit(1)
			}
		case "prg-nvram-size":
			if nesRom.Header20 == nil {
				println("PRG NVRAM is only available in NES 2.0 headers")
				os.Exit(1)
			}

			if strConvErr != nil {
				panic(err)
			}

			if paramInt > 15 {
				println("PRG NVRAM size is calculated as 64*x^X bytes, where X is 0-15")
				os.Exit(1)
			}

			nesRom.Header20.PRGNVRAMSize = uint8(paramInt)
		case "chr-rom-byte-size":
			if strConvErr != nil {
				panic(err)
			}

			if nesRom.Header20 != nil {
				err = NESTool.UpdateSizes(nesRom, nesRom.Header20.PRGROMCalculatedSize, paramInt)
			} else if nesRom.Header10 != nil {
				err = NESTool.UpdateSizes(nesRom, nesRom.Header10.PRGROMCalculatedSize, paramInt)
			} else {
				println("No valid ROM found")
				os.Exit(1)
			}

			if err != nil {
				panic(err)
			}
		case "chr-ram-size":
			if nesRom.Header20 == nil {
				println("CHR RAM is only available in NES 2.0 headers")
				os.Exit(1)
			}

			if strConvErr != nil {
				panic(err)
			}

			if paramInt > 255 {
				println("CHR RAM can have no more than 255 8KB units")
				os.Exit(1)
			}

			nesRom.Header20.CHRRAMSize = uint8(paramInt)
		case "chr-nvram-size":
			if nesRom.Header20 == nil {
				println("CHR NVRAM is only available in NES 2.0 headers")
				os.Exit(1)
			}

			if strConvErr != nil {
				panic(err)
			}

			if paramInt > 15 {
				println("CHR NVRAM size is calculated as 64*x^X bytes, where X is 0-15")
				os.Exit(1)
			}

			nesRom.Header20.CHRNVRAMSize = uint8(paramInt)
		case "number-of-misc-roms":
			if nesRom.Header20 == nil {
				println("Misc ROMs are only available in NES 2.0 ROMs")
				os.Exit(1)
			}

			if strConvErr != nil {
				panic(err)
			}

			if paramInt > 3 {
				println("ROM can have no more than 3 misc ROMs")
				os.Exit(1)
			}

			nesRom.Header20.MiscROMs = uint8(paramInt)
		case "has-trainer":
			var hasTrainer bool

			if *romFieldValue == "true" {
				hasTrainer = true
			} else if *romFieldValue == "false" {
				hasTrainer = false
			} else {
				println("has-trainer must be one of {true|false}")
				os.Exit(1)
			}

			if nesRom.Header20 != nil {
				nesRom.Header20.Trainer = hasTrainer
			} else if nesRom.Header10 != nil {
				nesRom.Header10.Trainer = hasTrainer
			} else {
				println("No valid ROM found")
				os.Exit(1)
			}
		case "mirroring-type":
			var mirroringType bool

			if *romFieldValue == "vertical" {
				mirroringType = true
			} else if *romFieldValue == "horizontal" {
				mirroringType = false
			} else {
				println("mirroring-type must be one of {horizontal|vertical}")
				os.Exit(1)
			}

			if nesRom.Header20 != nil {
				nesRom.Header20.MirroringType = mirroringType
			} else if nesRom.Header10 != nil {
				nesRom.Header10.MirroringType = mirroringType
			} else {
				println("No valid ROM found")
				os.Exit(1)
			}
		case "four-screen":
			var hasFourScreen bool

			if *romFieldValue == "true" {
				hasFourScreen = true
			} else if *romFieldValue == "false" {
				hasFourScreen = false
			} else {
				println("four-screen must be one of {true|false}")
				os.Exit(1)
			}

			if nesRom.Header20 != nil {
				nesRom.Header20.FourScreen = hasFourScreen
			} else if nesRom.Header10 != nil {
				nesRom.Header10.FourScreen = hasFourScreen
			} else {
				println("No valid ROM found")
				os.Exit(1)
			}
		case "has-battery":
			var hasBattery bool

			if *romFieldValue == "true" {
				hasBattery = true
			} else if *romFieldValue == "false" {
				hasBattery = false
			} else {
				println("has-battery must be one of {true|false}")
				os.Exit(1)
			}

			if nesRom.Header20 != nil {
				nesRom.Header20.Battery = hasBattery
			} else if nesRom.Header10 != nil {
				nesRom.Header10.Battery = hasBattery
			} else {
				println("No valid ROM found")
				os.Exit(1)
			}
		case "console-type":
			if nesRom.Header20 == nil {
				println("Console Type is only available in NES 2.0 headers")
				os.Exit(1)
			}

			if strConvErr != nil {
				panic(err)
			}

			if paramInt > 3 {
				println("Console Type must be a value from 0-3")
				os.Exit(1)
			} else if paramInt == 1 {
				nesRom.Header20.ExtendedConsoleType = 0
			} else if paramInt == 3 {
				nesRom.Header20.VsHardwareType = 0
				nesRom.Header20.VsPPUType = 0
			} else {
				nesRom.Header20.ExtendedConsoleType = 0
				nesRom.Header20.VsHardwareType = 0
				nesRom.Header20.VsPPUType = 0
			}

			nesRom.Header20.ConsoleType = uint8(paramInt)
		case "extended-console-type":
			if nesRom.Header20 == nil {
				println("Extended Console Type is only available in NES 2.0 headers")
				os.Exit(1)
			}

			if nesRom.Header20.ConsoleType != 3 {
				println("Extended Console Type is only valid when Console Type is 3")
				os.Exit(1)
			}

			if strConvErr != nil {
				panic(err)
			}

			if paramInt < 3 || paramInt > 15 {
				println("Extended Console Type must be a value from 3-15")
				os.Exit(1)
			}

			nesRom.Header20.ExtendedConsoleType = uint8(paramInt)
		case "mapper-number":
			if strConvErr != nil {
				panic(err)
			}

			if nesRom.Header20 != nil {
				if paramInt > 4095 {
					println("For NES 2.0 ROMs, mapper must be from 0-4095")
					os.Exit(1)
				}
				nesRom.Header20.Mapper = uint16(paramInt)
			} else if nesRom.Header10 != nil {
				if paramInt > 255 {
					println("For iNES ROMs, mapper must be from 0-255")
					os.Exit(1)
				}
				nesRom.Header10.Mapper = uint8(paramInt)
			} else {
				println("No valid ROM found")
				os.Exit(1)
			}

			if err != nil {
				panic(err)
			}
		case "submapper-number":
			if nesRom.Header20 == nil {
				println("Submappers are only available in NES 2.0 headers")
				os.Exit(1)
			}

			if strConvErr != nil {
				panic(err)
			}

			if paramInt > 15 {
				println("Submapper must be from 0-15")
				os.Exit(1)
			}

			nesRom.Header20.SubMapper = uint8(paramInt)
		case "cpu-ppu-timing":
			if nesRom.Header20 == nil {
				println("CPU/PPU Timing is only available in NES 2.0 headers")
				os.Exit(1)
			}

			if strConvErr != nil {
				panic(err)
			}

			if paramInt > 3 {
				println("CPU/PPU Timing must be from 0-3")
				os.Exit(1)
			}

			nesRom.Header20.CPUPPUTiming = uint8(paramInt)
		case "vs-hardware-type":
			if nesRom.Header20 == nil {
				println("Vs. Hardware Type is only available in NES 2.0 headers")
				os.Exit(1)
			}

			if nesRom.Header20.ConsoleType != 1 {
				println("Vs. Hardware Type is only valid when Console Type is 1")
				os.Exit(1)
			}

			if strConvErr != nil {
				panic(err)
			}

			if paramInt > 15 {
				println("Vs. Hardware Type must be from 0-15")
				os.Exit(1)
			}

			nesRom.Header20.VsHardwareType = uint8(paramInt)
		case "vs-ppu-type":
			if nesRom.Header20 == nil {
				println("Vs. PPU Type is only available in NES 2.0 headers")
				os.Exit(1)
			}

			if nesRom.Header20.ConsoleType != 1 {
				println("Vs. PPU Type is only valid when Console Type is 1")
				os.Exit(1)
			}

			if strConvErr != nil {
				panic(err)
			}

			if paramInt > 15 {
				println("Vs. PPU Type must be from 0-15")
				os.Exit(1)
			}

			nesRom.Header20.VsPPUType = uint8(paramInt)
		case "default-expansion":
			if nesRom.Header20 == nil {
				println("Default Hardware Expansion is only available in NES 2.0 headers")
				os.Exit(1)
			}

			if strConvErr != nil {
				panic(err)
			}

			if paramInt > 63 {
				println("Default Hardware Expansion must be from 0-63")
				os.Exit(1)
			}

			nesRom.Header20.DefaultExpansion = uint8(paramInt)
		case "vs-unisystem":
			if nesRom.Header10 == nil {
				println("Vs. Unisystem is only available in iNES headers")
				os.Exit(1)
			}

			var isVsUnisystem bool

			if *romFieldValue == "true" {
				isVsUnisystem = true
			} else if *romFieldValue == "false" {
				isVsUnisystem = false
			} else {
				println("vs-unisystem must be one of {true|false}")
				os.Exit(1)
			}

			nesRom.Header10.VsUnisystem = isVsUnisystem
		case "playchoice-10":
			if nesRom.Header10 == nil {
				println("PlayChoice 10 is only available in iNES headers")
				os.Exit(1)
			}

			var isPlayChoice10 bool

			if *romFieldValue == "true" {
				isPlayChoice10 = true
			} else if *romFieldValue == "false" {
				isPlayChoice10 = false
			} else {
				println("playchoice-10 must be one of {true|false}")
				os.Exit(1)
			}

			nesRom.Header10.PlayChoice10 = isPlayChoice10
		case "tv-system":
			if nesRom.Header10 == nil {
				println("TV System is only available in iNES headers")
				os.Exit(1)
			}

			var tvSystem bool

			if *romFieldValue == "pal" {
				tvSystem = true
			} else if *romFieldValue == "ntsc" {
				tvSystem = false
			} else {
				println("tv-system must be one of {ntsc|pal}")
				os.Exit(1)
			}

			nesRom.Header10.TVSystem = tvSystem
		default:
			printUsage()
			os.Exit(1)
		}

		nesRom.Name = outputFileName[:strings.LastIndex(outputFileName, ".")]
		nesRom.RelativePath = outputFileName

		err = FileTools.WriteROM(nesRom, true, false, true, outputFilePath)
		if err != nil {
			panic(err)
		}

		println("Finished writing " + *outputRom)
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
	println("")
	printFieldEditOptions()
}

// Show field edit usage options.
func printFieldEditOptions() {
	println("Valid fields for editing.")
	println("Descriptions marked with a * are NES 2.0-only fields.")
	println("Descriptions marked with a ! are iNES-only fields.")
	println("Values for assigned numbers can be found at:")
	println("https://www.nesdev.org/wiki/INES")
	println("https://www.nesdev.org/wiki/NES_2.0")
	println("")
	println("prg-rom-byte-size     :   The size of the PRG ROM in bytes")
	println("prg-ram-size          :   The size of the PRG RAM")
	println("                          For NES 2.0 ROMs, this is 64*2^X bytes, where X is the")
	println("                          number passed in as a parameter (0-15)")
	println("                          For iNES ROMs, this is the number of 8KB units (0-255)")
	println("prg-nvram-size        : * The number of bytes of PRG NVRAM, calculated as 64*2^X")
	println("                          where X is 0-15")
	println("chr-rom-byte-size     :   The size of the CHR ROM in bytes")
	println("chr-ram-size          : * The size of the CHR RAM as 64*2^X bytes (0-15)")
	println("chr-nvram-size        : * The number of bytes of CHR NVRAM, calculated as 64*2^X")
	println("                          where X is 0-15")
	println("number-of-misc-roms   : * The number of misc ROMs in the misc ROM blob (0-3)")
	println("has-trainer           :   Whether the ROM has a trainer {true|false}")
	println("mirroring-type        :   Nametable mirroring type {horizontal|vertical}")
	println("four-screen           :   Game uses four-screen mode {true|false}")
	println("has-battery           :   Whether the game has a battery")
	println("                          for save RAM {true|false}")
	println("console-type          : * Which console type the ROM is intended for (0-3)")
	println("extended-console-type : * Extended console type (3-15)")
	println("mapper-number         :   The mapper number for the ROM to use")
	println("                          (0-4095 for NES 2.0 ROMs and 0-255 for iNES ROMs)")
	println("submapper-number      : * The submapper number for the ROM to use (0-15)")
	println("cpu-ppu-timing        : * The CPU/PPU timing mode to use (0-3)")
	println("vs-hardware-type      : * The hardware type of the Vs. system (0-15)")
	println("vs-ppu-type           : * The PPU in the Vs. system (0-15)")
	println("default-expansion     : * The default hardware expansion to use (0-63)")
	println("vs-unisystem          : ! Whether the ROM is for a")
	println("                          Vs. system {true|false}")
	println("playchoice-10         : ! Whether the ROM is for a")
	println("                          PlayChoice 10 system {true|false}")
	println("tv-system             : ! The TV system the ROM is intended for {ntsc|pal}")
}
