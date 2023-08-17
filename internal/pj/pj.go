package pj

import (
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var indent = protojson.MarshalOptions{
	Indent: "  ",
}

func MarshalIndent(m proto.Message) ([]byte, error) {
	return indent.Marshal(m)
}
func Marshal(m proto.Message) ([]byte, error) {
	return protojson.Marshal(m)
}

func Unmarshal(data []byte, m proto.Message) error {
	return protojson.Unmarshal(data, m)
}

func UnmarshalLimit(m proto.Message, r io.Reader, limit int64) error {
	data, err := io.ReadAll(io.LimitReader(r, limit))
	if err != nil {
		return err
	}
	return Unmarshal(data, m)
}

const defaultLimit = 100 << 10

func UnmarshalDefault(m proto.Message, r io.Reader) error {
	return UnmarshalLimit(m, r, defaultLimit)
}
