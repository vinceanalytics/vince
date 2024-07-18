// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: vince/v1/config.proto

package v1

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	_ "google.golang.org/protobuf/types/known/emptypb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Version struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version string `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *Version) Reset() {
	*x = Version{}
	if protoimpl.UnsafeEnabled {
		mi := &file_vince_v1_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Version) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Version) ProtoMessage() {}

func (x *Version) ProtoReflect() protoreflect.Message {
	mi := &file_vince_v1_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Version.ProtoReflect.Descriptor instead.
func (*Version) Descriptor() ([]byte, []int) {
	return file_vince_v1_config_proto_rawDescGZIP(), []int{0}
}

func (x *Version) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data      string  `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Listen    string  `protobuf:"bytes,2,opt,name=listen,proto3" json:"listen,omitempty"`
	RateLimit float64 `protobuf:"fixed64,3,opt,name=rate_limit,json=rateLimit,proto3" json:"rate_limit,omitempty"`
	// Path to the geoiip database used to set web analytics country attribute.
	GeoipDbPath string `protobuf:"bytes,8,opt,name=geoip_db_path,json=geoipDbPath,proto3" json:"geoip_db_path,omitempty"`
	// How long data will be persisted, Older data is automatically deleted.
	RetentionPeriod *durationpb.Duration `protobuf:"bytes,10,opt,name=retention_period,json=retentionPeriod,proto3" json:"retention_period,omitempty"`
	AutoTls         bool                 `protobuf:"varint,11,opt,name=auto_tls,json=autoTls,proto3" json:"auto_tls,omitempty"`
	Acme            *Acme                `protobuf:"bytes,12,opt,name=acme,proto3" json:"acme,omitempty"`
	AuthToken       string               `protobuf:"bytes,13,opt,name=auth_token,json=authToken,proto3" json:"auth_token,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_vince_v1_config_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_vince_v1_config_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_vince_v1_config_proto_rawDescGZIP(), []int{1}
}

func (x *Config) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

func (x *Config) GetListen() string {
	if x != nil {
		return x.Listen
	}
	return ""
}

func (x *Config) GetRateLimit() float64 {
	if x != nil {
		return x.RateLimit
	}
	return 0
}

func (x *Config) GetGeoipDbPath() string {
	if x != nil {
		return x.GeoipDbPath
	}
	return ""
}

func (x *Config) GetRetentionPeriod() *durationpb.Duration {
	if x != nil {
		return x.RetentionPeriod
	}
	return nil
}

func (x *Config) GetAutoTls() bool {
	if x != nil {
		return x.AutoTls
	}
	return false
}

func (x *Config) GetAcme() *Acme {
	if x != nil {
		return x.Acme
	}
	return nil
}

func (x *Config) GetAuthToken() string {
	if x != nil {
		return x.AuthToken
	}
	return ""
}

type Acme struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Email  string `protobuf:"bytes,1,opt,name=email,proto3" json:"email,omitempty"`
	Domain string `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
}

func (x *Acme) Reset() {
	*x = Acme{}
	if protoimpl.UnsafeEnabled {
		mi := &file_vince_v1_config_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Acme) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Acme) ProtoMessage() {}

func (x *Acme) ProtoReflect() protoreflect.Message {
	mi := &file_vince_v1_config_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Acme.ProtoReflect.Descriptor instead.
func (*Acme) Descriptor() ([]byte, []int) {
	return file_vince_v1_config_proto_rawDescGZIP(), []int{2}
}

func (x *Acme) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *Acme) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

type Domain struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *Domain) Reset() {
	*x = Domain{}
	if protoimpl.UnsafeEnabled {
		mi := &file_vince_v1_config_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Domain) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Domain) ProtoMessage() {}

func (x *Domain) ProtoReflect() protoreflect.Message {
	mi := &file_vince_v1_config_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Domain.ProtoReflect.Descriptor instead.
func (*Domain) Descriptor() ([]byte, []int) {
	return file_vince_v1_config_proto_rawDescGZIP(), []int{3}
}

func (x *Domain) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type GetDomainResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Domains []*Domain `protobuf:"bytes,1,rep,name=domains,proto3" json:"domains,omitempty"`
}

func (x *GetDomainResponse) Reset() {
	*x = GetDomainResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_vince_v1_config_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDomainResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDomainResponse) ProtoMessage() {}

func (x *GetDomainResponse) ProtoReflect() protoreflect.Message {
	mi := &file_vince_v1_config_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDomainResponse.ProtoReflect.Descriptor instead.
func (*GetDomainResponse) Descriptor() ([]byte, []int) {
	return file_vince_v1_config_proto_rawDescGZIP(), []int{4}
}

func (x *GetDomainResponse) GetDomains() []*Domain {
	if x != nil {
		return x.Domains
	}
	return nil
}

type SendEventResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Dropped bool `protobuf:"varint,1,opt,name=dropped,proto3" json:"dropped,omitempty"`
}

func (x *SendEventResponse) Reset() {
	*x = SendEventResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_vince_v1_config_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendEventResponse) ProtoMessage() {}

func (x *SendEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_vince_v1_config_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendEventResponse.ProtoReflect.Descriptor instead.
func (*SendEventResponse) Descriptor() ([]byte, []int) {
	return file_vince_v1_config_proto_rawDescGZIP(), []int{5}
}

func (x *SendEventResponse) GetDropped() bool {
	if x != nil {
		return x.Dropped
	}
	return false
}

var File_vince_v1_config_proto protoreflect.FileDescriptor

var file_vince_v1_config_proto_rawDesc = []byte{
	0x0a, 0x15, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x76, 0x31, 0x1a, 0x1b, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70,
	0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x23, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0xba, 0x02, 0x0a, 0x06,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x1a, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x12, 0x23, 0x0a, 0x06, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x0b, 0xba, 0x48, 0x08, 0xc8, 0x01, 0x01, 0x72, 0x03, 0x80, 0x02, 0x01, 0x52,
	0x06, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x12, 0x25, 0x0a, 0x0a, 0x72, 0x61, 0x74, 0x65, 0x5f,
	0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x42, 0x06, 0xba, 0x48, 0x03,
	0xc8, 0x01, 0x01, 0x52, 0x09, 0x72, 0x61, 0x74, 0x65, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x22,
	0x0a, 0x0d, 0x67, 0x65, 0x6f, 0x69, 0x70, 0x5f, 0x64, 0x62, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x67, 0x65, 0x6f, 0x69, 0x70, 0x44, 0x62, 0x50, 0x61,
	0x74, 0x68, 0x12, 0x4c, 0x0a, 0x10, 0x72, 0x65, 0x74, 0x65, 0x6e, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52,
	0x0f, 0x72, 0x65, 0x74, 0x65, 0x6e, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64,
	0x12, 0x19, 0x0a, 0x08, 0x61, 0x75, 0x74, 0x6f, 0x5f, 0x74, 0x6c, 0x73, 0x18, 0x0b, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x07, 0x61, 0x75, 0x74, 0x6f, 0x54, 0x6c, 0x73, 0x12, 0x1c, 0x0a, 0x04, 0x61,
	0x63, 0x6d, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x76, 0x31, 0x2e, 0x41,
	0x63, 0x6d, 0x65, 0x52, 0x04, 0x61, 0x63, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x75, 0x74,
	0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61,
	0x75, 0x74, 0x68, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x4c, 0x0a, 0x04, 0x41, 0x63, 0x6d, 0x65,
	0x12, 0x20, 0x0a, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x0a, 0xba, 0x48, 0x07, 0xc8, 0x01, 0x01, 0x72, 0x02, 0x60, 0x01, 0x52, 0x05, 0x65, 0x6d, 0x61,
	0x69, 0x6c, 0x12, 0x22, 0x0a, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x0a, 0xba, 0x48, 0x07, 0xc8, 0x01, 0x01, 0x72, 0x02, 0x68, 0x01, 0x52, 0x06,
	0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x22, 0x24, 0x0a, 0x06, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e,
	0x12, 0x1a, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x06,
	0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x39, 0x0a, 0x11,
	0x47, 0x65, 0x74, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x24, 0x0a, 0x07, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x52, 0x07,
	0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x22, 0x2d, 0x0a, 0x11, 0x53, 0x65, 0x6e, 0x64, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x64, 0x72, 0x6f, 0x70, 0x70, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x64,
	0x72, 0x6f, 0x70, 0x70, 0x65, 0x64, 0x42, 0x6c, 0x0a, 0x06, 0x63, 0x6f, 0x6d, 0x2e, 0x76, 0x31,
	0x42, 0x0b, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a,
	0x2d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x69, 0x6e, 0x63,
	0x65, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x74, 0x73, 0x75, 0x2f, 0x67,
	0x65, 0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x31, 0xa2, 0x02,
	0x03, 0x56, 0x58, 0x58, 0xaa, 0x02, 0x02, 0x56, 0x31, 0xca, 0x02, 0x02, 0x56, 0x31, 0xe2, 0x02,
	0x0e, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x02, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_vince_v1_config_proto_rawDescOnce sync.Once
	file_vince_v1_config_proto_rawDescData = file_vince_v1_config_proto_rawDesc
)

func file_vince_v1_config_proto_rawDescGZIP() []byte {
	file_vince_v1_config_proto_rawDescOnce.Do(func() {
		file_vince_v1_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_vince_v1_config_proto_rawDescData)
	})
	return file_vince_v1_config_proto_rawDescData
}

var file_vince_v1_config_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_vince_v1_config_proto_goTypes = []interface{}{
	(*Version)(nil),             // 0: v1.Version
	(*Config)(nil),              // 1: v1.Config
	(*Acme)(nil),                // 2: v1.Acme
	(*Domain)(nil),              // 3: v1.Domain
	(*GetDomainResponse)(nil),   // 4: v1.GetDomainResponse
	(*SendEventResponse)(nil),   // 5: v1.SendEventResponse
	(*durationpb.Duration)(nil), // 6: google.protobuf.Duration
}
var file_vince_v1_config_proto_depIdxs = []int32{
	6, // 0: v1.Config.retention_period:type_name -> google.protobuf.Duration
	2, // 1: v1.Config.acme:type_name -> v1.Acme
	3, // 2: v1.GetDomainResponse.domains:type_name -> v1.Domain
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_vince_v1_config_proto_init() }
func file_vince_v1_config_proto_init() {
	if File_vince_v1_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_vince_v1_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Version); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_vince_v1_config_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_vince_v1_config_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Acme); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_vince_v1_config_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Domain); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_vince_v1_config_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDomainResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_vince_v1_config_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendEventResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_vince_v1_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_vince_v1_config_proto_goTypes,
		DependencyIndexes: file_vince_v1_config_proto_depIdxs,
		MessageInfos:      file_vince_v1_config_proto_msgTypes,
	}.Build()
	File_vince_v1_config_proto = out.File
	file_vince_v1_config_proto_rawDesc = nil
	file_vince_v1_config_proto_goTypes = nil
	file_vince_v1_config_proto_depIdxs = nil
}
