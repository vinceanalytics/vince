package notify

import (
	"context"

	v1 "github.com/vinceanalytics/vince/proto/v1"
)

type Notifier interface {
	Notify(context.Context, ...v1.Notice)
}
