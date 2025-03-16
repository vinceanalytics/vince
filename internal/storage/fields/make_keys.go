package fields

import v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"

type Data struct {
	Field     v1.Field
	Shard     uint64
	DataType  v1.DataType
	Container uint64
}
