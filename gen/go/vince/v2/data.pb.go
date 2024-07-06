// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: vince/v2/data.proto

package v2

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Data struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp      int64  `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Id             []byte `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Bounce         *bool  `protobuf:"varint,3,opt,name=bounce,proto3,oneof" json:"bounce,omitempty"`
	Session        bool   `protobuf:"varint,4,opt,name=session,proto3" json:"session,omitempty"`
	View           bool   `protobuf:"varint,5,opt,name=view,proto3" json:"view,omitempty"`
	Duration       int64  `protobuf:"varint,6,opt,name=duration,proto3" json:"duration,omitempty"`
	Browser        string `protobuf:"bytes,19,opt,name=browser,proto3" json:"browser,omitempty"`
	BrowserVersion string `protobuf:"bytes,20,opt,name=browser_version,json=browserVersion,proto3" json:"browser_version,omitempty"`
	City           string `protobuf:"bytes,26,opt,name=city,proto3" json:"city,omitempty"`
	Country        string `protobuf:"bytes,23,opt,name=country,proto3" json:"country,omitempty"`
	Device         string `protobuf:"bytes,18,opt,name=device,proto3" json:"device,omitempty"`
	Domain         string `protobuf:"bytes,25,opt,name=domain,proto3" json:"domain,omitempty"`
	EntryPage      string `protobuf:"bytes,9,opt,name=entry_page,json=entryPage,proto3" json:"entry_page,omitempty"`
	Event          string `protobuf:"bytes,7,opt,name=event,proto3" json:"event,omitempty"`
	ExitPage       string `protobuf:"bytes,10,opt,name=exit_page,json=exitPage,proto3" json:"exit_page,omitempty"`
	Host           string `protobuf:"bytes,27,opt,name=host,proto3" json:"host,omitempty"`
	Os             string `protobuf:"bytes,21,opt,name=os,proto3" json:"os,omitempty"`
	OsVersion      string `protobuf:"bytes,22,opt,name=os_version,json=osVersion,proto3" json:"os_version,omitempty"`
	Page           string `protobuf:"bytes,8,opt,name=page,proto3" json:"page,omitempty"`
	Referrer       string `protobuf:"bytes,12,opt,name=referrer,proto3" json:"referrer,omitempty"`
	Region         string `protobuf:"bytes,24,opt,name=region,proto3" json:"region,omitempty"`
	Source         string `protobuf:"bytes,11,opt,name=source,proto3" json:"source,omitempty"`
	UtmCampaign    string `protobuf:"bytes,15,opt,name=utm_campaign,json=utmCampaign,proto3" json:"utm_campaign,omitempty"`
	UtmContent     string `protobuf:"bytes,16,opt,name=utm_content,json=utmContent,proto3" json:"utm_content,omitempty"`
	UtmMedium      string `protobuf:"bytes,14,opt,name=utm_medium,json=utmMedium,proto3" json:"utm_medium,omitempty"`
	UtmSource      string `protobuf:"bytes,13,opt,name=utm_source,json=utmSource,proto3" json:"utm_source,omitempty"`
	UtmTerm        string `protobuf:"bytes,17,opt,name=utm_term,json=utmTerm,proto3" json:"utm_term,omitempty"`
	TenantId       string `protobuf:"bytes,28,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
}

func (x *Data) Reset() {
	*x = Data{}
	if protoimpl.UnsafeEnabled {
		mi := &file_vince_v2_data_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Data) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Data) ProtoMessage() {}

func (x *Data) ProtoReflect() protoreflect.Message {
	mi := &file_vince_v2_data_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Data.ProtoReflect.Descriptor instead.
func (*Data) Descriptor() ([]byte, []int) {
	return file_vince_v2_data_proto_rawDescGZIP(), []int{0}
}

func (x *Data) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *Data) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Data) GetBounce() bool {
	if x != nil && x.Bounce != nil {
		return *x.Bounce
	}
	return false
}

func (x *Data) GetSession() bool {
	if x != nil {
		return x.Session
	}
	return false
}

func (x *Data) GetView() bool {
	if x != nil {
		return x.View
	}
	return false
}

func (x *Data) GetDuration() int64 {
	if x != nil {
		return x.Duration
	}
	return 0
}

func (x *Data) GetBrowser() string {
	if x != nil {
		return x.Browser
	}
	return ""
}

func (x *Data) GetBrowserVersion() string {
	if x != nil {
		return x.BrowserVersion
	}
	return ""
}

func (x *Data) GetCity() string {
	if x != nil {
		return x.City
	}
	return ""
}

func (x *Data) GetCountry() string {
	if x != nil {
		return x.Country
	}
	return ""
}

func (x *Data) GetDevice() string {
	if x != nil {
		return x.Device
	}
	return ""
}

func (x *Data) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

func (x *Data) GetEntryPage() string {
	if x != nil {
		return x.EntryPage
	}
	return ""
}

func (x *Data) GetEvent() string {
	if x != nil {
		return x.Event
	}
	return ""
}

func (x *Data) GetExitPage() string {
	if x != nil {
		return x.ExitPage
	}
	return ""
}

func (x *Data) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *Data) GetOs() string {
	if x != nil {
		return x.Os
	}
	return ""
}

func (x *Data) GetOsVersion() string {
	if x != nil {
		return x.OsVersion
	}
	return ""
}

func (x *Data) GetPage() string {
	if x != nil {
		return x.Page
	}
	return ""
}

func (x *Data) GetReferrer() string {
	if x != nil {
		return x.Referrer
	}
	return ""
}

func (x *Data) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *Data) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *Data) GetUtmCampaign() string {
	if x != nil {
		return x.UtmCampaign
	}
	return ""
}

func (x *Data) GetUtmContent() string {
	if x != nil {
		return x.UtmContent
	}
	return ""
}

func (x *Data) GetUtmMedium() string {
	if x != nil {
		return x.UtmMedium
	}
	return ""
}

func (x *Data) GetUtmSource() string {
	if x != nil {
		return x.UtmSource
	}
	return ""
}

func (x *Data) GetUtmTerm() string {
	if x != nil {
		return x.UtmTerm
	}
	return ""
}

func (x *Data) GetTenantId() string {
	if x != nil {
		return x.TenantId
	}
	return ""
}

var File_vince_v2_data_proto protoreflect.FileDescriptor

var file_vince_v2_data_proto_rawDesc = []byte{
	0x0a, 0x13, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x32, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x76, 0x32, 0x22, 0xf6, 0x05, 0x0a, 0x04, 0x44, 0x61,
	0x74, 0x61, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x1b, 0x0a, 0x06, 0x62, 0x6f, 0x75, 0x6e, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08,
	0x48, 0x00, 0x52, 0x06, 0x62, 0x6f, 0x75, 0x6e, 0x63, 0x65, 0x88, 0x01, 0x01, 0x12, 0x18, 0x0a,
	0x07, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07,
	0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x76, 0x69, 0x65, 0x77, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x76, 0x69, 0x65, 0x77, 0x12, 0x1a, 0x0a, 0x08, 0x64,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x64,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x62, 0x72, 0x6f, 0x77, 0x73,
	0x65, 0x72, 0x18, 0x13, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x62, 0x72, 0x6f, 0x77, 0x73, 0x65,
	0x72, 0x12, 0x27, 0x0a, 0x0f, 0x62, 0x72, 0x6f, 0x77, 0x73, 0x65, 0x72, 0x5f, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x14, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x62, 0x72, 0x6f, 0x77,
	0x73, 0x65, 0x72, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x69,
	0x74, 0x79, 0x18, 0x1a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x69, 0x74, 0x79, 0x12, 0x18,
	0x0a, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x17, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x65, 0x76, 0x69,
	0x63, 0x65, 0x18, 0x12, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18, 0x19, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x65, 0x6e, 0x74, 0x72,
	0x79, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x65, 0x6e,
	0x74, 0x72, 0x79, 0x50, 0x61, 0x67, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x1b, 0x0a,
	0x09, 0x65, 0x78, 0x69, 0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x65, 0x78, 0x69, 0x74, 0x50, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f,
	0x73, 0x74, 0x18, 0x1b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x0e,
	0x0a, 0x02, 0x6f, 0x73, 0x18, 0x15, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x6f, 0x73, 0x12, 0x1d,
	0x0a, 0x0a, 0x6f, 0x73, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x16, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x6f, 0x73, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a,
	0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x67,
	0x65, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x18, 0x0c, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x12, 0x16, 0x0a,
	0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x18, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72,
	0x65, 0x67, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18,
	0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x21, 0x0a,
	0x0c, 0x75, 0x74, 0x6d, 0x5f, 0x63, 0x61, 0x6d, 0x70, 0x61, 0x69, 0x67, 0x6e, 0x18, 0x0f, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x75, 0x74, 0x6d, 0x43, 0x61, 0x6d, 0x70, 0x61, 0x69, 0x67, 0x6e,
	0x12, 0x1f, 0x0a, 0x0b, 0x75, 0x74, 0x6d, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18,
	0x10, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x75, 0x74, 0x6d, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x75, 0x74, 0x6d, 0x5f, 0x6d, 0x65, 0x64, 0x69, 0x75, 0x6d, 0x18,
	0x0e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x75, 0x74, 0x6d, 0x4d, 0x65, 0x64, 0x69, 0x75, 0x6d,
	0x12, 0x1d, 0x0a, 0x0a, 0x75, 0x74, 0x6d, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x0d,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x75, 0x74, 0x6d, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12,
	0x19, 0x0a, 0x08, 0x75, 0x74, 0x6d, 0x5f, 0x74, 0x65, 0x72, 0x6d, 0x18, 0x11, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x75, 0x74, 0x6d, 0x54, 0x65, 0x72, 0x6d, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x65,
	0x6e, 0x61, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x1c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74,
	0x65, 0x6e, 0x61, 0x6e, 0x74, 0x49, 0x64, 0x42, 0x09, 0x0a, 0x07, 0x5f, 0x62, 0x6f, 0x75, 0x6e,
	0x63, 0x65, 0x42, 0x6a, 0x0a, 0x06, 0x63, 0x6f, 0x6d, 0x2e, 0x76, 0x32, 0x42, 0x09, 0x44, 0x61,
	0x74, 0x61, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2d, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x61, 0x6e, 0x61, 0x6c, 0x79,
	0x74, 0x69, 0x63, 0x73, 0x2f, 0x74, 0x73, 0x75, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x6f, 0x2f,
	0x76, 0x69, 0x6e, 0x63, 0x65, 0x2f, 0x76, 0x32, 0xa2, 0x02, 0x03, 0x56, 0x58, 0x58, 0xaa, 0x02,
	0x02, 0x56, 0x32, 0xca, 0x02, 0x02, 0x56, 0x32, 0xe2, 0x02, 0x0e, 0x56, 0x32, 0x5c, 0x47, 0x50,
	0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x02, 0x56, 0x32, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_vince_v2_data_proto_rawDescOnce sync.Once
	file_vince_v2_data_proto_rawDescData = file_vince_v2_data_proto_rawDesc
)

func file_vince_v2_data_proto_rawDescGZIP() []byte {
	file_vince_v2_data_proto_rawDescOnce.Do(func() {
		file_vince_v2_data_proto_rawDescData = protoimpl.X.CompressGZIP(file_vince_v2_data_proto_rawDescData)
	})
	return file_vince_v2_data_proto_rawDescData
}

var file_vince_v2_data_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_vince_v2_data_proto_goTypes = []interface{}{
	(*Data)(nil), // 0: v2.Data
}
var file_vince_v2_data_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_vince_v2_data_proto_init() }
func file_vince_v2_data_proto_init() {
	if File_vince_v2_data_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_vince_v2_data_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Data); i {
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
	file_vince_v2_data_proto_msgTypes[0].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_vince_v2_data_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_vince_v2_data_proto_goTypes,
		DependencyIndexes: file_vince_v2_data_proto_depIdxs,
		MessageInfos:      file_vince_v2_data_proto_msgTypes,
	}.Build()
	File_vince_v2_data_proto = out.File
	file_vince_v2_data_proto_rawDesc = nil
	file_vince_v2_data_proto_goTypes = nil
	file_vince_v2_data_proto_depIdxs = nil
}
