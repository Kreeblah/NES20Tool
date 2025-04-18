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

// https://wiki.nesdev.com/w/index.php/NES_2.0
// https://wiki.nesdev.com/w/index.php/INES
// This implements the INES and NES 2.0 specifications, as described
// in the formats above.

package NESTool

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"
	"strconv"
	"strings"
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
	PlayChoice10          bool
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
	HeaderData   []byte
}

type NESROMError struct {
	Text string
}

func (r *NESROMError) Error() string {
	return r.Text
}

func (rom *NESROM) String() string {
	returnString := ""

	if rom.Header20 != nil {
		returnString = "ROM Header Version: NES 2.0\n"
	} else if rom.Header10 != nil {
		returnString = "ROM Header Version: iNES\n"
	} else {
		return ""
	}

	if rom.Name != "" {
		returnString = returnString + "ROM Name: " + rom.Name + "\n"
	} else if rom.Filename != "" {
		returnString = returnString + "ROM Filename: " + rom.Filename + "\n"
	} else if rom.RelativePath != "" {
		returnString = returnString + "ROM Relative Path: " + rom.RelativePath + "\n"
	}

	returnString = returnString + "ROM Size: " + strconv.Itoa(int(rom.Size)) + " bytes\n"

	crc32Bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crc32Bytes, rom.CRC32)
	returnString = returnString + "ROM CRC32: " + strings.ToUpper(hex.EncodeToString(crc32Bytes)) + "\n"

	returnString = returnString + "ROM MD5: " + strings.ToUpper(hex.EncodeToString(rom.MD5[:])) + "\n"
	returnString = returnString + "ROM SHA1: " + strings.ToUpper(hex.EncodeToString(rom.SHA1[:])) + "\n"
	returnString = returnString + "ROM SHA256: " + strings.ToUpper(hex.EncodeToString(rom.SHA256[:])) + "\n"

	if rom.Header20 != nil {
		returnString = returnString + "PRG ROM Size: " + strconv.Itoa(int(rom.Header20.PRGROMCalculatedSize)) + " bytes\n"

		prgSum16Bytes := make([]byte, 2)
		binary.BigEndian.PutUint16(prgSum16Bytes, rom.Header20.PRGROMSum16)
		returnString = returnString + "PRG ROM Sum16: " + strings.ToUpper(hex.EncodeToString(prgSum16Bytes)) + "\n"

		prgCrc32Bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(prgCrc32Bytes, rom.Header20.PRGROMCRC32)
		returnString = returnString + "PRG ROM CRC32: " + strings.ToUpper(hex.EncodeToString(prgCrc32Bytes)) + "\n"

		returnString = returnString + "PRG ROM MD5: " + strings.ToUpper(hex.EncodeToString(rom.Header20.PRGROMMD5[:])) + "\n"
		returnString = returnString + "PRG ROM SHA1: " + strings.ToUpper(hex.EncodeToString(rom.Header20.PRGROMSHA1[:])) + "\n"
		returnString = returnString + "PRG ROM SHA256: " + strings.ToUpper(hex.EncodeToString(rom.Header20.PRGROMSHA256[:])) + "\n"

		returnString = returnString + "CHR ROM Size: " + strconv.Itoa(int(rom.Header20.CHRROMCalculatedSize)) + " bytes\n"

		if rom.Header20.CHRROMCalculatedSize > 0 {
			chrSum16Bytes := make([]byte, 2)
			binary.BigEndian.PutUint16(chrSum16Bytes, rom.Header20.CHRROMSum16)
			returnString = returnString + "CHR ROM Sum16: " + strings.ToUpper(hex.EncodeToString(chrSum16Bytes)) + "\n"

			chrCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(chrCrc32Bytes, rom.Header20.CHRROMCRC32)
			returnString = returnString + "CHR ROM CRC32: " + strings.ToUpper(hex.EncodeToString(chrCrc32Bytes)) + "\n"

			returnString = returnString + "CHR ROM MD5: " + strings.ToUpper(hex.EncodeToString(rom.Header20.CHRROMMD5[:])) + "\n"
			returnString = returnString + "CHR ROM SHA1: " + strings.ToUpper(hex.EncodeToString(rom.Header20.CHRROMSHA1[:])) + "\n"
			returnString = returnString + "CHR ROM SHA256: " + strings.ToUpper(hex.EncodeToString(rom.Header20.CHRROMSHA256[:])) + "\n"
		}

		if !rom.Header20.Trainer {
			returnString = returnString + "Trainer Size: 0 bytes\n"
		} else {
			returnString = returnString + "Trainer Size: " + strconv.Itoa(int(rom.Header20.TrainerCalculatedSize)) + " bytes\n"

			if rom.Header20.TrainerCalculatedSize > 0 {
				trainerSum16Bytes := make([]byte, 2)
				binary.BigEndian.PutUint16(trainerSum16Bytes, rom.Header20.TrainerSum16)
				returnString = returnString + "Trainer Sum16: " + strings.ToUpper(hex.EncodeToString(trainerSum16Bytes)) + "\n"

				trainerCrc32Bytes := make([]byte, 4)
				binary.BigEndian.PutUint32(trainerCrc32Bytes, rom.Header20.TrainerCRC32)
				returnString = returnString + "Trainer CRC32: " + strings.ToUpper(hex.EncodeToString(trainerCrc32Bytes)) + "\n"

				returnString = returnString + "Trainer MD5: " + strings.ToUpper(hex.EncodeToString(rom.Header20.TrainerMD5[:])) + "\n"
				returnString = returnString + "Trainer SHA1: " + strings.ToUpper(hex.EncodeToString(rom.Header20.TrainerSHA1[:])) + "\n"
				returnString = returnString + "Trainer SHA256: " + strings.ToUpper(hex.EncodeToString(rom.Header20.TrainerSHA256[:])) + "\n"
			}
		}

		returnString = returnString + "Misc ROM Size: " + strconv.Itoa(int(rom.Header20.MiscROMCalculatedSize)) + " bytes\n"

		if rom.Header20.MiscROMCalculatedSize > 0 {
			returnString = returnString + "Number of misc ROMs: " + strconv.Itoa(int(rom.Header20.MiscROMs)) + "\n"

			miscRomSum16Bytes := make([]byte, 2)
			binary.BigEndian.PutUint16(miscRomSum16Bytes, rom.Header20.MiscROMSum16)
			returnString = returnString + "Misc ROM Sum16: " + strings.ToUpper(hex.EncodeToString(miscRomSum16Bytes)) + "\n"

			miscRomCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(miscRomCrc32Bytes, rom.Header20.MiscROMCRC32)
			returnString = returnString + "Misc ROM CRC32: " + strings.ToUpper(hex.EncodeToString(miscRomCrc32Bytes)) + "\n"

			returnString = returnString + "Misc ROM MD5: " + strings.ToUpper(hex.EncodeToString(rom.Header20.MiscROMMD5[:])) + "\n"
			returnString = returnString + "Misc ROM SHA1: " + strings.ToUpper(hex.EncodeToString(rom.Header20.MiscROMSHA1[:])) + "\n"
			returnString = returnString + "Misc ROM SHA256: " + strings.ToUpper(hex.EncodeToString(rom.Header20.MiscROMSHA256[:])) + "\n"
		}

		if rom.Header20.PRGRAMSize == 0 {
			returnString = returnString + "PRG RAM Size: 0 bytes\n"
		} else {
			var prgRamSizeBytes uint64 = 64
			for i := 0; uint8(i) < rom.Header20.PRGRAMSize; i++ {
				prgRamSizeBytes = prgRamSizeBytes << 1
			}

			returnString = returnString + "PRG RAM Size: " + strconv.Itoa(int(prgRamSizeBytes)) + " bytes\n"
		}

		if rom.Header20.PRGNVRAMSize == 0 {
			returnString = returnString + "PRG NVRAM Size: 0 bytes\n"
		} else {
			var prgNvramSizeBytes uint64 = 64
			for i := 0; uint8(i) < rom.Header20.PRGNVRAMSize; i++ {
				prgNvramSizeBytes = prgNvramSizeBytes << 1
			}

			returnString = returnString + "PRG NVRAM Size: " + strconv.Itoa(int(prgNvramSizeBytes)) + " bytes\n"
		}

		if rom.Header20.CHRRAMSize == 0 {
			returnString = returnString + "CHR RAM Size: 0 bytes\n"
		} else {
			var chrRamSizeBytes uint64 = 64
			for i := 0; uint8(i) < rom.Header20.CHRRAMSize; i++ {
				chrRamSizeBytes = chrRamSizeBytes << 1
			}

			returnString = returnString + "CHR RAM Size: " + strconv.Itoa(int(chrRamSizeBytes)) + " bytes\n"
		}

		if rom.Header20.CHRNVRAMSize == 0 {
			returnString = returnString + "CHR NVRAM Size: 0 bytes\n"
		} else {
			var chrNvramSizeBytes uint64 = 64
			for i := 0; uint8(i) < rom.Header20.CHRNVRAMSize; i++ {
				chrNvramSizeBytes = chrNvramSizeBytes << 1
			}

			returnString = returnString + "CHR NVRAM Size: " + strconv.Itoa(int(chrNvramSizeBytes)) + " bytes\n"
		}

		if !rom.Header20.MirroringType {
			returnString = returnString + "Mirroring Type: Horizontal or mapper-controlled\n"
		} else {
			returnString = returnString + "Mirroring Type: Vertical\n"
		}

		if !rom.Header20.Battery {
			returnString = returnString + "Battery Backup: No\n"
		} else {
			returnString = returnString + "Battery Backup: Yes\n"
		}

		if !rom.Header20.FourScreen {
			returnString = returnString + "Hard-wired Four Screen Mode: No\n"
		} else {
			returnString = returnString + "Hard-wired Four Screen Mode: Yes\n"
		}

		if rom.Header20.ConsoleType < 3 {
			if rom.Header20.ConsoleType == 0 {
				returnString = returnString + "Console Type: Regular NES/Famicom/Dendy\n"
			} else if rom.Header20.ConsoleType == 1 {
				returnString = returnString + "Console Type: Nintendo Vs. System\n"
				returnString = returnString + "Vs. PPU Type: " + getVsPPUTypeString(rom.Header20.VsPPUType) + "\n"
				returnString = returnString + "Vs. System Type: " + getVsSystemTypeString(rom.Header20.VsHardwareType) + "\n"
			} else {
				returnString = returnString + "Console Type: Playchoice 10\n"
			}
		} else {
			returnString = returnString + "Console Type: " + getConsoleTypeString(rom.Header20.ExtendedConsoleType) + "\n"
		}

		returnString = returnString + "Mapper: " + strconv.Itoa(int(rom.Header20.Mapper)) + "\n"
		returnString = returnString + "Submapper: " + strconv.Itoa(int(rom.Header20.SubMapper)) + "\n"

		if rom.Header20.CPUPPUTiming == 0 {
			returnString = returnString + "CPU/PPU Timing: RP2C02 (\"NTSC NES\")\n"
		} else if rom.Header20.CPUPPUTiming == 1 {
			returnString = returnString + "CPU/PPU Timing: RP2C07 (\"Licensed PAL NES\")\n"
		} else if rom.Header20.CPUPPUTiming == 2 {
			returnString = returnString + "CPU/PPU Timing: Multiple-region\n"
		} else if rom.Header20.CPUPPUTiming == 3 {
			returnString = returnString + "CPU/PPU Timing: UMC 6527P (\"Dendy\")\n"
		}

		returnString = returnString + "Default Expansion Device: " + getDefaultExpansionDeviceString(rom.Header20.DefaultExpansion) + "\n"

		if rom.HeaderData != nil {
			returnString = returnString + "ROM Header (Existing):   " + strings.ToUpper(hex.EncodeToString(rom.HeaderData)) + "\n"
		}

		calculatedHeaderBytes, err := EncodeNESROMHeader(rom, false, true)
		if err != nil {
			return ""
		}

		returnString = returnString + "ROM Header (Calculated): " + strings.ToUpper(hex.EncodeToString(calculatedHeaderBytes))
	} else if rom.Header10 != nil {
		returnString = returnString + "PRG ROM Size: " + strconv.Itoa(int(rom.Header10.PRGROMCalculatedSize)) + " bytes\n"

		prgSum16Bytes := make([]byte, 2)
		binary.BigEndian.PutUint16(prgSum16Bytes, rom.Header10.PRGROMSum16)
		returnString = returnString + "PRG ROM Sum16: " + strings.ToUpper(hex.EncodeToString(prgSum16Bytes)) + "\n"

		prgCrc32Bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(prgCrc32Bytes, rom.Header10.PRGROMCRC32)
		returnString = returnString + "PRG ROM CRC32: " + strings.ToUpper(hex.EncodeToString(prgCrc32Bytes)) + "\n"

		returnString = returnString + "PRG ROM MD5: " + strings.ToUpper(hex.EncodeToString(rom.Header10.PRGROMMD5[:])) + "\n"
		returnString = returnString + "PRG ROM SHA1: " + strings.ToUpper(hex.EncodeToString(rom.Header10.PRGROMSHA1[:])) + "\n"
		returnString = returnString + "PRG ROM SHA256: " + strings.ToUpper(hex.EncodeToString(rom.Header10.PRGROMSHA256[:])) + "\n"

		returnString = returnString + "CHR ROM Size: " + strconv.Itoa(int(rom.Header10.CHRROMCalculatedSize)) + " bytes\n"

		if rom.Header10.CHRROMCalculatedSize > 0 {
			chrSum16Bytes := make([]byte, 2)
			binary.BigEndian.PutUint16(chrSum16Bytes, rom.Header10.CHRROMSum16)
			returnString = returnString + "CHR ROM Sum16: " + strings.ToUpper(hex.EncodeToString(chrSum16Bytes)) + "\n"

			chrCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(chrCrc32Bytes, rom.Header10.CHRROMCRC32)
			returnString = returnString + "CHR ROM CRC32: " + strings.ToUpper(hex.EncodeToString(chrCrc32Bytes)) + "\n"

			returnString = returnString + "CHR ROM MD5: " + strings.ToUpper(hex.EncodeToString(rom.Header10.CHRROMMD5[:])) + "\n"
			returnString = returnString + "CHR ROM SHA1: " + strings.ToUpper(hex.EncodeToString(rom.Header10.CHRROMSHA1[:])) + "\n"
			returnString = returnString + "CHR ROM SHA256: " + strings.ToUpper(hex.EncodeToString(rom.Header10.CHRROMSHA256[:])) + "\n"
		}

		if !rom.Header10.Trainer {
			returnString = returnString + "Trainer Size: 0 bytes\n"
		} else {
			returnString = returnString + "Trainer Size: " + strconv.Itoa(int(rom.Header10.TrainerCalculatedSize)) + " bytes\n"

			if rom.Header10.TrainerCalculatedSize > 0 {
				trainerSum16Bytes := make([]byte, 2)
				binary.BigEndian.PutUint16(trainerSum16Bytes, rom.Header10.TrainerSum16)
				returnString = returnString + "Trainer Sum16: " + strings.ToUpper(hex.EncodeToString(trainerSum16Bytes)) + "\n"

				trainerCrc32Bytes := make([]byte, 4)
				binary.BigEndian.PutUint32(trainerCrc32Bytes, rom.Header10.TrainerCRC32)
				returnString = returnString + "Trainer CRC32: " + strings.ToUpper(hex.EncodeToString(trainerCrc32Bytes)) + "\n"

				returnString = returnString + "Trainer MD5: " + strings.ToUpper(hex.EncodeToString(rom.Header10.TrainerMD5[:])) + "\n"
				returnString = returnString + "Trainer SHA1: " + strings.ToUpper(hex.EncodeToString(rom.Header10.TrainerSHA1[:])) + "\n"
				returnString = returnString + "Trainer SHA256: " + strings.ToUpper(hex.EncodeToString(rom.Header10.TrainerSHA256[:])) + "\n"
			}
		}

		returnString = returnString + "PRG RAM Size: " + strconv.Itoa(8192*int(rom.Header10.PRGRAMSize)) + " bytes\n"

		if !rom.Header10.FourScreen {
			if !rom.Header10.MirroringType {
				returnString = returnString + "Mirroring Type: Horizontal (vertical arrangement) (CIRAM A10 = PPU A11)\n"
			} else {
				returnString = returnString + "Mirroring Type: Vertical (horizontal arrangement) (CIRAM A10 = PPU A10)\n"
			}
		} else {
			returnString = returnString + "Mirroring Type: N/A (Four-Screen VRAM)"
		}

		if !rom.Header10.Battery {
			returnString = returnString + "Battery Backup: No\n"
		} else {
			returnString = returnString + "Battery Backup: Yes\n"
		}

		if !rom.Header10.VsUnisystem {
			returnString = returnString + "Vs. Unisystem: No\n"
		} else {
			returnString = returnString + "Vs. Unisystem: Yes\n"
		}

		if !rom.Header10.PlayChoice10 {
			returnString = returnString + "Playchoice 10: No\n"
		} else {
			returnString = returnString + "Playchoice 10: Yes\n"
		}

		returnString = returnString + "Mapper: " + strconv.Itoa(int(rom.Header10.Mapper)) + "\n"

		if !rom.Header10.TVSystem {
			returnString = returnString + "TV System: NTSC\n"
		} else {
			returnString = returnString + "TV System: PAL\n"
		}

		if rom.HeaderData != nil {
			returnString = returnString + "ROM Header (Existing):   " + strings.ToUpper(hex.EncodeToString(rom.HeaderData)) + "\n"
		}

		calculatedHeaderBytes, err := EncodeNESROMHeader(rom, true, true)
		if err != nil {
			return ""
		}

		returnString = returnString + "ROM Header (Calculated): " + strings.ToUpper(hex.EncodeToString(calculatedHeaderBytes))
	} else {
		return ""
	}

	return returnString
}

// Read data from a byte slice and decode it into an NESROM struct
func DecodeNESROM(inputFile []byte, enableInes bool, preserveTrainer bool, relativeLocation string) (*NESROM, error) {
	headerVersion := 2
	fileSize := uint64(len(inputFile))
	rawROMBytes, rawHeaderBytes, rawTrainerBytes, _ := getStrippedRom(inputFile)

	romData := &NESROM{}

	// Metadata, including checksums
	romData.RelativePath = relativeLocation
	romData.CRC32 = crc32.ChecksumIEEE(rawROMBytes)
	romData.MD5 = md5.Sum(rawROMBytes)
	romData.SHA1 = sha1.Sum(rawROMBytes)
	romData.SHA256 = sha256.Sum256(rawROMBytes)
	romData.Size = uint64(len(rawROMBytes))
	romData.ROMData = rawROMBytes
	romData.HeaderData = rawHeaderBytes

	if preserveTrainer {
		romData.TrainerData = rawTrainerBytes
	}

	if fileSize < 16 {
		return romData, &NESROMError{Text: "File too small to be a headered NES ROM."}
	}

	if bytes.Compare(inputFile[0:4], []byte(NES_HEADER_MAGIC)) != 0 {
		return romData, &NESROMError{Text: "Unable to find NES magic."}
	}

	if (inputFile[7]&NES_20_AND_MASK) != NES_20_AND_MASK || (inputFile[7]|NES_20_OR_MASK) != NES_20_OR_MASK {
		if !enableInes {
			return romData, &NESROMError{Text: "Not an NES 2.0 ROM."}
		} else {
			headerVersion = 1
		}
	}

	if headerVersion == 2 {
		header20Data := &NES20Header{}
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

		if romData.MiscROMData != nil && len(romData.MiscROMData) > 0 {
			header20Data.MiscROMCalculatedSize = uint64(len(romData.MiscROMData))
		} else {
			header20Data.MiscROMCalculatedSize = 0
		}

		header20Data.PRGRAMSize = inputFile[10] & 0b00001111
		header20Data.PRGNVRAMSize = (inputFile[10] & 0b11110000) >> 4
		header20Data.CHRRAMSize = inputFile[11] & 0b00001111
		header20Data.CHRNVRAMSize = (inputFile[11] & 0b11110000) >> 4

		header20Data.MirroringType = (inputFile[6] & 0b00000001) == 0b00000001
		header20Data.Battery = (inputFile[6] & 0b00000010) == 0b00000010

		if !preserveTrainer {
			header20Data.Trainer = false
			header20Data.TrainerCalculatedSize = 0
		} else {
			header20Data.Trainer = (inputFile[6] & 0b00000100) == 0b00000100

			if header20Data.Trainer {
				header20Data.TrainerCalculatedSize = 512
			} else {
				header20Data.TrainerCalculatedSize = 0
			}
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
		header10Data := &NES10Header{}
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
			header10Data.TrainerCalculatedSize = 0
		} else {
			header10Data.Trainer = (inputFile[6] & 0b00000100) == 0b00000100

			if header10Data.Trainer {
				header10Data.TrainerCalculatedSize = 512
			} else {
				header10Data.TrainerCalculatedSize = 0
			}
		}

		header10Data.FourScreen = (inputFile[6] & 0b00001000) == 0b00001000
		header10Data.Mapper = ((inputFile[6] & 0b11110000) >> 4) | (inputFile[7] & 0b11110000)
		header10Data.VsUnisystem = (inputFile[7] & 0b00000001) == 0b00000001
		header10Data.PlayChoice10 = (inputFile[7] & 0b00000010) == 0b00000010
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

// Encode a byte slice from an NESROM struct
func EncodeNESROM(romModel *NESROM, enableInes bool, truncateRom bool, preserveTrainer bool) ([]byte, error) {
	headerVersion := 2

	if romModel.Header20 == nil {
		if romModel.Header10 == nil || !enableInes {
			return nil, &NESROMError{"Unable to find valid header on ROM model."}
		} else {
			headerVersion = 1
		}
	}

	headerBytes, err := EncodeNESROMHeader(romModel, enableInes, preserveTrainer)
	if err != nil {
		return nil, err
	}

	var rawRomBytes []byte

	romBytes := headerBytes

	if headerVersion == 2 {
		if truncateRom {
			rawRomBytes = romModel.ROMData[0:(romModel.Header20.PRGROMCalculatedSize + romModel.Header20.CHRROMCalculatedSize)]
		} else {
			rawRomBytes = romModel.ROMData
		}
	} else if headerVersion == 1 {
		if truncateRom && romModel.Header20.MiscROMs == 0 {
			rawRomBytes = romModel.ROMData[0:(romModel.Header10.PRGROMCalculatedSize + romModel.Header10.CHRROMCalculatedSize)]
		} else {
			rawRomBytes = romModel.ROMData
		}
	}

	if !preserveTrainer || (romModel.Header20 != nil && !romModel.Header20.Trainer) || (romModel.Header10 != nil && !romModel.Header10.Trainer) || (romModel.TrainerData != nil && len(romModel.TrainerData) != 512) {
		romBytes = append(romBytes, rawRomBytes...)
	} else {
		romBytes = append(romBytes, romModel.TrainerData...)
		romBytes = append(romBytes, rawRomBytes...)
	}

	return romBytes, nil
}

func EncodeNESROMHeader(romModel *NESROM, enableInes bool, preserveTrainer bool) ([]byte, error) {
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
		} else if romModel.Header20.CHRROMSizeExponent > 0 || romModel.Header20.CHRROMSizeMultiplier > 0 {
			headerBytes[5] = (romModel.Header20.CHRROMSizeExponent << 2) | romModel.Header20.CHRROMSizeMultiplier
			headerBytes[9] = headerBytes[9] | 0b11110000
		} else {
			headerBytes[5] = 0
			headerBytes[9] = headerBytes[9] & 0b00001111
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
		if romModel.Header10.PlayChoice10 {
			Flag7Byte = Flag7Byte | 0b00000010
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
	}

	return headerBytes, nil
}

// Take a total byte size for a PRG or CHR rom and factor it into the possible
// size notations for NES 2.0, giving preference to the non-exponential size
// calculation.
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
		sizeMultiplier = 3
	} else if romSize%5 == 0 {
		sizeMultiplier = 5
	} else if romSize%7 == 0 {
		sizeMultiplier = 7
	} else {
		sizeMultiplier = 1
	}

	var tempSize uint64

	if sizeMultiplier > 0 {
		tempSize = romSize / uint64(sizeMultiplier)
	} else {
		tempSize = romSize
	}

	sizeMultiplier = (sizeMultiplier - 1) >> 1

	sizeExponent = 0

	for tempSize > 1 {
		tempSize = tempSize >> 1
		sizeExponent = sizeExponent + 1
	}

	return 0, sizeExponent, sizeMultiplier
}

// Update size metadata based on the byte slice size, the total size of the segment in metadata, or the
// factored exponential size of the segment in metadata
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

// Update the checksums for the various elements
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

		if nesRom.CHRROMData != nil && len(nesRom.CHRROMData) > 0 {
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

		if nesRom.MiscROMData != nil && len(nesRom.MiscROMData) > 0 {
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

		if nesRom.TrainerData != nil && len(nesRom.TrainerData) > 0 {
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

		if nesRom.CHRROMData != nil && len(nesRom.CHRROMData) > 0 {
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

		if nesRom.TrainerData != nil && len(nesRom.TrainerData) > 0 {
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

// If we don't have any misc ROMs, then any extra data based the end of
// the CHR ROM is probably garbage that we can truncate.
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

// Get the ROM without the header data, along with any trainer data included.
func getStrippedRom(inputFile []byte) ([]byte, []byte, []byte, error) {
	fileSize := uint64(len(inputFile))
	if fileSize < 16 {
		return inputFile, nil, nil, &NESROMError{Text: "File too small to be a headered NES ROM."}
	}

	if bytes.Compare(inputFile[0:4], []byte(NES_HEADER_MAGIC)) != 0 {
		return inputFile, nil, nil, &NESROMError{Text: "Unable to find NES magic."}
	}

	hasTrainer := (inputFile[6] & 0b00000100) == 0b00000100

	if hasTrainer && fileSize < 528 {
		return inputFile, nil, nil, &NESROMError{Text: "Header indicates trainer data, but file too small for one."}
	}

	if !hasTrainer {
		return inputFile[16:fileSize], inputFile[0:16], nil, nil
	} else {
		return inputFile[528:fileSize], inputFile[0:16], inputFile[16:528], nil
	}
}

// Calculate the sum16 value used by Nintendo for production verification.
func calculateSum16(inputData []byte) (uint16, error) {
	if inputData == nil {
		return 0, &NESROMError{Text: "Cannot calculate sum16 for a null segment."}
	}

	var byteSum uint64 = 0
	for i := range inputData {
		byteSum = byteSum + uint64(inputData[i])
	}

	return uint16(byteSum), nil
}

// Get the PRG, CHR, and misc ROM data from the raw, headerless ROM data
func getSplitRomData(inputData []byte, prgRomSize uint64, chrRomSize uint64) ([]byte, []byte, []byte, error) {
	if inputData == nil {
		return nil, nil, nil, &NESROMError{Text: "Cannot extract PRGROM and/or CHRROM data from a null segment."}
	}

	if len(inputData) < int(prgRomSize+chrRomSize) {
		return nil, nil, nil, &NESROMError{Text: "Invalid PRGROM and/or CHRROM size(s) detected during extraction."}
	}

	prgRomData := inputData[0:prgRomSize]
	chrRomData := inputData[prgRomSize:(prgRomSize + chrRomSize)]
	miscRomData := make([]byte, 0)

	if prgRomSize+chrRomSize < uint64(len(inputData)) {
		miscRomData = inputData[(prgRomSize + chrRomSize):]
	}

	return prgRomData, chrRomData, miscRomData, nil
}

func getConsoleTypeString(consoleType uint8) string {
	switch consoleType {
	case 3:
		return "Regular Famiclone, but with CPU that supports Decimal Mode (e.g. Bit Corporation Creator)"
	case 4:
		return "V.R. Technology VT01 with monochrome palette"
	case 5:
		return "V.R. Technology VT01 with red/cyan STN palette"
	case 6:
		return "V.R. Technology VT02"
	case 7:
		return "V.R. Technology VT03"
	case 8:
		return "V.R. Technology VT09"
	case 9:
		return "V.R. Technology VT32"
	case 10:
		return "V.R. Technology VT369"
	case 11:
		return "UMC UM6578"
	default:
		return "Unknown/Undefined"
	}
}

func getVsPPUTypeString(vsPpuType uint8) string {
	switch vsPpuType {
	case 0:
		return "RP2C03B"
	case 1:
		return "RP2C03G"
	case 2:
		return "RP2C04-0001"
	case 3:
		return "RP2C04-0002"
	case 4:
		return "RP2C04-0003"
	case 5:
		return "RP2C04-0004"
	case 6:
		return "RC2C03B"
	case 7:
		return "RC2C03C"
	case 8:
		return "RC2C05-01 ($2002 AND $?? =$1B)"
	case 9:
		return "RC2C05-02 ($2002 AND $3F =$3D)"
	case 10:
		return "RC2C05-03 ($2002 AND $1F =$1C)"
	case 11:
		return "RC2C05-04 ($2002 AND $1F =$1B)"
	case 12:
		return "RC2C05-05 ($2002 AND $1F =unknown)"
	default:
		return "Unknown/Undefined"
	}
}

func getVsSystemTypeString(vsSystemType uint8) string {
	switch vsSystemType {
	case 0:
		return "Vs. Unisystem (normal)"
	case 1:
		return "Vs. Unisystem (RBI Baseball protection)"
	case 2:
		return "Vs. Unisystem (TKO Boxing protection)"
	case 3:
		return "Vs. Unisystem (Super Xevious protection)"
	case 4:
		return "Vs. Unisystem (Vs. Ice Climber Japan protection)"
	case 5:
		return "Vs. Dual System (normal)"
	case 6:
		return "Vs. Dual System (Raid on Bungeling Bay protection)"
	default:
		return "Unknown/Undefined"
	}
}

func getDefaultExpansionDeviceString(defaultExpansion uint8) string {
	switch defaultExpansion {
	case 0:
		return "Unspecified"
	case 1:
		return "Standard NES/Famicom controllers"
	case 2:
		return "NES Four Score/Satellite with two additional standard controllers"
	case 3:
		return "Famicom Four Players Adapter with two additional standard controllers"
	case 4:
		return "Vs. System"
	case 5:
		return "Vs. System with reversed inputs"
	case 6:
		return "Vs. Pinball (Japan)"
	case 7:
		return "Vs. Zapper"
	case 8:
		return "Zapper ($4017)"
	case 9:
		return "Two Zappers"
	case 10:
		return "Bandai Hyper Shot Lightgun"
	case 11:
		return "Power Pad Side A"
	case 12:
		return "Power Pad Side B"
	case 13:
		return "Family Trainer Side A"
	case 14:
		return "Family Trainer Side B"
	case 15:
		return "Arkanoid Vaus Controller (NES)"
	case 16:
		return "Arkanoid Vaus Controller (Famicom)"
	case 17:
		return "Two Vaus Controllers plus Famicom Data Recorder"
	case 18:
		return "Konami Hyper Shot Controller"
	case 19:
		return "Coconuts Pachinko Controller"
	case 20:
		return "Exciting Boxing Punching Bag (Blowup Doll)"
	case 21:
		return "Jissen Mahjong Controller"
	case 22:
		return "Party Tap"
	case 23:
		return "Oeka Kids Tablet"
	case 24:
		return "Sunsoft Barcode Battler"
	case 25:
		return "Miracle Piano Keyboard"
	case 26:
		return "Pokkun Moguraa (Whack-a-Mole Mat and Mallet)"
	case 27:
		return "Top Rider (Inflatable Bicycle)"
	case 28:
		return "Double-Fisted (Requires or allows use of two controllers by one player)"
	case 29:
		return "Famicom 3D System"
	case 30:
		return "Doremikko Keyboard"
	case 31:
		return "R.O.B. Gyro Set"
	case 32:
		return "Famicom Data Recorder (don't emulate keyboard)"
	case 33:
		return "ASCII Turbo File"
	case 34:
		return "IGS Storage Battle Box"
	case 35:
		return "Family BASIC Keyboard plus Famicom Data Recorder"
	case 36:
		return "Dongda PEC-586 Keyboard"
	case 37:
		return "Bit Corp. Bit-79 Keyboard"
	case 38:
		return "Subor Keyboard"
	case 39:
		return "Subor Keyboard plus mouse (3x8-bit protocol)"
	case 40:
		return "Subor Keyboard plus mouse (24-bit protocol)"
	case 41:
		return "SNES Mouse ($4017.d0)"
	case 42:
		return "Multicart"
	case 43:
		return "Two SNES controllers replacing the two standard NES controllers"
	case 44:
		return "RacerMate Bicycle"
	case 45:
		return "U-Force"
	case 46:
		return "R.O.B. Stack-Up"
	case 47:
		return "City Patrolman Lightgun"
	case 48:
		return "Sharp C1 Cassette Interface"
	case 49:
		return "Standard Controller with swapped Left-Right/Up-Down/B-A"
	case 50:
		return "Excalibor Sudoku Pad"
	case 51:
		return "ABL Pinball"
	case 52:
		return "Golden Nugget Casino extra buttons"
	default:
		return "Unknown/Undefined"
	}
}
