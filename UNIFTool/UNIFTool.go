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

// https://raw.githubusercontent.com/eteran/libunif/master/UNIF_current.txt
// This uses the most recent UNIF specification linked above as a reference.

package UNIFTool

import (
	"NES20Tool/NESTool"
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"strconv"
	"strings"
)

var (
	UNIF_MAGIC = "UNIF"
)

// Read a byte slice and copy the PRG, CHR, and base ROMs
// to an NESROM struct, then hash them.
func DecodeUNIFROM(inputFile []byte) (*NESTool.NESROM, error) {
	unifChunks, err := GetUNIFChunks(inputFile)
	if err != nil {
		return nil, err
	}

	unifVersion, err := GetUNIFVersion(inputFile)
	if err != nil {
		return nil, err
	}

	prgRomData, err := getRomData(unifChunks, unifVersion, "PRG", "PCK")
	if err != nil {
		return nil, err
	}

	chrRomData, err := getRomData(unifChunks, unifVersion, "CHR", "CCK")
	if err != nil {
		return nil, err
	}

	tempRom := &NESTool.NESROM{}

	if IsValidChunkNameForUnifVersion(unifVersion, "NAME") && unifChunks["NAME"] != nil {
		tempRom.Name = string(unifChunks["NAME"][0 : len(unifChunks["NAME"])-1])
	}

	tempRom.PRGROMData = prgRomData
	tempRom.CHRROMData = chrRomData
	tempRom.ROMData = append(prgRomData, chrRomData...)
	tempRom.Header20 = &NESTool.NES20Header{}

	err = NESTool.UpdateSizes(tempRom, NESTool.PRG_CANONICAL_SIZE_ROM, NESTool.CHR_CANONICAL_SIZE_ROM)
	if err != nil {
		return nil, err
	}

	err = NESTool.UpdateChecksums(tempRom)
	if err != nil {
		return nil, err
	}

	return tempRom, nil
}

// Validate UNIF header
func IsValidUNIFROM(inputData []byte) (bool, error) {
	// UNIF header
	if bytes.Compare(inputData[0:4], []byte(UNIF_MAGIC)) != 0 {
		return false, &NESTool.NESROMError{Text: "Not a valid UNIF ROM."}
	}

	return true, nil
}

// Get the UNIF standard version declared by the ROM
func GetUNIFVersion(inputData []byte) (uint32, error) {
	_, err := IsValidUNIFROM(inputData)
	if err != nil {
		return 0, err
	}

	// "DWORD" (32-bit little endian unsigned integer) version number
	return binary.LittleEndian.Uint32(inputData[4:8]), nil
}

// Get each of the UNIF chunks in the file
func GetUNIFChunks(inputData []byte) (map[string][]byte, error) {
	_, err := IsValidUNIFROM(inputData)
	if err != nil {
		return nil, err
	}

	unifChunks := make(map[string][]byte, 0)

	// Skipping the four-byte UNIF version number and the reserved bytes
	romPosition := uint64(32)

	for romPosition < uint64(len(inputData)) {
		chunkId, chunkData, tempPosition, err := getNextUNIFChunk(inputData, romPosition)
		if err != nil {
			return nil, err
		}

		unifChunks[chunkId] = chunkData

		romPosition = tempPosition
	}

	return unifChunks, nil
}

// Get the next UNIF chunk and return the name, the chunk data, and the offset to the next chunk.
func getNextUNIFChunk(inputData []byte, position uint64) (string, []byte, uint64, error) {
	if (position + 7) > uint64(len(inputData)) {
		return "", nil, 0, &NESTool.NESROMError{Text: "Not enough data to decode UNIF chunk."}
	}

	tempPosition := position

	// Workaround for malformed ROMs
	isMalformedROM := false

	// Skip the four-byte UNIF version number and the reserved bytes if at the beginning of the file
	if bytes.Compare(inputData[position:(position+4)], []byte(UNIF_MAGIC)) == 0 {
		tempPosition = 32
	}

	// Four-byte UTF-8 chunk ID
	chunkId := string(inputData[tempPosition:(tempPosition + 4)])

	// Workaround for malformed ROMs
	if chunkId == "\000DIN" {
		tempPosition = tempPosition + 1
		isMalformedROM = true
		chunkId = string(inputData[tempPosition:(tempPosition + 4)])
	}

	// "DWORD" (32-bit little endian unsigned integer) chunk length
	chunkLength := uint64(binary.LittleEndian.Uint32(inputData[(tempPosition + 4):(tempPosition + 8)]))

	// Workaround for malformed ROMs
	if isMalformedROM && chunkId == "DINF" && chunkLength == 0 {
		chunkLength = uint64(204)
	}

	if (tempPosition + 8 + chunkLength) > uint64(len(inputData)) {
		return "", nil, 0, &NESTool.NESROMError{Text: "Invalid UNIF chunk length."}
	}

	chunkData := inputData[(tempPosition + 8):(tempPosition + 8 + chunkLength)]

	return chunkId, chunkData, tempPosition + 8 + chunkLength, nil
}

// Build a PRG or CHR ROM from a map of UNIF chunks
func getRomData(unifChunks map[string][]byte, unifVersion uint32, romType string, checksumType string) ([]byte, error) {
	romData := make([]byte, 0)

	// PRG and CHR chunks are named PRG0, PRG1, etc. through PRGF.
	// Not all of these chunks necessarily exist and they're not
	// necessarily going to be in any particular order in the actual
	// UNIF file, but the complete PRG and/or CHR ROMs are the
	// concatenation of them all in numerical order by chunk name.
	// Additionally, checksums are not required to be included, so
	// we can only hard fail if one exists for a chunk and is a mismatch.
	for i := 0; i < 16; i++ {
		chunkStr := strings.ToUpper(strconv.FormatInt(int64(i), 16))
		if IsValidChunkNameForUnifVersion(unifVersion, romType+chunkStr) && unifChunks[romType+chunkStr] != nil {
			if IsValidChunkNameForUnifVersion(unifVersion, checksumType+chunkStr) && unifChunks[checksumType+chunkStr] != nil {
				calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+chunkStr])
				referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+chunkStr])

				if calculatedCrc32 != referenceCrc32 {
					return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + chunkStr}
				}
			}
			romData = append(romData, unifChunks[romType+chunkStr]...)
		}
	}

	return romData, nil
}

func IsValidChunkNameForUnifVersion(unifVersion uint32, chunkName string) bool {
	chunkNames := GetValidChunkNamesForUnifVersion(unifVersion)

	for _, testChunkName := range chunkNames {
		if testChunkName == chunkName {
			return true
		}
	}

	return false
}

func GetValidChunkNamesForUnifVersion(unifVersion uint32) []string {
	chunkNames := make([]string, 0)
	if unifVersion == 0 {
		return chunkNames
	}

	if unifVersion >= 1 {
		chunkNames = append(chunkNames, "MAPR")
		chunkNames = append(chunkNames, "READ")
		chunkNames = append(chunkNames, "NAME")
	}

	if unifVersion >= 2 {
		chunkNames = append(chunkNames, "DINF")

		// Workaround for malformed ROMs
		chunkNames = append(chunkNames, "\000DIN")
	}

	if unifVersion >= 4 {
		chunkNames = append(chunkNames, "PRG0")
		chunkNames = append(chunkNames, "PRG1")
		chunkNames = append(chunkNames, "PRG2")
		chunkNames = append(chunkNames, "PRG3")
		chunkNames = append(chunkNames, "PRG4")
		chunkNames = append(chunkNames, "PRG5")
		chunkNames = append(chunkNames, "PRG6")
		chunkNames = append(chunkNames, "PRG7")
		chunkNames = append(chunkNames, "PRG8")
		chunkNames = append(chunkNames, "PRG9")
		chunkNames = append(chunkNames, "PRGA")
		chunkNames = append(chunkNames, "PRGB")
		chunkNames = append(chunkNames, "PRGC")
		chunkNames = append(chunkNames, "PRGD")
		chunkNames = append(chunkNames, "PRGE")
		chunkNames = append(chunkNames, "PRGF")
		chunkNames = append(chunkNames, "CHR0")
		chunkNames = append(chunkNames, "CHR1")
		chunkNames = append(chunkNames, "CHR2")
		chunkNames = append(chunkNames, "CHR3")
		chunkNames = append(chunkNames, "CHR4")
		chunkNames = append(chunkNames, "CHR5")
		chunkNames = append(chunkNames, "CHR6")
		chunkNames = append(chunkNames, "CHR7")
		chunkNames = append(chunkNames, "CHR8")
		chunkNames = append(chunkNames, "CHR9")
		chunkNames = append(chunkNames, "CHRA")
		chunkNames = append(chunkNames, "CHRB")
		chunkNames = append(chunkNames, "CHRC")
		chunkNames = append(chunkNames, "CHRD")
		chunkNames = append(chunkNames, "CHRE")
		chunkNames = append(chunkNames, "CHRF")
	}

	if unifVersion >= 5 {
		chunkNames = append(chunkNames, "PCK0")
		chunkNames = append(chunkNames, "PCK1")
		chunkNames = append(chunkNames, "PCK2")
		chunkNames = append(chunkNames, "PCK3")
		chunkNames = append(chunkNames, "PCK4")
		chunkNames = append(chunkNames, "PCK5")
		chunkNames = append(chunkNames, "PCK6")
		chunkNames = append(chunkNames, "PCK7")
		chunkNames = append(chunkNames, "PCK8")
		chunkNames = append(chunkNames, "PCK9")
		chunkNames = append(chunkNames, "PCKA")
		chunkNames = append(chunkNames, "PCKB")
		chunkNames = append(chunkNames, "PCKC")
		chunkNames = append(chunkNames, "PCKD")
		chunkNames = append(chunkNames, "PCKE")
		chunkNames = append(chunkNames, "PCKF")
		chunkNames = append(chunkNames, "CCK0")
		chunkNames = append(chunkNames, "CCK1")
		chunkNames = append(chunkNames, "CCK2")
		chunkNames = append(chunkNames, "CCK3")
		chunkNames = append(chunkNames, "CCK4")
		chunkNames = append(chunkNames, "CCK5")
		chunkNames = append(chunkNames, "CCK6")
		chunkNames = append(chunkNames, "CCK7")
		chunkNames = append(chunkNames, "CCK8")
		chunkNames = append(chunkNames, "CCK9")
		chunkNames = append(chunkNames, "CCKA")
		chunkNames = append(chunkNames, "CCKB")
		chunkNames = append(chunkNames, "CCKC")
		chunkNames = append(chunkNames, "CCKD")
		chunkNames = append(chunkNames, "CCKE")
		chunkNames = append(chunkNames, "CCKF")
		chunkNames = append(chunkNames, "BATR")
		chunkNames = append(chunkNames, "VROR")
		chunkNames = append(chunkNames, "MIRR")
	}

	if unifVersion >= 7 {
		chunkNames = append(chunkNames, "CTRL")
	}

	return chunkNames
}
