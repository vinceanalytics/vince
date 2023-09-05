package b3

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/oklog/ulid/v2"
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
	ctx = SetReader(ctx, NewCache(b, o.CacheDir))
	return context.WithValue(ctx, b3Key{}, b), b
}

func Get(ctx context.Context) objstore.Bucket {
	return ctx.Value(b3Key{}).(objstore.Bucket)
}

type Reader interface {
	Read(ctx context.Context, id ulid.ULID, f func(io.ReaderAt) error) error
}

type readerKey struct{}

func SetReader(ctx context.Context, r Reader) context.Context {
	return context.WithValue(ctx, readerKey{}, r)
}

func GetReader(ctx context.Context) Reader {
	return ctx.Value(readerKey{}).(Reader)
}

type Cache struct {
	bucket objstore.Bucket
	dir    string
}

var _ Reader = (*Cache)(nil)

func NewCache(o objstore.Bucket, dir string) *Cache {
	return &Cache{
		bucket: o,
		dir:    dir,
	}
}

func (c *Cache) Read(ctx context.Context, id ulid.ULID, f func(io.ReaderAt) error) error {
	s := id.String()
	c.ensure(ctx, s)
	x, err := os.Open(filepath.Join(c.dir, s))
	if err != nil {
		return err
	}
	defer x.Close()
	return f(x)
}

func (c *Cache) ensure(ctx context.Context, id string) {
	file := filepath.Join(c.dir, id)
	_, err := os.Stat(file)
	if !os.IsNotExist(err) {
		// try to download from object store
		if r, err := c.bucket.Get(ctx, id); err == nil {
			f := must.Must(os.Open(file))("failed opening local bloc file")
			io.Copy(f, r)
			f.Close()
		}
	}
}
