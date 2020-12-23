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

// https://forums.nesdev.com/viewtopic.php?f=3&t=19940
// This implements the nes20db XML format for interchange
// with tools which support it.

package FileTools

import (
	"NES20Tool/NESTool"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	SHA1_ZERO_SUM = "DA39A3EE5E6B4B0D3255BFEF95601890AFD80709"
)

type NES20DBXML struct {
	XMLName xml.Name       `xml:"nes20db"`
	Text    string         `xml:",chardata"`
	Date    string         `xml:"date,attr"`
	Games   []*NES20DBGame `xml:"game"`
}

type NES20DBGame struct {
	Text   string `xml:",chardata"`
	Prgrom struct {
		Text  string `xml:",chardata"`
		Size  uint64 `xml:"size,attr"`
		Crc32 string `xml:"crc32,attr"`
		Sha1  string `xml:"sha1,attr"`
		Sum16 string `xml:"sum16,attr"`
	} `xml:"prgrom"`
	Chrrom struct {
		Text  string `xml:",chardata"`
		Size  uint64 `xml:"size,attr"`
		Crc32 string `xml:"crc32,attr"`
		Sha1  string `xml:"sha1,attr"`
		Sum16 string `xml:"sum16,attr"`
	} `xml:"chrrom"`
	Rom struct {
		Text  string `xml:",chardata"`
		Size  uint64 `xml:"size,attr"`
		Crc32 string `xml:"crc32,attr"`
		Sha1  string `xml:"sha1,attr"`
	} `xml:"rom"`
	Pcb struct {
		Text      string `xml:",chardata"`
		Mapper    uint16 `xml:"mapper,attr"`
		Submapper uint8  `xml:"submapper,attr"`
		Mirroring string `xml:"mirroring,attr"`
		Battery   uint8  `xml:"battery,attr"`
	} `xml:"pcb"`
	Console struct {
		Text   string `xml:",chardata"`
		Type   uint8  `xml:"type,attr"`
		Region uint8  `xml:"region,attr"`
	} `xml:"console"`
	Expansion struct {
		Text string `xml:",chardata"`
		Type uint8  `xml:"type,attr"`
	} `xml:"expansion"`
	Chrram struct {
		Text string `xml:",chardata"`
		Size uint32 `xml:"size,attr"`
	} `xml:"chrram"`
	Prgnvram struct {
		Text string `xml:",chardata"`
		Size uint32 `xml:"size,attr"`
	} `xml:"prgnvram"`
	Prgram struct {
		Text string `xml:",chardata"`
		Size uint32 `xml:"size,attr"`
	} `xml:"prgram"`
	Miscrom struct {
		Text   string `xml:",chardata"`
		Size   uint64 `xml:"size,attr"`
		Crc32  string `xml:"crc32,attr"`
		Sha1   string `xml:"sha1,attr"`
		Number uint8  `xml:"number,attr"`
	} `xml:"miscrom"`
	Vs struct {
		Text     string `xml:",chardata"`
		Hardware uint8  `xml:"hardware,attr"`
		Ppu      uint8  `xml:"ppu,attr"`
	} `xml:"vs"`
	Chrnvram struct {
		Text string `xml:",chardata"`
		Size uint32 `xml:"size,attr"`
	} `xml:"chrnvram"`
	Trainer struct {
		Text  string `xml:",chardata"`
		Size  uint16 `xml:"size,attr"`
		Crc32 string `xml:"crc32,attr"`
		Sha1  string `xml:"sha1,attr"`
	} `xml:"trainer"`
}

// Take a map of NES 2.0 ROMs and marshal an XML file in nes20db format from them
func MarshalNES20DBXMLFromROMMap(nesRoms map[string]*NESTool.NESROM) (string, error) {
	romXml := &NES20DBXML{}

	romXml.Date = time.Now().Format("2006-01-02")

	for index := range nesRoms {
		if nesRoms[index].Header20 != nil {
			tempGame := &NES20DBGame{}

			tempGame.Prgrom.Size = nesRoms[index].Header20.PRGROMCalculatedSize
			tempGame.Prgrom.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[index].Header20.PRGROMSHA1[:]))

			prgCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(prgCrc32Bytes, nesRoms[index].Header20.PRGROMCRC32)
			tempGame.Prgrom.Crc32 = strings.ToUpper(hex.EncodeToString(prgCrc32Bytes))

			prgSum16Bytes := make([]byte, 2)
			binary.BigEndian.PutUint16(prgSum16Bytes, nesRoms[index].Header20.PRGROMSum16)

			tempGame.Prgrom.Sum16 = strings.ToUpper(hex.EncodeToString(prgSum16Bytes))

			tempGame.Chrrom.Size = nesRoms[index].Header20.CHRROMCalculatedSize
			tempGame.Chrrom.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[index].Header20.CHRROMSHA1[:]))

			chrCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(chrCrc32Bytes, nesRoms[index].Header20.CHRROMCRC32)
			tempGame.Chrrom.Crc32 = strings.ToUpper(hex.EncodeToString(chrCrc32Bytes))

			chrSum16Bytes := make([]byte, 2)
			binary.BigEndian.PutUint16(chrSum16Bytes, nesRoms[index].Header20.CHRROMSum16)

			tempGame.Chrrom.Sum16 = strings.ToUpper(hex.EncodeToString(chrSum16Bytes))

			tempGame.Rom.Size = nesRoms[index].Size
			tempGame.Rom.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[index].SHA1[:]))

			romCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(romCrc32Bytes, nesRoms[index].CRC32)
			tempGame.Rom.Crc32 = strings.ToUpper(hex.EncodeToString(romCrc32Bytes))

			if nesRoms[index].Header20.Battery {
				tempGame.Pcb.Battery = 1
			} else {
				tempGame.Pcb.Battery = 0
			}

			tempGame.Pcb.Mapper = nesRoms[index].Header20.Mapper
			tempGame.Pcb.Submapper = nesRoms[index].Header20.SubMapper

			// This has weird special cases for mirroring
			if tempGame.Pcb.Mapper == 30 {
				if !nesRoms[index].Header20.FourScreen {
					if !nesRoms[index].Header20.MirroringType {
						tempGame.Pcb.Mirroring = "H"
					} else {
						tempGame.Pcb.Mirroring = "V"
					}
				} else {
					if !nesRoms[index].Header20.MirroringType {
						tempGame.Pcb.Mirroring = "1"
					} else {
						tempGame.Pcb.Mirroring = "4"
					}
				}
			} else if tempGame.Pcb.Mapper == 218 {
				if !nesRoms[index].Header20.FourScreen {
					if !nesRoms[index].Header20.MirroringType {
						tempGame.Pcb.Mirroring = "H"
					} else {
						tempGame.Pcb.Mirroring = "V"
					}
				} else {
					if !nesRoms[index].Header20.MirroringType {
						tempGame.Pcb.Mirroring = "0"
					} else {
						tempGame.Pcb.Mirroring = "1"
					}
				}
			} else {
				if !nesRoms[index].Header20.FourScreen {
					if !nesRoms[index].Header20.MirroringType {
						tempGame.Pcb.Mirroring = "H"
					} else {
						tempGame.Pcb.Mirroring = "V"
					}
				} else {
					if !nesRoms[index].Header20.MirroringType {
						tempGame.Pcb.Mirroring = "4"
					} else {
						return "", errors.New("Invalid mirroring type and four screen setting for mapper " + strconv.FormatUint(uint64(tempGame.Pcb.Mapper), 10) + " in ROM: " + nesRoms[index].Name)
					}
				}
			}

			tempGame.Console.Region = nesRoms[index].Header20.CPUPPUTiming

			if nesRoms[index].Header20.ConsoleType < 3 {
				tempGame.Console.Type = nesRoms[index].Header20.ConsoleType
			} else {
				tempGame.Console.Type = nesRoms[index].Header20.ExtendedConsoleType
			}

			tempGame.Expansion.Type = nesRoms[index].Header20.DefaultExpansion

			if nesRoms[index].Header20.CHRRAMSize > 0 {
				tempGame.Chrram.Size = 64 << nesRoms[index].Header20.CHRRAMSize
			}

			if nesRoms[index].Header20.PRGNVRAMSize > 0 {
				tempGame.Prgnvram.Size = 64 << nesRoms[index].Header20.PRGNVRAMSize
			}

			if nesRoms[index].Header20.PRGRAMSize > 0 {
				tempGame.Prgram.Size = 64 << nesRoms[index].Header20.PRGRAMSize
			}

			tempGame.Miscrom.Size = nesRoms[index].Header20.MiscROMCalculatedSize

			miscRomCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(miscRomCrc32Bytes, nesRoms[index].Header20.MiscROMCRC32)
			tempGame.Miscrom.Crc32 = strings.ToUpper(hex.EncodeToString(miscRomCrc32Bytes))

			tempGame.Miscrom.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[index].Header20.MiscROMSHA1[:]))

			tempGame.Miscrom.Number = nesRoms[index].Header20.MiscROMs

			tempGame.Vs.Hardware = nesRoms[index].Header20.VsHardwareType
			tempGame.Vs.Ppu = nesRoms[index].Header20.VsPPUType

			if nesRoms[index].Header20.CHRNVRAMSize > 0 {
				tempGame.Chrnvram.Size = 64 << nesRoms[index].Header20.CHRNVRAMSize
			}

			if nesRoms[index].TrainerData != nil && len(nesRoms[index].TrainerData) == 512 {
				tempGame.Trainer.Size = 512

				trainerCrc32Bytes := make([]byte, 4)
				binary.BigEndian.PutUint32(trainerCrc32Bytes, nesRoms[index].Header20.TrainerCRC32)
				tempGame.Trainer.Crc32 = strings.ToUpper(hex.EncodeToString(trainerCrc32Bytes))

				tempGame.Trainer.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[index].Header20.TrainerSHA1[:]))
			}

			romXml.Games = append(romXml.Games, tempGame)
		}
	}

	xmlBytes, err := xml.MarshalIndent(romXml, "", "\t")
	if err != nil {
		return "", err
	}

	returnString := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" + "\n" + string(xmlBytes)
	return returnString, nil
}

// Unmarshal an nes20db XML file to a map of NESROM structs, with their SHA1 checksum as the key
func UnmarshalNES20DBXMLToROMMap(xmlPayload string) (map[string]*NESTool.NESROM, error) {
	xmlStruct := &NES20DBXML{}
	err := xml.Unmarshal([]byte(xmlPayload), xmlStruct)
	if err != nil {
		return nil, err
	}

	romMap := make(map[string]*NESTool.NESROM)

	for index := range xmlStruct.Games {
		tempRom := &NESTool.NESROM{}

		tempRom.Size = xmlStruct.Games[index].Rom.Size

		crc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Rom.Crc32))
		if err == nil {
			tempRom.CRC32 = binary.BigEndian.Uint32(crc32Bytes)
		}

		sha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Rom.Sha1))
		if err == nil {
			copy(tempRom.SHA1[:], sha1Bytes)
		}

		tempRomHeader20 := &NESTool.NES20Header{}
		tempRom.Header20 = tempRomHeader20

		tempRom.Header20.PRGROMCalculatedSize = xmlStruct.Games[index].Prgrom.Size

		prgRomSum16Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Prgrom.Sum16))
		if err == nil {
			tempRom.Header20.PRGROMSum16 = binary.BigEndian.Uint16(prgRomSum16Bytes)
		}

		prgRomCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Prgrom.Crc32))
		if err == nil {
			tempRom.Header20.PRGROMCRC32 = binary.BigEndian.Uint32(prgRomCrc32Bytes)
		}

		prgRomSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Prgrom.Sha1))
		if err == nil {
			copy(tempRom.Header20.PRGROMSHA1[:], prgRomSha1Bytes)
		}

		tempRom.Header20.CHRROMCalculatedSize = xmlStruct.Games[index].Chrrom.Size

		if xmlStruct.Games[index].Chrrom.Size > 0 {
			chrRomSum16Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Chrrom.Sum16))
			if err == nil {
				tempRom.Header20.CHRROMSum16 = binary.BigEndian.Uint16(chrRomSum16Bytes)
			}

			chrRomCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Chrrom.Crc32))
			if err == nil {
				tempRom.Header20.CHRROMCRC32 = binary.BigEndian.Uint32(chrRomCrc32Bytes)
			}

			chrRomSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Chrrom.Sha1))
			if err == nil {
				copy(tempRom.Header20.CHRROMSHA1[:], chrRomSha1Bytes)
			}
		} else {
			tempRom.Header20.CHRROMSum16 = 0
			tempRom.Header20.CHRROMCRC32 = 0
			chrRomSha1ZeroBytes, err := hex.DecodeString(strings.ToLower(SHA1_ZERO_SUM))
			if err == nil {
				copy(tempRom.Header20.CHRROMSHA1[:], chrRomSha1ZeroBytes)
			}
		}

		tempRom.Header20.Mapper = xmlStruct.Games[index].Pcb.Mapper
		tempRom.Header20.SubMapper = xmlStruct.Games[index].Pcb.Submapper

		if tempRom.Header20.Mapper == 30 {
			if xmlStruct.Games[index].Pcb.Mirroring == "H" {
				tempRom.Header20.MirroringType = false
				tempRom.Header20.FourScreen = false
			} else if xmlStruct.Games[index].Pcb.Mirroring == "V" {
				tempRom.Header20.MirroringType = true
				tempRom.Header20.FourScreen = false
			} else if xmlStruct.Games[index].Pcb.Mirroring == "1" {
				tempRom.Header20.MirroringType = false
				tempRom.Header20.FourScreen = true
			} else if xmlStruct.Games[index].Pcb.Mirroring == "4" {
				tempRom.Header20.MirroringType = true
				tempRom.Header20.FourScreen = true
			}
		} else if tempRom.Header20.Mapper == 218 {
			if xmlStruct.Games[index].Pcb.Mirroring == "H" {
				tempRom.Header20.MirroringType = false
				tempRom.Header20.FourScreen = false
			} else if xmlStruct.Games[index].Pcb.Mirroring == "V" {
				tempRom.Header20.MirroringType = true
				tempRom.Header20.FourScreen = false
			} else if xmlStruct.Games[index].Pcb.Mirroring == "0" {
				tempRom.Header20.MirroringType = false
				tempRom.Header20.FourScreen = true
			} else if xmlStruct.Games[index].Pcb.Mirroring == "1" {
				tempRom.Header20.MirroringType = true
				tempRom.Header20.FourScreen = true
			}
		} else {
			if xmlStruct.Games[index].Pcb.Mirroring == "H" {
				tempRom.Header20.MirroringType = false
				tempRom.Header20.FourScreen = false
			} else if xmlStruct.Games[index].Pcb.Mirroring == "V" {
				tempRom.Header20.MirroringType = true
				tempRom.Header20.FourScreen = false
			} else if xmlStruct.Games[index].Pcb.Mirroring == "4" {
				tempRom.Header20.MirroringType = false
				tempRom.Header20.FourScreen = true
			}
		}

		if xmlStruct.Games[index].Pcb.Battery == 1 {
			tempRom.Header20.Battery = true
		} else {
			tempRom.Header20.Battery = false
		}

		tempRom.Header20.CPUPPUTiming = xmlStruct.Games[index].Console.Region

		if xmlStruct.Games[index].Console.Type < 4 {
			tempRom.Header20.ConsoleType = xmlStruct.Games[index].Console.Type
		} else {
			tempRom.Header20.ConsoleType = 3
			tempRom.Header20.ExtendedConsoleType = xmlStruct.Games[index].Console.Type
		}

		tempRom.Header20.DefaultExpansion = xmlStruct.Games[index].Expansion.Type

		if xmlStruct.Games[index].Chrram.Size > 0 {
			chrRamShifts := uint8(0)
			chrRamTempSize := xmlStruct.Games[index].Chrram.Size

			for chrRamTempSize > 64 {
				chrRamTempSize = chrRamTempSize >> 1

				chrRamShifts = chrRamShifts + 1
			}

			tempRom.Header20.CHRRAMSize = chrRamShifts
		} else {
			tempRom.Header20.CHRRAMSize = 0
		}

		if xmlStruct.Games[index].Prgnvram.Size > 0 {
			prgNvramShifts := uint8(0)
			prgNvramTempSize := xmlStruct.Games[index].Prgnvram.Size

			for prgNvramTempSize > 64 {
				prgNvramTempSize = prgNvramTempSize >> 1

				prgNvramShifts = prgNvramShifts + 1
			}

			tempRom.Header20.PRGNVRAMSize = prgNvramShifts
		} else {
			tempRom.Header20.PRGNVRAMSize = 0
		}

		if xmlStruct.Games[index].Prgram.Size > 0 {
			prgRamShifts := uint8(0)
			prgRamTempSize := xmlStruct.Games[index].Prgram.Size

			for prgRamTempSize > 64 {
				prgRamTempSize = prgRamTempSize >> 1

				prgRamShifts = prgRamShifts + 1
			}

			tempRom.Header20.PRGRAMSize = prgRamShifts
		} else {
			tempRom.Header20.PRGRAMSize = 0
		}

		if xmlStruct.Games[index].Miscrom.Number > 0 {
			tempRom.Header20.MiscROMs = xmlStruct.Games[index].Miscrom.Number
			tempRom.Header20.MiscROMCalculatedSize = xmlStruct.Games[index].Miscrom.Size

			miscRomCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Miscrom.Crc32))
			if err == nil {
				tempRom.Header20.MiscROMCRC32 = binary.BigEndian.Uint32(miscRomCrc32Bytes)
			}

			miscRomSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Miscrom.Sha1))
			if err == nil {
				copy(tempRom.Header20.MiscROMSHA1[:], miscRomSha1Bytes)
			}
		}

		tempRom.Header20.VsPPUType = xmlStruct.Games[index].Vs.Ppu
		tempRom.Header20.VsHardwareType = xmlStruct.Games[index].Vs.Hardware

		if xmlStruct.Games[index].Chrnvram.Size > 0 {
			chrNvramShifts := uint8(0)
			chrNvramTempSize := xmlStruct.Games[index].Chrnvram.Size

			for chrNvramTempSize > 64 {
				chrNvramTempSize = chrNvramTempSize >> 1

				chrNvramShifts = chrNvramShifts + 1
			}

			tempRom.Header20.CHRNVRAMSize = chrNvramShifts
		} else {
			tempRom.Header20.CHRNVRAMSize = 0
		}

		err = NESTool.UpdateSizes(tempRom, NESTool.PRG_CANONICAL_SIZE_CALCULATED, NESTool.CHR_CANONICAL_SIZE_CALCULATED)
		if err != nil {
			return nil, err
		}

		if xmlStruct.Games[index].Trainer.Size > 0 {
			tempRom.Header20.Trainer = true
			tempRom.Header20.TrainerCalculatedSize = 512

			trainerCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Trainer.Crc32))
			if err == nil {
				tempRom.Header20.TrainerCRC32 = binary.BigEndian.Uint32(trainerCrc32Bytes)
			}

			trainerSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.Games[index].Trainer.Sha1))
			if err == nil {
				copy(tempRom.Header20.TrainerSHA1[:], trainerSha1Bytes)
			}
		}

		romMap["SHA1:"+strings.ToUpper(xmlStruct.Games[index].Rom.Sha1)] = tempRom
	}

	return romMap, nil
}
