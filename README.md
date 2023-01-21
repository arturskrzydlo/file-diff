# file-diff


This is an implementation of [take home exercise](https://github.com/eqlabs/recruitment-exercises/blob/8e49a7b8cf9c415466876e852fbd862f74105ec6/rolling-hash.md)

This might be treated as library so there is no main function, it can be used in some other code. There were no clear requirements on that,
so I've focused on writing just a package which could be reused in some other place of the code.

Function for file chunking uses [Cyclic polynomial rolling hash algorithm](https://en.wikipedia.org/wiki/Rolling_hash), also known as BuzHash


### Usage

```go
file, err := os.Open("original_file")
if err != nil {
    return nil, fmt.Errorf("failed to open a file: %w", err)
}
defer file.Close()

updatedFile, err = os.Open("updated_file")
if err != nil {
    return nil, fmt.Errorf("failed to open a file: %w", err)
}
defer updatedFile.Close()

chunkSize := uint64(1024)
delta, err := filediff.FileDiff(originalFile, updatedFile, chunkSize)
```
This will produce delta with chunks which could be reused and changed ones

```go
// Delta represents the changes made to the original file
type Delta struct {
// Reused original chunks which can be reused. It excludes removed chunks
Reused []Chunk
// Changed chunks which has been modified or added. Chunks which needs to be sync with original file chunks
Changed []Chunk
}
```

### Running tests

It will run set of tests - intention was to have only blackbox tests, but verifying number of chunks violate this. This is due to
the fact chunks are dynamically sized and depends on few factors like chunk size and chosen hash algorithms (so implementation details).
Only real blackbox test would be when we would have complementary apply patch function, but it was out of the scope of this excercise

```shell
make tests
```

### Running benchmarks

Runs one benchmarks which shows how chunk size param to the function affects performance. With lower chunk size we are going to have more
chunks, and it means that apply patch function would need to operate on smaller number of data (and probably it will consume less transfer when doing it
over the network). On the other hand it can take significant amount of time more to produce that chunks. 

```shell
make benchmarks
```
### Possible improvements

* **Benchmarking** - firstly add more benchmarks which will allow to find well suited chunks size for different file sizes.
* **Automatic chunk size** - with such result we would be able to automatically adjust chunk size to give the best performance/chunk size. We could add it as an option to file diff function.
* **Performance** - few things could be done here to improve performance. Firstly verify if we are not doing any costly operation. Bit shifting of BuzHash algorithm should do the job anyway. We can also think about other rollin hash algorithms to choose most performant one.
Secondly we can verify if some parallelization could be done to hash calculation. At the moment I don't see such option, but I haven't analyzed it deeply. (we need to have hash from previous window as input to next hash calculation so parallelization might be not an option)
* **Storing result && command line solution** - depends on context of usage it might be useful to store delta on disk and also create cmd line from this filediff function. Also we could create cmd line which takes file names as params 
(then some modification would be needed to firstly read those files before passing them to FileDiff)

### Remarks

One requirement was not clear for me, and it might cause that this solution is aligned perfectly with requirements. `Only the exact differing locations should be added to the delta`. Because of nature of dynamically sized chunks and common file chunking implementations
I thought that it means locations of a chunks, not locations of a change. If it was location of a change some further processing would need to be applied - when detected changed chunks we would need to perform comparison between changed chunks (according to their offsets) with original file.