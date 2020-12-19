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
package FileTools

import (
	"NES20Tool/FDSTool"
	"NES20Tool/NES20Tool"
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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

func LoadROM(fileName string, enableInes bool, preserveTrainer bool, basePath string) (*NES20Tool.NESROM, error) {
	byteSlice, relativePath, err := LoadFile(fileName, basePath)
	if err != nil {
		return nil, err
	}

	decodedRom, err := NES20Tool.DecodeNESROM(byteSlice, enableInes, preserveTrainer, relativePath)
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

func LoadROMRecursive(basePath string, enableInes bool, preserveTrainers bool) ([]*NES20Tool.NESROM, error) {
	romSlice := make([]*NES20Tool.NESROM, 0)
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
				case *NES20Tool.NESROMError:
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

func LoadROMRecursiveMap(path string, enableInes bool, preserveTrainers bool) (map[[32]byte]*NES20Tool.NESROM, error) {
	romSlice, err := LoadROMRecursive(path, enableInes, preserveTrainers)
	if err != nil {
		switch err.(type) {
		case *NES20Tool.NESROMError:
			break
		default:
			return nil, err
		}
	}

	romMap := make(map[[32]byte]*NES20Tool.NESROM)

	for index := range romSlice {
		if romMap[romSlice[index].SHA256] == nil {
			romMap[romSlice[index].SHA256] = romSlice[index]
		}
	}

	return romMap, nil
}

//TODO: Determine a better way to identify duplicates based on archive/filesystem contents
func LoadFDSArchiveRecursiveMap(path string, generateChecksums bool) (map[[32]byte]*FDSTool.FDSArchiveFile, error) {
	archiveSlice, err := LoadFDSArchiveRecursive(path, generateChecksums)
	if err != nil {
		switch err.(type) {
		case *FDSTool.FDSError:
			break
		default:
			return nil, err
		}
	}

	archiveMap := make(map[[32]byte]*FDSTool.FDSArchiveFile)
	for index := range archiveSlice {
		if archiveMap[archiveSlice[index].SHA256] == nil {
			archiveMap[archiveSlice[index].SHA256] = archiveSlice[index]
		}
	}

	return archiveMap, nil
}

func WriteROM(romModel *NES20Tool.NESROM, enableInes bool, truncateRom bool, preserveTrainer bool, destinationBasePath string) error {
	nesRomBytes, err := NES20Tool.EncodeNESROM(romModel, enableInes, truncateRom, preserveTrainer)
	if err != nil {
		return err
	}

	if destinationBasePath == "" {
		return ioutil.WriteFile(romModel.Filename, nesRomBytes, 0644)
	} else {
		tempRomPath := destinationBasePath
		if tempRomPath[len(tempRomPath)-1] != os.PathSeparator {
			tempRomPath = tempRomPath + string(os.PathSeparator)
		}
		tempRomPath = tempRomPath + romModel.RelativePath
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

func WriteFDSArchive(archiveModel *FDSTool.FDSArchiveFile, writeFDSHeader bool, destinationBasePath string) error {
	fdsArchiveBytes, err := FDSTool.EncodeFDSArchive(archiveModel, writeFDSHeader, false, false, false)
	if err != nil {
		return err
	}

	if destinationBasePath == "" {
		return ioutil.WriteFile(archiveModel.Filename, fdsArchiveBytes, 0644)
	} else {
		tempRomPath := destinationBasePath
		if tempRomPath[len(tempRomPath)-1] != os.PathSeparator {
			tempRomPath = tempRomPath + string(os.PathSeparator)
		}
		tempRomPath = tempRomPath + archiveModel.RelativePath
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

func WriteStringToFile(dataString string, filePath string) error {
	return ioutil.WriteFile(filePath, []byte(dataString), 0644)
}
