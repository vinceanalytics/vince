package engine

import (
	"fmt"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/vinceanalytics/vince/internal/must"
)

var VinceFuncs = []sql.Function{
	sql.FunctionN{Name: TimeBucketName, Fn: NewTimeBucket},
}

const TimeBucketName = "time_bucket"

var base = must.Must(time.Parse(time.RFC822Z, time.RFC822Z))("failed creating base timestamp").UTC()

type TimeBucket struct {
	ts       sql.Expression
	interval *expression.Literal
	starting time.Time
	duration time.Duration
}

var _ sql.FunctionExpression = (*TimeBucket)(nil)

func NewTimeBucket(args ...sql.Expression) (sql.Expression, error) {
	switch len(args) {
	case 2:
		if !types.IsTimestampType(args[0].Type()) {
			return nil, fmt.Errorf("%s expects timestamp as first argument", TimeBucketName)
		}
		lit, ok := args[1].(*expression.Literal)
		if !ok {
			return nil, fmt.Errorf("%s expects duration as second argument", TimeBucketName)
		}
		val, _ := lit.Eval(nil, nil)
		d, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("%s expects duration as second argument", TimeBucketName)
		}
		dur, err := time.ParseDuration(d)
		if err != nil {
			return nil, fmt.Errorf("%s invalid duration", TimeBucketName)
		}
		return &TimeBucket{
			ts:       args[0],
			interval: lit,
			duration: dur,
		}, nil
	default:
		return nil, sql.ErrInvalidArgumentNumber.New(TimeBucketName, 2, len(args))
	}
}

// FunctionName implements sql.FunctionExpression
func (t *TimeBucket) FunctionName() string {
	return TimeBucketName
}

// Description implements sql.FunctionExpression
func (t *TimeBucket) Description() string {
	return "roll up timestamp in given interval."
}

// Children implements the sql.Expression interface.
func (t *TimeBucket) Children() []sql.Expression {
	return []sql.Expression{t.ts, t.interval}
}

// Resolved implements the sql.Expression interface.
func (t *TimeBucket) Resolved() bool {
	return t.ts.Resolved() && t.interval.Resolved()
}

// IsNullable implements the sql.Expression interface.
func (t *TimeBucket) IsNullable() bool {
	return false
}

// Type implements the sql.Expression interface.
func (t *TimeBucket) Type() sql.Type {
	return types.Timestamp
}

// WithChildren implements the Expression interface.
func (t *TimeBucket) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	return NewTimeBucket(children...)
}

// Eval implements the sql.Expression interface.
func (t *TimeBucket) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	val, err := t.ts.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	ts := val.(time.Time)
	if t.starting.IsZero() {
		t.starting = startingPoint(ts, t.duration)
	}
	if ts.After(t.starting) {
		t.starting = t.starting.Add(t.duration)
	}
	return t.starting, nil
}

func (t *TimeBucket) String() string {
	return fmt.Sprintf("%s(%s,%s)", t.FunctionName(), t.ts, t.interval)
}

// Calculates the first bucket that value a will fall into
func startingPoint(a time.Time, i time.Duration) time.Time {
	n := a.Sub(base)
	x := n / i
	return base.Add(i * x)
}
