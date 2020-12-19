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

type NES20Header struct {
	PRGROMSize           uint16
	PRGROMCalculatedSize uint32
	PRGROMSum16          uint16
	PRGROMSizeExponent   uint8
	PRGROMSizeMultiplier uint8
	CHRROMSize           uint16
	CHRROMCalculatedSize uint32
	CHRROMSum16          uint16
	CHRROMSizeExponent   uint8
	CHRROMSizeMultiplier uint8
	PRGRAMSize           uint8
	PRGNVRAMSize         uint8
	CHRRAMSize           uint8
	CHRNVRAMSize         uint8
	MirroringType        bool
	Battery              bool
	Trainer              bool
	FourScreen           bool
	ConsoleType          uint8
	Mapper               uint16
	SubMapper            uint8
	CPUPPUTiming         uint8
	VsHardwareType       uint8
	VsPPUType            uint8
	ExtendedConsoleType  uint8
	MiscROMs             uint8
	DefaultExpansion     uint8
}

type NES10Header struct {
	PRGROMSize    uint8
	PRGROMCalculatedSize uint32
	PRGROMSum16 uint16
	CHRROMSize    uint8
	CHRROMCalculatedSize uint32
	CHRROMSum16 uint16
	MirroringType bool
	Battery       bool
	Trainer       bool
	FourScreen    bool
	Mapper        uint8
	VsUnisystem   bool
	PRGRAMSize    uint8
	TVSystem      bool
}

type NESROM struct {
	Name         string
	Filename     string
	RelativePath string
	Size         uint64
	CRC32        uint32
	MD5          [16]byte
	SHA1         [20]byte
	SHA256       [32]byte
	Header20     *NES20Header
	Header10     *NES10Header
	ROMData      []byte
	TrainerData  []byte
}

type NESROMError struct {
	text string
}

func (r *NESROMError) Error() string {
	return r.text
}

func DecodeNESROM(inputFile []byte, enableInes bool, preserveTrainer bool, relativeLocation string) (*NESROM, error) {
	headerVersion := 2
	fileSize := uint64(len(inputFile))
	rawROMBytes, rawTrainerBytes, _ := getStrippedRom(inputFile)

	romData := &NESROM{}
	romData.RelativePath = relativeLocation
	romData.CRC32 = crc32.ChecksumIEEE(rawROMBytes)
	romData.MD5 = md5.Sum(rawROMBytes)
	romData.SHA1 = sha1.Sum(rawROMBytes)
	romData.SHA256 = sha256.Sum256(rawROMBytes)
	romData.Size = uint64(len(rawROMBytes))
	romData.ROMData = rawROMBytes

	if preserveTrainer {
		romData.TrainerData = rawTrainerBytes
	}

	if fileSize < 16 {
		return romData, &NESROMError{text: "File too small to be a headered NES ROM."}
	}

	if bytes.Compare(inputFile[0:4], []byte(NES_HEADER_MAGIC)) != 0 {
		return romData, &NESROMError{text: "Unable to find NES magic."}
	}

	if (inputFile[7]&NES_20_AND_MASK) != NES_20_AND_MASK || (inputFile[7]|NES_20_OR_MASK) != NES_20_OR_MASK {
		if !enableInes {
			return romData, &NESROMError{text: "Not an NES 2.0 ROM."}
		} else {
			headerVersion = 1
		}
	}

	header10Data := &NES10Header{}
	header20Data := &NES20Header{}

	if headerVersion == 2 {
		if inputFile[9]&0b00001111 != 0b00001111 {
			PRGROMBytes := make([]byte, 2)
			PRGROMBytes[0] = inputFile[4]
			PRGROMBytes[1] = inputFile[9] & 0b00001111
			header20Data.PRGROMSize = binary.LittleEndian.Uint16(PRGROMBytes)
			header20Data.PRGROMCalculatedSize = 16 * 1024 * uint32(header20Data.PRGROMSize)
		} else {
			header20Data.PRGROMSizeExponent = (inputFile[4] & 0b11111100) >> 2
			header20Data.PRGROMSizeMultiplier = inputFile[4] & 0b00000011
			header20Data.PRGROMCalculatedSize = (2 << header20Data.PRGROMSizeExponent) * uint32((header20Data.PRGROMSizeMultiplier * 2) + 1)
		}

		if inputFile[9]&0b11110000 != 0b11110000 {
			CHRROMBytes := make([]byte, 2)
			CHRROMBytes[0] = inputFile[5]
			CHRROMBytes[1] = (inputFile[9] & 0b11110000) >> 4
			header20Data.CHRROMSize = binary.LittleEndian.Uint16(CHRROMBytes)
			header20Data.CHRROMCalculatedSize = 8 * 1024 * uint32(header20Data.CHRROMSize)
		} else {
			header20Data.CHRROMSizeExponent = (inputFile[5] & 0b11111100) >> 2
			header20Data.CHRROMSizeMultiplier = inputFile[5] & 0b00000011
			header20Data.CHRROMCalculatedSize = (2 << header20Data.CHRROMSizeExponent) * uint32((header20Data.CHRROMSizeMultiplier * 2) + 1)
		}

		prgRomSum16, chrRomSum16, err := calculateSum16(romData.ROMData, header20Data.PRGROMCalculatedSize, header20Data.CHRROMCalculatedSize)
		if err != nil {
			return romData, err
		}

		header20Data.PRGROMSum16 = prgRomSum16
		header20Data.CHRROMSum16 = chrRomSum16

		header20Data.PRGRAMSize = inputFile[10] & 0b00001111
		header20Data.PRGNVRAMSize = (inputFile[10] & 0b11110000) >> 4
		header20Data.CHRRAMSize = inputFile[11] & 0b00001111
		header20Data.CHRNVRAMSize = (inputFile[11] & 0b11110000) >> 4

		header20Data.MirroringType = (inputFile[6] & 0b00000001) == 0b00000001
		header20Data.Battery = (inputFile[6] & 0b00000010) == 0b00000010

		if !preserveTrainer {
			header20Data.Trainer = false
		} else {
			header20Data.Trainer = (inputFile[6] & 0b00000100) == 0b00000100
		}

		header20Data.FourScreen = (inputFile[6] & 0b00001000) == 0b00001000
		header20Data.ConsoleType = inputFile[7] & 0b00000011

		MapperBytes := make([]byte, 2)
		MapperBytes[0] = ((inputFile[6] & 0b11110000) >> 4) | (inputFile[7] & 0b11110000)
		MapperBytes[1] = inputFile[8] & 0b00001111

		header20Data.Mapper = binary.LittleEndian.Uint16(MapperBytes)
		header20Data.SubMapper = (inputFile[8] & 0b11110000) >> 4
		header20Data.CPUPPUTiming = inputFile[12] & 0b00000011

		if header20Data.ConsoleType == 0 {
			header20Data.VsHardwareType = 0
			header20Data.VsPPUType = 0
			header20Data.ExtendedConsoleType = 0
		} else if header20Data.ConsoleType == 1 {
			header20Data.VsHardwareType = (inputFile[13] & 0b11110000) >> 4
			header20Data.VsPPUType = inputFile[13] & 0b00001111
			header20Data.ExtendedConsoleType = 0
		} else if header20Data.ConsoleType == 3 {
			header20Data.VsHardwareType = 0
			header20Data.VsPPUType = 0
			header20Data.ExtendedConsoleType = inputFile[13] & 0b00001111
		} else {
			header20Data.VsHardwareType = 0
			header20Data.VsPPUType = 0
			header20Data.ExtendedConsoleType = 0
		}

		header20Data.MiscROMs = inputFile[14] & 0b00000011
		header20Data.DefaultExpansion = inputFile[15] & 0b00111111

		romData.Header20 = header20Data
	} else if headerVersion == 1 {
		header10Data.PRGROMSize = inputFile[4]
		header10Data.PRGROMCalculatedSize = 16 * 1024 * uint32(header10Data.PRGROMSize)
		header10Data.CHRROMSize = inputFile[5]
		header10Data.CHRROMCalculatedSize = 8 * 1024 * uint32(header10Data.CHRROMSize)

		prgRomSum16, chrRomSum16, err := calculateSum16(romData.ROMData, header10Data.PRGROMCalculatedSize, header10Data.CHRROMCalculatedSize)
		if err != nil {
			return romData, err
		}

		header10Data.PRGROMSum16 = prgRomSum16
		header10Data.CHRROMSum16 = chrRomSum16

		header10Data.MirroringType = (inputFile[6] & 0b00000001) == 0b00000001
		header10Data.Battery = (inputFile[6] & 0b00000010) == 0b00000010

		if !preserveTrainer {
			header10Data.Trainer = false
		} else {
			header10Data.Trainer = (inputFile[6] & 0b00000100) == 0b00000100
		}

		header10Data.FourScreen = (inputFile[6] & 0b00001000) == 0b00001000
		header10Data.Mapper = ((inputFile[6] & 0b11110000) >> 4) | (inputFile[7] & 0b11110000)
		header10Data.VsUnisystem = (inputFile[7] & 0b00000001) == 0b00000001
		header10Data.PRGRAMSize = inputFile[8]
		header10Data.TVSystem = (inputFile[9] & 0b00000001) == 0b00000001

		romData.Header10 = header10Data
	}

	return romData, nil
}

func EncodeNESROM(romModel *NESROM, enableInes bool, truncateRom bool, preserveTrainer bool) ([]byte, error) {
	headerVersion := 2

	if romModel.Header20 == nil {
		if romModel.Header10 == nil || !enableInes {
			return nil, &NESROMError{"Unable to find valid header on ROM model."}
		} else {
			headerVersion = 1
		}
	}

	headerBytes := make([]byte, 16)
	for index := range NES_HEADER_MAGIC {
		headerBytes[index] = NES_HEADER_MAGIC[index]
	}

	var rawRomBytes []byte

	if headerVersion == 2 {
		headerBytes[9] = 0b00000000
		if romModel.Header20.PRGROMSize > 0 {
			PRGROMBytes := make([]byte, 2)
			binary.LittleEndian.PutUint16(PRGROMBytes, romModel.Header20.PRGROMSize)
			headerBytes[4] = PRGROMBytes[0]
			headerBytes[9] = headerBytes[9] | (PRGROMBytes[1] & 0b00001111)
		} else {
			headerBytes[4] = (romModel.Header20.PRGROMSizeExponent << 2) | romModel.Header20.PRGROMSizeMultiplier
			headerBytes[9] = headerBytes[9] | 0b00001111
		}

		if romModel.Header20.CHRROMSize > 0 {
			CHRROMBytes := make([]byte, 2)
			binary.LittleEndian.PutUint16(CHRROMBytes, romModel.Header20.CHRROMSize)
			headerBytes[5] = CHRROMBytes[0]
			headerBytes[9] = headerBytes[9] | ((CHRROMBytes[1] & 0b00001111) << 4)
		} else {
			headerBytes[5] = (romModel.Header20.CHRROMSizeExponent << 2) | romModel.Header20.CHRROMSizeMultiplier
			headerBytes[9] = headerBytes[9] | 0b11110000
		}

		MapperBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(MapperBytes, romModel.Header20.Mapper)

		var Flag6Byte byte = 0b00000000
		if romModel.Header20.MirroringType {
			Flag6Byte = Flag6Byte | 0b00000001
		}

		if romModel.Header20.Battery {
			Flag6Byte = Flag6Byte | 0b00000010
		}

		if romModel.Header20.Trainer {
			Flag6Byte = Flag6Byte | 0b00000100
		}

		if romModel.Header20.FourScreen {
			Flag6Byte = Flag6Byte | 0b00001000
		}

		Flag6Byte = Flag6Byte | ((MapperBytes[0] & 0b00001111) << 4)

		headerBytes[6] = Flag6Byte
		headerBytes[7] = (MapperBytes[0] & 0b11110000) | (romModel.Header20.ConsoleType & 0b00000011) | 0b00001000

		headerBytes[8] = MapperBytes[1] & 0b00001111
		headerBytes[8] = headerBytes[8] | ((romModel.Header20.SubMapper & 0b00001111) << 4)

		headerBytes[10] = 0b00000000
		if romModel.Header20.PRGRAMSize > 0 {
			headerBytes[10] = headerBytes[10] | romModel.Header20.PRGRAMSize
		}

		if romModel.Header20.PRGNVRAMSize > 0 {
			headerBytes[10] = headerBytes[10] | (romModel.Header20.PRGNVRAMSize << 4)
		}

		headerBytes[11] = 0b00000000
		if romModel.Header20.CHRRAMSize > 0 {
			headerBytes[11] = headerBytes[11] | romModel.Header20.CHRRAMSize
		}

		if romModel.Header20.CHRNVRAMSize > 0 {
			headerBytes[11] = headerBytes[11] | (romModel.Header20.CHRNVRAMSize << 4)
		}

		headerBytes[12] = 0b00000000 | romModel.Header20.CPUPPUTiming
		headerBytes[13] = 0b00000000
		if romModel.Header20.ConsoleType == 1 {
			headerBytes[13] = headerBytes[13] | ((romModel.Header20.VsHardwareType & 0b00001111) << 4)
			headerBytes[13] = headerBytes[13] | (romModel.Header20.VsPPUType & 0b00001111)
		} else if romModel.Header20.ConsoleType == 3 {
			headerBytes[13] = headerBytes[13] | (romModel.Header20.ExtendedConsoleType & 0b00001111)
		}

		headerBytes[14] = 0b00000011 & romModel.Header20.MiscROMs
		headerBytes[15] = 0b00111111 & romModel.Header20.DefaultExpansion

		if truncateRom {
			rawRomBytes = romModel.ROMData[0 : (romModel.Header20.PRGROMCalculatedSize + romModel.Header20.CHRROMCalculatedSize)]
		} else {
			rawRomBytes = romModel.ROMData
		}
	} else if headerVersion == 1 {
		headerBytes[4] = romModel.Header10.PRGROMSize
		headerBytes[5] = romModel.Header10.CHRROMSize

		var Flag6Byte byte = 0b00000000
		if romModel.Header10.MirroringType {
			Flag6Byte = Flag6Byte | 0b00000001
		}

		if romModel.Header10.Battery {
			Flag6Byte = Flag6Byte | 0b00000010
		}

		if romModel.Header10.Trainer {
			Flag6Byte = Flag6Byte | 0b00000100
		}

		if romModel.Header10.FourScreen {
			Flag6Byte = Flag6Byte | 0b00001000
		}

		Flag6Byte = Flag6Byte | ((romModel.Header10.Mapper & 0b00001111) << 4)
		headerBytes[6] = Flag6Byte

		var Flag7Byte byte = 0b00000000
		if romModel.Header10.VsUnisystem {
			Flag7Byte = Flag7Byte | 0b00000001
		}

		Flag7Byte = Flag7Byte | (romModel.Header10.Mapper & 0b11110000)
		headerBytes[7] = Flag7Byte

		headerBytes[8] = romModel.Header10.PRGRAMSize
		if !romModel.Header10.TVSystem {
			headerBytes[9] = 0
		} else {
			headerBytes[9] = 1
		}

		headerBytes[10] = 0
		headerBytes[11] = 0
		headerBytes[12] = 0
		headerBytes[13] = 0
		headerBytes[14] = 0
		headerBytes[15] = 0

		if truncateRom {
			rawRomBytes = romModel.ROMData[0 : (romModel.Header10.PRGROMCalculatedSize + romModel.Header10.CHRROMCalculatedSize)]
		} else {
			rawRomBytes = romModel.ROMData
		}
	}

	romBytes := headerBytes

	if !preserveTrainer {
		romBytes = append(romBytes, rawRomBytes...)
	} else {
		romBytes = append(romBytes, romModel.TrainerData...)
		romBytes = append(romBytes, rawRomBytes...)
	}

	return romBytes, nil
}

func getStrippedRom(inputFile []byte) ([]byte, []byte, error) {
	fileSize := uint64(len(inputFile))
	if fileSize < 16 {
		return inputFile, nil, &NESROMError{text: "File too small to be a headered NES ROM."}
	}

	if bytes.Compare(inputFile[0:4], []byte(NES_HEADER_MAGIC)) != 0 {
		return inputFile, nil, &NESROMError{text: "Unable to find NES magic."}
	}

	hasTrainer := (inputFile[6] & 0b00000100) == 0b00000100

	if hasTrainer && fileSize < 528 {
		return inputFile, nil, &NESROMError{text: "Header indicates trainer data, but file too small for one."}
	}

	if !hasTrainer {
		return inputFile[16:fileSize], nil, nil
	} else {
		return inputFile[528:fileSize], inputFile[16:528], nil
	}
}

func calculateSum16(inputData []byte, prgRomSize uint32, chrRomSize uint32) (uint16, uint16, error) {
	if inputData == nil {
		return 0, 0, &NESROMError{text: "Cannot calculate sum16 for a null segment."}
	}

	if len(inputData) < int(prgRomSize + chrRomSize) {
		return 0, 0, &NESROMError{text: "Invalid PRGROM and/or CHRROM size(s) detected during sum16 calculation."}
	}

	prgRomData, chrRomData, err := getPrgRomAndChrRomData(inputData, prgRomSize, chrRomSize)
	if err != nil {
		return 0, 0, err
	}

	var prgRomByteSum uint64 = 0
	for i := range prgRomData {
		prgRomByteSum = prgRomByteSum + uint64(prgRomData[i])
	}

	var chrRomByteSum uint64 = 0
	for i := range chrRomData {
		chrRomByteSum = chrRomByteSum + uint64(chrRomData[i])
	}

	return uint16(prgRomByteSum), uint16(chrRomByteSum), nil
}

func getPrgRomAndChrRomData(inputData []byte, prgRomSize uint32, chrRomSize uint32) ([]byte, []byte, error) {
	if inputData == nil {
		return nil, nil, &NESROMError{text: "Cannot extract PRGROM and/or CHRROM data from a null segment."}
	}

	if len(inputData) < int(prgRomSize + chrRomSize) {
		return nil, nil, &NESROMError{text: "Invalid PRGROM and/or CHRROM size(s) detected during extraction."}
	}

	prgRomData := inputData[0 : prgRomSize]
	chrRomData := inputData[prgRomSize : (prgRomSize + chrRomSize)]

	return prgRomData, chrRomData, nil
}