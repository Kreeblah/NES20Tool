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
	NES_HEADER_MAGIC                     = "\x4e\x45\x53\x1a"
	NES_20_AND_MASK                      = byte(0x08)
	NES_20_OR_MASK                       = byte(0xFB)
	ROM_TYPE_PRGROM               uint64 = 0
	ROM_TYPE_CHRROM               uint64 = 1
	PRG_CANONICAL_SIZE_ROM        uint64 = 0
	PRG_CANONICAL_SIZE_CALCULATED uint64 = 1
	PRG_CANONICAL_SIZE_FACTORED   uint64 = 2
	CHR_CANONICAL_SIZE_ROM        uint64 = 0
	CHR_CANONICAL_SIZE_CALCULATED uint64 = 1
	CHR_CANONICAL_SIZE_FACTORED   uint64 = 2
)

type NES20Header struct {
	PRGROMSize            uint16
	PRGROMCalculatedSize  uint64
	PRGROMSum16           uint16
	PRGROMCRC32           uint32
	PRGROMMD5             [16]byte
	PRGROMSHA1            [20]byte
	PRGROMSHA256          [32]byte
	PRGROMSizeExponent    uint8
	PRGROMSizeMultiplier  uint8
	CHRROMSize            uint16
	CHRROMCalculatedSize  uint64
	CHRROMSum16           uint16
	CHRROMCRC32           uint32
	CHRROMMD5             [16]byte
	CHRROMSHA1            [20]byte
	CHRROMSHA256          [32]byte
	CHRROMSizeExponent    uint8
	CHRROMSizeMultiplier  uint8
	MiscROMCalculatedSize uint64
	MiscROMSum16          uint16
	MiscROMCRC32          uint32
	MiscROMMD5            [16]byte
	MiscROMSHA1           [20]byte
	MiscROMSHA256         [32]byte
	TrainerCalculatedSize uint16
	TrainerSum16          uint16
	TrainerCRC32          uint32
	TrainerMD5            [16]byte
	TrainerSHA1           [20]byte
	TrainerSHA256         [32]byte
	PRGRAMSize            uint8
	PRGNVRAMSize          uint8
	CHRRAMSize            uint8
	CHRNVRAMSize          uint8
	MirroringType         bool
	Battery               bool
	Trainer               bool
	FourScreen            bool
	ConsoleType           uint8
	Mapper                uint16
	SubMapper             uint8
	CPUPPUTiming          uint8
	VsHardwareType        uint8
	VsPPUType             uint8
	ExtendedConsoleType   uint8
	MiscROMs              uint8
	DefaultExpansion      uint8
}

type NES10Header struct {
	PRGROMSize            uint8
	PRGROMCalculatedSize  uint64
	PRGROMSum16           uint16
	PRGROMCRC32           uint32
	PRGROMMD5             [16]byte
	PRGROMSHA1            [20]byte
	PRGROMSHA256          [32]byte
	CHRROMSize            uint8
	CHRROMCalculatedSize  uint64
	CHRROMSum16           uint16
	CHRROMCRC32           uint32
	CHRROMMD5             [16]byte
	CHRROMSHA1            [20]byte
	CHRROMSHA256          [32]byte
	TrainerCalculatedSize uint16
	TrainerSum16          uint16
	TrainerCRC32          uint32
	TrainerMD5            [16]byte
	TrainerSHA1           [20]byte
	TrainerSHA256         [32]byte
	MirroringType         bool
	Battery               bool
	Trainer               bool
	FourScreen            bool
	Mapper                uint8
	VsUnisystem           bool
	PRGRAMSize            uint8
	TVSystem              bool
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
	PRGROMData   []byte
	CHRROMData   []byte
	MiscROMData  []byte
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
			header20Data.PRGROMCalculatedSize = 16 * 1024 * uint64(header20Data.PRGROMSize)
		} else {
			header20Data.PRGROMSizeExponent = (inputFile[4] & 0b11111100) >> 2
			header20Data.PRGROMSizeMultiplier = inputFile[4] & 0b00000011
			header20Data.PRGROMCalculatedSize = (1 << header20Data.PRGROMSizeExponent) * uint64((header20Data.PRGROMSizeMultiplier*2)+1)
		}

		if inputFile[9]&0b11110000 != 0b11110000 {
			CHRROMBytes := make([]byte, 2)
			CHRROMBytes[0] = inputFile[5]
			CHRROMBytes[1] = (inputFile[9] & 0b11110000) >> 4
			header20Data.CHRROMSize = binary.LittleEndian.Uint16(CHRROMBytes)
			header20Data.CHRROMCalculatedSize = 8 * 1024 * uint64(header20Data.CHRROMSize)
		} else {
			header20Data.CHRROMSizeExponent = (inputFile[5] & 0b11111100) >> 2
			header20Data.CHRROMSizeMultiplier = inputFile[5] & 0b00000011
			header20Data.CHRROMCalculatedSize = (1 << header20Data.CHRROMSizeExponent) * uint64((header20Data.CHRROMSizeMultiplier*2)+1)
		}

		prgRomData, chrRomData, miscRomData, err := getSplitRomData(romData.ROMData, header20Data.PRGROMCalculatedSize, header20Data.CHRROMCalculatedSize)
		if err != nil {
			return romData, err
		}

		romData.PRGROMData = prgRomData
		romData.CHRROMData = chrRomData
		romData.MiscROMData = miscRomData

		romData.Header20.MiscROMCalculatedSize = uint64(len(romData.MiscROMData))

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

			header20Data.TrainerCalculatedSize = 512
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
		header10Data.PRGROMCalculatedSize = 16 * 1024 * uint64(header10Data.PRGROMSize)
		header10Data.CHRROMSize = inputFile[5]
		header10Data.CHRROMCalculatedSize = 8 * 1024 * uint64(header10Data.CHRROMSize)

		prgRomData, chrRomData, _, err := getSplitRomData(romData.ROMData, header10Data.PRGROMCalculatedSize, header10Data.CHRROMCalculatedSize)
		if err != nil {
			return romData, err
		}

		romData.PRGROMData = prgRomData
		romData.CHRROMData = chrRomData
		romData.MiscROMData = nil

		header10Data.MirroringType = (inputFile[6] & 0b00000001) == 0b00000001
		header10Data.Battery = (inputFile[6] & 0b00000010) == 0b00000010

		if !preserveTrainer {
			header10Data.Trainer = false
		} else {
			header10Data.Trainer = (inputFile[6] & 0b00000100) == 0b00000100

			header10Data.TrainerCalculatedSize = 512
		}

		header10Data.FourScreen = (inputFile[6] & 0b00001000) == 0b00001000
		header10Data.Mapper = ((inputFile[6] & 0b11110000) >> 4) | (inputFile[7] & 0b11110000)
		header10Data.VsUnisystem = (inputFile[7] & 0b00000001) == 0b00000001
		header10Data.PRGRAMSize = inputFile[8]
		header10Data.TVSystem = (inputFile[9] & 0b00000001) == 0b00000001

		romData.Header10 = header10Data
	}

	err := UpdateChecksums(romData)
	if err != nil {
		return romData, err
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

		if preserveTrainer && romModel.Header20.Trainer && len(romModel.TrainerData) == 512 {
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
			rawRomBytes = romModel.ROMData[0:(romModel.Header20.PRGROMCalculatedSize + romModel.Header20.CHRROMCalculatedSize)]
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

		if preserveTrainer && romModel.Header10.Trainer && len(romModel.TrainerData) == 512 {
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

		if truncateRom && romModel.Header20.MiscROMs == 0 {
			rawRomBytes = romModel.ROMData[0:(romModel.Header10.PRGROMCalculatedSize + romModel.Header10.CHRROMCalculatedSize)]
		} else {
			rawRomBytes = romModel.ROMData
		}
	}

	romBytes := headerBytes

	if !preserveTrainer || (!romModel.Header20.Trainer && !romModel.Header10.Trainer) || len(romModel.TrainerData) != 512 {
		romBytes = append(romBytes, rawRomBytes...)
	} else {
		romBytes = append(romBytes, romModel.TrainerData...)
		romBytes = append(romBytes, rawRomBytes...)
	}

	return romBytes, nil
}

func FactorRomSize(romSize uint64, romType uint64) (uint16, uint8, uint8) {
	var blockSize uint64
	var sizeExponent uint8
	var sizeMultiplier uint8

	if romSize == 0 {
		return 0, 0, 0
	}

	if romType == ROM_TYPE_PRGROM {
		blockSize = 16 * 1024
	} else if romType == ROM_TYPE_CHRROM {
		blockSize = 8 * 1024
	}

	if romSize%blockSize == 0 {
		return uint16(romSize / blockSize), 0, 0
	}

	if romSize%3 == 0 {
		sizeMultiplier = 2
	} else if romSize%5 == 0 {
		sizeMultiplier = 4
	} else if romSize%7 == 0 {
		sizeMultiplier = 6
	} else {
		sizeMultiplier = 0
	}

	tempSize := romSize / uint64(sizeMultiplier)
	sizeExponent = 0

	for tempSize > 0 {
		tempSize = tempSize << 1
		sizeExponent = sizeExponent + 1
	}

	return 0, sizeExponent, sizeMultiplier
}

func UpdateSizes(nesRom *NESROM, prgCanonicalSize uint64, chrCanonicalSize uint64) error {
	if nesRom.ROMData != nil {
		nesRom.Size = uint64(len(nesRom.ROMData))
	}

	if nesRom.Header20 != nil {
		if prgCanonicalSize == PRG_CANONICAL_SIZE_ROM && nesRom.PRGROMData != nil {
			nesRom.Header20.PRGROMCalculatedSize = uint64(len(nesRom.PRGROMData))
		}

		if (prgCanonicalSize == PRG_CANONICAL_SIZE_ROM && nesRom.PRGROMData != nil) || prgCanonicalSize == PRG_CANONICAL_SIZE_CALCULATED {
			nesRom.Header20.PRGROMSize, nesRom.Header20.PRGROMSizeExponent, nesRom.Header20.PRGROMSizeMultiplier = FactorRomSize(nesRom.Header20.PRGROMCalculatedSize, ROM_TYPE_PRGROM)
		}

		if prgCanonicalSize == PRG_CANONICAL_SIZE_FACTORED {
			nesRom.Header20.PRGROMCalculatedSize = (1 << nesRom.Header20.PRGROMSizeExponent) * uint64((nesRom.Header20.PRGROMSizeMultiplier*2)+1)
		}

		if chrCanonicalSize == CHR_CANONICAL_SIZE_ROM && nesRom.CHRROMData != nil {
			nesRom.Header20.CHRROMCalculatedSize = uint64(len(nesRom.CHRROMData))
		}

		if (chrCanonicalSize == CHR_CANONICAL_SIZE_ROM && nesRom.CHRROMData != nil) || chrCanonicalSize == CHR_CANONICAL_SIZE_CALCULATED {
			nesRom.Header20.CHRROMSize, nesRom.Header20.CHRROMSizeExponent, nesRom.Header20.CHRROMSizeMultiplier = FactorRomSize(nesRom.Header20.CHRROMCalculatedSize, ROM_TYPE_CHRROM)
		}

		if chrCanonicalSize == CHR_CANONICAL_SIZE_FACTORED {
			nesRom.Header20.CHRROMCalculatedSize = (1 << nesRom.Header20.CHRROMSizeExponent) * uint64((nesRom.Header20.CHRROMSizeMultiplier*2)+1)
		}

		if nesRom.Header20.MiscROMs > 0 && nesRom.MiscROMData != nil && len(nesRom.MiscROMData) > 0 {
			nesRom.Header20.MiscROMCalculatedSize = uint64(len(nesRom.MiscROMData))
		} else {
			nesRom.Header20.MiscROMCalculatedSize = 0
		}

		if nesRom.Header20.Trainer && nesRom.TrainerData != nil && len(nesRom.TrainerData) == 512 {
			nesRom.Header20.TrainerCalculatedSize = 512
		} else {
			nesRom.Header20.TrainerCalculatedSize = 0
		}
	}

	if nesRom.Header10 != nil {
		if prgCanonicalSize == PRG_CANONICAL_SIZE_ROM && nesRom.PRGROMData != nil {
			nesRom.Header10.PRGROMCalculatedSize = uint64(len(nesRom.PRGROMData))
		}

		if (prgCanonicalSize == PRG_CANONICAL_SIZE_ROM && nesRom.PRGROMData != nil) || prgCanonicalSize == PRG_CANONICAL_SIZE_CALCULATED {
			nesRom.Header10.PRGROMSize = uint8(nesRom.Header10.PRGROMCalculatedSize / (16 * 1024))
		}

		if prgCanonicalSize == PRG_CANONICAL_SIZE_FACTORED {
			nesRom.Header10.PRGROMCalculatedSize = uint64(nesRom.Header10.PRGROMSize) * (16 * 1024)
		}

		if chrCanonicalSize == CHR_CANONICAL_SIZE_ROM && nesRom.CHRROMData != nil {
			nesRom.Header10.CHRROMCalculatedSize = uint64(len(nesRom.CHRROMData))
		}

		if (chrCanonicalSize == CHR_CANONICAL_SIZE_ROM && nesRom.CHRROMData != nil) || chrCanonicalSize == CHR_CANONICAL_SIZE_CALCULATED {
			nesRom.Header10.CHRROMSize = uint8(nesRom.Header10.CHRROMCalculatedSize / (8 * 1024))
		}

		if chrCanonicalSize == CHR_CANONICAL_SIZE_FACTORED {
			nesRom.Header10.CHRROMCalculatedSize = uint64(nesRom.Header10.CHRROMSize) * (8 * 1024)
		}

		if nesRom.Header10.Trainer && nesRom.TrainerData != nil && len(nesRom.TrainerData) == 512 {
			nesRom.Header10.TrainerCalculatedSize = 512
		} else {
			nesRom.Header10.TrainerCalculatedSize = 0
		}
	}

	return nil
}

func UpdateChecksums(nesRom *NESROM) error {
	if nesRom.ROMData != nil {
		nesRom.CRC32 = crc32.ChecksumIEEE(nesRom.ROMData)
		nesRom.MD5 = md5.Sum(nesRom.ROMData)
		nesRom.SHA1 = sha1.Sum(nesRom.ROMData)
		nesRom.SHA256 = sha256.Sum256(nesRom.ROMData)
	}

	if nesRom.Header20 != nil {
		if nesRom.PRGROMData != nil {
			prgRomSum16, err := calculateSum16(nesRom.PRGROMData)
			if err != nil {
				return err
			}

			nesRom.Header20.PRGROMSum16 = prgRomSum16
			nesRom.Header20.PRGROMCRC32 = crc32.ChecksumIEEE(nesRom.PRGROMData)
			nesRom.Header20.PRGROMMD5 = md5.Sum(nesRom.PRGROMData)
			nesRom.Header20.PRGROMSHA1 = sha1.Sum(nesRom.PRGROMData)
			nesRom.Header20.PRGROMSHA256 = sha256.Sum256(nesRom.PRGROMData)
		}

		if nesRom.CHRROMData != nil {
			chrRomSum16, err := calculateSum16(nesRom.CHRROMData)
			if err != nil {
				return err
			}

			nesRom.Header20.CHRROMSum16 = chrRomSum16
			nesRom.Header20.CHRROMCRC32 = crc32.ChecksumIEEE(nesRom.CHRROMData)
			nesRom.Header20.CHRROMMD5 = md5.Sum(nesRom.CHRROMData)
			nesRom.Header20.CHRROMSHA1 = sha1.Sum(nesRom.CHRROMData)
			nesRom.Header20.CHRROMSHA256 = sha256.Sum256(nesRom.CHRROMData)
		}

		if nesRom.MiscROMData != nil {
			miscRomSum16, err := calculateSum16(nesRom.MiscROMData)
			if err != nil {
				return err
			}

			nesRom.Header20.MiscROMSum16 = miscRomSum16
			nesRom.Header20.MiscROMCRC32 = crc32.ChecksumIEEE(nesRom.MiscROMData)
			nesRom.Header20.MiscROMMD5 = md5.Sum(nesRom.MiscROMData)
			nesRom.Header20.MiscROMSHA1 = sha1.Sum(nesRom.MiscROMData)
			nesRom.Header20.MiscROMSHA256 = sha256.Sum256(nesRom.MiscROMData)
		}

		if nesRom.TrainerData != nil {
			trainerSum16, err := calculateSum16(nesRom.TrainerData)
			if err != nil {
				return err
			}

			nesRom.Header20.TrainerSum16 = trainerSum16
			nesRom.Header20.TrainerCRC32 = crc32.ChecksumIEEE(nesRom.TrainerData)
			nesRom.Header20.TrainerMD5 = md5.Sum(nesRom.TrainerData)
			nesRom.Header20.TrainerSHA1 = sha1.Sum(nesRom.TrainerData)
			nesRom.Header20.TrainerSHA256 = sha256.Sum256(nesRom.TrainerData)
		}
	}

	if nesRom.Header10 != nil {
		if nesRom.PRGROMData != nil {
			prgRomSum16, err := calculateSum16(nesRom.PRGROMData)
			if err != nil {
				return err
			}

			nesRom.Header10.PRGROMSum16 = prgRomSum16
			nesRom.Header10.PRGROMCRC32 = crc32.ChecksumIEEE(nesRom.PRGROMData)
			nesRom.Header10.PRGROMMD5 = md5.Sum(nesRom.PRGROMData)
			nesRom.Header10.PRGROMSHA1 = sha1.Sum(nesRom.PRGROMData)
			nesRom.Header10.PRGROMSHA256 = sha256.Sum256(nesRom.PRGROMData)
		}

		if nesRom.CHRROMData != nil {
			chrRomSum16, err := calculateSum16(nesRom.CHRROMData)
			if err != nil {
				return err
			}

			nesRom.Header10.CHRROMSum16 = chrRomSum16
			nesRom.Header10.CHRROMCRC32 = crc32.ChecksumIEEE(nesRom.CHRROMData)
			nesRom.Header10.CHRROMMD5 = md5.Sum(nesRom.CHRROMData)
			nesRom.Header10.CHRROMSHA1 = sha1.Sum(nesRom.CHRROMData)
			nesRom.Header10.CHRROMSHA256 = sha256.Sum256(nesRom.CHRROMData)
		}

		if nesRom.TrainerData != nil {
			trainerSum16, err := calculateSum16(nesRom.TrainerData)
			if err != nil {
				return err
			}

			nesRom.Header10.TrainerSum16 = trainerSum16
			nesRom.Header10.TrainerCRC32 = crc32.ChecksumIEEE(nesRom.TrainerData)
			nesRom.Header10.TrainerMD5 = md5.Sum(nesRom.TrainerData)
			nesRom.Header10.TrainerSHA1 = sha1.Sum(nesRom.TrainerData)
			nesRom.Header10.TrainerSHA256 = sha256.Sum256(nesRom.TrainerData)
		}
	}

	return nil
}

func TruncateROMDataAndSections(rom *NESROM) {
	if rom.Header20 != nil {
		if uint64(len(rom.PRGROMData)) > rom.Header20.PRGROMCalculatedSize {
			rom.PRGROMData = rom.PRGROMData[:rom.Header20.PRGROMCalculatedSize]
		}

		if uint64(len(rom.CHRROMData)) > rom.Header20.CHRROMCalculatedSize {
			rom.CHRROMData = rom.CHRROMData[:rom.Header20.CHRROMCalculatedSize]
		}

		rom.ROMData = rom.PRGROMData
		rom.ROMData = append(rom.ROMData, rom.CHRROMData...)

		return
	}

	if rom.Header10 != nil {
		if uint64(len(rom.PRGROMData)) > rom.Header10.PRGROMCalculatedSize {
			rom.PRGROMData = rom.PRGROMData[:rom.Header10.PRGROMCalculatedSize]
		}

		if uint64(len(rom.CHRROMData)) > rom.Header10.CHRROMCalculatedSize {
			rom.CHRROMData = rom.CHRROMData[:rom.Header10.CHRROMCalculatedSize]
		}

		rom.ROMData = rom.PRGROMData
		rom.ROMData = append(rom.ROMData, rom.CHRROMData...)

		return
	}
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

func calculateSum16(inputData []byte) (uint16, error) {
	if inputData == nil {
		return 0, &NESROMError{text: "Cannot calculate sum16 for a null segment."}
	}

	var byteSum uint64 = 0
	for i := range inputData {
		byteSum = byteSum + uint64(inputData[i])
	}

	return uint16(byteSum), nil
}

func getSplitRomData(inputData []byte, prgRomSize uint64, chrRomSize uint64) ([]byte, []byte, []byte, error) {
	if inputData == nil {
		return nil, nil, nil, &NESROMError{text: "Cannot extract PRGROM and/or CHRROM data from a null segment."}
	}

	if len(inputData) < int(prgRomSize+chrRomSize) {
		return nil, nil, nil, &NESROMError{text: "Invalid PRGROM and/or CHRROM size(s) detected during extraction."}
	}

	prgRomData := inputData[0:prgRomSize]
	chrRomData := inputData[prgRomSize:(prgRomSize + chrRomSize)]
	miscRomData := make([]byte, 0)

	if prgRomSize+chrRomSize < uint64(len(inputData)) {
		miscRomData = inputData[(prgRomSize + chrRomSize):]
	}

	return prgRomData, chrRomData, miscRomData, nil
}
