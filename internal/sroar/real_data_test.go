// +build real

package sroar

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// To run these benchmarks: go test -bench BenchmarkRealDataFastOr -run -

var realDatasets = []string{
	"census-income_srt", "census-income", "census1881_srt", "census1881",
	"dimension_003", "dimension_008", "dimension_033", "uscensus2000", "weather_sept_85_srt",
	"weather_sept_85", "wikileaks-noquotes_srt", "wikileaks-noquotes",
}

func getDataSetPath(dataset string) (string, error) {
	gopath, ok := os.LookupEnv("GOPATH")
	if !ok {
		return "", fmt.Errorf("GOPATH not set. It's required to locate real-roaring-dataset.")
	}

	basePath := path.Join(gopath, "src", "github.com", "RoaringBitmap", "real-roaring-datasets")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return "", fmt.Errorf("real-roaring-datasets does not exist. " +
			"Run `go get github.com/RoaringBitmap/real-roaring-datasets`")
	}

	datasetPath := path.Join(basePath, dataset+".zip")
	if _, err := os.Stat(datasetPath); os.IsNotExist(err) {
		return "", fmt.Errorf("dataset %s does not exist, tried path: %s",
			dataset, datasetPath)
	}
	return datasetPath, nil
}

func retrieveRealDataBitmaps(datasetName string, optimize bool) ([]*Bitmap, error) {
	datasetPath, err := getDataSetPath(datasetName)
	zipFile, err := zip.OpenReader(datasetPath)
	if err != nil {
		return nil, fmt.Errorf("error opening dataset %s zipfile, cause: %v", datasetPath, err)
	}
	defer zipFile.Close()

	bitmaps := make([]*Bitmap, len(zipFile.File))
	for i, f := range zipFile.File {
		res, err := processZipFile(f)
		if err != nil {
			return nil, errors.Wrap(err, "while processing zip file")
		}
		b := NewBitmap()
		for _, v := range res {
			b.Set(v)
		}
		bitmaps[i] = b
	}

	return bitmaps, nil
}

func processZipFile(f *zip.File) ([]uint64, error) {
	r, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to read bitmap file %s, cause: %v",
			f.Name, err)
	}

	buf := make([]byte, f.UncompressedSize)
	var bufStep uint64 = 32768 // apparently the largest buffer zip can read
	var totalReadBytes uint64

	for {
		var endOffset uint64
		if f.UncompressedSize64 < totalReadBytes+bufStep {
			endOffset = f.UncompressedSize64
		} else {
			endOffset = totalReadBytes + bufStep
		}

		readBytes, err := r.Read(buf[totalReadBytes:endOffset])
		totalReadBytes += uint64(readBytes)

		if err == io.EOF {
			r.Close()
			break
		} else if err != nil {
			r.Close()
			return nil, fmt.Errorf("could not read content of file %s , err: %v",
				f.Name, err)
		}
	}

	elemsAsBytes := bytes.Split(buf[:totalReadBytes], []byte{44}) // 44 is a comma

	var result []uint64
	for _, elemBytes := range elemsAsBytes {
		elemStr := strings.TrimSpace(string(elemBytes))

		e, err := strconv.ParseUint(elemStr, 10, 32)
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("could not parse %s as uint32. Reading %s, err: %v",
				elemStr, f.Name, err)
		}
		result = append(result, e)
	}
	return result, nil
}

func benchmarkRealDataAggregate(b *testing.B, aggregator func(b []*Bitmap) int) {
	for _, dataset := range realDatasets {
		once := false
		b.Run(dataset, func(b *testing.B) {
			bitmaps, err := retrieveRealDataBitmaps(dataset, true)
			if err != nil {
				b.Fatal(err)
			}
			if once {
				c := aggregator(bitmaps)
				b.Logf("Dataset: %s Got cardinality: %d\n", dataset, c)
				once = false
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				aggregator(bitmaps)
			}
		})
	}
}

func BenchmarkRealDataFastOr(b *testing.B) {
	benchmarkRealDataAggregate(b, func(bitmaps []*Bitmap) int {
		return FastOr(bitmaps...).GetCardinality()
	})
}
func BenchmarkRealDataFastParOr(b *testing.B) {
	benchmarkRealDataAggregate(b, func(bitmaps []*Bitmap) int {
		return FastParOr(4, bitmaps...).GetCardinality()
	})
}

func BenchmarkRealDataFastAnd(b *testing.B) {
	benchmarkRealDataAggregate(b, func(bitmaps []*Bitmap) int {
		return FastAnd(bitmaps...).GetCardinality()
	})
}

func TestOrRealData(t *testing.T) {
	test := func(t *testing.T, dataset string) {
		path, err := getDataSetPath(dataset)
		require.NoError(t, err)

		zipFile, err := zip.OpenReader(path)
		require.NoError(t, err)
		defer zipFile.Close()

		bitmaps := make([]*Bitmap, len(zipFile.File))
		valMap := make(map[uint64]struct{})

		res2 := NewBitmap()
		// For each file in the zip, create a new bitmap and check the created bitmap has correct
		// cardinality as well as it has all the elements.
		for i, f := range zipFile.File {
			vals, err := processZipFile(f)
			require.NoError(t, err)

			b := NewBitmap()
			for _, v := range vals {
				b.Set(v)
				res2.Set(v)
				valMap[v] = struct{}{}
			}
			require.Equal(t, len(vals), b.GetCardinality())
			for _, v := range vals {
				require.True(t, b.Contains(v))
			}
			bitmaps[i] = b
		}

		// Check that union operation is correct.
		res := FastOr(bitmaps...)

		t.Logf("Result: %s\n", res)
		require.Equal(t, len(valMap), res.GetCardinality())
		require.Equal(t, len(valMap), res2.GetCardinality())

		for k := range valMap {
			require.True(t, res.Contains(k))
			require.True(t, res2.Contains(k))
		}
	}

	for _, dataset := range realDatasets {
		t.Run(dataset, func(t *testing.T) { test(t, dataset) })
	}
}
