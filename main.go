package main

import "os"

func main() {

}

// Delta represents the changes made to the original file
type Delta struct {
	// Reused original chunks which can be reused
	Reused []Chunk
	// Changed chunks which will be used for patching
	Changed []Chunk
}

// Chunk represents a portion of the file
type Chunk struct {
	// Offset point to chunk position in the file
	Offset int64
	// Length define how long chunk is
	Length int64
	// Hash strong hash for the chunk
	Hash []byte
	// Data chunk data
	Data []byte
}

func FileDiff(original, updated *os.File, chunkSize int64, windowSize int64) (*Delta, error) {

	return nil, nil
}
