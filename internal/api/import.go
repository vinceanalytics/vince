package api

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/csv"
	"github.com/apache/arrow/go/v14/arrow/ipc"
	v1 "github.com/vinceanalytics/proto/gen/go/vince/import/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/px"
	"github.com/vinceanalytics/vince/internal/tokens"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ v1.ImportServer = (*API)(nil)

// Import implements v1.ImportServer . Accepts a schema together with a csv
// file. This will try to read the csv file using the provided schema into an
// arrow.Record which is stored uncompressed.
//
// Imports are stored globally, they are not namespaced by user.
func (API) Import(ctx context.Context, req *v1.ImportRequest) (*v1.ImportResponse, error) {
	me := tokens.GetAccount(ctx)
	key := keys.Import(req.Schema.Name)
	err := db.Update(ctx, func(txn db.Transaction) error {
		if txn.Has(key) {
			return status.Error(codes.AlreadyExists, "an import for the schema already exists")
		}
		r := csv.NewReader(bytes.NewReader(req.Csv), schema(req.Schema),
			csv.WithAllocator(entry.Pool),
			csv.WithChunk(-1),
		)
		r.Next()
		record := r.Record()
		var o bytes.Buffer
		w := ipc.NewWriter(&o,
			ipc.WithAllocator(entry.Pool),
			// We don't need to compress because we already compress with badger. We need
			// fast queries.
		)
		err := w.Write(record)
		if err != nil {
			slog.Error("failed to write import record",
				"err", err,
			)
			return E500()
		}
		return txn.Set(key, px.Encode(&v1.ImportData{
			Data:      o.Bytes(),
			CreatedBy: me.Name,
			CreatedAt: timestamppb.Now(),
		}), 0)
	})
	if err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "imports are not supported yet")
}

func schema(s *v1.Schema) *arrow.Schema {
	fields := make([]arrow.Field, len(s.Fields))
	for i := range s.Fields {
		fields[i] = arrow.Field{
			Name:     s.Fields[i].Name,
			Type:     dt(s.Fields[i].Type),
			Nullable: true,
		}
	}
	return arrow.NewSchema(fields, nil)
}

func dt(t v1.Schema_FieldType) arrow.DataType {
	switch t {
	case v1.Schema_INT64:
		return arrow.PrimitiveTypes.Int64
	case v1.Schema_FLOAT64:
		return arrow.PrimitiveTypes.Float64
	case v1.Schema_STRING:
		return arrow.BinaryTypes.String
	case v1.Schema_BOOL:
		return arrow.FixedWidthTypes.Boolean
	case v1.Schema_TIMESTAMP:
		return arrow.FixedWidthTypes.Timestamp_ms
	default:
		panic("unreachable")
	}
}
