package b3

import (
	"context"

	"github.com/thanos-io/objstore"
	"github.com/thanos-io/objstore/providers/azure"
	"github.com/thanos-io/objstore/providers/bos"
	"github.com/thanos-io/objstore/providers/cos"
	"github.com/thanos-io/objstore/providers/filesystem"
	"github.com/thanos-io/objstore/providers/gcs"
	"github.com/thanos-io/objstore/providers/obs"
	"github.com/thanos-io/objstore/providers/oss"
	"github.com/thanos-io/objstore/providers/s3"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/px"
)

func New(o *v1.BlockStore) (objstore.Bucket, error) {
	switch e := o.Provider.(type) {
	case *v1.BlockStore_Fs:
		return filesystem.NewBucket(e.Fs.Directory)
	case *v1.BlockStore_S3_:
		return s3.NewBucketWithConfig(nil, px.S3(e.S3), "")
	case *v1.BlockStore_Azure_:
		return azure.NewBucketWithConfig(nil, px.Azure(e.Azure), "")
	case *v1.BlockStore_Bos:
		return bos.NewBucketWithConfig(nil, px.Bos(e.Bos), "")
	case *v1.BlockStore_Cos:
		return cos.NewBucketWithConfig(nil, px.Cos(e.Cos), "")
	case *v1.BlockStore_Gcs:
		return gcs.NewBucketWithConfig(context.TODO(), nil, px.GCS(e.Gcs), "")
	case *v1.BlockStore_Obs:
		return obs.NewBucketWithConfig(nil, px.OBS(e.Obs))
	case *v1.BlockStore_Oss:
		return oss.NewBucketWithConfig(nil, px.OSS(e.Oss), "")
	default:
		panic("unreachable")
	}
}

type b3Key struct{}

func Open(ctx context.Context, o *v1.BlockStore) (context.Context, objstore.Bucket) {
	b := must.Must(New(o))("failed opening object storage")
	return context.WithValue(ctx, b3Key{}, b), b
}

func Get(ctx context.Context) objstore.Bucket {
	return ctx.Value(b3Key{}).(objstore.Bucket)
}
