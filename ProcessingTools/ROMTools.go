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
package ProcessingTools

import (
	"NES20Tool/FDSTool"
	"NES20Tool/NES20Tool"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"
)

var (
	HASH_TYPE_SUM16  uint64 = 1
	HASH_TYPE_CRC32  uint64 = 2
	HASH_TYPE_MD5    uint64 = 4
	HASH_TYPE_SHA1   uint64 = 8
	HASH_TYPE_SHA256 uint64 = 16
)

func MatchNESROM(testRom *NES20Tool.NESROM, templateRomMap map[string]*NES20Tool.NESROM, hashTypeTests uint64, enableInes bool) (*NES20Tool.NESROM, error) {
	if hashTypeTests&HASH_TYPE_SHA256 > 0 {
		if templateRomMap["SHA256:"+strings.ToUpper(hex.EncodeToString(testRom.SHA256[:]))] != nil {
			return templateRomMap["SHA256:"+strings.ToUpper(hex.EncodeToString(testRom.SHA256[:]))], nil
		}

		for index := range templateRomMap {
			if templateRomMap[index].Header20.PRGROMSHA256 == testRom.Header20.PRGROMSHA256 && templateRomMap[index].Header20.CHRROMSHA256 == testRom.Header20.CHRROMSHA256 {
				return templateRomMap[index], nil
			}

			if templateRomMap[index].Header20.PRGROMSHA256 == testRom.Header10.PRGROMSHA256 && templateRomMap[index].Header20.CHRROMSHA256 == testRom.Header10.CHRROMSHA256 {
				return templateRomMap[index], nil
			}

			if enableInes {
				if templateRomMap[index].Header10.PRGROMSHA256 == testRom.Header10.PRGROMSHA256 && templateRomMap[index].Header10.CHRROMSHA256 == testRom.Header10.CHRROMSHA256 {
					return templateRomMap[index], nil
				}

				if templateRomMap[index].Header10.PRGROMSHA256 == testRom.Header20.PRGROMSHA256 && templateRomMap[index].Header10.CHRROMSHA256 == testRom.Header20.CHRROMSHA256 {
					return templateRomMap[index], nil
				}
			}
		}
	}

	if hashTypeTests&HASH_TYPE_SHA1 > 0 {
		if templateRomMap["SHA1:"+strings.ToUpper(hex.EncodeToString(testRom.SHA1[:]))] != nil {
			return templateRomMap["SHA1:"+strings.ToUpper(hex.EncodeToString(testRom.SHA1[:]))], nil
		}

		for index := range templateRomMap {
			if templateRomMap[index].Header20.PRGROMSHA1 == testRom.Header20.PRGROMSHA1 && templateRomMap[index].Header20.CHRROMSHA1 == testRom.Header20.CHRROMSHA1 {
				return templateRomMap[index], nil
			}

			if templateRomMap[index].Header20.PRGROMSHA1 == testRom.Header10.PRGROMSHA1 && templateRomMap[index].Header20.CHRROMSHA1 == testRom.Header10.CHRROMSHA1 {
				return templateRomMap[index], nil
			}

			if enableInes {
				if templateRomMap[index].Header10.PRGROMSHA1 == testRom.Header10.PRGROMSHA1 && templateRomMap[index].Header10.CHRROMSHA1 == testRom.Header10.CHRROMSHA1 {
					return templateRomMap[index], nil
				}

				if templateRomMap[index].Header10.PRGROMSHA1 == testRom.Header20.PRGROMSHA1 && templateRomMap[index].Header10.CHRROMSHA1 == testRom.Header20.CHRROMSHA1 {
					return templateRomMap[index], nil
				}
			}
		}
	}

	if hashTypeTests&HASH_TYPE_MD5 > 0 {
		if templateRomMap["MD5:"+strings.ToUpper(hex.EncodeToString(testRom.MD5[:]))] != nil {
			return templateRomMap["MD5:"+strings.ToUpper(hex.EncodeToString(testRom.MD5[:]))], nil
		}

		for index := range templateRomMap {
			if templateRomMap[index].Header20.PRGROMMD5 == testRom.Header20.PRGROMMD5 && templateRomMap[index].Header20.CHRROMMD5 == testRom.Header20.CHRROMMD5 {
				return templateRomMap[index], nil
			}

			if templateRomMap[index].Header20.PRGROMMD5 == testRom.Header10.PRGROMMD5 && templateRomMap[index].Header20.CHRROMMD5 == testRom.Header10.CHRROMMD5 {
				return templateRomMap[index], nil
			}

			if enableInes {
				if templateRomMap[index].Header10.PRGROMMD5 == testRom.Header10.PRGROMMD5 && templateRomMap[index].Header10.CHRROMMD5 == testRom.Header10.CHRROMMD5 {
					return templateRomMap[index], nil
				}

				if templateRomMap[index].Header10.PRGROMMD5 == testRom.Header20.PRGROMMD5 && templateRomMap[index].Header10.CHRROMMD5 == testRom.Header20.CHRROMMD5 {
					return templateRomMap[index], nil
				}
			}
		}
	}

	if hashTypeTests&HASH_TYPE_CRC32 > 0 {
		testRomCrc32Bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(testRomCrc32Bytes, testRom.CRC32)
		if templateRomMap["CRC32:"+strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes))] != nil {
			return templateRomMap["CRC32:"+strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes))], nil
		}

		for index := range templateRomMap {
			if templateRomMap[index].Header20.PRGROMCRC32 == testRom.Header20.PRGROMCRC32 && templateRomMap[index].Header20.CHRROMCRC32 == testRom.Header20.CHRROMCRC32 {
				return templateRomMap[index], nil
			}

			if templateRomMap[index].Header20.PRGROMCRC32 == testRom.Header10.PRGROMCRC32 && templateRomMap[index].Header20.CHRROMCRC32 == testRom.Header10.CHRROMCRC32 {
				return templateRomMap[index], nil
			}

			if enableInes {
				if templateRomMap[index].Header10.PRGROMCRC32 == testRom.Header10.PRGROMCRC32 && templateRomMap[index].Header10.CHRROMCRC32 == testRom.Header10.CHRROMCRC32 {
					return templateRomMap[index], nil
				}

				if templateRomMap[index].Header10.PRGROMCRC32 == testRom.Header20.PRGROMCRC32 && templateRomMap[index].Header10.CHRROMCRC32 == testRom.Header20.CHRROMCRC32 {
					return templateRomMap[index], nil
				}
			}
		}
	}

	if hashTypeTests&HASH_TYPE_SUM16 > 0 {
		for index := range templateRomMap {
			if templateRomMap[index].Header20.PRGROMSum16 == testRom.Header20.PRGROMSum16 && templateRomMap[index].Header20.CHRROMSum16 == testRom.Header20.CHRROMSum16 {
				return templateRomMap[index], nil
			}

			if templateRomMap[index].Header20.PRGROMSum16 == testRom.Header10.PRGROMSum16 && templateRomMap[index].Header20.CHRROMSum16 == testRom.Header10.CHRROMSum16 {
				return templateRomMap[index], nil
			}

			if enableInes {
				if templateRomMap[index].Header10.PRGROMSum16 == testRom.Header10.PRGROMSum16 && templateRomMap[index].Header10.CHRROMSum16 == testRom.Header10.CHRROMSum16 {
					return templateRomMap[index], nil
				}

				if templateRomMap[index].Header10.PRGROMSum16 == testRom.Header20.PRGROMSum16 && templateRomMap[index].Header10.CHRROMSum16 == testRom.Header20.CHRROMSum16 {
					return templateRomMap[index], nil
				}
			}
		}
	}

	testRomCrc32Bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(testRomCrc32Bytes, testRom.CRC32)
	return nil, errors.New("No match found for NES ROM: " + testRom.Name + "\nCRC32: " + strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes)) + "\nSHA1: " + strings.ToUpper(hex.EncodeToString(testRom.SHA1[:])) + "\nSHA256: " + strings.ToUpper(hex.EncodeToString(testRom.SHA256[:])))
}

func UpdateNESROM(targetRom *NES20Tool.NESROM, templateRom *NES20Tool.NESROM, truncateRom bool, enableInes bool) error {
	if targetRom == nil || templateRom == nil {
		return errors.New("Missing target or template NES ROM for update.")
	}

	if templateRom.Header20 != nil {
		targetRom.Header20 = templateRom.Header20
	} else if enableInes && templateRom.Header10 != nil {
		targetRom.Header10 = templateRom.Header10
	} else {
		return errors.New("Unable to update ROM.")
	}

	targetRom.Name = templateRom.Name
	targetRom.SHA256 = templateRom.SHA256
	targetRom.SHA1 = templateRom.SHA1
	targetRom.MD5 = templateRom.MD5
	targetRom.CRC32 = templateRom.CRC32
	targetRom.RelativePath = templateRom.RelativePath

	if truncateRom {
		NES20Tool.TruncateROMDataAndSections(targetRom)
	}

	return nil
}

func ProcessNESROMs(testRomList []*NES20Tool.NESROM, templateRomMap map[string]*NES20Tool.NESROM, hashTypeTests uint64, truncateRoms bool, enableInes bool) []*NES20Tool.NESROM {
	returnRomList := make([]*NES20Tool.NESROM, 0)

	for index := range testRomList {
		tempRom, matchErr := MatchNESROM(testRomList[index], templateRomMap, hashTypeTests, enableInes)
		if matchErr == nil {
			updateErr := UpdateNESROM(testRomList[index], tempRom, truncateRoms, enableInes)
			if updateErr == nil {
				returnRomList = append(returnRomList, testRomList[index])
			}
		}
	}

	return returnRomList
}

func MatchFDSROM(testRom *FDSTool.FDSArchiveFile, romList map[string]*FDSTool.FDSArchiveFile, hashTypeTests uint64) (*FDSTool.FDSArchiveFile, error) {
	if hashTypeTests&HASH_TYPE_SHA256 > 0 {
		if romList["SHA256:"+strings.ToUpper(hex.EncodeToString(testRom.SHA256[:]))] != nil {
			return romList["SHA256:"+strings.ToUpper(hex.EncodeToString(testRom.SHA256[:]))], nil
		}
	}

	if hashTypeTests&HASH_TYPE_SHA1 > 0 {
		if romList["SHA1:"+strings.ToUpper(hex.EncodeToString(testRom.SHA1[:]))] != nil {
			return romList["SHA1:"+strings.ToUpper(hex.EncodeToString(testRom.SHA1[:]))], nil
		}
	}

	if hashTypeTests&HASH_TYPE_MD5 > 0 {
		if romList["MD5:"+strings.ToUpper(hex.EncodeToString(testRom.MD5[:]))] != nil {
			return romList["MD5:"+strings.ToUpper(hex.EncodeToString(testRom.MD5[:]))], nil
		}
	}

	if hashTypeTests&HASH_TYPE_CRC32 > 0 {
		testRomCrc32Bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(testRomCrc32Bytes, testRom.CRC32)
		if romList["CRC32:"+strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes))] != nil {
			return romList["CRC32:"+strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes))], nil
		}
	}

	testRomCrc32Bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(testRomCrc32Bytes, testRom.CRC32)
	return nil, errors.New("No match found for FDS ROM: " + testRom.Name + "\nCRC32: " + strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes)) + "\nSHA1: " + strings.ToUpper(hex.EncodeToString(testRom.SHA1[:])) + "\nSHA256: " + strings.ToUpper(hex.EncodeToString(testRom.SHA256[:])))
}

func UpdateFDSROM(targetRom *FDSTool.FDSArchiveFile, templateRom *FDSTool.FDSArchiveFile) error {
	if targetRom == nil || templateRom == nil {
		return errors.New("Missing target or template FDS ROM for update.")
	}

	targetRom.Name = templateRom.Name
	targetRom.RelativePath = templateRom.RelativePath

	return nil
}

func ProcessFDSROMs(testRomList []*FDSTool.FDSArchiveFile, templateRomMap map[string]*FDSTool.FDSArchiveFile, hashTypeTests uint64) []*FDSTool.FDSArchiveFile {
	returnRomList := make([]*FDSTool.FDSArchiveFile, 0)

	for index := range testRomList {
		tempRom, matchErr := MatchFDSROM(testRomList[index], templateRomMap, hashTypeTests)
		if matchErr == nil {
			updateErr := UpdateFDSROM(testRomList[index], tempRom)
			if updateErr == nil {
				returnRomList = append(returnRomList, testRomList[index])
			}
		}
	}

	return returnRomList
}
