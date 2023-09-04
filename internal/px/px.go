package px

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"
	"github.com/thanos-io/objstore/providers/azure"
	"github.com/thanos-io/objstore/providers/bos"
	"github.com/thanos-io/objstore/providers/cos"
	"github.com/thanos-io/objstore/providers/filesystem"
	"github.com/thanos-io/objstore/providers/gcs"
	"github.com/thanos-io/objstore/providers/obs"
	"github.com/thanos-io/objstore/providers/s3"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	configv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	"google.golang.org/protobuf/proto"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func Interface(v *v1.Query_Value) (val any) {
	switch e := v.Value.(type) {
	case *v1.Query_Value_Number:
		val = e.Number
	case *v1.Query_Value_Double:
		val = e.Double
	case *v1.Query_Value_String_:
		val = e.String_
	case *v1.Query_Value_Bool:
		val = e.Bool
	case *v1.Query_Value_Timestamp:
		val = e.Timestamp.AsTime()
	}
	return
}

func NewQueryValue(v any) *v1.Query_Value {
	switch e := v.(type) {
	case int64:
		return &v1.Query_Value{
			Value: &v1.Query_Value_Number{
				Number: e,
			},
		}
	case float64:
		return &v1.Query_Value{
			Value: &v1.Query_Value_Double{
				Double: e,
			},
		}
	case string:
		return &v1.Query_Value{
			Value: &v1.Query_Value_String_{
				String_: e,
			},
		}
	case bool:
		return &v1.Query_Value{
			Value: &v1.Query_Value_Bool{
				Bool: e,
			},
		}
	case time.Time:
		return &v1.Query_Value{
			Value: &v1.Query_Value_Timestamp{
				Timestamp: timestamppb.New(e),
			},
		}
	default:
		panic(fmt.Sprintf("unknown value type %#T", v))
	}
}

func ColumnIndex(c storev1.Column) int {
	if c <= storev1.Column_timestamp {
		return int(c)
	}
	return int(c - storev1.Column_browser)
}

func Decode(m proto.Message) func(val []byte) error {
	return func(val []byte) error {
		return proto.Unmarshal(val, m)
	}
}

func Azure(o *configv1.BlockStore_Azure) *azure.Config {
	var reader azure.ReaderConfig
	if o.ReaderConfig != nil {
		reader.MaxRetryRequests = int(o.ReaderConfig.MaxRetryRequests)
	}
	var pipe azure.PipelineConfig
	if o.PipelineConfig != nil {
		pipe.MaxRetryDelay = model.Duration(o.PipelineConfig.MaxTries)
		pipe.TryTimeout = model.Duration(o.PipelineConfig.TryTimeout.AsDuration())
		pipe.RetryDelay = model.Duration(o.PipelineConfig.RetryDelay.AsDuration())
		pipe.MaxRetryDelay = model.Duration(o.PipelineConfig.MaxRetryDelay.AsDuration())
	}
	return &azure.Config{
		StorageAccountName:      o.StorageAccount,
		StorageAccountKey:       o.StorageAccount,
		StorageConnectionString: o.StorageConnectionString,
		ContainerName:           o.Container,
		Endpoint:                o.Endpoint,
		UserAssignedID:          o.UserAssignedId,
		MaxRetries:              int(o.MaxRetries),
		ReaderConfig:            reader,
		PipelineConfig:          pipe,
	}
}

func Bos(o *configv1.BlockStore_BOS) *bos.Config {
	return &bos.Config{
		Bucket:    o.Bucket,
		Endpoint:  o.Endpoint,
		AccessKey: o.AccessKey,
		SecretKey: o.SecretKey,
	}
}

func Cos(o *configv1.BlockStore_COS) *cos.Config {
	return &cos.Config{
		Bucket:    o.Bucket,
		Region:    o.Region,
		AppId:     o.AppId,
		Endpoint:  o.Endpoint,
		SecretKey: o.SecretKey,
		SecretId:  o.SecretId,
	}
}

func Filesystem(o *configv1.BlockStore_Filesystem) *filesystem.Config {
	return &filesystem.Config{
		Directory: o.Directory,
	}
}

func GCS(o *configv1.BlockStore_GCS) *gcs.Config {
	return &gcs.Config{
		Bucket:         o.Bucket,
		ServiceAccount: o.ServiceAccount,
	}
}

func OBS(o *configv1.BlockStore_OBS) *obs.Config {
	return &obs.Config{
		Bucket:    o.Bucket,
		Endpoint:  o.Endpoint,
		AccessKey: o.AccessKey,
		SecretKey: o.SecretKey,
	}
}

func S3(o *configv1.BlockStore_S3) *s3.Config {
	var ss s3.SSEConfig
	if s := o.SseConfig; s != nil {
		ss.Type = s.Type
		ss.KMSKeyID = s.KmsKeyId
		ss.KMSEncryptionContext = s.KmsEncryptionContext
		ss.EncryptionKey = s.EncryptionKey
	}
	return &s3.Config{
		Bucket:             o.Bucket,
		Endpoint:           o.Endpoint,
		Region:             o.Region,
		AWSSDKAuth:         o.AwsSdkAuth,
		AccessKey:          o.AccessKey,
		Insecure:           o.Insecure,
		SignatureV2:        o.SignatureVersion2,
		SecretKey:          o.SecretKey,
		SessionToken:       o.SessionToken,
		PutUserMetadata:    o.PutUserMetadata,
		ListObjectsVersion: o.ListObjectsVersion,
		BucketLookupType:   s3.BucketLookupType(o.BucketLookupType),
		PartSize:           o.PartSize,
		SSEConfig:          ss,
		STSEndpoint:        o.StsEndpoint,
	}
}
