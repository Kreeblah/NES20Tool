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
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

func LoadROM(fileName string) (*NES20Tool.NESROM, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stats, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := stats.Size()
	byteSlice := make([]byte, size)

	bufr := bufio.NewReader(f)
	_, err = bufr.Read(byteSlice)

	if err != nil {
		return nil, err
	}

	decodedRom, err := NES20Tool.DecodeNESROM(byteSlice)
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

func LoadROMRecursive(path string) ([]*NES20Tool.NESROM, error) {
	//	romMap := make(map[[32]byte]*NES20Tool.NESROM)
	romSlice := make([]*NES20Tool.NESROM, 0)
	nesRegEx, err := regexp.Compile("^.+\\.nes$")
	if err != nil {
		return nil, err
	}

	fullPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		if !info.IsDir() && nesRegEx.MatchString(info.Name()) {
			tempRom, err := LoadROM(path)
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

func LoadROMRecursiveMap(path string) (map[[32]byte]*NES20Tool.NESROM, error) {
	romSlice, err := LoadROMRecursive(path)
	if err != nil {
		switch err.(type) {
		case *NES20Tool.NESROMError:
			break
		default:
			return nil, err
		}
	}

	romMap := make(map[[32]byte]*NES20Tool.NESROM)

	for index, _ := range romSlice {
		if romMap[romSlice[index].SHA256] == nil {
			romMap[romSlice[index].SHA256] = romSlice[index]
		}
	}

	return romMap, nil
}

func WriteROM(romModel *NES20Tool.NESROM) error {
	nesRomBytes, err := NES20Tool.EncodeNESROM(romModel)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(romModel.Filename, nesRomBytes, 0644)
}

func WriteStringToFile(dataString string, filePath string) error {
	return ioutil.WriteFile(filePath, []byte(dataString), 0644)
}
