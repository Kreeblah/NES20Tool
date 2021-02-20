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

// https://wiki.nesdev.com/w/index.php/FDS_file_format
// This file uses the FDS file format as a reference for
// the FDS header portion (which is trivial).

// https://wiki.nesdev.com/w/index.php/FDS_disk_format
// It also uses the FDS disk format for parsing the fields
// and data chunks of the FDS disks.  It also supports QD
// disks, the only difference seeming to be that they're
// 65,536 bytes instead of 65,500 bytes.

// http://forums.nesdev.com/viewtopic.php?p=194867
// On the off chance that anybody wants to calculate FDS
// CRCs, this also uses the algorithm mentioned here
// to do that.

package FDSTool

import (
	"github.com/Kreeblah/NES20Tool/NESTool"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"hash/crc32"
	"sort"
	"strconv"
	"time"
)

var (
	FDS_SIDE_SIZE              uint64 = 65500
	QD_SIDE_SIZE               uint64 = 65536
	FDS_DISK_INFO_BLOCK        uint64 = 1
	FDS_DISK_FILE_LAYOUT_BLOCK uint64 = 2
	FDS_FILE_HEADER_BLOCK      uint64 = 3
	FDS_FILE_DATA_BLOCK        uint64 = 4
	FDS_EPOCH                         = 1925
	FDS_HEADER_MAGIC                  = "\x46\x44\x53\x1a"
	FDS_HEADER_PADDING                = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
	FDS_MAGIC                         = "*NINTENDO-HVC*"
)

type FDSArchiveFile struct {
	Name         string
	Filename     string
	RelativePath string
	Size         uint64
	CRC32        uint32
	MD5          [16]byte
	SHA1         [20]byte
	SHA256       [32]byte
	ArchiveDisks []*FDSDisk
}

type FDSDisk struct {
	DiskNumber uint8
	DiskSides  []*FDSSide
}

type FDSSide struct {
	Size                   uint64
	CRC32                  uint32
	MD5                    [16]byte
	SHA1                   [20]byte
	SHA256                 [32]byte
	ManufacturerCode       uint8
	FDSGameName            string
	GameType               uint8
	RevisionNumber         uint8
	SideNumber             uint8
	DiskNumber             uint8
	DiskType               uint8
	Byte18                 uint8
	BootFileID             uint8
	Byte1A                 uint8
	Byte1B                 uint8
	Byte1C                 uint8
	Byte1D                 uint8
	Byte1E                 uint8
	ManufacturingDate      []byte
	CountryCode            uint8
	Byte23                 uint8
	Byte24                 uint8
	Byte25                 uint8
	Byte26                 uint8
	Byte27                 uint8
	Byte28                 uint8
	Byte29                 uint8
	Byte2A                 uint8
	Byte2B                 uint8
	RewriteDate            []byte
	Byte2F                 uint8
	Byte30                 uint8
	DiskWriterSerialNumber uint16
	Byte33                 uint8
	RewriteCount           uint8
	ActualDiskSide         uint8
	Byte36                 uint8
	Price                  uint8
	DiskInfoCRC            uint16
	FileTableCRC           uint16
	SideFiles              []*FDSFile
	UnallocatedSpace       []byte
	UnallocatedSpaceOffset uint16
}

type FDSFile struct {
	FileNumber             uint8
	FileIdentificationCode uint8
	FileName               string
	FileAddress            uint16
	FileSize               uint16
	FileType               uint8
	FileMetadataCRC        uint16
	FileData               *FDSFileData
}

type FDSFileData struct {
	Size        uint64
	CRC32       uint32
	MD5         [16]byte
	SHA1        [20]byte
	SHA256      [32]byte
	FileData    []byte
	FileDataCRC uint16
}

type FDSError struct {
	Text string
}

func (r *FDSError) Error() string {
	return r.Text
}

// Read a byte slice and attempt to decode it into an FDSArchiveFile structure
func DecodeFDSArchive(inputFile []byte, relativePath string, generateChecksums bool) (*FDSArchiveFile, error) {
	// Get all of the disk sides as byte slices
	sideByteSlices, err := GetStrippedDiskSideByteSlices(inputFile)
	if err != nil {
		return nil, err
	}

	// Find out how many slices we have
	numberOfSides := len(sideByteSlices)

	// Create the archive file
	tempArchive := &FDSArchiveFile{}
	sideArray := make([]*FDSSide, 0)

	// File location, size, and hashing metadata
	tempArchive.RelativePath = relativePath
	tempArchive.CRC32 = crc32.ChecksumIEEE(inputFile)
	tempArchive.MD5 = md5.Sum(inputFile)
	tempArchive.SHA1 = sha1.Sum(inputFile)
	tempArchive.SHA256 = sha256.Sum256(inputFile)
	tempArchive.Size = uint64(len(inputFile))

	// Decode and record metadata about each disk side
	for sliceIndex := 0; sliceIndex < numberOfSides; sliceIndex++ {
		tempSide, err := DecodeFDSSide(sideByteSlices[sliceIndex], generateChecksums)
		if err != nil {
			return nil, err
		}

		tempSide.CRC32 = crc32.ChecksumIEEE(sideByteSlices[sliceIndex])
		tempSide.MD5 = md5.Sum(sideByteSlices[sliceIndex])
		tempSide.SHA1 = sha1.Sum(sideByteSlices[sliceIndex])
		tempSide.SHA256 = sha256.Sum256(sideByteSlices[sliceIndex])
		tempSide.Size = uint64(len(sideByteSlices[sliceIndex]))

		sideArray = append(sideArray, tempSide)
	}

	// Assign disk numbers.  Each disk has two sides.
	diskNumbers := make([]uint8, 0)

	for sideIndex := range sideArray {
		hasDiskNumber := false
		for _, diskNumber := range diskNumbers {
			if sideArray[sideIndex].DiskNumber == diskNumber {
				hasDiskNumber = true
			}
		}

		if !hasDiskNumber {
			diskNumbers = append(diskNumbers, sideArray[sideIndex].DiskNumber)
		}
	}

	sort.Slice(diskNumbers, func(i int, j int) bool { return diskNumbers[i] < diskNumbers[j] })

	for diskIndex := range diskNumbers {
		tempArchive.ArchiveDisks = append(tempArchive.ArchiveDisks, &FDSDisk{})
		numberOfDisks := len(tempArchive.ArchiveDisks)
		tempArchive.ArchiveDisks[numberOfDisks-1].DiskNumber = diskNumbers[diskIndex]
		tempArchive.ArchiveDisks[numberOfDisks-1].DiskSides = make([]*FDSSide, 0)
	}

	for sideIndex := range sideArray {
		for diskIndex := range tempArchive.ArchiveDisks {
			if tempArchive.ArchiveDisks[diskIndex].DiskNumber == sideArray[sideIndex].DiskNumber {
				tempArchive.ArchiveDisks[diskIndex].DiskSides = append(tempArchive.ArchiveDisks[diskIndex].DiskSides, sideArray[sideIndex])
			}
		}
	}

	return tempArchive, nil
}

func DecodeFDSSide(inputSide []byte, generateChecksums bool) (*FDSSide, error) {
	// Validate the internal FDS header
	if inputSide[0x00] != uint8(FDS_DISK_INFO_BLOCK) {
		return nil, &FDSError{Text: "Unable to decode header for FDS side."}
	}

	readChecksums := false

	tempSide := &FDSSide{}
	var err error

	// Read in filesystem metadata for the disk side
	tempSide.ManufacturerCode = inputSide[0x0f]
	tempSide.FDSGameName = string(inputSide[0x10:0x13])
	tempSide.GameType = inputSide[0x13]
	tempSide.RevisionNumber = inputSide[0x14]
	tempSide.SideNumber = inputSide[0x15]
	tempSide.DiskNumber = inputSide[0x16]
	tempSide.DiskType = inputSide[0x17]
	tempSide.Byte18 = inputSide[0x18]
	tempSide.BootFileID = inputSide[0x19]
	tempSide.Byte1A = inputSide[0x1a]
	tempSide.Byte1B = inputSide[0x1b]
	tempSide.Byte1C = inputSide[0x1c]
	tempSide.Byte1D = inputSide[0x1d]
	tempSide.Byte1E = inputSide[0x1e]
	tempSide.ManufacturingDate = inputSide[0x1f:0x22]
	tempSide.CountryCode = inputSide[0x22]
	tempSide.Byte23 = inputSide[0x23]
	tempSide.Byte24 = inputSide[0x24]
	tempSide.Byte25 = inputSide[0x25]
	tempSide.Byte26 = inputSide[0x26]
	tempSide.Byte27 = inputSide[0x27]
	tempSide.Byte28 = inputSide[0x28]
	tempSide.Byte29 = inputSide[0x29]
	tempSide.Byte2A = inputSide[0x2a]
	tempSide.Byte2B = inputSide[0x2b]
	tempSide.RewriteDate = inputSide[0x2c:0x2f]
	tempSide.Byte2F = inputSide[0x2f]
	tempSide.Byte30 = inputSide[0x30]
	tempSide.DiskWriterSerialNumber = binary.LittleEndian.Uint16(inputSide[0x31:0x34])
	tempSide.Byte33 = inputSide[0x33]
	tempSide.RewriteCount = inputSide[0x34]
	tempSide.ActualDiskSide = inputSide[0x35]
	tempSide.Byte36 = inputSide[0x36]
	tempSide.Price = inputSide[0x37]

	// Generate a CRC for the initial metadata chunk
	diskInfoBytes := make([]byte, 0x38-0x00)
	copy(diskInfoBytes, inputSide[0:0x38])
	diskInfoBytes = append(diskInfoBytes, []byte{'\x00', '\x00'}...)
	testCrc, err := GenerateFDSBlockCRC(diskInfoBytes)
	if err != nil {
		return nil, err
	}

	// Read and update checksums as applicable
	if binary.LittleEndian.Uint16(inputSide[0x38:0x3a]) == testCrc && inputSide[0x3a] == '\x01' {
		readChecksums = true
		tempSide.DiskInfoCRC = testCrc
	} else if bytes.Compare(inputSide[0x38:0x3b], []byte{'\x00', '\x00', '\x01'}) == 0 {
		readChecksums = true
		if !generateChecksums {
			tempSide.DiskInfoCRC = binary.LittleEndian.Uint16(inputSide[0x38:0x3a])
		} else {
			tempSide.DiskInfoCRC = testCrc
		}
	} else if generateChecksums {
		tempSide.DiskInfoCRC = testCrc
	} else {
		tempSide.DiskInfoCRC = 0
	}

	var checksumOffset uint16 = 0
	if readChecksums {
		checksumOffset = 2
	}

	// Check for how many files are on the disk side
	if inputSide[0x38+checksumOffset] != uint8(FDS_DISK_FILE_LAYOUT_BLOCK) {
		return nil, &FDSError{Text: "Unable to determine number of files on FDS side."}
	}

	numberOfFiles := inputSide[0x39+checksumOffset]

	// More checksums
	if readChecksums {
		tempSide.FileTableCRC = binary.LittleEndian.Uint16(inputSide[0x38+checksumOffset : 0x3a+checksumOffset])
	} else {
		tempSide.FileTableCRC = 0
	}

	if generateChecksums {
		fileTableBytes := make([]byte, (0x3a+checksumOffset)-(0x38+checksumOffset))
		copy(fileTableBytes, inputSide[0x38+checksumOffset:0x3a+checksumOffset])
		fileTableBytes = append(fileTableBytes, []byte{'\x00', '\x00'}...)
		tempSide.FileTableCRC, err = GenerateFDSBlockCRC(fileTableBytes)
		if err != nil {
			return nil, err
		}
	}

	var currentIndex = 0x003a + (2 * checksumOffset)

	// Read each file on the disk side into a struct
	for fileIndex := 0; fileIndex < int(numberOfFiles); fileIndex++ {
		// Verify we're in the right block
		if inputSide[currentIndex] != uint8(FDS_FILE_HEADER_BLOCK) {
			return nil, &FDSError{Text: "Unable to read file header."}
		}

		tempFile := &FDSFile{}

		// File metadata
		tempFile.FileNumber = inputSide[currentIndex+1]
		tempFile.FileIdentificationCode = inputSide[currentIndex+2]
		tempFile.FileName = string(inputSide[currentIndex+3 : currentIndex+11])
		tempFile.FileAddress = binary.LittleEndian.Uint16(inputSide[currentIndex+11 : currentIndex+13])
		tempFile.FileSize = binary.LittleEndian.Uint16(inputSide[currentIndex+13 : currentIndex+15])
		tempFile.FileType = inputSide[currentIndex+15]

		// Checksums.  Again.
		if readChecksums {
			tempFile.FileMetadataCRC = binary.LittleEndian.Uint16(inputSide[currentIndex+16 : currentIndex+18])
		} else {
			tempFile.FileMetadataCRC = 0
		}

		if generateChecksums {
			fileMetadataBytes := make([]byte, (currentIndex+16)-currentIndex)
			copy(fileMetadataBytes, inputSide[currentIndex:currentIndex+16])
			fileMetadataBytes = append(fileMetadataBytes, []byte{'\x00', '\x00'}...)
			tempFile.FileMetadataCRC, err = GenerateFDSBlockCRC(fileMetadataBytes)
			if err != nil {
				return nil, err
			}
		}

		// Read in the file contents
		if inputSide[currentIndex+16+checksumOffset] != uint8(FDS_FILE_DATA_BLOCK) {
			return nil, &FDSError{Text: "Unable to read file data."}
		}

		tempFileData := &FDSFileData{}
		tempFileData.FileData = inputSide[currentIndex+17+checksumOffset : currentIndex+17+tempFile.FileSize+checksumOffset]

		tempFileData.CRC32 = crc32.ChecksumIEEE(tempFileData.FileData)
		tempFileData.MD5 = md5.Sum(tempFileData.FileData)
		tempFileData.SHA1 = sha1.Sum(tempFileData.FileData)
		tempFileData.SHA256 = sha256.Sum256(tempFileData.FileData)
		tempFileData.Size = uint64(len(tempFileData.FileData))

		if readChecksums {
			tempFileData.FileDataCRC = binary.LittleEndian.Uint16(inputSide[currentIndex+17+tempFile.FileSize+checksumOffset : currentIndex+17+tempFile.FileSize+checksumOffset+2])
		} else {
			tempFileData.FileDataCRC = 0
		}

		if generateChecksums {
			fileDataBytes := make([]byte, 1+len(tempFileData.FileData))
			fileDataBytes[0] = uint8(FDS_FILE_DATA_BLOCK)
			copy(fileDataBytes[1:], tempFileData.FileData)
			fileDataBytes = append(fileDataBytes, []byte{'\x00', '\x00'}...)
			tempFileData.FileDataCRC, err = GenerateFDSBlockCRC(fileDataBytes)
			if err != nil {
				return nil, err
			}
		}

		tempFile.FileData = tempFileData
		tempSide.SideFiles = append(tempSide.SideFiles, tempFile)

		currentIndex = currentIndex + 15 + 2 + tempFile.FileSize + (2 * checksumOffset)
	}

	if currentIndex < uint16(FDS_SIDE_SIZE) {
		tempSide.UnallocatedSpace = inputSide[currentIndex:FDS_SIDE_SIZE]
		tempSide.UnallocatedSpaceOffset = currentIndex
	}

	return tempSide, nil
}

// Turn an FDSArchiveFile struct into a byte slice that can be written to disk as a .fds file
func EncodeFDSArchive(inputArchive *FDSArchiveFile, writeHeader bool, writeChecksums bool, generateChecksums bool, writeQd bool) ([]byte, error) {
	archiveBytes := make([]byte, 0)

	// Write an FDS header if requested.  Don't do this unless you know you need to, though.
	if writeHeader {
		var numberOfSides uint8 = 0
		for diskIndex := range inputArchive.ArchiveDisks {
			numberOfSides = numberOfSides + uint8(len(inputArchive.ArchiveDisks[diskIndex].DiskSides))
		}
		archiveBytes = append(archiveBytes, []byte(FDS_HEADER_MAGIC)...)
		archiveBytes = append(archiveBytes, numberOfSides)
		archiveBytes = append(archiveBytes, []byte(FDS_HEADER_PADDING)...)
	}

	// Encode and append each disk side
	for diskIndex := range inputArchive.ArchiveDisks {
		for sideIndex := range inputArchive.ArchiveDisks[diskIndex].DiskSides {
			sideBytes, err := EncodeFDSSide(inputArchive.ArchiveDisks[diskIndex].DiskSides[sideIndex], writeChecksums, generateChecksums, writeQd)
			if err != nil {
				return nil, err
			}

			archiveBytes = append(archiveBytes, sideBytes...)
		}
	}

	return archiveBytes, nil
}

// Turn an FDSSide struct into something that can be appended to the data
// in an FDSArchiveFile struct
func EncodeFDSSide(inputSide *FDSSide, writeChecksums bool, generateChecksums bool, writeQd bool) ([]byte, error) {
	sideSlice := make([]byte, 0)

	// Lots of metadata in the info block
	sideSlice = append(sideSlice, byte(FDS_DISK_INFO_BLOCK))
	sideSlice = append(sideSlice, []byte(FDS_MAGIC)...)
	sideSlice = append(sideSlice, inputSide.ManufacturerCode)
	sideSlice = append(sideSlice, []byte(inputSide.FDSGameName)...)
	sideSlice = append(sideSlice, inputSide.GameType)
	sideSlice = append(sideSlice, inputSide.RevisionNumber)
	sideSlice = append(sideSlice, inputSide.SideNumber)
	sideSlice = append(sideSlice, inputSide.DiskNumber)
	sideSlice = append(sideSlice, inputSide.DiskType)
	sideSlice = append(sideSlice, inputSide.Byte18)
	sideSlice = append(sideSlice, inputSide.BootFileID)
	sideSlice = append(sideSlice, inputSide.Byte1A)
	sideSlice = append(sideSlice, inputSide.Byte1B)
	sideSlice = append(sideSlice, inputSide.Byte1C)
	sideSlice = append(sideSlice, inputSide.Byte1D)
	sideSlice = append(sideSlice, inputSide.Byte1E)
	sideSlice = append(sideSlice, inputSide.ManufacturingDate...)
	sideSlice = append(sideSlice, inputSide.CountryCode)
	sideSlice = append(sideSlice, inputSide.Byte23)
	sideSlice = append(sideSlice, inputSide.Byte24)
	sideSlice = append(sideSlice, inputSide.Byte25)
	sideSlice = append(sideSlice, inputSide.Byte26)
	sideSlice = append(sideSlice, inputSide.Byte27)
	sideSlice = append(sideSlice, inputSide.Byte28)
	sideSlice = append(sideSlice, inputSide.Byte29)
	sideSlice = append(sideSlice, inputSide.Byte2A)
	sideSlice = append(sideSlice, inputSide.Byte2B)
	sideSlice = append(sideSlice, inputSide.RewriteDate...)
	sideSlice = append(sideSlice, inputSide.Byte2F)
	sideSlice = append(sideSlice, inputSide.Byte30)
	diskWriterSerialNumberBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(diskWriterSerialNumberBytes, inputSide.DiskWriterSerialNumber)
	sideSlice = append(sideSlice, diskWriterSerialNumberBytes...)
	sideSlice = append(sideSlice, inputSide.Byte33)
	sideSlice = append(sideSlice, inputSide.RewriteCount)
	sideSlice = append(sideSlice, inputSide.ActualDiskSide)
	sideSlice = append(sideSlice, inputSide.Byte36)
	sideSlice = append(sideSlice, inputSide.Price)

	// Yet more checksums
	if writeChecksums {
		tempCrc16 := inputSide.DiskInfoCRC
		if generateChecksums {
			diskInfoBytes := make([]byte, len(sideSlice))
			copy(diskInfoBytes, sideSlice)
			diskInfoBytes = append(diskInfoBytes, []byte{'\x00', '\x00'}...)
			tempCrc16Generated, err := GenerateFDSBlockCRC(diskInfoBytes)
			if err != nil {
				return nil, &FDSError{Text: "Unable to generate disk info CRC."}
			}

			tempCrc16 = tempCrc16Generated
		}

		crcBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(crcBytes, tempCrc16)

		sideSlice = append(sideSlice, crcBytes...)
	}

	// Now for the layout . . .
	fileLayoutSlice := make([]byte, 2)
	fileLayoutSlice[0] = byte(FDS_DISK_FILE_LAYOUT_BLOCK)
	fileLayoutSlice[1] = uint8(len(inputSide.SideFiles))

	// . . . and checksums
	if writeChecksums {
		tempCrc16 := inputSide.FileTableCRC
		if generateChecksums {
			fileLayoutBytes := make([]byte, len(fileLayoutSlice))
			copy(fileLayoutBytes, fileLayoutSlice)
			fileLayoutBytes = append(fileLayoutBytes, []byte{'\x00', '\x00'}...)
			tempCrc16Generated, err := GenerateFDSBlockCRC(fileLayoutBytes)
			if err != nil {
				return nil, &FDSError{Text: "Unable to generate disk info CRC."}
			}

			tempCrc16 = tempCrc16Generated
		}

		crcBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(crcBytes, tempCrc16)

		fileLayoutSlice = append(fileLayoutSlice, crcBytes...)
	}

	sideSlice = append(sideSlice, fileLayoutSlice...)

	// Finally, each of the files and their metadata (and checksums)
	for index := range inputSide.SideFiles {
		fileHeaderSlice := make([]byte, 0)

		fileHeaderSlice = append(fileHeaderSlice, byte(FDS_FILE_HEADER_BLOCK))
		fileHeaderSlice = append(fileHeaderSlice, inputSide.SideFiles[index].FileNumber)
		fileHeaderSlice = append(fileHeaderSlice, inputSide.SideFiles[index].FileIdentificationCode)
		fileHeaderSlice = append(fileHeaderSlice, []byte(inputSide.SideFiles[index].FileName)...)
		fileAddressBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(fileAddressBytes, inputSide.SideFiles[index].FileAddress)
		fileHeaderSlice = append(fileHeaderSlice, fileAddressBytes...)
		fileSizeBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(fileSizeBytes, inputSide.SideFiles[index].FileSize)
		fileHeaderSlice = append(fileHeaderSlice, fileSizeBytes...)
		fileHeaderSlice = append(fileHeaderSlice, inputSide.SideFiles[index].FileType)

		if writeChecksums {
			tempCrc16 := inputSide.SideFiles[index].FileMetadataCRC
			if generateChecksums {
				fileHeaderBytes := make([]byte, len(fileHeaderSlice))
				copy(fileHeaderBytes, fileHeaderSlice)
				fileHeaderBytes = append(fileHeaderBytes, []byte{'\x00', '\x00'}...)
				tempCrc16Generated, err := GenerateFDSBlockCRC(fileHeaderBytes)
				if err != nil {
					return nil, &FDSError{Text: "Unable to generate disk info CRC."}
				}

				tempCrc16 = tempCrc16Generated
			}

			crcBytes := make([]byte, 2)
			binary.LittleEndian.PutUint16(crcBytes, tempCrc16)

			fileHeaderSlice = append(fileHeaderSlice, crcBytes...)
		}

		sideSlice = append(sideSlice, fileHeaderSlice...)

		fileDataSlice := make([]byte, 0)
		fileDataSlice = append(fileDataSlice, byte(FDS_FILE_DATA_BLOCK))
		fileDataSlice = append(fileDataSlice, inputSide.SideFiles[index].FileData.FileData...)

		if writeChecksums {
			tempCrc16 := inputSide.SideFiles[index].FileData.FileDataCRC
			if generateChecksums {
				fileDataBytes := make([]byte, len(fileDataSlice))
				copy(fileDataBytes, fileDataSlice)
				fileDataBytes = append(fileDataBytes, []byte{'\x00', '\x00'}...)
				tempCrc16Generated, err := GenerateFDSBlockCRC(fileDataBytes)
				if err != nil {
					return nil, &FDSError{Text: "Unable to generate disk info CRC."}
				}

				tempCrc16 = tempCrc16Generated
			}

			crcBytes := make([]byte, 2)
			binary.LittleEndian.PutUint16(crcBytes, tempCrc16)

			fileDataSlice = append(fileDataSlice, crcBytes...)
		}

		sideSlice = append(sideSlice, fileDataSlice...)
	}

	// Fill in any unallocated space.  Sometimes games have important
	// data in here.  Regardless, the side needs to be 65,500 bytes for
	// an FDS disk, or 65,536 bytes for QD disk
	unallocatedSpaceBytes := inputSide.UnallocatedSpace
	tempSideSliceLength := uint16(len(sideSlice))

	if tempSideSliceLength < inputSide.UnallocatedSpaceOffset {
		fileGapSize := inputSide.UnallocatedSpaceOffset - tempSideSliceLength
		fileGapBytes := make([]byte, fileGapSize)
		for fileGapIndex := 0; fileGapIndex < len(fileGapBytes); fileGapIndex++ {
			fileGapBytes[fileGapIndex] = 0x00
		}

		sideSlice = append(sideSlice, fileGapBytes...)
	} else if tempSideSliceLength > inputSide.UnallocatedSpaceOffset {
		if tempSideSliceLength < (inputSide.UnallocatedSpaceOffset + uint16(len(inputSide.UnallocatedSpace))) {
			unallocatedSpaceBytes = unallocatedSpaceBytes[tempSideSliceLength-inputSide.UnallocatedSpaceOffset:]
		}
	}

	sideSlice = append(sideSlice, unallocatedSpaceBytes...)
	sideSliceLength := uint64(len(sideSlice))

	if writeQd {
		if sideSliceLength < QD_SIDE_SIZE {
			fillBytes := make([]byte, QD_SIDE_SIZE-sideSliceLength)
			for fillIndex := 0; fillIndex < len(fillBytes); fillIndex++ {
				fillBytes[fillIndex] = '\x00'
			}

			sideSlice = append(sideSlice, fillBytes...)
		} else if sideSliceLength > QD_SIDE_SIZE {
			sideSlice = sideSlice[:QD_SIDE_SIZE]
		}
	} else {
		if sideSliceLength < FDS_SIDE_SIZE {
			fillBytes := make([]byte, FDS_SIDE_SIZE-sideSliceLength)
			for fillIndex := 0; fillIndex < len(fillBytes); fillIndex++ {
				fillBytes[fillIndex] = '\x00'
			}

			sideSlice = append(sideSlice, fillBytes...)
		} else if sideSliceLength > FDS_SIDE_SIZE {
			sideSlice = sideSlice[:FDS_SIDE_SIZE]
		}
	}

	return sideSlice, nil
}

// Generate a CRC for given block of data.  Few, if any,
// FDS implementations actually use these.
func GenerateFDSBlockCRC(rawBlock []byte) (uint16, error) {
	var workingCrc uint16 = 0x8000
	dataSize := len(rawBlock)

	if dataSize < 3 {
		return 0, &FDSError{Text: "Data too small to be a valid FDS block."}
	}

	workingBlock := make([]byte, dataSize)

	copy(workingBlock, rawBlock)

	workingBlock[dataSize-2] = 0
	workingBlock[dataSize-1] = 0

	for index := 0; index < dataSize; index++ {
		var tempByte byte = 0
		if index < dataSize-2 {
			tempByte = workingBlock[index]
		}

		for bitIndex := 0; bitIndex < 8; bitIndex++ {
			var workingBit uint16 = (uint16(tempByte) >> bitIndex) & 0b0000000000000001
			var carryBit = workingCrc & 0b0000000000000001
			workingCrc = (workingCrc >> 1) | (workingBit << 15)

			if carryBit == 0b0000000000000001 {
				workingCrc = workingCrc ^ 0x8408
			}
		}
	}

	return workingCrc, nil
}

// Get a slice of byte slices containing each of the disk sides.  If there's
// an FDS file header, strip it out.
func GetStrippedDiskSideByteSlices(inputFile []byte) ([][]byte, error) {
	isQd := false

	fileSize := len(inputFile)

	if fileSize < 2 {
		return nil, &FDSError{Text: "File too small to be a valid FDS archive."}
	}

	if bytes.Compare(inputFile[0:4], []byte(FDS_HEADER_MAGIC)) == 0 {
		if (uint64(fileSize-16) % FDS_SIDE_SIZE) != 0 {
			if (uint64(fileSize-16) % QD_SIDE_SIZE) != 0 {
				return nil, &FDSError{Text: "File is not a valid FDS or QD archive.  " + strconv.Itoa(fileSize-16) + " should be divisible by " + strconv.FormatUint(FDS_SIDE_SIZE, 10) + " for an FDS archive or " + strconv.FormatUint(QD_SIDE_SIZE, 10) + " for a QD archive."}
			} else {
				isQd = true
			}
		}
		return getDiskSideByteSlices(inputFile[16:fileSize], isQd)
	} else {
		if (uint64(fileSize) % FDS_SIDE_SIZE) != 0 {
			if (uint64(fileSize) % QD_SIDE_SIZE) != 0 {
				return nil, &FDSError{Text: "File is not a valid FDS or QD archive.  " + strconv.Itoa(fileSize) + " should be divisible by " + strconv.FormatUint(FDS_SIDE_SIZE, 10) + " for an FDS archive or " + strconv.FormatUint(QD_SIDE_SIZE, 10) + " for a QD archive."}
			} else {
				isQd = true
			}
		}
		return getDiskSideByteSlices(inputFile, isQd)
	}
}

// Get a slice of byte slices of disk sides
func getDiskSideByteSlices(inputFile []byte, isQd bool) ([][]byte, error) {
	sideSize := FDS_SIDE_SIZE
	if isQd {
		sideSize = QD_SIDE_SIZE
	}

	archiveSize := uint64(len(inputFile))
	archiveSides := archiveSize / sideSize

	sideByteSlices := make([][]byte, 0)

	for index := 0; uint64(index) < archiveSides; index++ {
		byteOffset := uint64(index) * sideSize
		if bytes.Compare(inputFile[byteOffset+1:byteOffset+15], []byte(FDS_MAGIC)) != 0 {
			return nil, &FDSError{Text: "Unable to identify side " + strconv.Itoa(index) + " as an FDS or QD disk side.  File is not a valid FDS or QD archive."}
		}
		sideByteSlices = append(sideByteSlices, inputFile[byteOffset:byteOffset+sideSize])
	}

	return sideByteSlices, nil
}

// Decode the FDS date format.  It's BCD-encoded, with an epoch set to
// the beginning of the Showa period.
func DecodeFDSDateFormat(dateBytes []byte) time.Time {
	year := decodeBcdByte(dateBytes[0])
	if year < 83 {
		year = year + uint16(FDS_EPOCH)
	} else {
		year = year + uint16(FDS_EPOCH) - 25
	}

	month := decodeBcdByte(dateBytes[1])
	day := decodeBcdByte(dateBytes[2])

	return time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
}

// Encode dates to the FDS date format
func EncodeFDSDateFormat(dateTime time.Time) ([]byte, error) {
	year, err := encodeBcdByte(uint8(dateTime.Year() - FDS_EPOCH))
	if err != nil {
		return nil, err
	}

	monthInt := uint8(dateTime.Month())
	month, err := encodeBcdByte(monthInt)
	if err != nil {
		return nil, err
	}

	day, err := encodeBcdByte(uint8(dateTime.Day()))
	if err != nil {
		return nil, err
	}

	return []byte{year, month, day}, nil
}

// Turns a BCD byte into an unsigned 16-bit integer
func decodeBcdByte(bcdByte byte) uint16 {
	leastSignificantDigit := bcdByte & 0x0f
	mostSignificantDigit := (bcdByte & 0xf0) >> 4
	return uint16((10 * mostSignificantDigit) + leastSignificantDigit)
}

// Encodes values 0 - 99 to a BCD byte
func encodeBcdByte(bcdInt uint8) (byte, error) {
	if bcdInt > 99 {
		return 0, &NESTool.NESROMError{Text: strconv.Itoa(int(bcdInt)) + " is too large to be packed into a BCD byte."}
	}
	leastSignificantNibble := bcdInt % 10
	mostSignificantNibble := bcdInt / 10
	return (mostSignificantNibble << 4) | leastSignificantNibble, nil
}
