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

package segment

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type dataTest struct {
	name       string
	input      []byte
	sliceStart int
	sliceEnd   int
}

func TestData(t *testing.T) {
	testCases := []dataTest{
		{
			name:  "simple",
			input: []byte("simple"),
		},
		{
			name:  "kila",
			input: bytes.Repeat([]byte{0}, 1024),
		},
		{
			name:  "mega",
			input: bytes.Repeat([]byte{'m'}, 1024*1024),
		},
		{
			name:     "simple-sliced",
			input:    []byte("simple"),
			sliceEnd: 4,
		},
		{
			name:       "kila",
			input:      bytes.Repeat([]byte{0}, 1024),
			sliceStart: 4,
			sliceEnd:   1024 - 24,
		},
		{
			name:       "mega",
			input:      bytes.Repeat([]byte{'m'}, 1024*1024),
			sliceStart: 27,
			sliceEnd:   (1024 * 1024) - 48,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testDataMem(t, testCase)
		})
	}

	tmpDir, err := ioutil.TempDir("", "data-test")
	if err != nil {
		t.Fatal(err)
	}

	// repeat using files
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name+"-file", func(t *testing.T) {
			testDataFile(t, tmpDir, testCase)
		})
	}
}

func testDataFile(t *testing.T, tmpDir string, testCase dataTest) {
	filePath := filepath.Join(tmpDir, testCase.name)

	err := ioutil.WriteFile(filePath, testCase.input, 0600)
	if err != nil {
		t.Fatalf("error creating tmp file: %v", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("error opening tmp file: %v", err)
	}
	cleanup := func() {
		_ = f.Close()
	}
	data, err := NewDataFile(f)
	if err != nil {
		cleanup()
		t.Fatalf("error creating data file: %v", err)
	}
	var buf bytes.Buffer
	n, err := data.WriteTo(&buf)
	if err != nil {
		t.Errorf("error writing data: %v", err)
	}
	if n != int64(len(testCase.input)) {
		t.Errorf("write %d bytes, expected: %d", n, len(testCase.input))
	}
	err = f.Close()
	if err != nil {
		t.Errorf("error closing tmp file: %v", err)
	}
}

func testDataMem(t *testing.T, testCase dataTest) {
	data := NewDataBytes(testCase.input)
	if testCase.sliceStart != 0 || testCase.sliceEnd != 0 {
		data = data.Slice(testCase.sliceStart, testCase.sliceEnd)
	}
	var buf bytes.Buffer
	n, err := data.WriteTo(&buf)
	if err != nil {
		t.Errorf("error writing data: %v", err)
	}
	expect := testCase.input
	if testCase.sliceStart != 0 || testCase.sliceEnd != 0 {
		expect = expect[testCase.sliceStart:testCase.sliceEnd]
	}
	if n != int64(len(expect)) {
		t.Errorf("write %d bytes, expected: %d", n, len(expect))
	}
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Errorf("expected %v, got %v", expect, buf.Bytes())
	}
}
