package engine

import (
	"testing"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
)

func TestSchema(t *testing.T) {

	t.Run("creates a sql schema and arrow schema", func(t *testing.T) {
		source := "vince"
		ts := createSchema(source, []v1.Column{
			v1.Column_id,        // int64
			v1.Column_timestamp, // time.Time
			v1.Column_duration,  // time.Duration
			v1.Column_event,     // time.Duration
		})
		expectSql := sql.Schema{
			{
				Name:   v1.Column_id.String(),
				Source: source,
				Type:   types.Int64,
			},
			{
				Name:   v1.Column_timestamp.String(),
				Source: source,
				Type:   types.Timestamp,
			},
			{
				Name:   v1.Column_duration.String(),
				Source: source,
				Type:   types.Int64,
			},
			{
				Name:   v1.Column_event.String(),
				Source: source,
				Type:   types.Text,
			},
		}
		if !expectSql.Equals(ts.sql) {
			t.Error("mismatch sal schema")
		}
		expectedArrow := arrow.NewSchema([]arrow.Field{
			{
				Name: v1.Column_id.String(),
				Type: arrow.PrimitiveTypes.Int64,
			},
			{
				Name: v1.Column_timestamp.String(),
				Type: arrow.FixedWidthTypes.Timestamp_ms,
			},
			{
				Name: v1.Column_duration.String(),
				Type: arrow.FixedWidthTypes.Duration_ms,
			},
			{
				Name: v1.Column_event.String(),
				Type: arrow.BinaryTypes.String,
			},
		}, nil)
		if !expectedArrow.Equal(ts.arrow) {
			t.Errorf("mismatch arrow schema \n %s \n %s ", expectedArrow, ts.arrow)
		}
	})
}
