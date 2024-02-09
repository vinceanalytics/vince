package ice

import (
	"bytes"
	"testing"
)

func TestChunkedDocumentCoder(t *testing.T) {
	tests := []struct {
		chunkSize        uint64
		docNums          []uint64
		metas            [][]byte
		datas            [][]byte
		expected         []byte
		expectedChunkNum int
	}{
		{
			chunkSize: 1,
			docNums:   []uint64{0},
			metas:     [][]byte{{0}},
			datas:     [][]byte{[]byte("bluge")},
			expected: []byte{
				0x28, 0xb5, 0x2f, 0xfd, 0x4, 0x0, 0x41,
				0x0, 0x0, 0x1, 0x5, 0x0, 0x62, 0x6c, 0x75, 0x67, 0x65, 0x2b, 0x30, 0x97, 0x33, 0x0, 0x15, 0x15,
				0x0, 0x0, 0x0, 0x3, 0x0, 0x0, 0x0, 0x3,
			},
			expectedChunkNum: 3, // left, chunk, right
		},
		{
			chunkSize: 1,
			docNums:   []uint64{0, 1},
			metas:     [][]byte{{0}, {1}},
			datas:     [][]byte{[]byte("upside"), []byte("scorch")},
			expected: []byte{
				0x28, 0xb5, 0x2f, 0xfd, 0x4, 0x0, 0x49,
				0x0, 0x0, 0x1, 0x6, 0x0, 0x75, 0x70, 0x73, 0x69, 0x64, 0x65,
				0x36, 0x6e, 0x7e, 0x39, 0x28, 0xb5, 0x2f, 0xfd, 0x4, 0x0, 0x49,
				0x0, 0x0, 0x1, 0x6, 0x1, 0x73, 0x63, 0x6f, 0x72, 0x63, 0x68,
				0x8f, 0x83, 0xa3, 0x37, 0x0, 0x16, 0x2c, 0x2c,
				0x0, 0x0, 0x0, 0x4, 0x0, 0x0, 0x0, 0x4,
			},
			expectedChunkNum: 4, // left, chunk, chunk, right
		},
	}

	for _, test := range tests {
		var actual bytes.Buffer
		cic := newChunkedDocumentCoder(test.chunkSize, &actual)
		for i, docNum := range test.docNums {
			_, err := cic.Add(docNum, test.metas[i], test.datas[i])
			if err != nil {
				t.Fatalf("error adding to documentcoder: %v", err)
			}
		}
		err := cic.Write()
		if err != nil {
			t.Fatalf("error writing: %v", err)
		}
		if !bytes.Equal(test.expected, actual.Bytes()) {
			t.Errorf("got:%s, expected:%s", actual.String(), string(test.expected))
		}
		if test.expectedChunkNum != cic.Len() {
			t.Errorf("got:%d, expected:%d", cic.Len(), test.expectedChunkNum)
		}
	}
}

func TestChunkedDocumentCoders(t *testing.T) {
	chunkSize := uint64(2)
	docNums := []uint64{0, 1, 2, 3, 4, 5}
	metas := [][]byte{
		{0},
		{1},
		{2},
		{3},
		{4},
		{5},
	}
	datas := [][]byte{
		[]byte("scorch"),
		[]byte("does"),
		[]byte("better"),
		[]byte("than"),
		[]byte("upside"),
		[]byte("down"),
	}
	chunkNum := 5 // left, chunk, chunk, chunk, right

	var actual1, actual2 bytes.Buffer
	// chunkedDocumentCoder that writes out at the end
	cic1 := newChunkedDocumentCoder(chunkSize, &actual1)
	// chunkedContentCoder that writes out in chunks
	cic2 := newChunkedDocumentCoder(chunkSize, &actual2)

	for i, docNum := range docNums {
		_, err := cic1.Add(docNum, metas[i], datas[i])
		if err != nil {
			t.Fatalf("error adding to documentcoder: %v", err)
		}
		_, err = cic2.Add(docNum, metas[i], datas[i])
		if err != nil {
			t.Fatalf("error adding to documentcoder: %v", err)
		}
	}

	err := cic1.Write()
	if err != nil {
		t.Fatalf("error writing: %v", err)
	}
	err = cic2.Write()
	if err != nil {
		t.Fatalf("error writing: %v", err)
	}

	if !bytes.Equal(actual1.Bytes(), actual2.Bytes()) {
		t.Errorf("%s != %s", actual1.String(), actual2.String())
	}
	if chunkNum != cic1.Len() {
		t.Errorf("got:%d, expected:%d", cic1.Len(), chunkNum)
	}
	if chunkNum != cic2.Len() {
		t.Errorf("got:%d, expected:%d", cic2.Len(), chunkNum)
	}
}
