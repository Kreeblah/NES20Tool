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
	"strconv"
	"strings"
)

func MarshalDBFileFromROMMap(nesRoms map[string]*NESTool.NESROM, enableInes bool) (string, error) {
	dbString := ""

	for index := range nesRoms {
		if nesRoms[index].Header20 != nil && nesRoms[index].Header20.ConsoleType == 0 {
			dbString = dbString + strconv.Itoa(int(nesRoms[index].Header20.PRGROMCalculatedSize)) + "^^"
			dbString = dbString + strconv.Itoa(int(nesRoms[index].Header20.CHRROMCalculatedSize)) + "^^"

			prgCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(prgCrc32Bytes, nesRoms[index].Header20.PRGROMCRC32)
			dbString = dbString + strings.ToUpper(hex.EncodeToString(prgCrc32Bytes)) + "^^"

			if nesRoms[index].Header20.CHRROMCalculatedSize > 0 {
				chrCrc32Bytes := make([]byte, 4)
				binary.BigEndian.PutUint32(chrCrc32Bytes, nesRoms[index].Header20.CHRROMCRC32)
				dbString = dbString + strings.ToUpper(hex.EncodeToString(chrCrc32Bytes)) + "^^"
			} else {
				dbString = dbString + "^^"
			}

			if nesRoms[index].Name != "" {
				dbString = dbString + nesRoms[index].Name + "^^"
			} else if nesRoms[index].RelativePath != "" {
				tempName := nesRoms[index].RelativePath
				tempName = tempName[strings.LastIndex(tempName, string(os.PathSeparator)) + 1:]
				tempName = tempName[:strings.LastIndex(tempName, ".nes")]
				dbString = dbString + tempName + "^^"
			} else {
				dbString = dbString + "^^"
			}

			headerBytes, err := NESTool.EncodeNESROMHeader(nesRoms[index], false, false)
			if err != nil {
				return "", err
			}
			dbString = dbString + strings.ToUpper(hex.EncodeToString(headerBytes)) + "\n"
		} else if nesRoms[index].Header10 != nil && enableInes {
			dbString = dbString + strconv.Itoa(int(nesRoms[index].Header10.PRGROMCalculatedSize)) + "^^"
			dbString = dbString + strconv.Itoa(int(nesRoms[index].Header10.CHRROMCalculatedSize)) + "^^"

			prgCrc32Bytes := make([]byte, 4)
			binary.BigEndian.PutUint32(prgCrc32Bytes, nesRoms[index].Header10.PRGROMCRC32)
			dbString = dbString + strings.ToUpper(hex.EncodeToString(prgCrc32Bytes)) + "^^"

			if nesRoms[index].Header10.CHRROMCalculatedSize > 0 {
				chrCrc32Bytes := make([]byte, 4)
				binary.BigEndian.PutUint32(chrCrc32Bytes, nesRoms[index].Header10.CHRROMCRC32)
				dbString = dbString + strings.ToUpper(hex.EncodeToString(chrCrc32Bytes)) + "^^"
			} else {
				dbString = dbString + "^^"
			}

			if nesRoms[index].Name != "" {
				dbString = dbString + nesRoms[index].Name + "^^"
			} else if nesRoms[index].RelativePath != "" {
				tempName := nesRoms[index].RelativePath
				tempName = tempName[strings.LastIndex(tempName, string(os.PathSeparator)) + 1:]
				tempName = tempName[:strings.LastIndex(tempName, ".nes")]
				dbString = dbString + tempName + "^^"
			} else {
				dbString = dbString + "^^"
			}

			headerBytes, err := NESTool.EncodeNESROMHeader(nesRoms[index], true, false)
			if err != nil {
				return "", err
			}
			dbString = dbString + strings.ToUpper(hex.EncodeToString(headerBytes)) + "\n"
		}
	}

	return dbString, nil
}