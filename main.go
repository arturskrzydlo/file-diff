package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

const mask = (1 << 6) - 1

const WINDOW_SIZE = 63
const BH_ROTATE = WINDOW_SIZE % 32
const BH_ROTATE_COMP = 32 - BH_ROTATE

var hashes [256]int
var currentHash int

func main() {

}

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

// TODO: signature could be a map type
type signature map[string]Chunk

type RollingHash interface {
	rollingHash(data []byte, chunkSize uint64) uint64
}

func FileDiff(original, updated *os.File) (*Delta, error) {
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

	originalFileSignature := createSignature(data)

	fileInfo, _ = updated.Stat()
	fileSize = fileInfo.Size()
	updatedData := make([]byte, fileSize)

	_, err = io.ReadFull(updated, updatedData)
	if err != nil {
		return nil, err
	}

	delta := getDelta(originalFileSignature, updatedData)
	endTime := time.Now()
	fmt.Printf("Finished in %d", (endTime.Sub(startTime)).Milliseconds())

	return delta, nil

}

func getDelta(originalFileSignature signature, updatedFileData []byte) *Delta {
	updatedFileSignature := createSignature(updatedFileData)
	reusedFileChunks := make([]Chunk, 0)
	changedFileChunks := make([]Chunk, 0)

	for hash, updatedFileChunk := range updatedFileSignature {
		if chunk, ok := originalFileSignature[hash]; ok {
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

func createSignature(data []byte) signature {
	signatureChunks := make(signature)
	resetHash(data, WINDOW_SIZE)

	previousSplitPosition := 0
	newBytePosition := 0
	for i := 0; i < len(data); i++ {
		if i+WINDOW_SIZE >= len(data) {
			newBytePosition = len(data) - 1
		} else {
			newBytePosition = i + WINDOW_SIZE
		}
		roll(data[i], data[newBytePosition])
		if shouldSplit() {
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
func init() {
	rand.Seed(42)
	for i := 0; i < 256; i++ {
		hashes[i] = rand.Int()
	}
}

func resetHash(input []byte, pos int) {
	currentHash = 0
	if pos > len(input) {
		pos = len(input)
	}
	for i := WINDOW_SIZE; i > 0; i-- {
		currentHash = (currentHash<<1 | currentHash>>31) ^ hashes[input[pos-i]+128]
	}
}

func roll(oldByte byte, newByte byte) {
	oldHash := hashes[oldByte+128]
	currentHash = (currentHash<<1 | currentHash>>31) ^ (oldHash<<BH_ROTATE | oldHash>>BH_ROTATE_COMP) ^ hashes[newByte+128]
}

func shouldSplit() bool {
	return (currentHash & mask) == 0
}
