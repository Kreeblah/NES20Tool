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

// https://raw.githubusercontent.com/eteran/libunif/master/UNIF_current.txt
// This uses the most recent UNIF specification linked above as a reference.

package UNIFTool

import (
	"NES20Tool/NES20Tool"
	"bytes"
	"encoding/binary"
)

var (
	UNIF_MAGIC = "UNIF"
)

// Read a byte slice and copy the PRG, CHR, and base ROMs
// to an NESROM struct, then hash them.
func DecodeUNIFROM(inputFile []byte) (*NES20Tool.NESROM, error) {
	// UNIF header
	if bytes.Compare(inputFile[0:4], []byte(UNIF_MAGIC)) != 0 {
		return nil, &NES20Tool.NESROMError{Text: "Not a valid UNIF ROM."}
	}

	unifChunks, err := GetUNIFChunks(inputFile)
	if err != nil {
		return nil, err
	}

	prgRomData := getRomData(unifChunks, "PRG")
	chrRomData := getRomData(unifChunks, "CHR")

	tempRom := &NES20Tool.NESROM{}

	tempRom.PRGROMData = prgRomData
	tempRom.CHRROMData = chrRomData
	tempRom.ROMData = append(prgRomData, chrRomData...)
	tempRom.Header20 = &NES20Tool.NES20Header{}

	err = NES20Tool.UpdateSizes(tempRom, NES20Tool.PRG_CANONICAL_SIZE_ROM, NES20Tool.CHR_CANONICAL_SIZE_ROM)
	if err != nil {
		return nil, err
	}

	err = NES20Tool.UpdateChecksums(tempRom)
	if err != nil {
		return nil, err
	}

	return tempRom, nil
}

// Get each of the UNIF chunks in the file
func GetUNIFChunks(inputData []byte) (map[string][]byte, error) {
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
		return "", nil, 0, &NES20Tool.NESROMError{Text: "Not enough data to decode UNIF chunk."}
	}

	tempPosition := position

	// Skip the four-byte UNIF version number and the reserved bytes if at the beginning of the file
	if bytes.Compare(inputData[position: (position + 4)], []byte(UNIF_MAGIC)) == 0 {
		tempPosition = 32
	}

	// Four-byte UTF-8 chunk ID
	chunkId := string(inputData[tempPosition: (tempPosition + 4)])

	// "DWORD" (32-bit little endian unsigned integer) chunk length
	chunkLength := uint64(binary.LittleEndian.Uint32(inputData[(tempPosition + 4) : (tempPosition + 8)]))

	if (tempPosition + 8 + chunkLength) > uint64(len(inputData)) {
		return "", nil, 0, &NES20Tool.NESROMError{Text: "Invalid UNIF chunk length."}
	}

	chunkData := inputData[(tempPosition + 8) : (tempPosition + 8 + chunkLength)]

	return chunkId, chunkData, tempPosition + 8 + chunkLength, nil
}

// Build a PRG or CHR ROM from a map of UNIF chunks
func getRomData(unifChunks map[string][]byte, romType string) []byte {
	romData := make([]byte, 0)

	// PRG and CHR chunks are named PRG0, PRG1, etc. through PRGF.
	// Not all of these chunks necessarily exist, but the complete
	// PRG and/or CHR ROMs are the concatenation of them all in
	// numerical order.
	if unifChunks[romType + "0"] != nil {
		romData = append(romData, unifChunks[romType + "0"]...)
	}

	if unifChunks[romType + "1"] != nil {
		romData = append(romData, unifChunks[romType + "1"]...)
	}

	if unifChunks[romType + "2"] != nil {
		romData = append(romData, unifChunks[romType + "2"]...)
	}

	if unifChunks[romType + "3"] != nil {
		romData = append(romData, unifChunks[romType + "3"]...)
	}

	if unifChunks[romType + "4"] != nil {
		romData = append(romData, unifChunks[romType + "4"]...)
	}

	if unifChunks[romType + "5"] != nil {
		romData = append(romData, unifChunks[romType + "5"]...)
	}

	if unifChunks[romType + "6"] != nil {
		romData = append(romData, unifChunks[romType + "6"]...)
	}

	if unifChunks[romType + "7"] != nil {
		romData = append(romData, unifChunks[romType + "7"]...)
	}

	if unifChunks[romType + "8"] != nil {
		romData = append(romData, unifChunks[romType + "8"]...)
	}

	if unifChunks[romType + "9"] != nil {
		romData = append(romData, unifChunks[romType + "9"]...)
	}

	if unifChunks[romType + "A"] != nil {
		romData = append(romData, unifChunks[romType + "A"]...)
	}

	if unifChunks[romType + "B"] != nil {
		romData = append(romData, unifChunks[romType + "B"]...)
	}

	if unifChunks[romType + "C"] != nil {
		romData = append(romData, unifChunks[romType + "C"]...)
	}

	if unifChunks[romType + "D"] != nil {
		romData = append(romData, unifChunks[romType + "D"]...)
	}

	if unifChunks[romType + "E"] != nil {
		romData = append(romData, unifChunks[romType + "E"]...)
	}

	if unifChunks[romType + "F"] != nil {
		romData = append(romData, unifChunks[romType + "F"]...)
	}

	return romData
}