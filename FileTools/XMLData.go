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

package FileTools

import (
	"NES20Tool/FDSTool"
	"NES20Tool/NESTool"
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
	Header20   *NES20XMLFields `xml:"nes20"`
	Header10   *NES10XMLFields `xml:"ines"`
	FDSArchive *FDSXMLFields   `xml:"fds"`
}

type NES20XMLFields struct {
	Prgrom struct {
		Text           string `xml:",chardata"`
		Size           uint16 `xml:"size,attr"`
		SizeExponent   uint8  `xml:"sizeExponent,attr"`
		SizeMultiplier uint8  `xml:"sizeMultiplier,attr"`
		Sum16          string `xml:"sum16,attr"`
		Crc32          string `xml:"crc32,attr"`
		Md5            string `xml:"md5,attr"`
		Sha1           string `xml:"sha1,attr"`
		Sha256         string `xml:"sha256,attr"`
	} `xml:"prgrom"`
	Chrrom struct {
		Text           string `xml:",chardata"`
		Size           uint16 `xml:"size,attr"`
		SizeExponent   uint8  `xml:"sizeExponent,attr"`
		SizeMultiplier uint8  `xml:"sizeMultiplier,attr"`
		Sum16          string `xml:"sum16,attr"`
		Crc32          string `xml:"crc32,attr"`
		Md5            string `xml:"md5,attr"`
		Sha1           string `xml:"sha1,attr"`
		Sha256         string `xml:"sha256,attr"`
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
		Text   string `xml:",chardata"`
		Value  bool   `xml:"value,attr"`
		Size   uint16 `xml:"size,attr"`
		Sum16  string `xml:"sum16,attr"`
		Crc32  string `xml:"crc32,attr"`
		Md5    string `xml:"md5,attr"`
		Sha1   string `xml:"sha1,attr"`
		Sha256 string `xml:"sha256,attr"`
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
		Text   string `xml:",chardata"`
		Value  uint8  `xml:"value,attr"`
		Size   uint64 `xml:"size,attr"`
		Sum16  string `xml:"sum16,attr"`
		Crc32  string `xml:"crc32,attr"`
		Md5    string `xml:"md5,attr"`
		Sha1   string `xml:"sha1,attr"`
		Sha256 string `xml:"sha256,attr"`
	} `xml:"miscRoms"`
	DefaultExpansion struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"defaultExpansion"`
}

type NES10XMLFields struct {
	Prgrom struct {
		Text   string `xml:",chardata"`
		Size   uint8  `xml:"size,attr"`
		Sum16  string `xml:"sum16,attr"`
		Crc32  string `xml:"crc32,attr"`
		Md5    string `xml:"md5,attr"`
		Sha1   string `xml:"sha1,attr"`
		Sha256 string `xml:"sha256,attr"`
	} `xml:"prgrom"`
	Chrrom struct {
		Text   string `xml:",chardata"`
		Size   uint8  `xml:"size,attr"`
		Sum16  string `xml:"sum16,attr"`
		Crc32  string `xml:"crc32,attr"`
		Md5    string `xml:"md5,attr"`
		Sha1   string `xml:"sha1,attr"`
		Sha256 string `xml:"sha256,attr"`
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
		Text   string `xml:",chardata"`
		Value  bool   `xml:"value,attr"`
		Size   uint16 `xml:"size,attr"`
		Sum16  string `xml:"sum16,attr"`
		Crc32  string `xml:"crc32,attr"`
		Md5    string `xml:"md5,attr"`
		Sha1   string `xml:"sha1,attr"`
		Sha256 string `xml:"sha256,attr"`
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
	PlayChoice10 struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"value,attr"`
	} `xml:"playChoice10"`
	Prgram struct {
		Text string `xml:",chardata"`
		Size uint8  `xml:"size,attr"`
	} `xml:"prgram"`
	TvSystem struct {
		Text  string `xml:",chardata"`
		Value bool   `xml:"size,attr"`
	} `xml:"tvSystem"`
}

type FDSXMLFields struct {
	Text           string              `xml:",chardata"`
	FDSArchiveDisk []*FDSDiskXMLFields `xml:"fdsDisk"`
}

type FDSDiskXMLFields struct {
	Text       string              `xml:",chardata"`
	DiskNumber uint8               `xml:"diskNumber,attr"`
	FDSSide    []*FDSSideXMLFields `xml:"fdsSide"`
}

type FDSSideXMLFields struct {
	Text             string `xml:",chardata"`
	Size             uint64 `xml:"size,attr"`
	Crc32            string `xml:"crc32,attr"`
	Md5              string `xml:"md5,attr"`
	Sha1             string `xml:"sha1,attr"`
	Sha256           string `xml:"sha256,attr"`
	ManufacturerCode struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"manufacturerCode"`
	FdsGameName struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"fdsGameName"`
	GameType struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"gameType"`
	RevisionNumber struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"revisionNumber"`
	SideNumber struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"sideNumber"`
	DiskNumber struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"diskNumber"`
	DiskType struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"diskType"`
	Byte18 struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte18"`
	BootFileId struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"bootFileId"`
	Byte1a struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte1a"`
	Byte1b struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte1b"`
	Byte1c struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte1c"`
	Byte1d struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte1d"`
	Byte1e struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte1e"`
	ManufacturingDate struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"manufacturingDate"`
	CountryCode struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"countryCode"`
	Byte23 struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte23"`
	Byte24 struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte24"`
	Byte25 struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte25"`
	Byte26 struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte26"`
	Byte27 struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte27"`
	Byte28 struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte28"`
	Byte29 struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte29"`
	Byte2a struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte2a"`
	Byte2b struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte2b"`
	RewriteDate struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"rewriteDate"`
	Byte2f struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte2f"`
	Byte30 struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte30"`
	DiskWriterSerialNumber struct {
		Text  string `xml:",chardata"`
		Value uint16 `xml:"value,attr"`
	} `xml:"diskWriterSerialNumber"`
	Byte33 struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte33"`
	RewriteCount struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"rewriteCount"`
	ActualDiskSide struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"actualDiskSide"`
	Byte36 struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"byte36"`
	Price struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"price"`
	DiskInfoCrc struct {
		Text  string `xml:",chardata"`
		Value uint16 `xml:"value,attr"`
	} `xml:"diskInfoCrc"`
	FileTableCrc struct {
		Text  string `xml:",chardata"`
		Value uint16 `xml:"value,attr"`
	} `xml:"fileTableCrc"`
	UnallocatedSpace struct {
		Text string `xml:",chardata"`
	} `xml:"unallocatedSpace"`
	UnallocatedSpaceOffset struct {
		Text  string `xml:",chardata"`
		Value uint16 `xml:"value,attr"`
	} `xml:"unallocatedSpaceOffset"`
	FDSFile []*FDSFileXMLFields `xml:"fdsFile"`
}

type FDSFileXMLFields struct {
	Text       string `xml:",chardata"`
	FileNumber struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"fileNumber"`
	FileIdentificationCode struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"fileIdentificationCode"`
	FileName struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"fileName"`
	FileAddress struct {
		Text  string `xml:",chardata"`
		Value uint16 `xml:"value,attr"`
	} `xml:"fileAddress"`
	FileSize struct {
		Text  string `xml:",chardata"`
		Value uint16 `xml:"value,attr"`
	} `xml:"fileSize"`
	FileType struct {
		Text  string `xml:",chardata"`
		Value uint8  `xml:"value,attr"`
	} `xml:"fileType"`
	FileMetadataCrc struct {
		Text  string `xml:",chardata"`
		Value uint16 `xml:"value,attr"`
	} `xml:"fileMetadataCrc"`
	FileData struct {
		Text        string `xml:",chardata"`
		Size        uint64 `xml:"size,attr"`
		FileDataCrc uint16 `xml:"fdsCrc,attr"`
		Crc32       string `xml:"crc32,attr"`
		Md5         string `xml:"md5,attr"`
		Sha1        string `xml:"sha1,attr"`
		Sha256      string `xml:"sha256,attr"`
	} `xml:"fileData"`
}

// Marshal maps of NESROM and FDSArchiveFile structs to XML
func MarshalXMLFromROMMap(nesRoms map[string]*NESTool.NESROM, fdsArchives map[string]*FDSTool.FDSArchiveFile, enableInes bool, preserveTrainer bool, enableOrganization bool) (string, error) {
	romXml := &NESXML{}

	for key := range nesRoms {
		if nesRoms[key].Header20 != nil {
			tempXmlRom := &NESXMLROM{}
			tempXmlHeader20 := &NES20XMLFields{}
			tempXmlRom.Header20 = tempXmlHeader20

			tempXmlRom.Name = nesRoms[key].Name
			tempXmlRom.Size = nesRoms[key].Size

			crc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(crc32Bytes, nesRoms[key].CRC32)
			tempXmlRom.Crc32 = strings.ToUpper(hex.EncodeToString(crc32Bytes))

			tempXmlRom.Md5 = strings.ToUpper(hex.EncodeToString(nesRoms[key].MD5[:]))
			tempXmlRom.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[key].SHA1[:]))
			tempXmlRom.Sha256 = strings.ToUpper(hex.EncodeToString(nesRoms[key].SHA256[:]))

			if enableOrganization {
				tempRelativePath := nesRoms[key].RelativePath
				if tempRelativePath[0] == os.PathSeparator {
					tempRelativePath = tempRelativePath[1:]
				}
				tempRelativePath = strings.Replace(tempRelativePath, string(os.PathSeparator), "/", -1)
				tempXmlRom.RelativePath = tempRelativePath
			}

			if preserveTrainer {
				tempXmlRom.Header20.Trainer.Value = nesRoms[key].Header20.Trainer

				if nesRoms[key].Header20.Trainer {
					trainerSum16Bytes := make([]byte, 2)
					binary.BigEndian.PutUint16(trainerSum16Bytes, nesRoms[key].Header20.TrainerSum16)
					tempXmlRom.Header20.Trainer.Sum16 = strings.ToUpper(hex.EncodeToString(trainerSum16Bytes))

					trainerCrc32Bytes := make([]byte, 4)
					binary.BigEndian.PutUint32(trainerCrc32Bytes, nesRoms[key].Header20.TrainerCRC32)
					tempXmlRom.Header20.Trainer.Crc32 = strings.ToUpper(hex.EncodeToString(trainerCrc32Bytes))

					tempXmlRom.Header20.Trainer.Md5 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.TrainerMD5[:]))
					tempXmlRom.Header20.Trainer.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.TrainerSHA1[:]))
					tempXmlRom.Header20.Trainer.Sha256 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.TrainerSHA256[:]))

					// Trainers will always be 512 bytes long if they exist, so enforce that.
					if len(nesRoms[key].TrainerData) == 512 {
						tempXmlRom.Header20.Trainer.Size = 512
						tempXmlRom.TrainerData.Text = strings.ToUpper(hex.EncodeToString(nesRoms[key].TrainerData))
					}
				}
			} else {
				tempXmlRom.Header20.Trainer.Value = false
			}

			if nesRoms[key].Header20.PRGROMSize > 0 {
				tempXmlRom.Header20.Prgrom.Size = nesRoms[key].Header20.PRGROMSize
			} else {
				tempXmlRom.Header20.Prgrom.SizeExponent = nesRoms[key].Header20.PRGROMSizeExponent
				tempXmlRom.Header20.Prgrom.SizeMultiplier = nesRoms[key].Header20.PRGROMSizeMultiplier
			}

			prgSum16Bytes := make([]byte, 2)
			binary.BigEndian.PutUint16(prgSum16Bytes, nesRoms[key].Header20.PRGROMSum16)
			tempXmlRom.Header20.Prgrom.Sum16 = strings.ToUpper(hex.EncodeToString(prgSum16Bytes))

			prgCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(prgCrc32Bytes, nesRoms[key].Header20.PRGROMCRC32)
			tempXmlRom.Header20.Prgrom.Crc32 = strings.ToUpper(hex.EncodeToString(prgCrc32Bytes))

			tempXmlRom.Header20.Prgrom.Md5 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.PRGROMMD5[:]))
			tempXmlRom.Header20.Prgrom.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.PRGROMSHA1[:]))
			tempXmlRom.Header20.Prgrom.Sha256 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.PRGROMSHA256[:]))

			if nesRoms[key].Header20.CHRROMSize > 0 {
				tempXmlRom.Header20.Chrrom.Size = nesRoms[key].Header20.CHRROMSize
			} else if nesRoms[key].Header20.CHRROMSizeExponent > 0 {
				tempXmlRom.Header20.Chrrom.SizeExponent = nesRoms[key].Header20.CHRROMSizeExponent
				tempXmlRom.Header20.Chrrom.SizeMultiplier = nesRoms[key].Header20.CHRROMSizeMultiplier
			}

			chrSum16Bytes := make([]byte, 2)
			binary.BigEndian.PutUint16(chrSum16Bytes, nesRoms[key].Header20.CHRROMSum16)

			tempXmlRom.Header20.Chrrom.Sum16 = strings.ToUpper(hex.EncodeToString(chrSum16Bytes))

			chrCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(chrCrc32Bytes, nesRoms[key].Header20.CHRROMCRC32)
			tempXmlRom.Header20.Chrrom.Crc32 = strings.ToUpper(hex.EncodeToString(chrCrc32Bytes))

			tempXmlRom.Header20.Chrrom.Md5 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.CHRROMMD5[:]))
			tempXmlRom.Header20.Chrrom.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.CHRROMSHA1[:]))
			tempXmlRom.Header20.Chrrom.Sha256 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.CHRROMSHA256[:]))

			tempXmlRom.Header20.Prgram.Size = nesRoms[key].Header20.PRGRAMSize
			tempXmlRom.Header20.Prgnvram.Size = nesRoms[key].Header20.PRGNVRAMSize
			tempXmlRom.Header20.Chrram.Size = nesRoms[key].Header20.CHRRAMSize
			tempXmlRom.Header20.Chrnvram.Size = nesRoms[key].Header20.CHRNVRAMSize
			tempXmlRom.Header20.MirroringType.Value = nesRoms[key].Header20.MirroringType
			tempXmlRom.Header20.Battery.Value = nesRoms[key].Header20.Battery
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

			if nesRoms[key].Header20.MiscROMs > 0 {
				tempXmlRom.Header20.MiscRoms.Size = nesRoms[key].Header20.MiscROMCalculatedSize

				miscRomSum16Bytes := make([]byte, 2)
				binary.BigEndian.PutUint16(miscRomSum16Bytes, nesRoms[key].Header20.MiscROMSum16)
				tempXmlRom.Header20.MiscRoms.Sum16 = strings.ToUpper(hex.EncodeToString(miscRomSum16Bytes))

				miscRomCrc32Bytes := make([]byte, 4)
				binary.BigEndian.PutUint32(miscRomCrc32Bytes, nesRoms[key].Header20.MiscROMCRC32)
				tempXmlRom.Header20.MiscRoms.Crc32 = strings.ToUpper(hex.EncodeToString(miscRomCrc32Bytes))

				tempXmlRom.Header20.MiscRoms.Md5 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.MiscROMMD5[:]))
				tempXmlRom.Header20.MiscRoms.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.MiscROMSHA1[:]))
				tempXmlRom.Header20.MiscRoms.Sha256 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header20.MiscROMSHA256[:]))
			}

			romXml.XMLROMs = append(romXml.XMLROMs, tempXmlRom)
		} else if enableInes && nesRoms[key].Header10 != nil {
			tempXmlRom := &NESXMLROM{}
			tempXmlHeader10 := &NES10XMLFields{}
			tempXmlRom.Header10 = tempXmlHeader10

			tempXmlRom.Name = nesRoms[key].Name
			tempXmlRom.Size = nesRoms[key].Size

			crc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(crc32Bytes, nesRoms[key].CRC32)
			tempXmlRom.Crc32 = strings.ToUpper(hex.EncodeToString(crc32Bytes))

			tempXmlRom.Md5 = strings.ToUpper(hex.EncodeToString(nesRoms[key].MD5[:]))
			tempXmlRom.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[key].SHA1[:]))
			tempXmlRom.Sha256 = strings.ToUpper(hex.EncodeToString(nesRoms[key].SHA256[:]))

			if enableOrganization {
				tempRelativePath := nesRoms[key].RelativePath
				if tempRelativePath[0] == os.PathSeparator {
					tempRelativePath = tempRelativePath[1:]
				}
				tempRelativePath = strings.Replace(tempRelativePath, string(os.PathSeparator), "/", -1)
				tempXmlRom.RelativePath = tempRelativePath
			}

			if preserveTrainer {
				tempXmlRom.Header10.Trainer.Value = nesRoms[key].Header10.Trainer

				if nesRoms[key].Header10.Trainer {
					trainerSum16Bytes := make([]byte, 2)
					binary.BigEndian.PutUint16(trainerSum16Bytes, nesRoms[key].Header10.TrainerSum16)
					tempXmlRom.Header10.Trainer.Sum16 = strings.ToUpper(hex.EncodeToString(trainerSum16Bytes))

					trainerCrc32Bytes := make([]byte, 4)
					binary.BigEndian.PutUint32(trainerCrc32Bytes, nesRoms[key].Header10.TrainerCRC32)
					tempXmlRom.Header10.Trainer.Crc32 = strings.ToUpper(hex.EncodeToString(trainerCrc32Bytes))

					tempXmlRom.Header10.Trainer.Md5 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header10.TrainerMD5[:]))
					tempXmlRom.Header10.Trainer.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header10.TrainerSHA1[:]))
					tempXmlRom.Header10.Trainer.Sha256 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header10.TrainerSHA256[:]))

					// Trainers will always be 512 bytes long if they exist, so enforce that.
					if len(nesRoms[key].TrainerData) == 512 {
						tempXmlRom.Header10.Trainer.Size = 512
						tempXmlRom.TrainerData.Text = strings.ToUpper(hex.EncodeToString(nesRoms[key].TrainerData))
					}
				}
			} else {
				tempXmlRom.Header10.Trainer.Value = nesRoms[key].Header10.Trainer
			}

			tempXmlRom.Header10.Prgrom.Size = nesRoms[key].Header10.PRGROMSize

			prgSum16Bytes := make([]byte, 2)
			binary.BigEndian.PutUint16(prgSum16Bytes, nesRoms[key].Header10.PRGROMSum16)
			tempXmlRom.Header10.Prgrom.Sum16 = strings.ToUpper(hex.EncodeToString(prgSum16Bytes))

			prgCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(prgCrc32Bytes, nesRoms[key].Header10.PRGROMCRC32)
			tempXmlRom.Header10.Prgrom.Crc32 = strings.ToUpper(hex.EncodeToString(prgCrc32Bytes))

			tempXmlRom.Header10.Prgrom.Md5 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header10.PRGROMMD5[:]))
			tempXmlRom.Header10.Prgrom.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header10.PRGROMSHA1[:]))
			tempXmlRom.Header10.Prgrom.Sha256 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header10.PRGROMSHA256[:]))

			tempXmlRom.Header10.Chrrom.Size = nesRoms[key].Header10.CHRROMSize

			chrSum16Bytes := make([]byte, 2)
			binary.BigEndian.PutUint16(chrSum16Bytes, nesRoms[key].Header10.CHRROMSum16)
			tempXmlRom.Header10.Chrrom.Sum16 = strings.ToUpper(hex.EncodeToString(chrSum16Bytes))

			chrCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(chrCrc32Bytes, nesRoms[key].Header10.CHRROMCRC32)
			tempXmlRom.Header10.Chrrom.Crc32 = strings.ToUpper(hex.EncodeToString(chrCrc32Bytes))

			tempXmlRom.Header10.Chrrom.Md5 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header10.CHRROMMD5[:]))
			tempXmlRom.Header10.Chrrom.Sha1 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header10.CHRROMSHA1[:]))
			tempXmlRom.Header10.Chrrom.Sha256 = strings.ToUpper(hex.EncodeToString(nesRoms[key].Header10.CHRROMSHA256[:]))

			tempXmlRom.Header10.MirroringType.Value = nesRoms[key].Header10.MirroringType
			tempXmlRom.Header10.Battery.Value = nesRoms[key].Header10.Battery
			tempXmlRom.Header10.FourScreen.Value = nesRoms[key].Header10.FourScreen
			tempXmlRom.Header10.Mapper.Value = nesRoms[key].Header10.Mapper
			tempXmlRom.Header10.VsUnisystem.Value = nesRoms[key].Header10.VsUnisystem
			tempXmlRom.Header10.PlayChoice10.Value = nesRoms[key].Header10.PlayChoice10
			tempXmlRom.Header10.Prgram.Size = nesRoms[key].Header10.PRGRAMSize
			tempXmlRom.Header10.TvSystem.Value = nesRoms[key].Header10.TVSystem

			romXml.XMLROMs = append(romXml.XMLROMs, tempXmlRom)
		}
	}

	for key := range fdsArchives {
		tempXmlRom := &NESXMLROM{}
		tempFdsArchive := &FDSXMLFields{}

		tempXmlRom.Name = fdsArchives[key].Name
		tempXmlRom.Size = fdsArchives[key].Size

		if enableOrganization {
			tempRelativePath := fdsArchives[key].RelativePath
			if tempRelativePath[0] == os.PathSeparator {
				tempRelativePath = tempRelativePath[1:]
			}
			tempRelativePath = strings.Replace(tempRelativePath, string(os.PathSeparator), "/", -1)
			tempXmlRom.RelativePath = tempRelativePath
		}

		crc32Bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(crc32Bytes, fdsArchives[key].CRC32)
		tempXmlRom.Crc32 = strings.ToUpper(hex.EncodeToString(crc32Bytes))
		tempXmlRom.Md5 = strings.ToUpper(hex.EncodeToString(fdsArchives[key].MD5[:]))
		tempXmlRom.Sha1 = strings.ToUpper(hex.EncodeToString(fdsArchives[key].SHA1[:]))
		tempXmlRom.Sha256 = strings.ToUpper(hex.EncodeToString(fdsArchives[key].SHA256[:]))

		for diskKey := range fdsArchives[key].ArchiveDisks {
			tempDisk := &FDSDiskXMLFields{}
			tempDisk.DiskNumber = fdsArchives[key].ArchiveDisks[diskKey].DiskNumber

			for sideKey := range fdsArchives[key].ArchiveDisks[diskKey].DiskSides {
				tempSide := &FDSSideXMLFields{}
				tempSide.Size = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Size

				sideCrc32Bytes := make([]byte, 4)
				binary.BigEndian.PutUint32(sideCrc32Bytes, fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].CRC32)
				tempSide.Crc32 = strings.ToUpper(hex.EncodeToString(sideCrc32Bytes))
				tempSide.Md5 = strings.ToUpper(hex.EncodeToString(fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].MD5[:]))
				tempSide.Sha1 = strings.ToUpper(hex.EncodeToString(fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SHA1[:]))
				tempSide.Sha256 = strings.ToUpper(hex.EncodeToString(fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SHA256[:]))

				tempSide.ManufacturerCode.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].ManufacturerCode
				tempSide.FdsGameName.Value = strings.ToUpper(hex.EncodeToString([]byte(fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].FDSGameName)))
				tempSide.GameType.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].GameType
				tempSide.RevisionNumber.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].RevisionNumber
				tempSide.SideNumber.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideNumber
				tempSide.DiskNumber.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].DiskNumber
				tempSide.DiskType.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].DiskType
				tempSide.Byte18.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte18
				tempSide.BootFileId.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].BootFileID
				tempSide.Byte1a.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte1A
				tempSide.Byte1b.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte1B
				tempSide.Byte1c.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte1C
				tempSide.Byte1d.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte1D
				tempSide.Byte1e.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte1E
				tempSide.ManufacturingDate.Value = strings.ToUpper(hex.EncodeToString(fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].ManufacturingDate))
				tempSide.CountryCode.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].CountryCode
				tempSide.Byte23.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte23
				tempSide.Byte24.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte24
				tempSide.Byte25.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte25
				tempSide.Byte26.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte26
				tempSide.Byte27.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte27
				tempSide.Byte28.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte28
				tempSide.Byte29.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte29
				tempSide.Byte2a.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte2A
				tempSide.Byte2b.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte2B
				tempSide.RewriteDate.Value = strings.ToUpper(hex.EncodeToString(fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].RewriteDate))
				tempSide.Byte2f.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte2F
				tempSide.Byte30.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte30
				tempSide.DiskWriterSerialNumber.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].DiskWriterSerialNumber
				tempSide.Byte33.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte33
				tempSide.RewriteCount.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].RewriteCount
				tempSide.ActualDiskSide.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].ActualDiskSide
				tempSide.Byte36.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Byte36
				tempSide.Price.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].Price
				tempSide.DiskInfoCrc.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].DiskInfoCRC
				tempSide.FileTableCrc.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].FileTableCRC
				tempSide.UnallocatedSpaceOffset.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].UnallocatedSpaceOffset
				tempSide.UnallocatedSpace.Text = strings.ToUpper(hex.EncodeToString(fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].UnallocatedSpace))

				for fileKey := range fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles {
					tempFile := &FDSFileXMLFields{}

					tempFile.FileNumber.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileNumber
					tempFile.FileIdentificationCode.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileIdentificationCode
					tempFile.FileName.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileName
					tempFile.FileAddress.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileAddress
					tempFile.FileSize.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileSize
					tempFile.FileType.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileType
					tempFile.FileMetadataCrc.Value = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileMetadataCRC
					tempFile.FileData.Size = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileData.Size

					fileCrc32Bytes := make([]byte, 4)
					binary.BigEndian.PutUint32(fileCrc32Bytes, fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileData.CRC32)
					tempFile.FileData.Crc32 = strings.ToUpper(hex.EncodeToString(fileCrc32Bytes))
					tempFile.FileData.Md5 = strings.ToUpper(hex.EncodeToString(fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileData.MD5[:]))
					tempFile.FileData.Sha1 = strings.ToUpper(hex.EncodeToString(fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileData.SHA1[:]))
					tempFile.FileData.Sha256 = strings.ToUpper(hex.EncodeToString(fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileData.SHA256[:]))
					tempFile.FileData.FileDataCrc = fdsArchives[key].ArchiveDisks[diskKey].DiskSides[sideKey].SideFiles[fileKey].FileData.FileDataCRC

					tempSide.FDSFile = append(tempSide.FDSFile, tempFile)
				}

				tempDisk.FDSSide = append(tempDisk.FDSSide, tempSide)
			}

			tempFdsArchive.FDSArchiveDisk = append(tempFdsArchive.FDSArchiveDisk, tempDisk)
		}

		tempXmlRom.FDSArchive = tempFdsArchive
		romXml.XMLROMs = append(romXml.XMLROMs, tempXmlRom)
	}

	xmlBytes, err := xml.MarshalIndent(romXml, "", "  ")
	if err != nil {
		return "", err
	}

	return string(xmlBytes), nil
}

// Unmarshal an XML file to maps of NESROM and FDSArchiveFile structs
func UnmarshalXMLToROMMap(xmlPayload string, enableInes bool, preserveTrainer bool, enableOrganization bool) (map[string]*NESTool.NESROM, map[string]*FDSTool.FDSArchiveFile, error) {
	xmlStruct := &NESXML{}
	err := xml.Unmarshal([]byte(xmlPayload), xmlStruct)
	if err != nil {
		return nil, nil, err
	}

	romMap := make(map[string]*NESTool.NESROM)
	archiveMap := make(map[string]*FDSTool.FDSArchiveFile)

	for index := range xmlStruct.XMLROMs {
		tempRom := &NESTool.NESROM{}

		crc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Crc32))
		if err == nil {
			tempRom.CRC32 = binary.BigEndian.Uint32(crc32Bytes)
		}

		md5Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Md5))
		if err == nil {
			copy(tempRom.MD5[:], md5Bytes)
		}

		sha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Sha1))
		if err == nil {
			copy(tempRom.SHA1[:], sha1Bytes)
		}

		sha256Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Sha256))
		if err == nil {
			copy(tempRom.SHA256[:], sha256Bytes)
		}

		tempRom.Name = xmlStruct.XMLROMs[index].Name
		tempRom.Size = xmlStruct.XMLROMs[index].Size

		if enableOrganization {
			tempRelativePath := xmlStruct.XMLROMs[index].RelativePath
			tempRelativePath = strings.Replace(tempRelativePath, "/", string(os.PathSeparator), -1)
			tempRelativePath = strings.Replace(tempRelativePath, "&amp;", "&", -1)
			tempRom.RelativePath = tempRelativePath
		}

		if xmlStruct.XMLROMs[index].Header20 != nil {
			tempRomHeader20 := &NESTool.NES20Header{}
			tempRom.Header20 = tempRomHeader20

			if xmlStruct.XMLROMs[index].Header20.Prgrom.Size > 0 {
				tempRom.Header20.PRGROMSize = xmlStruct.XMLROMs[index].Header20.Prgrom.Size
			} else {
				tempRom.Header20.PRGROMSizeExponent = xmlStruct.XMLROMs[index].Header20.Prgrom.SizeExponent
				tempRom.Header20.PRGROMSizeMultiplier = xmlStruct.XMLROMs[index].Header20.Prgrom.SizeMultiplier
			}

			prgRomSum16Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Prgrom.Sum16))
			if err == nil {
				tempRom.Header20.PRGROMSum16 = binary.BigEndian.Uint16(prgRomSum16Bytes)
			}

			prgRomCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Prgrom.Crc32))
			if err == nil {
				tempRom.Header20.PRGROMCRC32 = binary.BigEndian.Uint32(prgRomCrc32Bytes)
			}

			prgRomMd5Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Prgrom.Md5))
			if err == nil {
				copy(tempRom.Header20.PRGROMMD5[:], prgRomMd5Bytes)
			}

			prgRomSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Prgrom.Sha1))
			if err == nil {
				copy(tempRom.Header20.PRGROMSHA1[:], prgRomSha1Bytes)
			}

			prgRomSha256Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Prgrom.Sha256))
			if err == nil {
				copy(tempRom.Header20.PRGROMSHA256[:], prgRomSha256Bytes)
			}

			if xmlStruct.XMLROMs[index].Header20.Chrrom.Size > 0 {
				tempRom.Header20.CHRROMSize = xmlStruct.XMLROMs[index].Header20.Chrrom.Size
			} else if xmlStruct.XMLROMs[index].Header20.Chrrom.SizeExponent > 0 {
				tempRom.Header20.CHRROMSizeExponent = xmlStruct.XMLROMs[index].Header20.Chrrom.SizeExponent
				tempRom.Header20.CHRROMSizeMultiplier = xmlStruct.XMLROMs[index].Header20.Chrrom.SizeMultiplier
			}

			chrRomSum16Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Chrrom.Sum16))
			if err == nil {
				tempRom.Header20.CHRROMSum16 = binary.BigEndian.Uint16(chrRomSum16Bytes)
			}

			chrRomCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Chrrom.Crc32))
			if err == nil {
				tempRom.Header20.CHRROMCRC32 = binary.BigEndian.Uint32(chrRomCrc32Bytes)
			}

			chrRomMd5Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Chrrom.Md5))
			if err == nil {
				copy(tempRom.Header20.CHRROMMD5[:], chrRomMd5Bytes)
			}

			chrRomSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Chrrom.Sha1))
			if err == nil {
				copy(tempRom.Header20.CHRROMSHA1[:], chrRomSha1Bytes)
			}

			chrRomSha256Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Chrrom.Sha256))
			if err == nil {
				copy(tempRom.Header20.CHRROMSHA256[:], chrRomSha256Bytes)
			}

			tempRom.Header20.PRGRAMSize = xmlStruct.XMLROMs[index].Header20.Prgram.Size
			tempRom.Header20.PRGNVRAMSize = xmlStruct.XMLROMs[index].Header20.Prgnvram.Size
			tempRom.Header20.CHRRAMSize = xmlStruct.XMLROMs[index].Header20.Chrram.Size
			tempRom.Header20.CHRNVRAMSize = xmlStruct.XMLROMs[index].Header20.Chrnvram.Size
			tempRom.Header20.MirroringType = xmlStruct.XMLROMs[index].Header20.MirroringType.Value
			tempRom.Header20.Battery = xmlStruct.XMLROMs[index].Header20.Battery.Value
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

			err = NESTool.UpdateSizes(tempRom, NESTool.PRG_CANONICAL_SIZE_FACTORED, NESTool.CHR_CANONICAL_SIZE_FACTORED)
			if err != nil {
				return nil, nil, err
			}

			if preserveTrainer {
				tempRom.Header20.Trainer = xmlStruct.XMLROMs[index].Header20.Trainer.Value

				if xmlStruct.XMLROMs[index].Header20.Trainer.Value {
					trainerSum16Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Trainer.Sum16))
					if err == nil {
						tempRom.Header20.TrainerSum16 = binary.BigEndian.Uint16(trainerSum16Bytes)
					}

					trainerCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Trainer.Crc32))
					if err == nil {
						tempRom.Header20.TrainerCRC32 = binary.BigEndian.Uint32(trainerCrc32Bytes)
					}

					trainerMd5Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Trainer.Md5))
					if err == nil {
						copy(tempRom.Header20.TrainerMD5[:], trainerMd5Bytes)
					}

					trainerSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Trainer.Sha1))
					if err == nil {
						copy(tempRom.Header20.TrainerSHA1[:], trainerSha1Bytes)
					}

					trainerSha256Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.Trainer.Sha256))
					if err == nil {
						copy(tempRom.Header20.TrainerSHA256[:], trainerSha256Bytes)
					}

					if len(xmlStruct.XMLROMs[index].TrainerData.Text) == 2*512 {
						tempRom.Header20.TrainerCalculatedSize = 512

						trainerDataBytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].TrainerData.Text))
						if err == nil {
							tempRom.TrainerData = trainerDataBytes
						}
					}
				}
			} else {
				tempRom.Header20.Trainer = false
				tempRom.Header20.TrainerCalculatedSize = 0
			}

			if xmlStruct.XMLROMs[index].Header20.MiscRoms.Value > 0 {
				tempRom.Header20.MiscROMs = xmlStruct.XMLROMs[index].Header20.MiscRoms.Value
				tempRom.Header20.MiscROMCalculatedSize = xmlStruct.XMLROMs[index].Header20.MiscRoms.Size

				miscRomSum16Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.MiscRoms.Sum16))
				if err == nil {
					tempRom.Header20.MiscROMSum16 = binary.BigEndian.Uint16(miscRomSum16Bytes)
				}

				miscRomCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.MiscRoms.Crc32))
				if err == nil {
					tempRom.Header20.MiscROMCRC32 = binary.BigEndian.Uint32(miscRomCrc32Bytes)
				}

				miscRomMd5Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.MiscRoms.Md5))
				if err == nil {
					copy(tempRom.Header20.MiscROMMD5[:], miscRomMd5Bytes)
				}

				miscRomSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.MiscRoms.Sha1))
				if err == nil {
					copy(tempRom.Header20.MiscROMSHA1[:], miscRomSha1Bytes)
				}

				miscRomSha256Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header20.MiscRoms.Sha256))
				if err == nil {
					copy(tempRom.Header20.MiscROMSHA256[:], miscRomSha256Bytes)
				}
			} else {
				tempRom.Header20.MiscROMs = 0
				tempRom.Header20.MiscROMCalculatedSize = 0
			}

			romMap["SHA256:"+strings.ToUpper(hex.EncodeToString(tempRom.SHA256[:]))] = tempRom
		} else if enableInes && xmlStruct.XMLROMs[index].Header10 != nil {
			tempRomHeader10 := &NESTool.NES10Header{}
			tempRom.Header10 = tempRomHeader10

			tempRom.Header10.PRGROMSize = xmlStruct.XMLROMs[index].Header10.Prgrom.Size

			prgRomSum16Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Prgrom.Sum16))
			if err == nil {
				tempRom.Header10.PRGROMSum16 = binary.BigEndian.Uint16(prgRomSum16Bytes)
			}

			prgRomCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Prgrom.Crc32))
			if err == nil {
				tempRom.Header10.PRGROMCRC32 = binary.BigEndian.Uint32(prgRomCrc32Bytes)
			}

			prgRomMd5Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Prgrom.Md5))
			if err == nil {
				copy(tempRom.Header10.PRGROMMD5[:], prgRomMd5Bytes)
			}

			prgRomSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Prgrom.Sha1))
			if err == nil {
				copy(tempRom.Header10.PRGROMSHA1[:], prgRomSha1Bytes)
			}

			prgRomSha256Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Prgrom.Sha256))
			if err == nil {
				copy(tempRom.Header10.PRGROMSHA256[:], prgRomSha256Bytes)
			}

			tempRom.Header10.CHRROMSize = xmlStruct.XMLROMs[index].Header10.Chrrom.Size

			chrRomSum16Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Chrrom.Sum16))
			if err == nil {
				tempRom.Header10.CHRROMSum16 = binary.BigEndian.Uint16(chrRomSum16Bytes)
			}

			chrRomCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Chrrom.Crc32))
			if err == nil {
				tempRom.Header10.CHRROMCRC32 = binary.BigEndian.Uint32(chrRomCrc32Bytes)
			}

			chrRomMd5Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Chrrom.Md5))
			if err == nil {
				copy(tempRom.Header10.CHRROMMD5[:], chrRomMd5Bytes)
			}

			chrRomSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Chrrom.Sha1))
			if err == nil {
				copy(tempRom.Header10.CHRROMSHA1[:], chrRomSha1Bytes)
			}

			chrRomSha256Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Chrrom.Sha256))
			if err == nil {
				copy(tempRom.Header10.CHRROMSHA256[:], chrRomSha256Bytes)
			}

			tempRom.Header10.MirroringType = xmlStruct.XMLROMs[index].Header10.MirroringType.Value
			tempRom.Header10.Battery = xmlStruct.XMLROMs[index].Header10.Battery.Value
			tempRom.Header10.Trainer = xmlStruct.XMLROMs[index].Header10.Trainer.Value
			tempRom.Header10.FourScreen = xmlStruct.XMLROMs[index].Header10.FourScreen.Value
			tempRom.Header10.Mapper = xmlStruct.XMLROMs[index].Header10.Mapper.Value
			tempRom.Header10.VsUnisystem = xmlStruct.XMLROMs[index].Header10.VsUnisystem.Value
			tempRom.Header10.PlayChoice10 = xmlStruct.XMLROMs[index].Header10.PlayChoice10.Value
			tempRom.Header10.PRGRAMSize = xmlStruct.XMLROMs[index].Header10.Prgram.Size
			tempRom.Header10.TVSystem = xmlStruct.XMLROMs[index].Header10.TvSystem.Value

			err = NESTool.UpdateSizes(tempRom, NESTool.PRG_CANONICAL_SIZE_FACTORED, NESTool.CHR_CANONICAL_SIZE_FACTORED)
			if err != nil {
				return nil, nil, err
			}

			if preserveTrainer {
				tempRom.Header10.Trainer = xmlStruct.XMLROMs[index].Header10.Trainer.Value

				if xmlStruct.XMLROMs[index].Header10.Trainer.Value {
					trainerSum16Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Trainer.Sum16))
					if err == nil {
						tempRom.Header10.TrainerSum16 = binary.BigEndian.Uint16(trainerSum16Bytes)
					}

					trainerCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Trainer.Crc32))
					if err == nil {
						tempRom.Header10.TrainerCRC32 = binary.BigEndian.Uint32(trainerCrc32Bytes)
					}

					trainerMd5Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Trainer.Md5))
					if err == nil {
						copy(tempRom.Header10.TrainerMD5[:], trainerMd5Bytes)
					}

					trainerSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Trainer.Sha1))
					if err == nil {
						copy(tempRom.Header10.TrainerSHA1[:], trainerSha1Bytes)
					}

					trainerSha256Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].Header10.Trainer.Sha256))
					if err == nil {
						copy(tempRom.Header10.TrainerSHA256[:], trainerSha256Bytes)
					}

					if len(xmlStruct.XMLROMs[index].TrainerData.Text) == 2*512 {
						tempRom.Header10.TrainerCalculatedSize = 512

						trainerDataBytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].TrainerData.Text))
						if err == nil {
							tempRom.TrainerData = trainerDataBytes
						}
					}
				}
			} else {
				tempRom.Header10.Trainer = false
				tempRom.Header10.TrainerCalculatedSize = 0
			}

			romMap["SHA256:"+strings.ToUpper(hex.EncodeToString(tempRom.SHA256[:]))] = tempRom
		} else if xmlStruct.XMLROMs[index].FDSArchive != nil {
			tempArchive := &FDSTool.FDSArchiveFile{}
			tempArchive.CRC32 = binary.BigEndian.Uint32(crc32Bytes)
			copy(tempArchive.MD5[:], md5Bytes)
			copy(tempArchive.SHA1[:], sha1Bytes)
			copy(tempArchive.SHA256[:], sha256Bytes)

			tempArchive.Name = xmlStruct.XMLROMs[index].Name
			tempArchive.Size = xmlStruct.XMLROMs[index].Size

			if enableOrganization {
				tempRelativePath := xmlStruct.XMLROMs[index].RelativePath
				tempRelativePath = strings.Replace(tempRelativePath, "/", string(os.PathSeparator), -1)
				tempArchive.RelativePath = tempRelativePath
			}

			for diskKey := range xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk {
				tempDisk := &FDSTool.FDSDisk{}
				tempDisk.DiskNumber = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].DiskNumber

				for sideKey := range xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide {
					tempSide := &FDSTool.FDSSide{}

					tempSide.Size = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Size

					sideCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Crc32))
					if err == nil {
						tempSide.CRC32 = binary.BigEndian.Uint32(sideCrc32Bytes)
					}

					sideMd5Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Md5))
					if err == nil {
						copy(tempSide.MD5[:], sideMd5Bytes)
					}

					sideSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Sha1))
					if err == nil {
						copy(tempSide.SHA1[:], sideSha1Bytes)
					}

					sideSha256Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Sha256))
					if err == nil {
						copy(tempSide.SHA256[:], sideSha256Bytes)
					}

					tempSide.ManufacturerCode = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].ManufacturerCode.Value

					fdsGameNameString, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FdsGameName.Value))
					if err == nil {
						tempSide.FDSGameName = string(fdsGameNameString)
					}

					tempSide.GameType = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].GameType.Value
					tempSide.RevisionNumber = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].RevisionNumber.Value
					tempSide.SideNumber = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].SideNumber.Value
					tempSide.DiskNumber = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].DiskNumber.Value
					tempSide.DiskType = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].DiskType.Value
					tempSide.Byte18 = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte18.Value
					tempSide.BootFileID = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].BootFileId.Value
					tempSide.Byte1A = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte1a.Value
					tempSide.Byte1B = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte1b.Value
					tempSide.Byte1C = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte1c.Value
					tempSide.Byte1D = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte1d.Value
					tempSide.Byte1E = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte1e.Value

					sideManufacturingDateBytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].ManufacturingDate.Value))
					if err == nil {
						tempSide.ManufacturingDate = sideManufacturingDateBytes
					}

					tempSide.CountryCode = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].CountryCode.Value
					tempSide.Byte23 = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte23.Value
					tempSide.Byte24 = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte24.Value
					tempSide.Byte25 = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte25.Value
					tempSide.Byte26 = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte26.Value
					tempSide.Byte27 = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte27.Value
					tempSide.Byte28 = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte28.Value
					tempSide.Byte29 = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte29.Value
					tempSide.Byte2A = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte2a.Value
					tempSide.Byte2B = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte2b.Value

					sideRewriteDateBytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].RewriteDate.Value))
					if err == nil {
						tempSide.RewriteDate = sideRewriteDateBytes
					}

					tempSide.Byte2F = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte2f.Value
					tempSide.Byte30 = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte30.Value
					tempSide.DiskWriterSerialNumber = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].DiskWriterSerialNumber.Value
					tempSide.Byte33 = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte33.Value
					tempSide.RewriteCount = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].RewriteCount.Value
					tempSide.ActualDiskSide = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].ActualDiskSide.Value
					tempSide.Byte36 = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Byte36.Value
					tempSide.Price = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].Price.Value
					tempSide.DiskInfoCRC = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].DiskInfoCrc.Value
					tempSide.FileTableCRC = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FileTableCrc.Value
					tempSide.UnallocatedSpaceOffset = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].UnallocatedSpaceOffset.Value

					sideUnallocatedSpaceBytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].UnallocatedSpace.Text))
					if err == nil {
						tempSide.UnallocatedSpace = sideUnallocatedSpaceBytes
					}

					for fileKey := range xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile {
						tempFile := &FDSTool.FDSFile{}

						tempFile.FileNumber = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileNumber.Value
						tempFile.FileIdentificationCode = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileIdentificationCode.Value
						tempFile.FileName = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileName.Value
						tempFile.FileAddress = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileAddress.Value
						tempFile.FileSize = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileSize.Value
						tempFile.FileType = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileType.Value
						tempFile.FileMetadataCRC = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileMetadataCrc.Value

						tempFileData := &FDSTool.FDSFileData{}

						tempFileData.Size = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileData.Size
						fileDataCrc32Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileData.Crc32))
						if err == nil {
							tempFileData.CRC32 = binary.BigEndian.Uint32(fileDataCrc32Bytes)
						}

						fileDataMd5Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileData.Md5))
						if err == nil {
							copy(tempFileData.MD5[:], fileDataMd5Bytes)
						}

						fileDataSha1Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileData.Sha1))
						if err == nil {
							copy(tempFileData.SHA1[:], fileDataSha1Bytes)
						}

						fileDataSha256Bytes, err := hex.DecodeString(strings.ToLower(xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileData.Sha256))
						if err == nil {
							copy(tempFileData.SHA256[:], fileDataSha256Bytes)
						}

						tempFileData.FileDataCRC = xmlStruct.XMLROMs[index].FDSArchive.FDSArchiveDisk[diskKey].FDSSide[sideKey].FDSFile[fileKey].FileData.FileDataCrc

						tempFile.FileData = tempFileData
						tempSide.SideFiles = append(tempSide.SideFiles, tempFile)
					}

					tempDisk.DiskSides = append(tempDisk.DiskSides, tempSide)
				}

				tempArchive.ArchiveDisks = append(tempArchive.ArchiveDisks, tempDisk)
			}

			archiveMap["SHA256:"+strings.ToUpper(hex.EncodeToString(tempArchive.SHA256[:]))] = tempArchive
		}
	}

	return romMap, archiveMap, nil
}
