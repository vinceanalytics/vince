//  Copyright (c) 2020 The Bluge Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ice

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/blevesearch/mmap-go"
	"github.com/vinceanalytics/vince/segment"
)

// various helpers to test with files, even though
// ice no longer knows about files itself

func setupTestDir(t *testing.T) (path string, cleanup func()) {
	path, err := ioutil.TempDir("", "ice-test")
	if err != nil {
		t.Fatalf("error creating tmp dir: %v", err)
	}
	return path, func() {
		err2 := os.RemoveAll(path)
		if err2 != nil {
			t.Fatalf("error removing '%s': %v", path, err2)
		}
	}
}

type closeFunc func() error

var noCloseFunc = func() error {
	return nil
}

func encodeNorm(_ string, numTerms int) float32 {
	return math.Float32frombits(uint32(numTerms))
}

func decodeNorm(_ string, encodeNorm float32) int {
	return int(math.Float32bits(encodeNorm))
}

type segmentBuilder func() (*Segment, error)

func createDiskSegment(builder segmentBuilder, path string) (*Segment, closeFunc, error) {
	memSeg, err := builder()
	if err != nil {
		return nil, nil, err
	}

	err = persistToFile(memSeg, path)
	if err != nil {
		return nil, nil, err
	}
	seg, closeF, err := openFromFile(path)
	if err != nil {
		return nil, nil, err
	}
	return seg, closeF, nil
}

func openFromFile(path string) (*Segment, closeFunc, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, noCloseFunc, err
	}
	mm, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		// mmap failed, try to close the file
		_ = f.Close()
		return nil, noCloseFunc, err
	}

	closeFunc := func() error {
		err2 := mm.Unmap()
		// try to close file even if unmap failed
		err3 := f.Close()
		if err2 == nil {
			// try to return first error
			err2 = err3
		}
		return err2
	}

	data := segment.NewDataBytes(mm)

	seg, err := load(data)
	if err != nil {
		_ = closeFunc()
		return nil, noCloseFunc, fmt.Errorf("error loading segment: %v", err)
	}

	return seg, closeFunc, nil
}

func persistToFile(sb *Segment, path string) error {
	flag := os.O_RDWR | os.O_CREATE

	f, err := os.OpenFile(path, flag, 0600)
	if err != nil {
		return err
	}

	cleanup := func() {
		_ = f.Close()
		_ = os.Remove(path)
	}

	br := bufio.NewWriter(f)
	_, err = sb.WriteTo(br, nil)
	if err != nil {
		cleanup()
		return err
	}

	err = br.Flush()
	if err != nil {
		cleanup()
		return err
	}

	err = f.Sync()
	if err != nil {
		cleanup()
		return err
	}

	err = f.Close()
	if err != nil {
		cleanup()
		return err
	}

	return nil
}

var DefaultFileMergerBufferSize = 1024 * 1024

func mergeSegments(segments []segment.Segment, drops []*roaring.Bitmap, path string) (uint64, error) {
	flag := os.O_RDWR | os.O_CREATE

	f, err := os.OpenFile(path, flag, 0600)
	if err != nil {
		return 0, err
	}

	cleanup := func() {
		_ = f.Close()
		_ = os.Remove(path)
	}

	// buffer the output
	br := bufio.NewWriterSize(f, DefaultFileMergerBufferSize)
	_, count, err := merge(segments, drops, br, nil)
	if err != nil {
		cleanup()
		return 0, err
	}

	err = br.Flush()
	if err != nil {
		cleanup()
		return 0, err
	}

	err = f.Sync()
	if err != nil {
		cleanup()
		return 0, err
	}

	err = f.Close()
	if err != nil {
		cleanup()
		return 0, err
	}

	return count, nil
}
