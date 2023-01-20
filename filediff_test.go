package filediff

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The best test would be to have a tool to sync and apply patch on original file.
// Because it was out of the scope these test are focusing on detecting changes and removals, but it might be a bit brittle
// if we change hashing algo or window size. It's not perfect because though because it depends on some internal changes.
// We have variable chunk sizes, so it depends fully on those params
//
// I've added function which resurrects updated file from delta. It was out of scope, so I did it really simple
// to show option which could be tested. With such function it would be much easier to test, and we wouldn't need to test number of chunks which fully depends
// on  FileDiff internals (rolling hash algorithm and window size)
func TestFileDiff(t *testing.T) {

	assert := assert.New(t)

	testCases := map[string]struct {
		originalFile  []byte
		updatedFile   []byte
		changedChunks int
		reusedChunks  int
	}{
		"should be able to detect no change": {
			originalFile: []byte("Hello everyone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay"),
			// change everyone -> evbryone,
			updatedFile:   []byte("Hello everyone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay"),
			changedChunks: 0,
			reusedChunks:  2,
		},
		"should be able to detect one change in file which is chunked to two chunks": {
			originalFile: []byte("Hello everyone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay"),
			// change everyone -> evbryone,
			updatedFile:   []byte("Hello evbryone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay"),
			changedChunks: 1,
			reusedChunks:  1,
		},
		"should be able to detect two changes in file which is chunked to two chunks": {
			originalFile: []byte("Hello everyone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay"),
			// change everyone -> evbryone, and is -> IS
			updatedFile:   []byte("Hello evbryone, this will be a very short text about nothing. Its only purpose IS for testing. Testing should be sufficient. Yay"),
			changedChunks: 2,
			reusedChunks:  0,
		},
		"should be able to detect chunk removal at the end of the file": {
			originalFile:  []byte("Hello everyone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay"),
			updatedFile:   []byte("Hello everyone, this will be a very short text about nothing. It"),
			changedChunks: 1,
			reusedChunks:  0,
		},
		"should be able to detect chunk removal at the beginning of the file": {
			originalFile:  []byte("Hello everyone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay"),
			updatedFile:   []byte("s only purpose is for testing. Testing should be sufficient. Yay"),
			changedChunks: 1,
			reusedChunks:  0,
		},
		"should be able to detect chunk switch changes": {
			originalFile:  []byte("Hello everyone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay"),
			updatedFile:   []byte("s only purpose is for testing. Testing should be sufficient. YayHello everyone, this will be a very short text about nothing. It"),
			changedChunks: 2,
			reusedChunks:  0,
		},
		"should be able to detect additions at the end of the file": {
			originalFile:  []byte("Hello everyone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay"),
			updatedFile:   []byte("Hello everyone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay it's really exciting"),
			changedChunks: 2,
			reusedChunks:  1,
		},
		"should be able to detect additions at the beginning of the file": {
			originalFile:  []byte("Hello everyone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay"),
			updatedFile:   []byte("It's really exciting. Hello everyone, this will be a very short text about nothing. Its only purpose is for testing. Testing should be sufficient. Yay"),
			changedChunks: 1,
			reusedChunks:  1,
		},
		// sliding window for which changes are detected is set to 64. Let's generate longer text to have many chunks
		"should detect two changes, one addition at the end and one removal in long text": {
			originalFile: []byte("Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient" +
				" montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel," +
				" aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt." +
				" Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem" +
				" ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue." +
				" Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque" +
				" sed ipsum. Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus. Donec vitae sapien ut libero venenatis faucibus." +
				" Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt. Duis leo. Sed fringilla mauris sit amet nibh. Donec sodales sagittis magna. Sed consequat, leo eget " +
				" bibendum sodales, augue velit cursus nunc, quis gravida magna mi a libero. Fusce vulputate eleifend sapien. Vestibulum purus quam, scelerisque ut, mollis sed, nonummy id," +
				" metus. Nullam accumsan lorem in dui. Cras ultricies mi eu turpis hendrerit fringilla. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae;" +
				" In ac dui quis mi consectetuer lacinia. Nam pretium turpis et arcu. Duis arcu tortor, suscipit eget, imperdiet nec, imperdiet iaculis, ipsum. Sed aliquam ultrices mauris." +
				" Integer ante arcu, accumsan a, consectetuer eget, posuere ut, mauris. Praesent adipiscing. Phasellus ullamcorper ipsum rutrum nunc. Nunc nonummy metus. Vestibulum volutpat" +
				" pretium libero. Cras id dui. Aenean ut eros et nisl sagittis vestibulum. Nullam nulla eros, ultricies sit amet, nonummy id, imperdiet feugiat, pede. Sed lectus. Donec mollis" +
				" hendrerit risus. Phasellus nec sem in justo pellentesque facilisis. Etiam imperdiet imperdiet orci. Nunc nec neque. Phasellus leo dolor, tempus non, auctor et, hendrerit quis, nisi." +
				" Curabitur ligula sapien, tincidunt non, euismod vitae, posuere imperdiet, leo. Maecenas malesuada. Praesent congue erat at massa. Sed cursus turpis vitae tortor. Donec posuere" +
				" vulputate arcu. Phasellus accumsan cursus velit. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Sed aliquam, nisi quis porttitor congue," +
				" elit erat euismod orci, ac placerat dolor lectus quis orci. Phasellus consectetuer vestibulum elit. Aenean tellus metus, bibendum sed, posuere ac, mattis non, nunc. Vestibulum" +
				" fringilla pede sit amet augue. In turpis. Pellentesque posuere. Praesent turpis. Aenean posuere, tortor sed cursus feugiat, nunc augue blandit nunc, eu sollicitudin urna dolor sagittis lacus. " +
				" Donec elit libero, sodales nec, volutpat a, suscipit non, turpis. Nullam sagittis. Suspendisse pulvinar, augue ac venenatis condimentum, sem libero volutpat nibh, nec pellentesque velit" +
				" pede quis nunc. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Fusce id purus. Ut varius tincidunt libero. Phasellus dolor. Maecenas vestibulum mollis" +
				" diam. Pellentesque ut neque. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. In dui magna, posuere eget, vestibulum et, tempor auctor, justo. In ac felis quis tortor malesuada pretium." +
				" Pellentesque auctor neque nec urna. Proin sapien ipsum, porta a, auctor quis, euismod ut, mi. Aenean viverra rhoncus pede. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas." +
				" Ut non enim eleifend felis pretium feugiat. Vivamus quis mi. Phasellus a est. Phasellus magna. In hac habitasse platea dictumst. Curabitur at lacus ac velit ornare lobortis. Cura"),
			// changes are named (to ease searching) - FIRST, SECOND, ADDITION, removed 'ultrices' from 'et ultrices posuere'
			updatedFile: []byte("Lorem ipsum FIRST sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient" +
				" montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel," +
				" aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt." +
				" Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem" +
				" ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue." +
				" Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque" +
				" sed ipsum. Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus. Donec vitae sapien ut libero venenatis faucibus." +
				" Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt. Duis leo. Sed fringilla mauris sit amet nibh. Donec sodales sagittis magna. Sed consequat, leo eget " +
				" bibendum sodales, augue velit cursus nunc, quis gravida magna mi a libero. Fusce vulputate eleifend sapien. Vestibulum purus quam, scelerisque ut, mollis sed, nonummy id," +
				" metus. Nullam accumsan lorem in dui. Cras SECONDies mi eu turpis hendrerit fringilla. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae;" +
				" In ac dui quis mi consectetuer lacinia. Nam pretium turpis et arcu. Duis arcu tortor, suscipit eget, imperdiet nec, imperdiet iaculis, ipsum. Sed aliquam ultrices mauris." +
				" Integer ante arcu, accumsan a, consectetuer eget, posuere ut, mauris. Praesent adipiscing. Phasellus ullamcorper ipsum rutrum nunc. Nunc nonummy metus. Vestibulum volutpat" +
				" pretium libero. Cras id dui. Aenean ut eros et nisl sagittis vestibulum. Nullam nulla eros, ultricies sit amet, nonummy id, imperdiet feugiat, pede. Sed lectus. Donec mollis" +
				" hendrerit risus. Phasellus nec sem in justo pellentesque facilisis. Etiam imperdiet imperdiet orci. Nunc nec neque. Phasellus leo dolor, tempus non, auctor et, hendrerit quis, nisi." +
				" Curabitur ligula sapien, tincidunt non, euismod vitae, posuere imperdiet, leo. Maecenas malesuada. Praesent congue erat at massa. Sed cursus turpis vitae tortor. Donec posuere" +
				" vulputate arcu. Phasellus accumsan cursus velit. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Sed aliquam, nisi quis porttitor congue," +
				" elit erat euismod orci, ac placerat dolor lectus quis orci. Phasellus consectetuer vestibulum elit. Aenean tellus metus, bibendum sed, posuere ac, mattis non, nunc. Vestibulum" +
				" fringilla pede sit amet augue. In turpis. Pellentesque posuere. Praesent turpis. Aenean posuere, tortor sed cursus feugiat, nunc augue blandit nunc, eu sollicitudin urna dolor sagittis lacus. " +
				" Donec elit libero, sodales nec, volutpat a, suscipit non, turpis. Nullam sagittis. Suspendisse pulvinar, augue ac venenatis condimentum, sem libero volutpat nibh, nec pellentesque velit" +
				" pede quis nunc. Vestibulum ante ipsum primis in faucibus orci luctus et posuere cubilia Curae; Fusce id purus. Ut varius tincidunt libero. Phasellus dolor. Maecenas vestibulum mollis" +
				" diam. Pellentesque ut neque. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. In dui magna, posuere eget, vestibulum et, tempor auctor, justo. In ac felis quis tortor malesuada pretium." +
				" Pellentesque auctor neque nec urna. Proin sapien ipsum, porta a, auctor quis, euismod ut, mi. Aenean viverra rhoncus pede. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas." +
				" Ut non enim eleifend felis pretium feugiat. Vivamus quis mi. Phasellus a est. Phasellus magna. In hac habitasse platea dictumst. Curabitur at lacus ac velit ornare lobortis. Cura ADDITION"),
			// there are 4 changes - two changes, one removal and one addition. But because addition is at the end of the file
			// it affects chunking (more character and split) and there were two chunks instead of one, so in total 5 chunks.
			// Original file has 45 chunks so four of them were not in updated file.
			changedChunks: 5,
			reusedChunks:  41,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// given
			chunkSize := uint64(64)
			// original file
			originalFile, err := createTempTestFile(tc.originalFile, "original")
			defer originalFile.Close()
			defer os.Remove(originalFile.Name())
			assert.NoError(err)

			// updated file
			updatedFile, err := createTempTestFile(tc.updatedFile, "updated")
			defer updatedFile.Close()
			defer os.Remove(updatedFile.Name())
			assert.NoError(err)

			// when
			delta, err := FileDiff(originalFile, updatedFile, chunkSize)

			//then
			assert.NoError(err)
			assert.NotNil(delta)
			assert.Equal(tc.changedChunks, len(delta.Changed))
			assert.Equal(tc.reusedChunks, len(delta.Reused))
			frankenstein := frankensteinFunc(delta)
			assert.Equal(tc.updatedFile, frankenstein)
		})
	}

	t.Run("should be able to detect chunk removals at the beginning of the file", func(t *testing.T) {
		// given
		original, err := os.Open("go.zip")
		if err != nil {
			return
		}
		defer original.Close()

		updated, err := os.Open("go_2.zip")
		if err != nil {
			return
		}
		defer updated.Close()

		fileInfo, _ := updated.Stat()
		fileSize := fileInfo.Size()

		data := make([]byte, fileSize)

		_, err = io.ReadFull(updated, data)
		assert.NoError(err)

		updated, err = os.Open("go_2.zip")
		if err != nil {
			return
		}

		// when
		delta, err := FileDiff(original, updated, 8388608)
		assert.NoError(err)
		assert.NotNil(delta)

		// figure out how to test it
		assert.True(len(delta.Changed) > 0)
		assert.True(len(delta.Reused) > 0)
		bytes := frankensteinFunc(delta)
		assert.Equal(data, bytes)
	})
}

func createTempTestFile(fileContent []byte, name string) (file *os.File, err error) {
	// Write the original and updated bytes to temporary files
	file, err = os.CreateTemp("", name)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if err != nil {
			removalErr := os.Remove(file.Name())
			if removalErr != nil {
				err = fmt.Errorf("%w, error removing file on error", err)
			}
		}
	}()

	if _, err = file.Write(fileContent); err != nil {
		return nil, fmt.Errorf("failed to write file on disk: %w", err)
	}
	if err = file.Close(); err != nil {
		return nil, fmt.Errorf("failed to close file: %w", err)
	}

	// Open the file
	file, err = os.OpenFile(file.Name(), os.O_RDWR, os.ModeAppend)
	if err != nil {
		return nil, fmt.Errorf("failed to open a file: %w", err)
	}
	defer func() {
		if err != nil {
			closeErr := file.Close()
			if closeErr != nil {
				err = fmt.Errorf("%w, error closing file on error", err)
			}
		}
	}()
	return file, nil
}

// it recreates updated file from pieces (delta)
// this is super simple function, not optimized nor well-designed (definitely not a patching function)
func frankensteinFunc(delta *Delta) []byte {

	combinedChunks := append(delta.Reused, delta.Changed...)
	sort.Slice(combinedChunks, func(i, j int) bool {
		return combinedChunks[i].Offset < combinedChunks[j].Offset
	})

	recreatedFile := make([]byte, 0)
	for _, chunk := range combinedChunks {
		recreatedFile = append(recreatedFile, chunk.Data...)
	}

	return recreatedFile
}

func BenchmarkFileDiff(b *testing.B) {
	// generate completely different files
	original, err := createTempTestFile([]byte("initial content"), "file_diff_benchmark_orig")
	require.NoError(b, err)
	defer original.Close()

	_, err = io.CopyN(original, rand.Reader, 10*1000*1024) // generate 10MB file
	require.NoError(b, err)

	updated, err := createTempTestFile([]byte("initial content"), "file_diff_benchmark_updated")
	require.NoError(b, err)
	defer updated.Close()

	_, err = io.CopyN(updated, rand.Reader, 10*1000*1024) // generate 10MB file
	require.NoError(b, err)

	var table = []struct {
		chunkSize uint64
	}{
		{chunkSize: 64},
		// 1KB
		{chunkSize: 1024},
		// 8 MB
		{chunkSize: 8388608},
		// 128 MB
		{chunkSize: 134217728},
	}

	b.ResetTimer()
	for _, v := range table {
		b.Run(fmt.Sprintf("chunk_size_%d", v.chunkSize), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				FileDiff(original, updated, v.chunkSize)
			}
		})
	}
}
