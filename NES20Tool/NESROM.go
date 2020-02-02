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
package NES20Tool

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"hash/crc32"
)

var (
	NES_HEADER_MAGIC = "\x4e\x45\x53\x1a"
	NES_20_AND_MASK  = byte(0x08)
	NES_20_OR_MASK   = byte(0xFB)
)

type NESHeader struct {
	PRGROMSize          uint16
	CHRROMSize          uint16
	PRGRAMSize          uint8
	PRGNVRAMSize        uint8
	CHRRAMSize          uint8
	CHRNVRAMSize        uint8
	MirroringType       bool
	Battery             bool
	Trainer             bool
	FourScreen          bool
	ConsoleType         uint8
	Mapper              uint16
	SubMapper           uint8
	CPUPPUTiming        uint8
	VsHardwareType      uint8
	VsPPUType           uint8
	ExtendedConsoleType uint8
	MiscROMs            uint8
	DefaultExpansion    uint8
}

type NESROM struct {
	Name     string
	Filename string
	Size     uint64
	CRC32    uint32
	MD5      [16]byte
	SHA1     [20]byte
	SHA256   [32]byte
	Header   *NESHeader
	ROMData  []byte
}

type NESROMError struct {
	text string
}

func (r *NESROMError) Error() string {
	return r.text
}

func DecodeNESROM(inputFile []byte) (*NESROM, error) {
	fileSize := uint64(len(inputFile))
	rawROMBytes, _ := getStrippedRom(inputFile)

	romData := &NESROM{}
	romData.CRC32 = crc32.ChecksumIEEE(rawROMBytes)
	romData.MD5 = md5.Sum(rawROMBytes)
	romData.SHA1 = sha1.Sum(rawROMBytes)
	romData.SHA256 = sha256.Sum256(rawROMBytes)
	romData.ROMData = rawROMBytes

	if fileSize > 15 {
		romData.Size = fileSize - 16
	} else {
		romData.Size = fileSize
		return romData, &NESROMError{text: "File too small to be a headered NES ROM."}
	}

	if bytes.Compare(inputFile[0:4], []byte(NES_HEADER_MAGIC)) != 0 {
		return romData, &NESROMError{text: "Unable to find NES magic."}
	}

	if (inputFile[7]&NES_20_AND_MASK) != NES_20_AND_MASK || (inputFile[7]|NES_20_OR_MASK) != NES_20_OR_MASK {
		return romData, &NESROMError{text: "Not an NES 2.0 ROM."}
	}

	headerData := &NESHeader{}

	PRGROMBytes := make([]byte, 2)
	PRGROMBytes[0] = inputFile[4]
	PRGROMBytes[1] = inputFile[9] & 0b00001111
	headerData.PRGROMSize = binary.LittleEndian.Uint16(PRGROMBytes)

	CHRROMBytes := make([]byte, 2)
	CHRROMBytes[0] = inputFile[5]
	CHRROMBytes[1] = (inputFile[9] & 0b11110000) >> 4
	headerData.CHRROMSize = binary.LittleEndian.Uint16(CHRROMBytes)

	headerData.PRGRAMSize = inputFile[10] & 0b00001111
	headerData.PRGNVRAMSize = (inputFile[10] & 0b11110000) >> 4
	headerData.CHRRAMSize = inputFile[11] & 0b00001111
	headerData.CHRNVRAMSize = (inputFile[11] & 0b11110000) >> 4

	headerData.MirroringType = (inputFile[6] & 0b00000001) == 0b00000001
	headerData.Battery = (inputFile[6] & 0b00000010) == 0b00000010
	headerData.Trainer = (inputFile[6] & 0b00000100) == 0b00000100
	headerData.FourScreen = (inputFile[6] & 0b00001000) == 0b00001000
	headerData.ConsoleType = inputFile[7] & 0b00000011

	MapperBytes := make([]byte, 2)
	MapperBytes[0] = ((inputFile[6] & 0b11110000) >> 4) | (inputFile[7] & 0b11110000)
	MapperBytes[1] = inputFile[8] & 0b00001111

	headerData.Mapper = binary.LittleEndian.Uint16(MapperBytes)
	headerData.SubMapper = (inputFile[8] & 0b11110000) >> 4
	headerData.CPUPPUTiming = inputFile[12] & 0b00000011

	if headerData.ConsoleType == 0 {
		headerData.VsHardwareType = 0
		headerData.VsPPUType = 0
		headerData.ExtendedConsoleType = 0
	} else if headerData.ConsoleType == 1 {
		headerData.VsHardwareType = (inputFile[13] & 0b11110000) >> 4
		headerData.VsPPUType = inputFile[13] & 0b00001111
		headerData.ExtendedConsoleType = 0
	} else if headerData.ConsoleType == 3 {
		headerData.VsHardwareType = 0
		headerData.VsPPUType = 0
		headerData.ExtendedConsoleType = inputFile[13] & 0b00001111
	} else {
		headerData.VsHardwareType = 0
		headerData.VsPPUType = 0
		headerData.ExtendedConsoleType = 0
	}

	headerData.MiscROMs = inputFile[14] & 0b00000011
	headerData.DefaultExpansion = inputFile[15] & 0b00111111

	romData.Header = headerData

	return romData, nil
}

func EncodeNESROM(romModel *NESROM) ([]byte, error) {
	headerBytes := make([]byte, 16)
	for index, _ := range NES_HEADER_MAGIC {
		headerBytes[index] = NES_HEADER_MAGIC[index]
	}

	PRGROMBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(PRGROMBytes, romModel.Header.PRGROMSize)
	headerBytes[4] = PRGROMBytes[0]

	CHRROMBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(CHRROMBytes, romModel.Header.CHRROMSize)
	headerBytes[5] = CHRROMBytes[0]

	MapperBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(MapperBytes, romModel.Header.Mapper)

	var Flag6Byte byte = 0b00000000
	if romModel.Header.MirroringType {
		Flag6Byte = Flag6Byte | 0b00000001
	}

	if romModel.Header.Battery {
		Flag6Byte = Flag6Byte | 0b00000010
	}

	if romModel.Header.Trainer {
		Flag6Byte = Flag6Byte | 0b00000100
	}

	if romModel.Header.FourScreen {
		Flag6Byte = Flag6Byte | 0b00001000
	}

	Flag6Byte = Flag6Byte | ((MapperBytes[0] & 0b00001111) << 4)

	headerBytes[6] = Flag6Byte
	headerBytes[7] = (MapperBytes[0] & 0b11110000) | (romModel.Header.ConsoleType & 0b00000011) | 0b00001000

	headerBytes[8] = MapperBytes[1] & 0b00001111
	headerBytes[8] = headerBytes[8] | ((romModel.Header.SubMapper & 0b00001111) << 4)

	headerBytes[9] = PRGROMBytes[1] & 0b00001111
	headerBytes[9] = headerBytes[9] | ((CHRROMBytes[1] & 0b00001111) << 4)

	headerBytes[10] = 0b00000000
	if romModel.Header.PRGRAMSize > 0 {
		headerBytes[10] = headerBytes[10] | romModel.Header.PRGRAMSize
	}

	if romModel.Header.PRGNVRAMSize > 0 {
		headerBytes[10] = headerBytes[10] | (romModel.Header.PRGNVRAMSize << 4)
	}

	headerBytes[11] = 0b00000000
	if romModel.Header.CHRRAMSize > 0 {
		headerBytes[11] = headerBytes[11] | romModel.Header.CHRRAMSize
	}

	if romModel.Header.CHRNVRAMSize > 0 {
		headerBytes[11] = headerBytes[11] | (romModel.Header.CHRNVRAMSize << 4)
	}

	headerBytes[12] = 0b00000000 | romModel.Header.CPUPPUTiming
	headerBytes[13] = 0b00000000
	if romModel.Header.ConsoleType == 1 {
		headerBytes[13] = headerBytes[13] | ((romModel.Header.VsHardwareType & 0b00001111) << 4)
		headerBytes[13] = headerBytes[13] | (romModel.Header.VsPPUType & 0b00001111)
	} else if romModel.Header.ConsoleType == 3 {
		headerBytes[13] = headerBytes[13] | (romModel.Header.ExtendedConsoleType & 0b00001111)
	}

	headerBytes[14] = 0b00000011 & romModel.Header.MiscROMs
	headerBytes[15] = 0b00111111 & romModel.Header.DefaultExpansion

	romBytes := append(headerBytes, romModel.ROMData...)

	return romBytes, nil
}

func getStrippedRom(inputFile []byte) ([]byte, error) {
	fileSize := uint64(len(inputFile))
	if fileSize < 16 {
		return inputFile, &NESROMError{text: "File too small to be a headered NES ROM."}
	}

	if bytes.Compare(inputFile[0:4], []byte(NES_HEADER_MAGIC)) != 0 {
		return inputFile, &NESROMError{text: "Unable to find NES magic."}
	}

	return inputFile[16:fileSize], nil
}
