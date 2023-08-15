package pj

import (
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
	return proto.Unmarshal(data, m)
}
