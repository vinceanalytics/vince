package notify

import (
	"context"

	v1 "github.com/vinceanalytics/proto/gen/go/vince/api/v1"
)

type Notifier interface {
	Notify(context.Context, ...v1.Notice)
}
