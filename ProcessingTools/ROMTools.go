/*
   Copyright 2021, Christopher Gelatt

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

package ProcessingTools

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/Kreeblah/NES20Tool/FDSTool"
	"github.com/Kreeblah/NES20Tool/NESTool"
	"strings"
)

var (
	HASH_TYPE_SUM16  uint64 = 1
	HASH_TYPE_CRC32  uint64 = 2
	HASH_TYPE_MD5    uint64 = 4
	HASH_TYPE_SHA1   uint64 = 8
	HASH_TYPE_SHA256 uint64 = 16
)

// Match a given ROM to a template ROM based on the hashing algorithm(s) specified.
// Multiple hash types can be used via a bitwise OR operation.  Higher-complexity
// algorithms are preferred over lower-complexity algorithms.
func MatchNESROM(testRom *NESTool.NESROM, templateRomMap map[string]*NESTool.NESROM, hashTypeTests uint64, enableInes bool) (*NESTool.NESROM, error) {
	if hashTypeTests&HASH_TYPE_SHA256 > 0 {
		if templateRomMap["SHA256:"+strings.ToUpper(hex.EncodeToString(testRom.SHA256[:]))] != nil {
			return templateRomMap["SHA256:"+strings.ToUpper(hex.EncodeToString(testRom.SHA256[:]))], nil
		}

		for index := range templateRomMap {
			if templateRomMap[index].Header20 != nil {
				if testRom.Header20 != nil {
					if templateRomMap[index].Header20.PRGROMSHA256 == testRom.Header20.PRGROMSHA256 && ((templateRomMap[index].Header20.CHRROMSize == 0 && testRom.Header20.CHRROMSize == 0) || templateRomMap[index].Header20.CHRROMSHA256 == testRom.Header20.CHRROMSHA256) {
						return templateRomMap[index], nil
					}
				}

				if testRom.Header10 != nil {
					if templateRomMap[index].Header20.PRGROMSHA256 == testRom.Header10.PRGROMSHA256 && ((templateRomMap[index].Header20.CHRROMSize == 0 && testRom.Header10.CHRROMSize == 0) || templateRomMap[index].Header20.CHRROMSHA256 == testRom.Header10.CHRROMSHA256) {
						return templateRomMap[index], nil
					}
				}
			} else if enableInes && templateRomMap[index].Header10 != nil {
				if testRom.Header20 != nil {
					if templateRomMap[index].Header10.PRGROMSHA256 == testRom.Header20.PRGROMSHA256 && ((templateRomMap[index].Header10.CHRROMSize == 0 && testRom.Header20.CHRROMSize == 0) || templateRomMap[index].Header10.CHRROMSHA256 == testRom.Header20.CHRROMSHA256) {
						return templateRomMap[index], nil
					}
				}

				if testRom.Header10 != nil {
					if templateRomMap[index].Header10.PRGROMSHA256 == testRom.Header10.PRGROMSHA256 && ((templateRomMap[index].Header10.CHRROMSize == 0 && testRom.Header10.CHRROMSize == 0) || templateRomMap[index].Header10.CHRROMSHA256 == testRom.Header10.CHRROMSHA256) {
						return templateRomMap[index], nil
					}
				}
			}
		}
	}

	if hashTypeTests&HASH_TYPE_SHA1 > 0 {
		if templateRomMap["SHA1:"+strings.ToUpper(hex.EncodeToString(testRom.SHA1[:]))] != nil {
			return templateRomMap["SHA1:"+strings.ToUpper(hex.EncodeToString(testRom.SHA1[:]))], nil
		}

		for index := range templateRomMap {
			if templateRomMap[index].Header20 != nil {
				if testRom.Header20 != nil {
					if templateRomMap[index].Header20.PRGROMSHA1 == testRom.Header20.PRGROMSHA1 && ((templateRomMap[index].Header20.CHRROMSize == 0 && testRom.Header20.CHRROMSize == 0) || templateRomMap[index].Header20.CHRROMSHA1 == testRom.Header20.CHRROMSHA1) {
						return templateRomMap[index], nil
					}
				}

				if testRom.Header10 != nil {
					if templateRomMap[index].Header20.PRGROMSHA1 == testRom.Header10.PRGROMSHA1 && ((templateRomMap[index].Header20.CHRROMSize == 0 && testRom.Header10.CHRROMSize == 0) || templateRomMap[index].Header20.CHRROMSHA1 == testRom.Header10.CHRROMSHA1) {
						return templateRomMap[index], nil
					}
				}
			} else if enableInes && templateRomMap[index].Header10 != nil {
				if testRom.Header20 != nil {
					if templateRomMap[index].Header10.PRGROMSHA1 == testRom.Header20.PRGROMSHA1 && ((templateRomMap[index].Header10.CHRROMSize == 0 && testRom.Header20.CHRROMSize == 0) || templateRomMap[index].Header10.CHRROMSHA1 == testRom.Header20.CHRROMSHA1) {
						return templateRomMap[index], nil
					}
				}

				if testRom.Header10 != nil {
					if templateRomMap[index].Header10.PRGROMSHA1 == testRom.Header10.PRGROMSHA1 && ((templateRomMap[index].Header10.CHRROMSize == 0 && testRom.Header10.CHRROMSize == 0) || templateRomMap[index].Header10.CHRROMSHA1 == testRom.Header10.CHRROMSHA1) {
						return templateRomMap[index], nil
					}
				}
			}
		}
	}

	if hashTypeTests&HASH_TYPE_MD5 > 0 {
		if templateRomMap["MD5:"+strings.ToUpper(hex.EncodeToString(testRom.MD5[:]))] != nil {
			return templateRomMap["MD5:"+strings.ToUpper(hex.EncodeToString(testRom.MD5[:]))], nil
		}

		for index := range templateRomMap {
			if templateRomMap[index].Header20 != nil {
				if testRom.Header20 != nil {
					if templateRomMap[index].Header20.PRGROMMD5 == testRom.Header20.PRGROMMD5 && ((templateRomMap[index].Header20.CHRROMSize == 0 && testRom.Header20.CHRROMSize == 0) || templateRomMap[index].Header20.CHRROMMD5 == testRom.Header20.CHRROMMD5) {
						return templateRomMap[index], nil
					}
				}

				if testRom.Header10 != nil {
					if templateRomMap[index].Header20.PRGROMMD5 == testRom.Header10.PRGROMMD5 && ((templateRomMap[index].Header20.CHRROMSize == 0 && testRom.Header10.CHRROMSize == 0) || templateRomMap[index].Header20.CHRROMMD5 == testRom.Header10.CHRROMMD5) {
						return templateRomMap[index], nil
					}
				}
			} else if enableInes && templateRomMap[index].Header10 != nil {
				if testRom.Header20 != nil {
					if templateRomMap[index].Header10.PRGROMMD5 == testRom.Header20.PRGROMMD5 && ((templateRomMap[index].Header10.CHRROMSize == 0 && testRom.Header20.CHRROMSize == 0) || templateRomMap[index].Header10.CHRROMMD5 == testRom.Header20.CHRROMMD5) {
						return templateRomMap[index], nil
					}
				}

				if testRom.Header10 != nil {
					if templateRomMap[index].Header10.PRGROMMD5 == testRom.Header10.PRGROMMD5 && ((templateRomMap[index].Header10.CHRROMSize == 0 && testRom.Header10.CHRROMSize == 0) || templateRomMap[index].Header10.CHRROMMD5 == testRom.Header10.CHRROMMD5) {
						return templateRomMap[index], nil
					}
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
			if templateRomMap[index].Header20 != nil {
				if testRom.Header20 != nil {
					if templateRomMap[index].Header20.PRGROMCRC32 == testRom.Header20.PRGROMCRC32 && ((templateRomMap[index].Header20.CHRROMSize == 0 && testRom.Header20.CHRROMSize == 0) || templateRomMap[index].Header20.CHRROMCRC32 == testRom.Header20.CHRROMCRC32) {
						return templateRomMap[index], nil
					}
				}

				if testRom.Header10 != nil {
					if templateRomMap[index].Header20.PRGROMCRC32 == testRom.Header10.PRGROMCRC32 && ((templateRomMap[index].Header20.CHRROMSize == 0 && testRom.Header10.CHRROMSize == 0) || templateRomMap[index].Header20.CHRROMCRC32 == testRom.Header10.CHRROMCRC32) {
						return templateRomMap[index], nil
					}
				}
			} else if enableInes && templateRomMap[index].Header10 != nil {
				if testRom.Header20 != nil {
					if templateRomMap[index].Header10.PRGROMCRC32 == testRom.Header20.PRGROMCRC32 && ((templateRomMap[index].Header10.CHRROMSize == 0 && testRom.Header20.CHRROMSize == 0) || templateRomMap[index].Header10.CHRROMCRC32 == testRom.Header20.CHRROMCRC32) {
						return templateRomMap[index], nil
					}
				}

				if testRom.Header10 != nil {
					if templateRomMap[index].Header10.PRGROMCRC32 == testRom.Header10.PRGROMCRC32 && ((templateRomMap[index].Header10.CHRROMSize == 0 && testRom.Header10.CHRROMSize == 0) || templateRomMap[index].Header10.CHRROMCRC32 == testRom.Header10.CHRROMCRC32) {
						return templateRomMap[index], nil
					}
				}
			}
		}
	}

	if hashTypeTests&HASH_TYPE_SUM16 > 0 {
		for index := range templateRomMap {
			if templateRomMap[index].Header20 != nil {
				if testRom.Header20 != nil {
					if templateRomMap[index].Header20.PRGROMSum16 == testRom.Header20.PRGROMSum16 && ((templateRomMap[index].Header20.CHRROMSize == 0 && testRom.Header20.CHRROMSize == 0) || templateRomMap[index].Header20.CHRROMSum16 == testRom.Header20.CHRROMSum16) {
						return templateRomMap[index], nil
					}
				}

				if testRom.Header10 != nil {
					if templateRomMap[index].Header20.PRGROMSum16 == testRom.Header10.PRGROMSum16 && ((templateRomMap[index].Header20.CHRROMSize == 0 && testRom.Header10.CHRROMSize == 0) || templateRomMap[index].Header20.CHRROMSum16 == testRom.Header10.CHRROMSum16) {
						return templateRomMap[index], nil
					}
				}
			} else if enableInes && templateRomMap[index].Header10 != nil {
				if testRom.Header20 != nil {
					if templateRomMap[index].Header10.PRGROMSum16 == testRom.Header20.PRGROMSum16 && ((templateRomMap[index].Header10.CHRROMSize == 0 && testRom.Header20.CHRROMSize == 0) || templateRomMap[index].Header10.CHRROMSum16 == testRom.Header20.CHRROMSum16) {
						return templateRomMap[index], nil
					}
				}

				if testRom.Header10 != nil {
					if templateRomMap[index].Header10.PRGROMSum16 == testRom.Header10.PRGROMSum16 && ((templateRomMap[index].Header10.CHRROMSize == 0 && testRom.Header10.CHRROMSize == 0) || templateRomMap[index].Header10.CHRROMSum16 == testRom.Header10.CHRROMSum16) {
						return templateRomMap[index], nil
					}
				}
			}
		}
	}

	testRomCrc32Bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(testRomCrc32Bytes, testRom.CRC32)
	return nil, errors.New("No match found for NES ROM: " + testRom.Name + "\nCRC32: " + strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes)) + "\nSHA1: " + strings.ToUpper(hex.EncodeToString(testRom.SHA1[:])) + "\nSHA256: " + strings.ToUpper(hex.EncodeToString(testRom.SHA256[:])))
}

// Update an NES ROM with info from a given template ROM.
func UpdateNESROM(targetRom *NESTool.NESROM, templateRom *NESTool.NESROM, truncateRom bool, organizeRoms bool, enableInes bool) error {
	if targetRom == nil || templateRom == nil {
		return errors.New("Missing target or template NES ROM for update.")
	}

	if templateRom.Header20 != nil {
		targetRom.Header20 = templateRom.Header20
		targetRom.Header10 = nil
	} else if enableInes && templateRom.Header10 != nil {
		targetRom.Header10 = templateRom.Header10
		targetRom.Header20 = nil
	} else {
		return errors.New("Unable to update ROM.")
	}

	if templateRom.Name != "" && (organizeRoms || targetRom.Name == "") {
		targetRom.Name = templateRom.Name
	}

	if templateRom.RelativePath != "" && (organizeRoms || targetRom.RelativePath == "") {
		targetRom.RelativePath = templateRom.RelativePath
	}

	targetRom.SHA256 = templateRom.SHA256
	targetRom.SHA1 = templateRom.SHA1
	targetRom.MD5 = templateRom.MD5
	targetRom.CRC32 = templateRom.CRC32

	if truncateRom {
		NESTool.TruncateROMDataAndSections(targetRom)
	}

	return nil
}

// Match and update a ROM to a template ROM from a map of potential template ROMs
func ProcessNESROMs(testRomList []*NESTool.NESROM, templateRomMap map[string]*NESTool.NESROM, hashTypeTests uint64, truncateRoms bool, organizeRoms bool, enableInes bool) []*NESTool.NESROM {
	returnRomList := make([]*NESTool.NESROM, 0)

	for index := range testRomList {
		tempRom, matchErr := MatchNESROM(testRomList[index], templateRomMap, hashTypeTests, enableInes)
		if matchErr == nil {
			updateErr := UpdateNESROM(testRomList[index], tempRom, truncateRoms, organizeRoms, enableInes)
			if updateErr == nil {
				returnRomList = append(returnRomList, testRomList[index])
			}
		}
	}

	return returnRomList
}

// Match an FDS ROM to a template ROM based on the specified hashing algorithm(s).
// Algorithms can be bitwise ORed together to use multiple in order of descending
// complexity.
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

// Update an FDS ROM with metadata for where to be written for organizational purposes.
// If a better way of determining matching FDS ROMs is developed in the future, this
// may become more robust.
func UpdateFDSROM(targetRom *FDSTool.FDSArchiveFile, templateRom *FDSTool.FDSArchiveFile, organizeRoms bool) error {
	if targetRom == nil || templateRom == nil {
		return errors.New("Missing target or template FDS ROM for update.")
	}

	if organizeRoms {
		targetRom.Name = templateRom.Name
		targetRom.RelativePath = templateRom.RelativePath
	}

	return nil
}

// Match and update an FDS ROM from a map of potential template FDS ROMs
func ProcessFDSROMs(testRomList []*FDSTool.FDSArchiveFile, templateRomMap map[string]*FDSTool.FDSArchiveFile, hashTypeTests uint64, organizeRoms bool) []*FDSTool.FDSArchiveFile {
	returnRomList := make([]*FDSTool.FDSArchiveFile, 0)

	for index := range testRomList {
		tempRom, matchErr := MatchFDSROM(testRomList[index], templateRomMap, hashTypeTests)
		if matchErr == nil {
			updateErr := UpdateFDSROM(testRomList[index], tempRom, organizeRoms)
			if updateErr == nil {
				returnRomList = append(returnRomList, testRomList[index])
			}
		}
	}

	return returnRomList
}
