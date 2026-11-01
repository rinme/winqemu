package unattend

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	FloppySectorSize = 512
	FloppyTotalSectors = 2880
	FloppyImageSize = FloppySectorSize * FloppyTotalSectors // 1,474,560 bytes
	ReservedSectors = 1
	NumFATs = 2
	SectorsPerFAT = 9
	RootDirectoryEntries = 224
	BytesPerDirEntry = 32
	RootDirSectors = (RootDirectoryEntries * BytesPerDirEntry) / FloppySectorSize // 14
	FirstDataSector = ReservedSectors + (NumFATs * SectorsPerFAT) + RootDirSectors // 33
)

// BuildFAT12Image creates a 1.44MB raw floppy disk image in memory containing a single file.
func BuildFAT12Image(filename string, data []byte) ([]byte, error) {
	img := make([]byte, FloppyImageSize)

	// Sector 0: BPB (BIOS Parameter Block)
	img[0] = 0xEB // JMP short
	img[1] = 0x3C
	img[2] = 0x90 // NOP
	copy(img[3:11], []byte("MSWIN4.1"))
	binary.LittleEndian.PutUint16(img[11:13], FloppySectorSize) // Bytes per sector (512)
	img[13] = 1                                                // Sectors per cluster
	binary.LittleEndian.PutUint16(img[14:16], ReservedSectors)  // Reserved sectors (1)
	img[16] = NumFATs                                          // Number of FATs (2)
	binary.LittleEndian.PutUint16(img[17:19], RootDirectoryEntries) // Max root dir entries (224)
	binary.LittleEndian.PutUint16(img[19:21], FloppyTotalSectors)   // Total sectors (2880)
	img[21] = 0xF0                                             // Media descriptor (3.5" 1.44MB)
	binary.LittleEndian.PutUint16(img[22:24], SectorsPerFAT)    // Sectors per FAT (9)
	binary.LittleEndian.PutUint16(img[24:26], 18)               // Sectors per track
	binary.LittleEndian.PutUint16(img[26:28], 2)                // Number of heads
	binary.LittleEndian.PutUint32(img[28:32], 0)                // Hidden sectors
	binary.LittleEndian.PutUint32(img[32:36], 0)                // Large total sectors

	// Extended BPB
	img[36] = 0x00               // Drive number
	img[37] = 0x00               // Reserved
	img[38] = 0x29               // Extended boot signature
	binary.LittleEndian.PutUint32(img[39:43], 0x12345678) // Volume ID
	copy(img[43:54], []byte("WINQEMU    ")) // Volume label (11 bytes)
	copy(img[54:62], []byte("FAT12   "))    // File system type (8 bytes)

	// Boot signature
	img[510] = 0x55
	img[511] = 0xAA

	// FAT 1 & 2 initialization
	fat1Offset := ReservedSectors * FloppySectorSize
	fat2Offset := (ReservedSectors + SectorsPerFAT) * FloppySectorSize

	// First two FAT12 entries are 0xFF0 and 0xFFF
	img[fat1Offset] = 0xF0
	img[fat1Offset+1] = 0xFF
	img[fat1Offset+2] = 0xFF

	img[fat2Offset] = 0xF0
	img[fat2Offset+1] = 0xFF
	img[fat2Offset+2] = 0xFF

	// Calculate required clusters
	numClusters := (len(data) + FloppySectorSize - 1) / FloppySectorSize
	if numClusters == 0 {
		numClusters = 1
	}

	// Chain clusters in FAT
	for c := 0; c < numClusters; c++ {
		clusterNum := 2 + c
		var nextVal uint16
		if c == numClusters-1 {
			nextVal = 0x0FFF // End of cluster chain
		} else {
			nextVal = uint16(clusterNum + 1)
		}

		// Set FAT12 entry
		setFAT12Entry(img[fat1Offset:fat1Offset+SectorsPerFAT*FloppySectorSize], clusterNum, nextVal)
		setFAT12Entry(img[fat2Offset:fat2Offset+SectorsPerFAT*FloppySectorSize], clusterNum, nextVal)
	}

	// Copy file data to data area (cluster 2 starts at FirstDataSector)
	dataStart := FirstDataSector * FloppySectorSize
	copy(img[dataStart:dataStart+len(data)], data)

	// Root Directory Entry
	rootDirStart := (ReservedSectors + NumFATs*SectorsPerFAT) * FloppySectorSize
	dirEntry := img[rootDirStart : rootDirStart+BytesPerDirEntry]

	// 8.3 filename
	base, ext := split83(filename)
	copy(dirEntry[0:8], []byte(fmt.Sprintf("%-8s", base)))
	copy(dirEntry[8:11], []byte(fmt.Sprintf("%-3s", ext)))
	dirEntry[11] = 0x20 // Archive attribute

	now := time.Now()
	// Date & Time in DOS format
	dosTime := uint16(now.Hour()<<11 | now.Minute()<<5 | now.Second()/2)
	dosDate := uint16((now.Year()-1980)<<9 | int(now.Month())<<5 | now.Day())
	binary.LittleEndian.PutUint16(dirEntry[22:24], dosTime)
	binary.LittleEndian.PutUint16(dirEntry[24:26], dosDate)

	binary.LittleEndian.PutUint16(dirEntry[26:28], 2)                     // Starting cluster
	binary.LittleEndian.PutUint32(dirEntry[28:32], uint32(len(data)))      // File size in bytes

	return img, nil
}

func setFAT12Entry(fat []byte, cluster int, value uint16) {
	byteOffset := (cluster * 3) / 2
	if cluster%2 == 0 {
		fat[byteOffset] = byte(value & 0xFF)
		fat[byteOffset+1] = (fat[byteOffset+1] & 0xF0) | byte((value>>8)&0x0F)
	} else {
		fat[byteOffset] = (fat[byteOffset] & 0x0F) | byte((value<<4)&0xF0)
		fat[byteOffset+1] = byte((value >> 4) & 0xFF)
	}
}

func split83(filename string) (string, string) {
	parts := strings.Split(filename, ".")
	name := strings.ToUpper(parts[0])
	ext := ""
	if len(parts) > 1 {
		ext = strings.ToUpper(parts[1])
	}
	if len(name) > 8 {
		name = name[:8]
	}
	if len(ext) > 3 {
		ext = ext[:3]
	}
	return name, ext
}

// WriteUnattendFiles writes both the plain autounattend.xml and a .vfd / .img disk image into targetDir.
func WriteUnattendFiles(targetDir, xmlContent string) (xmlPath, diskPath string, err error) {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create directory: %w", err)
	}

	xmlPath = filepath.Join(targetDir, "autounattend.xml")
	if err := os.WriteFile(xmlPath, []byte(xmlContent), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write XML: %w", err)
	}

	imgBytes, err := BuildFAT12Image("autounattend.xml", []byte(xmlContent))
	if err != nil {
		return "", "", fmt.Errorf("failed to build floppy image: %w", err)
	}

	diskPath = filepath.Join(targetDir, "unattend.vfd")
	if err := os.WriteFile(diskPath, imgBytes, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write disk image: %w", err)
	}

	return xmlPath, diskPath, nil
}
