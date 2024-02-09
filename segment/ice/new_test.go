//  Copyright (c) 2020 Couchbase, Inc.
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
	"path/filepath"
	"testing"

	"github.com/vinceanalytics/vince/segment"
)

func TestBuild(t *testing.T) {
	path, cleanup := setupTestDir(t)
	defer cleanup()

	sb, err := buildTestSegment()
	if err != nil {
		t.Fatal(err)
	}
	segPath := filepath.Join(path, "segment.ice")
	err = persistToFile(sb, segPath)
	if err != nil {
		t.Fatal(err)
	}
}

func buildTestSegment() (*Segment, error) {
	doc := &FakeDocument{
		NewFakeField("_id", "a", true, false, false),
		NewFakeField("name", "wow", true, true, false),
		NewFakeField("desc", "some thing", true, true, false),
		NewFakeField("tag", "cold", true, true, false),
		NewFakeField("tag", "dark", true, true, false),
	}
	doc.FakeComposite("_all", []string{"_id"})

	results := []segment.Document{
		doc,
	}

	seg, _, err := newWithChunkMode(results, encodeNorm, defaultChunkMode)
	return seg.(*Segment), err
}

func buildTestSegmentMulti() (*Segment, error) {
	results := buildTestAnalysisResultsMulti()

	seg, _, err := newWithChunkMode(results, encodeNorm, defaultChunkMode)
	return seg.(*Segment), err
}

func buildTestSegmentMultiWithChunkFactor(chunkFactor uint32) (*Segment, error) {
	results := buildTestAnalysisResultsMulti()

	seg, _, err := newWithChunkMode(results, encodeNorm, chunkFactor)
	return seg.(*Segment), err
}

func buildTestSegmentMultiWithDifferentFields(includeDocA, includeDocB bool) (*Segment, error) {
	results := buildTestAnalysisResultsMultiWithDifferentFields(includeDocA, includeDocB)

	seg, _, err := newWithChunkMode(results, encodeNorm, defaultChunkMode)
	return seg.(*Segment), err
}

func buildTestAnalysisResultsMulti() []segment.Document {
	doc := &FakeDocument{
		NewFakeField("_id", "a", true, false, false),
		NewFakeField("name", "wow", true, true, false),
		NewFakeField("desc", "some thing", true, true, false),
		NewFakeField("tag", "dark", true, true, false),
	}
	doc.FakeComposite("_all", []string{"_id"})

	doc2 := &FakeDocument{
		NewFakeField("_id", "b", true, false, false),
		NewFakeField("name", "who", true, true, false),
		NewFakeField("desc", "some thing", true, true, false),
		NewFakeField("tag", "cold", true, true, false),
		NewFakeField("tag", "dark", true, true, false),
	}
	doc2.FakeComposite("_all", []string{"_id"})

	results := []segment.Document{
		doc, doc2,
	}

	return results
}

func buildTestAnalysisResultsMultiWithDifferentFields(includeDocA, includeDocB bool) []segment.Document {
	var results []segment.Document

	if includeDocA {
		doc := &FakeDocument{
			NewFakeField("_id", "a", false, false, false),
			NewFakeField("name", "ABC", false, false, true),
			NewFakeField("dept", "ABC dept", false, false, true),
			NewFakeField("manages.id", "XYZ", false, false, true),
			NewFakeField("manages.count", "1", false, false, true),
		}
		doc.FakeComposite("_all", []string{"_id"})
		results = append(results, doc)
	}

	if includeDocB {
		doc := &FakeDocument{
			NewFakeField("_id", "b", true, false, false),
			NewFakeField("name", "XYZ", false, false, true),
			NewFakeField("dept", "ABC dept", false, false, true),
			NewFakeField("reportsTo.id", "ABC", false, false, true),
		}
		doc.FakeComposite("_all", []string{"_id"})
		results = append(results, doc)
	}

	return results
}

func buildTestSegmentWithDefaultFieldMapping(chunkFactor uint32) (
	*Segment, []string, error) {
	doc := &FakeDocument{
		NewFakeField("_id", "a", true, false, false),
		NewFakeField("name", "wow", false, false, true),
		NewFakeField("desc", "some thing", false, false, true),
		NewFakeField("tag", "cold", false, false, true),
	}
	doc.FakeComposite("_all", []string{"_id"})

	var fields []string
	fields = append(fields, "_id", "name", "desc", "tag")

	results := []segment.Document{
		doc,
	}

	sb, _, err := newWithChunkMode(results, encodeNorm, chunkFactor)

	return sb.(*Segment), fields, err
}
