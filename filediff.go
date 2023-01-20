package filediff

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"file-diff/hash"
)

// Delta represents the changes made to the original file
type Delta struct {
	// Reused original chunks which can be reused. It excludes removed chunks
	Reused []Chunk
	// Changed chunks which has been modified or added. Chunks which needs to be sync with original file chunks
	Changed []Chunk
}

// Chunk represents a portion of the file
type Chunk struct {
	// Offset point to starting chunk position in the file
	Offset int
	// Length define how long chunk is
	Length int
	// Hash strong hash for the chunk
	Hash string
	// Data chunk data
	Data []byte
}

type signature map[string]Chunk

func FileDiff(original, updated *os.File, chunkSize uint64) (*Delta, error) {
	if !isPowerOfTwo(chunkSize) {
		return nil, errors.New("chunkSize parameter must be a power of two")
	}
	// Get the file size
	fileInfo, _ := original.Stat()
	fileSize := fileInfo.Size()

	data := make([]byte, fileSize)

	// Read a chunk from the original file // TODO: it returns int so it depends on os. Does it mean any constraint how big file can be read ?
	_, err := io.ReadFull(original, data)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	originalFileSignature := createSignature(data, chunkSize)

	fileInfo, _ = updated.Stat()
	fileSize = fileInfo.Size()
	updatedData := make([]byte, fileSize)

	_, err = io.ReadFull(updated, updatedData)
	if err != nil {
		return nil, err
	}

	delta := getDelta(originalFileSignature, updatedData, chunkSize)
	endTime := time.Now()
	fmt.Printf("Finished in %d", (endTime.Sub(startTime)).Milliseconds())

	return delta, nil

}

func getDelta(originalFileSignature signature, updatedFileData []byte, chunkSize uint64) *Delta {
	updatedFileSignature := createSignature(updatedFileData, chunkSize)
	reusedFileChunks := make([]Chunk, 0)
	changedFileChunks := make([]Chunk, 0)

	for sigHash, updatedFileChunk := range updatedFileSignature {
		if chunk, ok := originalFileSignature[sigHash]; ok {
			reusedFileChunks = append(reusedFileChunks, chunk)
			continue
		}

		changedFileChunks = append(changedFileChunks, updatedFileChunk)
	}

	return &Delta{
		Reused:  reusedFileChunks,
		Changed: changedFileChunks,
	}
}

func createSignature(data []byte, chunkSize uint64) signature {
	signatureChunks := make(signature)
	buzHash := hash.NewBuzHash()
	buzHash.ResetHash(data, hash.WindowSize)

	previousSplitPosition := 0
	newBytePosition := 0
	for i := 0; i < len(data); i++ {
		// if this will be potential last chunk, end position
		// is just end of a data slice
		if i+hash.WindowSize >= len(data) {
			newBytePosition = len(data) - 1
		} else {
			newBytePosition = i + hash.WindowSize
		}
		currentHash := buzHash.RollingHash(data[i], data[newBytePosition])
		if shouldSplit(currentHash, chunkSize) {
			addNewChunkToSignature(data[previousSplitPosition:i], previousSplitPosition, signatureChunks)
			previousSplitPosition = i
		}

		// if last element
		if i == len(data)-1 {
			addNewChunkToSignature(data[previousSplitPosition:], previousSplitPosition, signatureChunks)
		}
	}

	return signatureChunks
}

func addNewChunkToSignature(chunkData []byte, offset int, fileSig signature) {
	strongHash := sha256.Sum256(chunkData)
	strongHashString := hex.EncodeToString(strongHash[:])
	if _, ok := fileSig[strongHashString]; !ok {
		fileSig[strongHashString] = Chunk{
			Offset: offset,
			Length: len(chunkData),
			Data:   chunkData,
			Hash:   hex.EncodeToString(strongHash[:]),
		}
	}
}

func shouldSplit(hash int, chunkSize uint64) bool {
	bitShift := math.Log2(float64(chunkSize))
	mask := (1 << int(bitShift)) - 1
	return (hash & mask) == 0
}

func isPowerOfTwo(x uint64) bool {
	return x > 0 && (x&(x-1)) == 0
}
