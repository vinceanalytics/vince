package be

import (
	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/compute/exprs"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/entry"
)

var extSet = exprs.NewDefaultExtensionSet()

// Full table scan for sites record.
var Base = must.Must(exprs.ToSubstraitType(
	arrow.StructOf(entry.Schema.Fields()...),
	false, extSet,
))("base type")
