package columns

import (
	"github.com/vinceanalytics/vince/camel"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
)

var Event = camel.Case(v1.Filters_Event.String())
var ID = camel.Case(v1.Filters_ID.String())
var Session = camel.Case(v1.Filters_Session.String())
var Bounce = camel.Case(v1.Filters_Bounce.String())
var Duration = camel.Case(v1.Filters_Duration.String())
var Timestamp = camel.Case(v1.Filters_Timestamp.String())
var View = camel.Case(v1.Filters_View.String())
