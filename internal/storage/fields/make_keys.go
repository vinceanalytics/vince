package fields

import v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"

type Data struct {
	Shard     uint64
	Container uint64
	Field     v1.Field
	DataType  v1.DataType
}
