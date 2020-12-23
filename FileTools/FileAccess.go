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

package FileTools

import (
	"NES20Tool/FDSTool"
	"NES20Tool/NESTool"
	"NES20Tool/ProcessingTools"
	"NES20Tool/UNIFTool"
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Read in a file from disk
func LoadFile(fileName string, basePath string) ([]byte, string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, "", err
	}

	defer func() {
		err := f.Close()
		if err != nil {
			panic(errors.New("Unable to close file: " + fileName))
		}
	}()

	stats, err := f.Stat()
	if err != nil {
		return nil, "", err
	}

	size := stats.Size()
	byteSlice := make([]byte, size)

	bufr := bufio.NewReader(f)
	_, err = bufr.Read(byteSlice)

	if err != nil {
		return nil, "", err
	}

	relativePath := ""
	if basePath != "" {
		relativePath = strings.TrimPrefix(fileName, basePath)
		if relativePath[0] == os.PathSeparator {
			relativePath = relativePath[1:]
		}
	}

	return byteSlice, relativePath, nil
}

// Read in an INES or NES 2.0 ROM and decode it into an NESROM struct
func LoadROM(fileName string, enableInes bool, preserveTrainer bool, basePath string) (*NESTool.NESROM, error) {
	byteSlice, relativePath, err := LoadFile(fileName, basePath)
	if err != nil {
		return nil, err
	}

	decodedRom, err := NESTool.DecodeNESROM(byteSlice, enableInes, preserveTrainer, relativePath)
	if decodedRom != nil {
		decodedRom.Filename = fileName
		tempName := filepath.Base(fileName)
		tempNameLen := len(tempName)
		if tempName[tempNameLen-4:] == ".nes" {
			tempName = tempName[0 : tempNameLen-4]
		}

		decodedRom.Name = tempName
	}

	return decodedRom, err
}

// Read in a UNIF ROM and decode it into an NESROM struct
func LoadUNIF(fileName string, basePath string) (*NESTool.NESROM, error) {
	byteSlice, _, err := LoadFile(fileName, basePath)
	if err != nil {
		return nil, err
	}

	decodedRom, err := UNIFTool.DecodeUNIFROM(byteSlice)
	if decodedRom != nil {
		decodedRom.Filename = fileName
		tempName := filepath.Base(fileName)
		tempNameLen := len(tempName)
		if tempName[tempNameLen-5:] == ".unif" {
			tempName = tempName[0 : tempNameLen-5]
		} else if tempName[tempNameLen-4:] == ".unf" {
			tempName = tempName[0 : tempNameLen-4]
		}

		decodedRom.Name = tempName
	}

	return decodedRom, err
}

// Read in an FDS file and decode it into an FDSArchiveFile struct
func LoadFDSArchive(fileName string, basePath string, generateChecksums bool) (*FDSTool.FDSArchiveFile, error) {
	byteSlice, relativePath, err := LoadFile(fileName, basePath)
	if err != nil {
		return nil, err
	}

	decodedArchive, err := FDSTool.DecodeFDSArchive(byteSlice, relativePath, generateChecksums)
	if decodedArchive != nil {
		decodedArchive.Filename = fileName
		tempName := filepath.Base(fileName)
		tempNameLen := len(tempName)
		if tempName[tempNameLen-4:] == ".fds" {
			tempName = tempName[0 : tempNameLen-4]
		}

		decodedArchive.Name = tempName
	}

	return decodedArchive, nil
}

// Read in INES and NES 2.0 files recursively from a given path
func LoadROMRecursive(basePath string, enableInes bool, preserveTrainers bool) ([]*NESTool.NESROM, error) {
	romSlice := make([]*NESTool.NESROM, 0)
	nesRegEx, err := regexp.Compile("^.+\\.nes$")
	if err != nil {
		return nil, err
	}

	fullPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		if !info.IsDir() && nesRegEx.MatchString(info.Name()) {
			tempRom, err := LoadROM(path, enableInes, preserveTrainers, basePath)
			if err != nil {
				switch err.(type) {
				case *NESTool.NESROMError:
					break
				default:
					return err
				}
			}

			if tempRom != nil {
				println("Loading ROM: " + path)
				romSlice = append(romSlice, tempRom)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return romSlice, nil
}

// Read in UNIF files recursively from a given base path
func LoadUNIFRecursive(basePath string) ([]*NESTool.NESROM, error) {
	romSlice := make([]*NESTool.NESROM, 0)
	unifRegEx, err := regexp.Compile("^.+\\.unif$")
	if err != nil {
		return nil, err
	}

	unfRegEx, err := regexp.Compile("^.+\\.unf$")
	if err != nil {
		return nil, err
	}

	fullPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		if !info.IsDir() && (unifRegEx.MatchString(info.Name()) || unfRegEx.MatchString(info.Name())) {
			tempRom, err := LoadUNIF(path, basePath)
			if err != nil {
				switch err.(type) {
				case *NESTool.NESROMError:
					break
				default:
					return err
				}
			}

			if tempRom != nil {
				println("Loading ROM: " + path)
				romSlice = append(romSlice, tempRom)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return romSlice, nil
}

// Read in FDS files recursively from a given path
func LoadFDSArchiveRecursive(basePath string, generateChecksums bool) ([]*FDSTool.FDSArchiveFile, error) {
	archiveSlice := make([]*FDSTool.FDSArchiveFile, 0)
	fdsRegEx, err := regexp.Compile("^.+\\.fds$")
	if err != nil {
		return nil, err
	}

	fullPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		if !info.IsDir() && fdsRegEx.MatchString(info.Name()) {
			tempArchive, err := LoadFDSArchive(path, basePath, generateChecksums)
			if err != nil {
				switch err.(type) {
				case *FDSTool.FDSError:
					break
				default:
					return err
				}
			}

			if tempArchive != nil {
				println("Loading FDS archive: " + path)
				archiveSlice = append(archiveSlice, tempArchive)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return archiveSlice, nil
}

// Read in INES and NES 2.0 ROMs recursively and add them to a map, with checksums as keys
func LoadROMRecursiveMap(basePath string, enableInes bool, preserveTrainers bool, hashTypes uint64) (map[string]*NESTool.NESROM, error) {
	romSlice, err := LoadROMRecursive(basePath, enableInes, preserveTrainers)
	if err != nil {
		switch err.(type) {
		case *NESTool.NESROMError:
			break
		default:
			return nil, err
		}
	}

	romMap := make(map[string]*NESTool.NESROM)

	for index := range romSlice {
		if hashTypes&ProcessingTools.HASH_TYPE_SHA256 > 0 {
			if romMap["SHA256:"+strings.ToUpper(hex.EncodeToString(romSlice[index].SHA256[:]))] == nil {
				romMap["SHA256:"+strings.ToUpper(hex.EncodeToString(romSlice[index].SHA256[:]))] = romSlice[index]
			}
		}

		if hashTypes&ProcessingTools.HASH_TYPE_SHA1 > 0 {
			if romMap["SHA1:"+strings.ToUpper(hex.EncodeToString(romSlice[index].SHA1[:]))] == nil {
				romMap["SHA1:"+strings.ToUpper(hex.EncodeToString(romSlice[index].SHA1[:]))] = romSlice[index]
			}
		}

		if hashTypes&ProcessingTools.HASH_TYPE_MD5 > 0 {
			if romMap["MD5:"+strings.ToUpper(hex.EncodeToString(romSlice[index].MD5[:]))] == nil {
				romMap["MD5:"+strings.ToUpper(hex.EncodeToString(romSlice[index].MD5[:]))] = romSlice[index]
			}
		}

		if hashTypes&ProcessingTools.HASH_TYPE_CRC32 > 0 {
			testRomCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(testRomCrc32Bytes, romSlice[index].CRC32)

			if romMap["CRC32:"+strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes))] == nil {
				romMap["CRC32:"+strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes))] = romSlice[index]
			}
		}
	}

	return romMap, nil
}

// Read in UNIF files recursively and add them to a map, with checksums as keys
func LoadUNIFRecursiveMap(basePath string, hashTypes uint64) (map[string]*NESTool.NESROM, error) {
	romSlice, err := LoadUNIFRecursive(basePath)
	if err != nil {
		switch err.(type) {
		case *NESTool.NESROMError:
			break
		default:
			return nil, err
		}
	}

	romMap := make(map[string]*NESTool.NESROM)

	for index := range romSlice {
		if hashTypes&ProcessingTools.HASH_TYPE_SHA256 > 0 {
			if romMap["SHA256:"+strings.ToUpper(hex.EncodeToString(romSlice[index].SHA256[:]))] == nil {
				romMap["SHA256:"+strings.ToUpper(hex.EncodeToString(romSlice[index].SHA256[:]))] = romSlice[index]
			}
		}

		if hashTypes&ProcessingTools.HASH_TYPE_SHA1 > 0 {
			if romMap["SHA1:"+strings.ToUpper(hex.EncodeToString(romSlice[index].SHA1[:]))] == nil {
				romMap["SHA1:"+strings.ToUpper(hex.EncodeToString(romSlice[index].SHA1[:]))] = romSlice[index]
			}
		}

		if hashTypes&ProcessingTools.HASH_TYPE_MD5 > 0 {
			if romMap["MD5:"+strings.ToUpper(hex.EncodeToString(romSlice[index].MD5[:]))] == nil {
				romMap["MD5:"+strings.ToUpper(hex.EncodeToString(romSlice[index].MD5[:]))] = romSlice[index]
			}
		}

		if hashTypes&ProcessingTools.HASH_TYPE_CRC32 > 0 {
			testRomCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(testRomCrc32Bytes, romSlice[index].CRC32)

			if romMap["CRC32:"+strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes))] == nil {
				romMap["CRC32:"+strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes))] = romSlice[index]
			}
		}
	}

	return romMap, nil
}

// Read in FDS files and add them to a map, with checksums as keys
//TODO: Determine a better way to identify duplicates based on archive/filesystem contents
func LoadFDSArchiveRecursiveMap(basePath string, generateChecksums bool, hashTypes uint64) (map[string]*FDSTool.FDSArchiveFile, error) {
	archiveSlice, err := LoadFDSArchiveRecursive(basePath, generateChecksums)
	if err != nil {
		switch err.(type) {
		case *FDSTool.FDSError:
			break
		default:
			return nil, err
		}
	}

	archiveMap := make(map[string]*FDSTool.FDSArchiveFile)
	for index := range archiveSlice {
		if hashTypes&ProcessingTools.HASH_TYPE_SHA256 > 0 {
			if archiveMap["SHA256:"+strings.ToUpper(hex.EncodeToString(archiveSlice[index].SHA256[:]))] == nil {
				archiveMap["SHA256:"+strings.ToUpper(hex.EncodeToString(archiveSlice[index].SHA256[:]))] = archiveSlice[index]
			}
		}

		if hashTypes&ProcessingTools.HASH_TYPE_SHA1 > 0 {
			if archiveMap["SHA1:"+strings.ToUpper(hex.EncodeToString(archiveSlice[index].SHA1[:]))] == nil {
				archiveMap["SHA1:"+strings.ToUpper(hex.EncodeToString(archiveSlice[index].SHA1[:]))] = archiveSlice[index]
			}
		}

		if hashTypes&ProcessingTools.HASH_TYPE_MD5 > 0 {
			if archiveMap["MD5:"+strings.ToUpper(hex.EncodeToString(archiveSlice[index].MD5[:]))] == nil {
				archiveMap["MD5:"+strings.ToUpper(hex.EncodeToString(archiveSlice[index].MD5[:]))] = archiveSlice[index]
			}
		}

		if hashTypes&ProcessingTools.HASH_TYPE_CRC32 > 0 {
			testRomCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(testRomCrc32Bytes, archiveSlice[index].CRC32)

			if archiveMap["CRC32:"+strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes))] == nil {
				archiveMap["CRC32:"+strings.ToUpper(hex.EncodeToString(testRomCrc32Bytes))] = archiveSlice[index]
			}
		}
	}

	return archiveMap, nil
}

// Encode and write an NES ROM to disk
func WriteROM(romModel *NESTool.NESROM, enableInes bool, truncateRom bool, preserveTrainer bool, destinationBasePath string) error {
	nesRomBytes, err := NESTool.EncodeNESROM(romModel, enableInes, truncateRom, preserveTrainer)
	if err != nil {
		return err
	}

	if destinationBasePath == "" {
		tempFilename := romModel.Filename
		if tempFilename == "" {
			tempFilename = romModel.Name + ".nes"
		}

		return ioutil.WriteFile(tempFilename, nesRomBytes, 0644)
	} else {
		tempRelativePath := romModel.RelativePath
		if tempRelativePath == "" {
			tempRelativePath = romModel.Name + ".nes"
		}

		tempRomPath := destinationBasePath
		if tempRomPath[len(tempRomPath)-1] != os.PathSeparator {
			tempRomPath = tempRomPath + string(os.PathSeparator)
		}
		tempRomPath = tempRomPath + tempRelativePath
		directoryPath := tempRomPath[0:strings.LastIndex(tempRomPath, string(os.PathSeparator))]

		defer func() {
			err := os.MkdirAll(directoryPath, os.ModeDir|0770)
			if err != nil {
				panic(errors.New("Unable to create directory: " + directoryPath))
			}
		}()

		return ioutil.WriteFile(tempRomPath, nesRomBytes, 0644)
	}
}

// Encode and write an FDS archive to disk
func WriteFDSArchive(archiveModel *FDSTool.FDSArchiveFile, writeFDSHeader bool, destinationBasePath string) error {
	fdsArchiveBytes, err := FDSTool.EncodeFDSArchive(archiveModel, writeFDSHeader, false, false, false)
	if err != nil {
		return err
	}

	if destinationBasePath == "" {
		tempFilename := archiveModel.Filename
		if tempFilename == "" {
			tempFilename = archiveModel.Name + ".nes"
		}

		return ioutil.WriteFile(tempFilename, fdsArchiveBytes, 0644)
	} else {
		tempRelativePath := archiveModel.RelativePath
		if tempRelativePath == "" {
			tempRelativePath = archiveModel.Name + ".nes"
		}

		tempRomPath := destinationBasePath
		if tempRomPath[len(tempRomPath)-1] != os.PathSeparator {
			tempRomPath = tempRomPath + string(os.PathSeparator)
		}
		tempRomPath = tempRomPath + tempRelativePath
		directoryPath := tempRomPath[0:strings.LastIndex(tempRomPath, string(os.PathSeparator))]

		defer func() {
			err := os.MkdirAll(directoryPath, os.ModeDir|0770)
			if err != nil {
				panic(errors.New("Unable to create directory: " + directoryPath))
			}
		}()

		return ioutil.WriteFile(tempRomPath, fdsArchiveBytes, 0644)
	}
}

// Write a string to a file (used for XML generation)
func WriteStringToFile(dataString string, filePath string) error {
	return ioutil.WriteFile(filePath, []byte(dataString), 0644)
}
