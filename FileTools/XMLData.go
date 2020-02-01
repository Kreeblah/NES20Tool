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
)

type NES20XML struct {
	XMLName xml.Name     `xml:"nes20"`
	Text    string       `xml:",chardata"`
	XMLROMs []*NESXMLROM `xml:"rom"`
}

type NESXMLROM struct {
	Text   string `xml:",chardata"`
	Name   string `xml:"name,attr"`
	Size   uint64 `xml:"size,attr"`
	Crc32  string `xml:"crc32,attr"`
	Md5    string `xml:"md5,attr"`
	Sha1   string `xml:"sha1,attr"`
	Sha256 string `xml:"sha256,attr"`
	Prgrom struct {
		Text string `xml:",chardata"`
		Size uint16 `xml:"size,attr"`
	} `xml:"prgrom"`
	Chrrom struct {
		Text string `xml:",chardata"`
		Size uint16 `xml:"size,attr"`
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

func MarshalXMLFromROMMap(nesRoms map[[32]byte]*NES20Tool.NESROM) (string, error) {
	romXml := &NES20XML{}
	for key, _ := range nesRoms {
		if nesRoms[key].Header != nil {
			tempXmlRom := &NESXMLROM{}
			tempXmlRom.Name = nesRoms[key].Name
			tempXmlRom.Size = nesRoms[key].Size

			crc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(crc32Bytes, nesRoms[key].CRC32)
			tempXmlRom.Crc32 = hex.EncodeToString(crc32Bytes)

			tempXmlRom.Md5 = hex.EncodeToString(nesRoms[key].MD5[:])
			tempXmlRom.Sha1 = hex.EncodeToString(nesRoms[key].SHA1[:])
			tempXmlRom.Sha256 = hex.EncodeToString(nesRoms[key].SHA256[:])
			tempXmlRom.Prgrom.Size = nesRoms[key].Header.PRGROMSize
			tempXmlRom.Chrrom.Size = nesRoms[key].Header.CHRROMSize
			tempXmlRom.Prgram.Size = nesRoms[key].Header.PRGRAMSize
			tempXmlRom.Prgnvram.Size = nesRoms[key].Header.PRGNVRAMSize
			tempXmlRom.Chrram.Size = nesRoms[key].Header.CHRRAMSize
			tempXmlRom.Chrnvram.Size = nesRoms[key].Header.CHRNVRAMSize
			tempXmlRom.MirroringType.Value = nesRoms[key].Header.MirroringType
			tempXmlRom.Battery.Value = nesRoms[key].Header.Battery
			tempXmlRom.Trainer.Value = nesRoms[key].Header.Trainer
			tempXmlRom.FourScreen.Value = nesRoms[key].Header.FourScreen
			tempXmlRom.ConsoleType.Value = nesRoms[key].Header.ConsoleType
			tempXmlRom.Mapper.Value = nesRoms[key].Header.Mapper
			tempXmlRom.SubMapper.Value = nesRoms[key].Header.SubMapper
			tempXmlRom.CpuPpuTiming.Value = nesRoms[key].Header.CPUPPUTiming
			tempXmlRom.VsHardwareType.Value = nesRoms[key].Header.VsHardwareType
			tempXmlRom.VsPpuType.Value = nesRoms[key].Header.VsPPUType
			tempXmlRom.ExtendedConsoleType.Value = nesRoms[key].Header.ExtendedConsoleType
			tempXmlRom.MiscRoms.Value = nesRoms[key].Header.MiscROMs
			tempXmlRom.DefaultExpansion.Value = nesRoms[key].Header.DefaultExpansion

			romXml.XMLROMs = append(romXml.XMLROMs, tempXmlRom)
		}
	}

	xmlBytes, err := xml.MarshalIndent(romXml, "", "  ")
	if err != nil {
		return "", err
	}

	return string(xmlBytes), nil
}

func UnmarshalXMLToROMMap(xmlPayload string) (map[[32]byte]*NES20Tool.NESROM, error) {
	xmlStruct := &NES20XML{}
	err := xml.Unmarshal([]byte(xmlPayload), xmlStruct)
	if err != nil {
		return nil, err
	}

	romMap := make(map[[32]byte]*NES20Tool.NESROM)

	for index, _ := range xmlStruct.XMLROMs {
		tempRom := &NES20Tool.NESROM{}
		tempRomHeader := &NES20Tool.NESHeader{}
		tempRom.Header = tempRomHeader

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
		tempRom.Header.PRGROMSize = xmlStruct.XMLROMs[index].Prgrom.Size
		tempRom.Header.CHRROMSize = xmlStruct.XMLROMs[index].Chrrom.Size
		tempRom.Header.PRGRAMSize = xmlStruct.XMLROMs[index].Prgram.Size
		tempRom.Header.PRGNVRAMSize = xmlStruct.XMLROMs[index].Prgnvram.Size
		tempRom.Header.CHRRAMSize = xmlStruct.XMLROMs[index].Chrram.Size
		tempRom.Header.CHRNVRAMSize = xmlStruct.XMLROMs[index].Chrnvram.Size
		tempRom.Header.MirroringType = xmlStruct.XMLROMs[index].MirroringType.Value
		tempRom.Header.Battery = xmlStruct.XMLROMs[index].Battery.Value
		tempRom.Header.Trainer = xmlStruct.XMLROMs[index].Trainer.Value
		tempRom.Header.FourScreen = xmlStruct.XMLROMs[index].FourScreen.Value
		tempRom.Header.ConsoleType = xmlStruct.XMLROMs[index].ConsoleType.Value
		tempRom.Header.Mapper = xmlStruct.XMLROMs[index].Mapper.Value
		tempRom.Header.SubMapper = xmlStruct.XMLROMs[index].SubMapper.Value
		tempRom.Header.CPUPPUTiming = xmlStruct.XMLROMs[index].CpuPpuTiming.Value
		tempRom.Header.VsHardwareType = xmlStruct.XMLROMs[index].VsHardwareType.Value
		tempRom.Header.VsPPUType = xmlStruct.XMLROMs[index].VsPpuType.Value
		tempRom.Header.ExtendedConsoleType = xmlStruct.XMLROMs[index].ExtendedConsoleType.Value
		tempRom.Header.MiscROMs = xmlStruct.XMLROMs[index].MiscRoms.Value
		tempRom.Header.DefaultExpansion = xmlStruct.XMLROMs[index].DefaultExpansion.Value

		romMap[tempRom.SHA256] = tempRom
	}

	return romMap, nil
}
