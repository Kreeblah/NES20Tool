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

// https://raw.githubusercontent.com/eteran/libunif/master/UNIF_current.txt
// This uses the most recent UNIF specification linked above as a reference.

package UNIFTool

import (
	"NES20Tool/NESTool"
	"bytes"
	"encoding/binary"
	"hash/crc32"
)

var (
	UNIF_MAGIC = "UNIF"
)

// Read a byte slice and copy the PRG, CHR, and base ROMs
// to an NESROM struct, then hash them.
func DecodeUNIFROM(inputFile []byte) (*NESTool.NESROM, error) {
	// UNIF header
	if bytes.Compare(inputFile[0:4], []byte(UNIF_MAGIC)) != 0 {
		return nil, &NESTool.NESROMError{Text: "Not a valid UNIF ROM."}
	}

	unifChunks, err := GetUNIFChunks(inputFile)
	if err != nil {
		return nil, err
	}

	prgRomData, err := getRomData(unifChunks, "PRG", "PCK")
	if err != nil {
		return nil, err
	}

	chrRomData, err := getRomData(unifChunks, "CHR", "CCK")
	if err != nil {
		return nil, err
	}

	tempRom := &NESTool.NESROM{}

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

// Get each of the UNIF chunks in the file
func GetUNIFChunks(inputData []byte) (map[string][]byte, error) {
	// UNIF header
	if bytes.Compare(inputData[0:4], []byte(UNIF_MAGIC)) != 0 {
		return nil, &NESTool.NESROMError{Text: "Not a valid UNIF ROM."}
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

	// Skip the four-byte UNIF version number and the reserved bytes if at the beginning of the file
	if bytes.Compare(inputData[position:(position+4)], []byte(UNIF_MAGIC)) == 0 {
		tempPosition = 32
	}

	// Four-byte UTF-8 chunk ID
	chunkId := string(inputData[tempPosition:(tempPosition + 4)])

	// "DWORD" (32-bit little endian unsigned integer) chunk length
	chunkLength := uint64(binary.LittleEndian.Uint32(inputData[(tempPosition + 4):(tempPosition + 8)]))

	if (tempPosition + 8 + chunkLength) > uint64(len(inputData)) {
		return "", nil, 0, &NESTool.NESROMError{Text: "Invalid UNIF chunk length."}
	}

	chunkData := inputData[(tempPosition + 8):(tempPosition + 8 + chunkLength)]

	return chunkId, chunkData, tempPosition + 8 + chunkLength, nil
}

// Build a PRG or CHR ROM from a map of UNIF chunks
func getRomData(unifChunks map[string][]byte, romType string, checksumType string) ([]byte, error) {
	romData := make([]byte, 0)

	// PRG and CHR chunks are named PRG0, PRG1, etc. through PRGF.
	// Not all of these chunks necessarily exist and they're not
	// necessarily going to be in any particular order in the actual
	// UNIF file, but the complete PRG and/or CHR ROMs are the
	// concatenation of them all in numerical order by chunk name.
	// Additionally, checksums are not required to be included, so
	// we can only hard fail if one exists for a chunk and is a mismatch.
	if unifChunks[romType+"0"] != nil {
		if unifChunks[checksumType+"0"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"0"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"0"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "0"}
			}
		}
		romData = append(romData, unifChunks[romType+"0"]...)
	}

	if unifChunks[romType+"1"] != nil {
		if unifChunks[checksumType+"1"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"1"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"1"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "1"}
			}
		}
		romData = append(romData, unifChunks[romType+"1"]...)
	}

	if unifChunks[romType+"2"] != nil {
		if unifChunks[checksumType+"2"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"2"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"2"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "2"}
			}
		}
		romData = append(romData, unifChunks[romType+"2"]...)
	}

	if unifChunks[romType+"3"] != nil {
		if unifChunks[checksumType+"3"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"3"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"3"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "3"}
			}
		}
		romData = append(romData, unifChunks[romType+"3"]...)
	}

	if unifChunks[romType+"4"] != nil {
		if unifChunks[checksumType+"4"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"4"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"4"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "4"}
			}
		}
		romData = append(romData, unifChunks[romType+"4"]...)
	}

	if unifChunks[romType+"5"] != nil {
		if unifChunks[checksumType+"5"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"5"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"5"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "5"}
			}
		}
		romData = append(romData, unifChunks[romType+"5"]...)
	}

	if unifChunks[romType+"6"] != nil {
		if unifChunks[checksumType+"6"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"6"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"6"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "6"}
			}
		}
		romData = append(romData, unifChunks[romType+"6"]...)
	}

	if unifChunks[romType+"7"] != nil {
		if unifChunks[checksumType+"7"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"7"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"7"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "7"}
			}
		}
		romData = append(romData, unifChunks[romType+"7"]...)
	}

	if unifChunks[romType+"8"] != nil {
		if unifChunks[checksumType+"8"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"8"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"8"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "8"}
			}
		}
		romData = append(romData, unifChunks[romType+"8"]...)
	}

	if unifChunks[romType+"9"] != nil {
		if unifChunks[checksumType+"9"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"9"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"9"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "9"}
			}
		}
		romData = append(romData, unifChunks[romType+"9"]...)
	}

	if unifChunks[romType+"A"] != nil {
		if unifChunks[checksumType+"A"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"A"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"A"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "A"}
			}
		}
		romData = append(romData, unifChunks[romType+"A"]...)
	}

	if unifChunks[romType+"B"] != nil {
		if unifChunks[checksumType+"B"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"B"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"B"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "B"}
			}
		}
		romData = append(romData, unifChunks[romType+"B"]...)
	}

	if unifChunks[romType+"C"] != nil {
		if unifChunks[checksumType+"C"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"C"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"C"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "C"}
			}
		}
		romData = append(romData, unifChunks[romType+"C"]...)
	}

	if unifChunks[romType+"D"] != nil {
		if unifChunks[checksumType+"D"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"D"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"D"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "D"}
			}
		}
		romData = append(romData, unifChunks[romType+"D"]...)
	}

	if unifChunks[romType+"E"] != nil {
		if unifChunks[checksumType+"E"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"E"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"E"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "E"}
			}
		}
		romData = append(romData, unifChunks[romType+"E"]...)
	}

	if unifChunks[romType+"F"] != nil {
		if unifChunks[checksumType+"F"] != nil {
			calculatedCrc32 := crc32.ChecksumIEEE(unifChunks[romType+"F"])
			referenceCrc32 := binary.LittleEndian.Uint32(unifChunks[checksumType+"F"])

			if calculatedCrc32 != referenceCrc32 {
				return nil, &NESTool.NESROMError{Text: "Checksum mismatch for chunk " + romType + "F"}
			}
		}
		romData = append(romData, unifChunks[romType+"F"]...)
	}

	return romData, nil
}
