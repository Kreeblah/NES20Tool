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

// This is a file for generating flat file databases for lookups
// of NES 2.0 headers with Sanni's cart reader.
// https://github.com/sanni/cartreader

package FileTools

import (
	"NES20Tool/NESTool"
	"encoding/binary"
	"encoding/hex"
	"os"
	"sort"
	"strconv"
	"strings"
)

func MarshalDBFileFromROMMap(nesRoms map[string]*NESTool.NESROM, enableInes bool) (string, error) {
	dbString := ""
	tempString := ""
	romNameMap := make(map[string]string)

	for index := range nesRoms {
		if nesRoms[index].Header20 != nil && nesRoms[index].Header20.ConsoleType == 0 {
			tempName := ""

			tempString = strconv.Itoa(int(nesRoms[index].Header20.PRGROMCalculatedSize)) + "^^"
			tempString = tempString + strconv.Itoa(int(nesRoms[index].Header20.CHRROMCalculatedSize)) + "^^"

			prgCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(prgCrc32Bytes, nesRoms[index].Header20.PRGROMCRC32)
			tempString = tempString + strings.ToUpper(hex.EncodeToString(prgCrc32Bytes)) + "^^"

			if nesRoms[index].Header20.CHRROMCalculatedSize > 0 {
				chrCrc32Bytes := make([]byte, 4)
				binary.BigEndian.PutUint32(chrCrc32Bytes, nesRoms[index].Header20.CHRROMCRC32)
				tempString = tempString + strings.ToUpper(hex.EncodeToString(chrCrc32Bytes)) + "^^"
			} else {
				tempString = tempString + "^^"
			}

			if nesRoms[index].Name != "" {
				tempString = tempString + nesRoms[index].Name + "^^"
			} else if nesRoms[index].RelativePath != "" {
				tempName = nesRoms[index].RelativePath
				tempName = tempName[strings.LastIndex(tempName, string(os.PathSeparator)) + 1:]
				tempName = tempName[:strings.LastIndex(tempName, ".nes")]
				tempName = strings.Replace(tempName, "&amp;", "&", -1)
				tempName = strings.Replace(tempName, "&gt;", ">", -1)
				tempName = strings.Replace(tempName, "&lt;", "<", -1)
				tempString = tempString + tempName + "^^"
			} else {
				tempString = tempString + "^^"
			}

			headerBytes, err := NESTool.EncodeNESROMHeader(nesRoms[index], false, false)
			if err != nil {
				return "", err
			}
			tempString = tempString + strings.ToUpper(hex.EncodeToString(headerBytes))

			tempLength := len(tempString)
			if tempLength < 255 {
				for i := 0; i < (255 - tempLength); i++ {
					tempString = tempString + "\000"
				}
			}

			tempString = tempString + "\000"

			romNameMap[tempName] = tempString
		} else if enableInes && nesRoms[index].Header10 != nil && nesRoms[index].Header10.VsUnisystem == false {
			tempName := ""

			tempString = strconv.Itoa(int(nesRoms[index].Header10.PRGROMCalculatedSize)) + "^^"
			tempString = tempString + strconv.Itoa(int(nesRoms[index].Header10.CHRROMCalculatedSize)) + "^^"

			prgCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(prgCrc32Bytes, nesRoms[index].Header10.PRGROMCRC32)
			tempString = tempString + strings.ToUpper(hex.EncodeToString(prgCrc32Bytes)) + "^^"

			if nesRoms[index].Header10.CHRROMCalculatedSize > 0 {
				chrCrc32Bytes := make([]byte, 4)
				binary.BigEndian.PutUint32(chrCrc32Bytes, nesRoms[index].Header10.CHRROMCRC32)
				tempString = tempString + strings.ToUpper(hex.EncodeToString(chrCrc32Bytes)) + "^^"
			} else {
				tempString = tempString + "^^"
			}

			if nesRoms[index].Name != "" {
				tempString = tempString + nesRoms[index].Name + "^^"
			} else if nesRoms[index].RelativePath != "" {
				tempName = nesRoms[index].RelativePath
				tempName = tempName[strings.LastIndex(tempName, string(os.PathSeparator)) + 1:]
				tempName = tempName[:strings.LastIndex(tempName, ".nes")]
				tempName = strings.Replace(tempName, "&amp;", "&", -1)
				tempName = strings.Replace(tempName, "&gt;", ">", -1)
				tempName = strings.Replace(tempName, "&lt;", "<", -1)
				tempString = tempString + tempName + "^^"
			} else {
				tempString = tempString + "^^"
			}

			headerBytes, err := NESTool.EncodeNESROMHeader(nesRoms[index], true, false)
			if err != nil {
				return "", err
			}
			tempString = tempString + strings.ToUpper(hex.EncodeToString(headerBytes))

			tempLength := len(tempString)
			if tempLength < 255 {
				for i := 0; i < (255 - tempLength); i++ {
					tempString = tempString + "\000"
				}
			}

			tempString = tempString + "\000"

			romNameMap[tempName] = tempString
		}
	}

	keys := make([]string, 0, len(romNameMap))
	for k := range romNameMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		dbString = dbString + romNameMap[k]
	}

	return dbString, nil
}