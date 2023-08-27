package v1

import (
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"io"
	"slices"
	"time"

	"google.golang.org/protobuf/proto"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func (s *StoreKey) Parts() []string {
	if s.Namespace == "" {
		s.Namespace = "vince"
	}
	return []string{
		s.Namespace, s.Prefix.String(),
	}
}

func (s *Site_Key) Parts() []string {
	return append(s.Store.Parts(), s.Domain)
}

func (s *Block_Key) Parts() []string {
	return append(s.Store.Parts(), s.Kind.String(), s.Domain, s.Uid)
}

func (s *Account_Key) Parts() []string {
	return append(s.Store.Parts(), s.Name)
}

func (s *Token_Key) Parts() []string {
	return append(s.Store.Parts(), fmt.Sprint(s.Hash))
}

func (s *Raft_Log_Key) Parts() []string {
	var k string
	if s.Index != -1 {
		k = fmt.Sprint(s.Index)
	}
	return append(s.Store.Parts(), k)
}

func (s *Raft_Stable_Key) Parts() []string {
	return append(s.Store.Parts(), base32.StdEncoding.EncodeToString(s.Key))
}

func (s *Raft_Snapshot_Key) Parts() []string {
	return append(s.Store.Parts(), s.Mode.String(), s.Id)
}

func (v *Query_Value) Interface() (val any) {
	switch e := v.Value.(type) {
	case *Query_Value_Number:
		val = e.Number
	case *Query_Value_Double:
		val = e.Double
	case *Query_Value_String_:
		val = e.String_
	case *Query_Value_Bool:
		val = e.Bool
	case *Query_Value_Timestamp:
		val = e.Timestamp.AsTime()
	}
	return
}

func NewQueryValue(v any) *Query_Value {
	switch e := v.(type) {
	case int64:
		return &Query_Value{
			Value: &Query_Value_Number{
				Number: e,
			},
		}
	case float64:
		return &Query_Value{
			Value: &Query_Value_Double{
				Double: e,
			},
		}
	case string:
		return &Query_Value{
			Value: &Query_Value_String_{
				String_: e,
			},
		}
	case bool:
		return &Query_Value{
			Value: &Query_Value_Bool{
				Bool: e,
			},
		}
	case time.Time:
		return &Query_Value{
			Value: &Query_Value_Timestamp{
				Timestamp: timestamppb.New(e),
			},
		}
	default:
		fmt.Printf("======= %v %#T\n", v, v)
		panic(fmt.Sprintf("unknown value type %#T", v))
	}
}

func (c Column) Index() int {
	if c <= Column_timestamp {
		return int(c)
	}
	return int(c - Column_browser)
}

func WriteTo(w io.Writer, m proto.Message) error {
	if err := binary.Write(w, binary.LittleEndian, uint64(proto.Size(m))); err != nil {
		return err
	}
	buf, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

type Reader struct {
	buf []byte
	r   io.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:   r,
		buf: make([]byte, 0, 4<<10),
	}
}

func (r *Reader) Reset(rd io.Reader) {
	r.r = rd
	r.buf = r.buf[:0]
}

func (r *Reader) Read(m proto.Message) error {
	var sz uint64
	err := binary.Read(r.r, binary.LittleEndian, &sz)
	if err != nil {
		return err
	}
	r.buf = slices.Grow(r.buf, int(sz))
	if _, err = io.ReadFull(r.r, r.buf[:sz]); err != nil {
		return err
	}
	return proto.Unmarshal(r.buf[:sz], m)
}
