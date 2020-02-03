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
package FileTools

import (
	"NES20Tool/NES20Tool"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"os"
	"strings"
)

type NESXML struct {
	XMLName xml.Name     `xml:"nesroms"`
	Text    string       `xml:",chardata"`
	XMLROMs []*NESXMLROM `xml:"rom"`
}

type NESXMLROM struct {
	Text         string `xml:",chardata"`
	Name         string `xml:"name,attr"`
	Size         uint64 `xml:"size,attr"`
	RelativePath string `xml:"relativePath,attr"`
	Crc32        string `xml:"crc32,attr"`
	Md5          string `xml:"md5,attr"`
	Sha1         string `xml:"sha1,attr"`
	Sha256       string `xml:"sha256,attr"`
	TrainerData  struct {
		Text string `xml:",chardata"`
	} `xml:"trainerData"`
	Header20 *NES20XMLFields `xml:"nes20"`
	Header10 *NES10XMLFields `xml:"ines"`
}

type NES20XMLFields struct {
	Prgrom struct {
		Text           string `xml:",chardata"`
		Size           uint16 `xml:"size,attr"`
		SizeExponent   uint8  `xml:"sizeExponent,attr"`
		SizeMultiplier uint8  `xml:"sizeMultiplier,attr"`
	} `xml:"prgrom"`
	Chrrom struct {
		Text           string `xml:",chardata"`
		Size           uint16 `xml:"size,attr"`
		SizeExponent   uint8  `xml:"sizeExponent,attr"`
		SizeMultiplier uint8  `xml:"sizeMultiplier,attr"`
	} `xml:"chrrom"`
	Prgram struct {
		Text string `xml:",chardata"`
		Size uint8  `xml:"size,attr"`
	} `xml:"prgram"`
	Prgnvram struct {
		Text string `xml:",chardata"`
		Size uint8  `xml:"size,attr"`
	} `xml:"prgnvram"`
	Chrram struct {
		Text string `xml:",chardata"`
		Size uint8  `xml:"size,attr"`
	} `xml:"chrram"`
	Chrnvram struct {
		Text string `xml:",chardata"`
		Size uint8  `xml:"size,attr"`
	} `xml:"chrnvram"`
	MirroringType struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"value,attr"`
	} `xml:"mirroringType"`
	Battery struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"value,attr"`
	} `xml:"battery"`
	Trainer struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"value,attr"`
	} `xml:"trainer"`
	FourScreen struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"value,attr"`
	} `xml:"fourScreen"`
	ConsoleType struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"consoleType"`
	Mapper struct {
		Text  string `xml:",chardata"`
		Value uint16 `xml:"value,attr"`
	} `xml:"mapper"`
	SubMapper struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"subMapper"`
	CpuPpuTiming struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"cpuPpuTiming"`
	VsHardwareType struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"vsHardwareType"`
	VsPpuType struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"vsPpuType"`
	ExtendedConsoleType struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"extendedConsoleType"`
	MiscRoms struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"miscRoms"`
	DefaultExpansion struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"defaultExpansion"`
}

type NES10XMLFields struct {
	Prgrom struct {
		Text string `xml:",chardata"`
		Size uint8  `xml:"size,attr"`
	} `xml:"prgrom"`
	Chrrom struct {
		Text string `xml:",chardata"`
		Size uint8  `xml:"size,attr"`
	} `xml:"chrrom"`
	MirroringType struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"value,attr"`
	} `xml:"mirroringType"`
	Battery struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"value,attr"`
	} `xml:"battery"`
	Trainer struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"value,attr"`
	} `xml:"trainer"`
	FourScreen struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"value,attr"`
	} `xml:"fourScreen"`
	Mapper struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"mapper"`
	VsUnisystem struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"value,attr"`
	} `xml:"vsUnisystem"`
	Prgram struct {
		Text string `xml:",chardata"`
		Size uint8  `xml:"size,attr"`
	} `xml:"prgram"`
	TvSystem struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"size,attr"`
	} `xml:"tvSystem"`
}

func MarshalXMLFromROMMap(nesRoms map[[32]byte]*NES20Tool.NESROM, enableInes bool, preserveTrainer bool, enableOrganization bool) (string, error) {
	romXml := &NESXML{}
	for key, _ := range nesRoms {
		if nesRoms[key].Header20 != nil {
			tempXmlRom := &NESXMLROM{}
			tempXmlHeader20 := &NES20XMLFields{}
			tempXmlRom.Header20 = tempXmlHeader20

			tempXmlRom.Name = nesRoms[key].Name
			tempXmlRom.Size = nesRoms[key].Size

			crc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(crc32Bytes, nesRoms[key].CRC32)
			tempXmlRom.Crc32 = hex.EncodeToString(crc32Bytes)

			tempXmlRom.Md5 = hex.EncodeToString(nesRoms[key].MD5[:])
			tempXmlRom.Sha1 = hex.EncodeToString(nesRoms[key].SHA1[:])
			tempXmlRom.Sha256 = hex.EncodeToString(nesRoms[key].SHA256[:])

			if enableOrganization {
				tempRelativePath := nesRoms[key].RelativePath
				if tempRelativePath[0] == os.PathSeparator {
					tempRelativePath = tempRelativePath[1:]
				}
				tempRelativePath = strings.Replace(tempRelativePath, string(os.PathSeparator), "/", -1)
				tempXmlRom.RelativePath = tempRelativePath
			}

			if preserveTrainer && nesRoms[key].TrainerData != nil {
				tempXmlRom.TrainerData.Text = hex.EncodeToString(nesRoms[key].TrainerData)
			}

			if nesRoms[key].Header20.PRGROMSize > 0 {
				tempXmlRom.Header20.Prgrom.Size = nesRoms[key].Header20.PRGROMSize
			} else {
				tempXmlRom.Header20.Prgrom.SizeExponent = nesRoms[key].Header20.PRGROMSizeExponent
				tempXmlRom.Header20.Prgrom.SizeMultiplier = nesRoms[key].Header20.PRGROMSizeMultiplier
			}

			if nesRoms[key].Header20.CHRROMSize > 0 {
				tempXmlRom.Header20.Chrrom.Size = nesRoms[key].Header20.CHRROMSize
			} else {
				tempXmlRom.Header20.Chrrom.SizeExponent = nesRoms[key].Header20.CHRROMSizeExponent
				tempXmlRom.Header20.Chrrom.SizeMultiplier = nesRoms[key].Header20.CHRROMSizeMultiplier
			}

			tempXmlRom.Header20.Prgram.Size = nesRoms[key].Header20.PRGRAMSize
			tempXmlRom.Header20.Prgnvram.Size = nesRoms[key].Header20.PRGNVRAMSize
			tempXmlRom.Header20.Chrram.Size = nesRoms[key].Header20.CHRRAMSize
			tempXmlRom.Header20.Chrnvram.Size = nesRoms[key].Header20.CHRNVRAMSize
			tempXmlRom.Header20.MirroringType.Value = nesRoms[key].Header20.MirroringType
			tempXmlRom.Header20.Battery.Value = nesRoms[key].Header20.Battery
			tempXmlRom.Header20.Trainer.Value = nesRoms[key].Header20.Trainer
			tempXmlRom.Header20.FourScreen.Value = nesRoms[key].Header20.FourScreen
			tempXmlRom.Header20.ConsoleType.Value = nesRoms[key].Header20.ConsoleType
			tempXmlRom.Header20.Mapper.Value = nesRoms[key].Header20.Mapper
			tempXmlRom.Header20.SubMapper.Value = nesRoms[key].Header20.SubMapper
			tempXmlRom.Header20.CpuPpuTiming.Value = nesRoms[key].Header20.CPUPPUTiming
			tempXmlRom.Header20.VsHardwareType.Value = nesRoms[key].Header20.VsHardwareType
			tempXmlRom.Header20.VsPpuType.Value = nesRoms[key].Header20.VsPPUType
			tempXmlRom.Header20.ExtendedConsoleType.Value = nesRoms[key].Header20.ExtendedConsoleType
			tempXmlRom.Header20.MiscRoms.Value = nesRoms[key].Header20.MiscROMs
			tempXmlRom.Header20.DefaultExpansion.Value = nesRoms[key].Header20.DefaultExpansion

			romXml.XMLROMs = append(romXml.XMLROMs, tempXmlRom)
		} else if enableInes && nesRoms[key].Header10 != nil {
			tempXmlRom := &NESXMLROM{}
			tempXmlHeader10 := &NES10XMLFields{}
			tempXmlRom.Header10 = tempXmlHeader10

			tempXmlRom.Name = nesRoms[key].Name
			tempXmlRom.Size = nesRoms[key].Size

			crc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(crc32Bytes, nesRoms[key].CRC32)
			tempXmlRom.Crc32 = hex.EncodeToString(crc32Bytes)

			tempXmlRom.Md5 = hex.EncodeToString(nesRoms[key].MD5[:])
			tempXmlRom.Sha1 = hex.EncodeToString(nesRoms[key].SHA1[:])
			tempXmlRom.Sha256 = hex.EncodeToString(nesRoms[key].SHA256[:])

			if enableOrganization {
				tempRelativePath := nesRoms[key].RelativePath
				if tempRelativePath[0] == os.PathSeparator {
					tempRelativePath = tempRelativePath[1:]
				}
				tempRelativePath = strings.Replace(tempRelativePath, string(os.PathSeparator), "/", -1)
				tempXmlRom.RelativePath = tempRelativePath
			}

			if preserveTrainer && nesRoms[key].TrainerData != nil {
				tempXmlRom.TrainerData.Text = hex.EncodeToString(nesRoms[key].TrainerData)
			}

			tempNesRom := nesRoms[key]
			if tempNesRom == nil {
			}

			tempXmlRom.Header10.Prgrom.Size = nesRoms[key].Header10.PRGROMSize
			tempXmlRom.Header10.Chrrom.Size = nesRoms[key].Header10.CHRROMSize
			tempXmlRom.Header10.MirroringType.Value = nesRoms[key].Header10.MirroringType
			tempXmlRom.Header10.Battery.Value = nesRoms[key].Header10.Battery
			tempXmlRom.Header10.Trainer.Value = nesRoms[key].Header10.Trainer
			tempXmlRom.Header10.FourScreen.Value = nesRoms[key].Header10.FourScreen
			tempXmlRom.Header10.Mapper.Value = nesRoms[key].Header10.Mapper
			tempXmlRom.Header10.VsUnisystem.Value = nesRoms[key].Header10.VsUnisystem
			tempXmlRom.Header10.Prgram.Size = nesRoms[key].Header10.PRGRAMSize
			tempXmlRom.Header10.TvSystem.Value = nesRoms[key].Header10.TVSystem

			romXml.XMLROMs = append(romXml.XMLROMs, tempXmlRom)
		}
	}

	xmlBytes, err := xml.MarshalIndent(romXml, "", "  ")
	if err != nil {
		return "", err
	}

	return string(xmlBytes), nil
}

func UnmarshalXMLToROMMap(xmlPayload string, enableInes bool, preserveTrainer bool, enableOrganization bool) (map[[32]byte]*NES20Tool.NESROM, error) {
	xmlStruct := &NESXML{}
	err := xml.Unmarshal([]byte(xmlPayload), xmlStruct)
	if err != nil {
		return nil, err
	}

	romMap := make(map[[32]byte]*NES20Tool.NESROM)

	for index, _ := range xmlStruct.XMLROMs {
		tempRom := &NES20Tool.NESROM{}

		crc32Bytes, err := hex.DecodeString(xmlStruct.XMLROMs[index].Crc32)
		if err == nil {
			tempRom.CRC32 = binary.BigEndian.Uint32(crc32Bytes)
		}

		md5Bytes, err := hex.DecodeString(xmlStruct.XMLROMs[index].Md5)
		if err == nil {
			copy(tempRom.MD5[:], md5Bytes)
		}

		sha1Bytes, err := hex.DecodeString(xmlStruct.XMLROMs[index].Sha1)
		if err == nil {
			copy(tempRom.SHA1[:], sha1Bytes)
		}

		sha256Bytes, err := hex.DecodeString(xmlStruct.XMLROMs[index].Sha256)
		if err == nil {
			copy(tempRom.SHA256[:], sha256Bytes)
		}

		tempRom.Name = xmlStruct.XMLROMs[index].Name
		tempRom.Size = xmlStruct.XMLROMs[index].Size

		if enableOrganization {
			tempRelativePath := xmlStruct.XMLROMs[index].RelativePath
			tempRelativePath = strings.Replace(tempRelativePath, "/", string(os.PathSeparator), -1)
			tempRom.RelativePath = tempRelativePath
		}

		if preserveTrainer && xmlStruct.XMLROMs[index].TrainerData.Text != "" {
			trainerDataBytes, err := hex.DecodeString(xmlStruct.XMLROMs[index].TrainerData.Text)
			if err == nil {
				tempRom.TrainerData = trainerDataBytes
			}
		}

		if xmlStruct.XMLROMs[index].Header20 != nil {
			tempRomHeader20 := &NES20Tool.NES20Header{}
			tempRom.Header20 = tempRomHeader20

			if xmlStruct.XMLROMs[index].Header20.Prgrom.Size > 0 {
				tempRom.Header20.PRGROMSize = xmlStruct.XMLROMs[index].Header20.Prgrom.Size
			} else {
				tempRom.Header20.PRGROMSizeExponent = xmlStruct.XMLROMs[index].Header20.Prgrom.SizeExponent
				tempRom.Header20.PRGROMSizeMultiplier = xmlStruct.XMLROMs[index].Header20.Prgrom.SizeMultiplier
			}

			if xmlStruct.XMLROMs[index].Header20.Chrrom.Size > 0 {
				tempRom.Header20.CHRROMSize = xmlStruct.XMLROMs[index].Header20.Chrrom.Size
			} else {
				tempRom.Header20.CHRROMSizeExponent = xmlStruct.XMLROMs[index].Header20.Chrrom.SizeExponent
				tempRom.Header20.CHRROMSizeMultiplier = xmlStruct.XMLROMs[index].Header20.Chrrom.SizeMultiplier
			}

			tempRom.Header20.PRGRAMSize = xmlStruct.XMLROMs[index].Header20.Prgram.Size
			tempRom.Header20.PRGNVRAMSize = xmlStruct.XMLROMs[index].Header20.Prgnvram.Size
			tempRom.Header20.CHRRAMSize = xmlStruct.XMLROMs[index].Header20.Chrram.Size
			tempRom.Header20.CHRNVRAMSize = xmlStruct.XMLROMs[index].Header20.Chrnvram.Size
			tempRom.Header20.MirroringType = xmlStruct.XMLROMs[index].Header20.MirroringType.Value
			tempRom.Header20.Battery = xmlStruct.XMLROMs[index].Header20.Battery.Value
			tempRom.Header20.Trainer = xmlStruct.XMLROMs[index].Header20.Trainer.Value
			tempRom.Header20.FourScreen = xmlStruct.XMLROMs[index].Header20.FourScreen.Value
			tempRom.Header20.ConsoleType = xmlStruct.XMLROMs[index].Header20.ConsoleType.Value
			tempRom.Header20.Mapper = xmlStruct.XMLROMs[index].Header20.Mapper.Value
			tempRom.Header20.SubMapper = xmlStruct.XMLROMs[index].Header20.SubMapper.Value
			tempRom.Header20.CPUPPUTiming = xmlStruct.XMLROMs[index].Header20.CpuPpuTiming.Value
			tempRom.Header20.VsHardwareType = xmlStruct.XMLROMs[index].Header20.VsHardwareType.Value
			tempRom.Header20.VsPPUType = xmlStruct.XMLROMs[index].Header20.VsPpuType.Value
			tempRom.Header20.ExtendedConsoleType = xmlStruct.XMLROMs[index].Header20.ExtendedConsoleType.Value
			tempRom.Header20.MiscROMs = xmlStruct.XMLROMs[index].Header20.MiscRoms.Value
			tempRom.Header20.DefaultExpansion = xmlStruct.XMLROMs[index].Header20.DefaultExpansion.Value
		} else if enableInes && xmlStruct.XMLROMs[index].Header10 != nil {
			tempRomHeader10 := &NES20Tool.NES10Header{}
			tempRom.Header10 = tempRomHeader10

			tempRom.Header10.PRGROMSize = xmlStruct.XMLROMs[index].Header10.Prgrom.Size
			tempRom.Header10.CHRROMSize = xmlStruct.XMLROMs[index].Header10.Chrrom.Size
			tempRom.Header10.MirroringType = xmlStruct.XMLROMs[index].Header10.MirroringType.Value
			tempRom.Header10.Battery = xmlStruct.XMLROMs[index].Header10.Battery.Value
			tempRom.Header10.Trainer = xmlStruct.XMLROMs[index].Header10.Trainer.Value
			tempRom.Header10.FourScreen = xmlStruct.XMLROMs[index].Header10.FourScreen.Value
			tempRom.Header10.Mapper = xmlStruct.XMLROMs[index].Header10.Mapper.Value
			tempRom.Header10.VsUnisystem = xmlStruct.XMLROMs[index].Header10.VsUnisystem.Value
			tempRom.Header10.PRGRAMSize = xmlStruct.XMLROMs[index].Header10.Prgram.Size
			tempRom.Header10.TVSystem = xmlStruct.XMLROMs[index].Header10.TvSystem.Value
		}

		romMap[tempRom.SHA256] = tempRom
	}

	return romMap, nil
}
