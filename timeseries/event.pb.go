// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: event.proto

package timeseries

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

type PROPS int32

const (
	PROPS_base           PROPS = 0
	PROPS_event          PROPS = 1
	PROPS_page           PROPS = 2
	PROPS_entryPage      PROPS = 3
	PROPS_exitPage       PROPS = 4
	PROPS_referrer       PROPS = 5
	PROPS_utmMedium      PROPS = 6
	PROPS_utmSource      PROPS = 7
	PROPS_utmCampaign    PROPS = 8
	PROPS_utmContent     PROPS = 9
	PROPS_utmTerm        PROPS = 10
	PROPS_utmDevice      PROPS = 11
	PROPS_utmBrowser     PROPS = 12
	PROPS_browserVersion PROPS = 13
	PROPS_os             PROPS = 14
	PROPS_osVersion      PROPS = 15
	PROPS_country        PROPS = 16
	PROPS_region         PROPS = 17
	PROPS_city           PROPS = 18
)

// Enum value maps for PROPS.
var (
	PROPS_name = map[int32]string{
		0:  "base",
		1:  "event",
		2:  "page",
		3:  "entryPage",
		4:  "exitPage",
		5:  "referrer",
		6:  "utmMedium",
		7:  "utmSource",
		8:  "utmCampaign",
		9:  "utmContent",
		10: "utmTerm",
		11: "utmDevice",
		12: "utmBrowser",
		13: "browserVersion",
		14: "os",
		15: "osVersion",
		16: "country",
		17: "region",
		18: "city",
	}
	PROPS_value = map[string]int32{
		"base":           0,
		"event":          1,
		"page":           2,
		"entryPage":      3,
		"exitPage":       4,
		"referrer":       5,
		"utmMedium":      6,
		"utmSource":      7,
		"utmCampaign":    8,
		"utmContent":     9,
		"utmTerm":        10,
		"utmDevice":      11,
		"utmBrowser":     12,
		"browserVersion": 13,
		"os":             14,
		"osVersion":      15,
		"country":        16,
		"region":         17,
		"city":           18,
	}
)

func (x PROPS) Enum() *PROPS {
	p := new(PROPS)
	*p = x
	return p
}

func (x PROPS) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PROPS) Descriptor() protoreflect.EnumDescriptor {
	return file_event_proto_enumTypes[0].Descriptor()
}

func (PROPS) Type() protoreflect.EnumType {
	return &file_event_proto_enumTypes[0]
}

func (x PROPS) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PROPS.Descriptor instead.
func (PROPS) EnumDescriptor() ([]byte, []int) {
	return file_event_proto_rawDescGZIP(), []int{0}
}

type METRIC_TYPE int32

const (
	METRIC_TYPE_visitors      METRIC_TYPE = 0
	METRIC_TYPE_views         METRIC_TYPE = 1
	METRIC_TYPE_events        METRIC_TYPE = 2
	METRIC_TYPE_visits        METRIC_TYPE = 3
	METRIC_TYPE_bounce_rate   METRIC_TYPE = 4
	METRIC_TYPE_visitDuration METRIC_TYPE = 5
	METRIC_TYPE_viewsPerVisit METRIC_TYPE = 6
)

// Enum value maps for METRIC_TYPE.
var (
	METRIC_TYPE_name = map[int32]string{
		0: "visitors",
		1: "views",
		2: "events",
		3: "visits",
		4: "bounce_rate",
		5: "visitDuration",
		6: "viewsPerVisit",
	}
	METRIC_TYPE_value = map[string]int32{
		"visitors":      0,
		"views":         1,
		"events":        2,
		"visits":        3,
		"bounce_rate":   4,
		"visitDuration": 5,
		"viewsPerVisit": 6,
	}
)

func (x METRIC_TYPE) Enum() *METRIC_TYPE {
	p := new(METRIC_TYPE)
	*p = x
	return p
}

func (x METRIC_TYPE) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (METRIC_TYPE) Descriptor() protoreflect.EnumDescriptor {
	return file_event_proto_enumTypes[1].Descriptor()
}

func (METRIC_TYPE) Type() protoreflect.EnumType {
	return &file_event_proto_enumTypes[1]
}

func (x METRIC_TYPE) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use METRIC_TYPE.Descriptor instead.
func (METRIC_TYPE) EnumDescriptor() ([]byte, []int) {
	return file_event_proto_rawDescGZIP(), []int{1}
}

type Entries struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Events []*Entry `protobuf:"bytes,1,rep,name=events,proto3" json:"events,omitempty"`
}

func (x *Entries) Reset() {
	*x = Entries{}
	if protoimpl.UnsafeEnabled {
		mi := &file_event_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entries) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entries) ProtoMessage() {}

func (x *Entries) ProtoReflect() protoreflect.Message {
	mi := &file_event_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Entries.ProtoReflect.Descriptor instead.
func (*Entries) Descriptor() ([]byte, []int) {
	return file_event_proto_rawDescGZIP(), []int{0}
}

func (x *Entries) GetEvents() []*Entry {
	if x != nil {
		return x.Events
	}
	return nil
}

type Entry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp              int64    `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Name                   string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Domain                 string   `protobuf:"bytes,3,opt,name=domain,proto3" json:"domain,omitempty"`
	Id                     *Session `protobuf:"bytes,5,opt,name=id,proto3" json:"id,omitempty"`
	Hostname               string   `protobuf:"bytes,6,opt,name=hostname,proto3" json:"hostname,omitempty"`
	Pathname               string   `protobuf:"bytes,7,opt,name=pathname,proto3" json:"pathname,omitempty"`
	Referrer               string   `protobuf:"bytes,8,opt,name=referrer,proto3" json:"referrer,omitempty"`
	ReferrerSource         string   `protobuf:"bytes,9,opt,name=referrer_source,json=referrerSource,proto3" json:"referrer_source,omitempty"`
	CountryCode            string   `protobuf:"bytes,10,opt,name=country_code,json=countryCode,proto3" json:"country_code,omitempty"`
	Subdivision1Code       string   `protobuf:"bytes,11,opt,name=subdivision1_code,json=subdivision1Code,proto3" json:"subdivision1_code,omitempty"`
	Subdivision2Code       string   `protobuf:"bytes,12,opt,name=subdivision2_code,json=subdivision2Code,proto3" json:"subdivision2_code,omitempty"`
	CityGeoNameId          uint32   `protobuf:"varint,13,opt,name=city_geo_name_id,json=cityGeoNameId,proto3" json:"city_geo_name_id,omitempty"`
	ScreenSize             string   `protobuf:"bytes,14,opt,name=screen_size,json=screenSize,proto3" json:"screen_size,omitempty"`
	OperatingSystem        string   `protobuf:"bytes,15,opt,name=operating_system,json=operatingSystem,proto3" json:"operating_system,omitempty"`
	Browser                string   `protobuf:"bytes,16,opt,name=browser,proto3" json:"browser,omitempty"`
	UtmMedium              string   `protobuf:"bytes,17,opt,name=utm_medium,json=utmMedium,proto3" json:"utm_medium,omitempty"`
	UtmSource              string   `protobuf:"bytes,18,opt,name=utm_source,json=utmSource,proto3" json:"utm_source,omitempty"`
	UtmCampaign            string   `protobuf:"bytes,19,opt,name=utm_campaign,json=utmCampaign,proto3" json:"utm_campaign,omitempty"`
	BrowserVersion         string   `protobuf:"bytes,20,opt,name=browser_version,json=browserVersion,proto3" json:"browser_version,omitempty"`
	OperatingSystemVersion string   `protobuf:"bytes,21,opt,name=operating_system_version,json=operatingSystemVersion,proto3" json:"operating_system_version,omitempty"`
	UtmContent             string   `protobuf:"bytes,22,opt,name=utm_content,json=utmContent,proto3" json:"utm_content,omitempty"`
	UtmTerm                string   `protobuf:"bytes,23,opt,name=utm_term,json=utmTerm,proto3" json:"utm_term,omitempty"`
	TransferredFrom        string   `protobuf:"bytes,24,opt,name=transferred_from,json=transferredFrom,proto3" json:"transferred_from,omitempty"`
	EntryPage              string   `protobuf:"bytes,25,opt,name=entry_page,json=entryPage,proto3" json:"entry_page,omitempty"`
	ExitPage               string   `protobuf:"bytes,26,opt,name=exit_page,json=exitPage,proto3" json:"exit_page,omitempty"`
	PageViews              int32    `protobuf:"varint,27,opt,name=page_views,json=pageViews,proto3" json:"page_views,omitempty"`
	Events                 int32    `protobuf:"varint,28,opt,name=events,proto3" json:"events,omitempty"`
	Sign                   int32    `protobuf:"varint,29,opt,name=sign,proto3" json:"sign,omitempty"`
	IsBounce               bool     `protobuf:"varint,30,opt,name=is_bounce,json=isBounce,proto3" json:"is_bounce,omitempty"`
	Duration               float64  `protobuf:"fixed64,31,opt,name=duration,proto3" json:"duration,omitempty"`
	Start                  int64    `protobuf:"varint,32,opt,name=start,proto3" json:"start,omitempty"`
	HourIndex              int32    `protobuf:"varint,33,opt,name=hour_index,json=hourIndex,proto3" json:"hour_index,omitempty"`
}

func (x *Entry) Reset() {
	*x = Entry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_event_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entry) ProtoMessage() {}

func (x *Entry) ProtoReflect() protoreflect.Message {
	mi := &file_event_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Entry.ProtoReflect.Descriptor instead.
func (*Entry) Descriptor() ([]byte, []int) {
	return file_event_proto_rawDescGZIP(), []int{1}
}

func (x *Entry) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *Entry) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Entry) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

func (x *Entry) GetId() *Session {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Entry) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *Entry) GetPathname() string {
	if x != nil {
		return x.Pathname
	}
	return ""
}

func (x *Entry) GetReferrer() string {
	if x != nil {
		return x.Referrer
	}
	return ""
}

func (x *Entry) GetReferrerSource() string {
	if x != nil {
		return x.ReferrerSource
	}
	return ""
}

func (x *Entry) GetCountryCode() string {
	if x != nil {
		return x.CountryCode
	}
	return ""
}

func (x *Entry) GetSubdivision1Code() string {
	if x != nil {
		return x.Subdivision1Code
	}
	return ""
}

func (x *Entry) GetSubdivision2Code() string {
	if x != nil {
		return x.Subdivision2Code
	}
	return ""
}

func (x *Entry) GetCityGeoNameId() uint32 {
	if x != nil {
		return x.CityGeoNameId
	}
	return 0
}

func (x *Entry) GetScreenSize() string {
	if x != nil {
		return x.ScreenSize
	}
	return ""
}

func (x *Entry) GetOperatingSystem() string {
	if x != nil {
		return x.OperatingSystem
	}
	return ""
}

func (x *Entry) GetBrowser() string {
	if x != nil {
		return x.Browser
	}
	return ""
}

func (x *Entry) GetUtmMedium() string {
	if x != nil {
		return x.UtmMedium
	}
	return ""
}

func (x *Entry) GetUtmSource() string {
	if x != nil {
		return x.UtmSource
	}
	return ""
}

func (x *Entry) GetUtmCampaign() string {
	if x != nil {
		return x.UtmCampaign
	}
	return ""
}

func (x *Entry) GetBrowserVersion() string {
	if x != nil {
		return x.BrowserVersion
	}
	return ""
}

func (x *Entry) GetOperatingSystemVersion() string {
	if x != nil {
		return x.OperatingSystemVersion
	}
	return ""
}

func (x *Entry) GetUtmContent() string {
	if x != nil {
		return x.UtmContent
	}
	return ""
}

func (x *Entry) GetUtmTerm() string {
	if x != nil {
		return x.UtmTerm
	}
	return ""
}

func (x *Entry) GetTransferredFrom() string {
	if x != nil {
		return x.TransferredFrom
	}
	return ""
}

func (x *Entry) GetEntryPage() string {
	if x != nil {
		return x.EntryPage
	}
	return ""
}

func (x *Entry) GetExitPage() string {
	if x != nil {
		return x.ExitPage
	}
	return ""
}

func (x *Entry) GetPageViews() int32 {
	if x != nil {
		return x.PageViews
	}
	return 0
}

func (x *Entry) GetEvents() int32 {
	if x != nil {
		return x.Events
	}
	return 0
}

func (x *Entry) GetSign() int32 {
	if x != nil {
		return x.Sign
	}
	return 0
}

func (x *Entry) GetIsBounce() bool {
	if x != nil {
		return x.IsBounce
	}
	return false
}

func (x *Entry) GetDuration() float64 {
	if x != nil {
		return x.Duration
	}
	return 0
}

func (x *Entry) GetStart() int64 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *Entry) GetHourIndex() int32 {
	if x != nil {
		return x.HourIndex
	}
	return 0
}

type Session struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId    uint64 `protobuf:"varint,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	SessionId uint64 `protobuf:"varint,2,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
}

func (x *Session) Reset() {
	*x = Session{}
	if protoimpl.UnsafeEnabled {
		mi := &file_event_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Session) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Session) ProtoMessage() {}

func (x *Session) ProtoReflect() protoreflect.Message {
	mi := &file_event_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Session.ProtoReflect.Descriptor instead.
func (*Session) Descriptor() ([]byte, []int) {
	return file_event_proto_rawDescGZIP(), []int{2}
}

func (x *Session) GetUserId() uint64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

func (x *Session) GetSessionId() uint64 {
	if x != nil {
		return x.SessionId
	}
	return 0
}

var File_event_proto protoreflect.FileDescriptor

var file_event_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x65, 0x72, 0x69, 0x65, 0x73, 0x22, 0x34, 0x0a, 0x07, 0x45, 0x6e, 0x74,
	0x72, 0x69, 0x65, 0x73, 0x12, 0x29, 0x0a, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x65, 0x72, 0x69, 0x65,
	0x73, 0x2e, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x22,
	0x9f, 0x08, 0x0a, 0x05, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x64,
	0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x64, 0x6f, 0x6d,
	0x61, 0x69, 0x6e, 0x12, 0x23, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x65, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x53, 0x65, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x73, 0x74,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x74, 0x68, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x61, 0x74, 0x68, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x12, 0x27, 0x0a, 0x0f,
	0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18,
	0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x53,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79,
	0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x72, 0x79, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x2b, 0x0a, 0x11, 0x73, 0x75, 0x62, 0x64,
	0x69, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x31, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x0b, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x10, 0x73, 0x75, 0x62, 0x64, 0x69, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e,
	0x31, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x2b, 0x0a, 0x11, 0x73, 0x75, 0x62, 0x64, 0x69, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x32, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x10, 0x73, 0x75, 0x62, 0x64, 0x69, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x32, 0x43, 0x6f,
	0x64, 0x65, 0x12, 0x27, 0x0a, 0x10, 0x63, 0x69, 0x74, 0x79, 0x5f, 0x67, 0x65, 0x6f, 0x5f, 0x6e,
	0x61, 0x6d, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0d, 0x63, 0x69,
	0x74, 0x79, 0x47, 0x65, 0x6f, 0x4e, 0x61, 0x6d, 0x65, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x73,
	0x63, 0x72, 0x65, 0x65, 0x6e, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x73, 0x63, 0x72, 0x65, 0x65, 0x6e, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x29, 0x0a, 0x10,
	0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d,
	0x18, 0x0f, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6e,
	0x67, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x12, 0x18, 0x0a, 0x07, 0x62, 0x72, 0x6f, 0x77, 0x73,
	0x65, 0x72, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x62, 0x72, 0x6f, 0x77, 0x73, 0x65,
	0x72, 0x12, 0x1d, 0x0a, 0x0a, 0x75, 0x74, 0x6d, 0x5f, 0x6d, 0x65, 0x64, 0x69, 0x75, 0x6d, 0x18,
	0x11, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x75, 0x74, 0x6d, 0x4d, 0x65, 0x64, 0x69, 0x75, 0x6d,
	0x12, 0x1d, 0x0a, 0x0a, 0x75, 0x74, 0x6d, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x12,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x75, 0x74, 0x6d, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12,
	0x21, 0x0a, 0x0c, 0x75, 0x74, 0x6d, 0x5f, 0x63, 0x61, 0x6d, 0x70, 0x61, 0x69, 0x67, 0x6e, 0x18,
	0x13, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x75, 0x74, 0x6d, 0x43, 0x61, 0x6d, 0x70, 0x61, 0x69,
	0x67, 0x6e, 0x12, 0x27, 0x0a, 0x0f, 0x62, 0x72, 0x6f, 0x77, 0x73, 0x65, 0x72, 0x5f, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x14, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x62, 0x72, 0x6f,
	0x77, 0x73, 0x65, 0x72, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x38, 0x0a, 0x18, 0x6f,
	0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x5f,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x15, 0x20, 0x01, 0x28, 0x09, 0x52, 0x16, 0x6f,
	0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1f, 0x0a, 0x0b, 0x75, 0x74, 0x6d, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x18, 0x16, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x75, 0x74, 0x6d, 0x43,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x75, 0x74, 0x6d, 0x5f, 0x74, 0x65,
	0x72, 0x6d, 0x18, 0x17, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x75, 0x74, 0x6d, 0x54, 0x65, 0x72,
	0x6d, 0x12, 0x29, 0x0a, 0x10, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x72, 0x65, 0x64,
	0x5f, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x18, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x65, 0x72, 0x72, 0x65, 0x64, 0x46, 0x72, 0x6f, 0x6d, 0x12, 0x1d, 0x0a, 0x0a,
	0x65, 0x6e, 0x74, 0x72, 0x79, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x18, 0x19, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x50, 0x61, 0x67, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x65,
	0x78, 0x69, 0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x18, 0x1a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x65, 0x78, 0x69, 0x74, 0x50, 0x61, 0x67, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x61, 0x67, 0x65,
	0x5f, 0x76, 0x69, 0x65, 0x77, 0x73, 0x18, 0x1b, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x70, 0x61,
	0x67, 0x65, 0x56, 0x69, 0x65, 0x77, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x73, 0x18, 0x1c, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12,
	0x12, 0x0a, 0x04, 0x73, 0x69, 0x67, 0x6e, 0x18, 0x1d, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x73,
	0x69, 0x67, 0x6e, 0x12, 0x1b, 0x0a, 0x09, 0x69, 0x73, 0x5f, 0x62, 0x6f, 0x75, 0x6e, 0x63, 0x65,
	0x18, 0x1e, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x69, 0x73, 0x42, 0x6f, 0x75, 0x6e, 0x63, 0x65,
	0x12, 0x1a, 0x0a, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x1f, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05,
	0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x20, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x68, 0x6f, 0x75, 0x72, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78,
	0x18, 0x21, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x68, 0x6f, 0x75, 0x72, 0x49, 0x6e, 0x64, 0x65,
	0x78, 0x22, 0x41, 0x0a, 0x07, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x17, 0x0a, 0x07,
	0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x75,
	0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x73, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x49, 0x64, 0x2a, 0x8a, 0x02, 0x0a, 0x05, 0x50, 0x52, 0x4f, 0x50, 0x53, 0x12, 0x08,
	0x0a, 0x04, 0x62, 0x61, 0x73, 0x65, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x10, 0x02, 0x12, 0x0d, 0x0a,
	0x09, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x50, 0x61, 0x67, 0x65, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08,
	0x65, 0x78, 0x69, 0x74, 0x50, 0x61, 0x67, 0x65, 0x10, 0x04, 0x12, 0x0c, 0x0a, 0x08, 0x72, 0x65,
	0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x10, 0x05, 0x12, 0x0d, 0x0a, 0x09, 0x75, 0x74, 0x6d, 0x4d,
	0x65, 0x64, 0x69, 0x75, 0x6d, 0x10, 0x06, 0x12, 0x0d, 0x0a, 0x09, 0x75, 0x74, 0x6d, 0x53, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x10, 0x07, 0x12, 0x0f, 0x0a, 0x0b, 0x75, 0x74, 0x6d, 0x43, 0x61, 0x6d,
	0x70, 0x61, 0x69, 0x67, 0x6e, 0x10, 0x08, 0x12, 0x0e, 0x0a, 0x0a, 0x75, 0x74, 0x6d, 0x43, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x10, 0x09, 0x12, 0x0b, 0x0a, 0x07, 0x75, 0x74, 0x6d, 0x54, 0x65,
	0x72, 0x6d, 0x10, 0x0a, 0x12, 0x0d, 0x0a, 0x09, 0x75, 0x74, 0x6d, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x10, 0x0b, 0x12, 0x0e, 0x0a, 0x0a, 0x75, 0x74, 0x6d, 0x42, 0x72, 0x6f, 0x77, 0x73, 0x65,
	0x72, 0x10, 0x0c, 0x12, 0x12, 0x0a, 0x0e, 0x62, 0x72, 0x6f, 0x77, 0x73, 0x65, 0x72, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x10, 0x0d, 0x12, 0x06, 0x0a, 0x02, 0x6f, 0x73, 0x10, 0x0e, 0x12,
	0x0d, 0x0a, 0x09, 0x6f, 0x73, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x10, 0x0f, 0x12, 0x0b,
	0x0a, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x10, 0x10, 0x12, 0x0a, 0x0a, 0x06, 0x72,
	0x65, 0x67, 0x69, 0x6f, 0x6e, 0x10, 0x11, 0x12, 0x08, 0x0a, 0x04, 0x63, 0x69, 0x74, 0x79, 0x10,
	0x12, 0x2a, 0x75, 0x0a, 0x0b, 0x4d, 0x45, 0x54, 0x52, 0x49, 0x43, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x12, 0x0c, 0x0a, 0x08, 0x76, 0x69, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x73, 0x10, 0x00, 0x12, 0x09,
	0x0a, 0x05, 0x76, 0x69, 0x65, 0x77, 0x73, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x73, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x76, 0x69, 0x73, 0x69, 0x74, 0x73, 0x10,
	0x03, 0x12, 0x0f, 0x0a, 0x0b, 0x62, 0x6f, 0x75, 0x6e, 0x63, 0x65, 0x5f, 0x72, 0x61, 0x74, 0x65,
	0x10, 0x04, 0x12, 0x11, 0x0a, 0x0d, 0x76, 0x69, 0x73, 0x69, 0x74, 0x44, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x10, 0x05, 0x12, 0x11, 0x0a, 0x0d, 0x76, 0x69, 0x65, 0x77, 0x73, 0x50, 0x65,
	0x72, 0x56, 0x69, 0x73, 0x69, 0x74, 0x10, 0x06, 0x42, 0x25, 0x5a, 0x23, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x65, 0x72, 0x6e, 0x65, 0x73, 0x74, 0x2f, 0x76,
	0x69, 0x6e, 0x63, 0x65, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x65, 0x72, 0x69, 0x65, 0x73, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_event_proto_rawDescOnce sync.Once
	file_event_proto_rawDescData = file_event_proto_rawDesc
)

func file_event_proto_rawDescGZIP() []byte {
	file_event_proto_rawDescOnce.Do(func() {
		file_event_proto_rawDescData = protoimpl.X.CompressGZIP(file_event_proto_rawDescData)
	})
	return file_event_proto_rawDescData
}

var file_event_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_event_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_event_proto_goTypes = []interface{}{
	(PROPS)(0),       // 0: timeseries.PROPS
	(METRIC_TYPE)(0), // 1: timeseries.METRIC_TYPE
	(*Entries)(nil),  // 2: timeseries.Entries
	(*Entry)(nil),    // 3: timeseries.Entry
	(*Session)(nil),  // 4: timeseries.Session
}
var file_event_proto_depIdxs = []int32{
	3, // 0: timeseries.Entries.events:type_name -> timeseries.Entry
	4, // 1: timeseries.Entry.id:type_name -> timeseries.Session
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_event_proto_init() }
func file_event_proto_init() {
	if File_event_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_event_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Entries); i {
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
		file_event_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Entry); i {
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
		file_event_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Session); i {
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
			RawDescriptor: file_event_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_event_proto_goTypes,
		DependencyIndexes: file_event_proto_depIdxs,
		EnumInfos:         file_event_proto_enumTypes,
		MessageInfos:      file_event_proto_msgTypes,
	}.Build()
	File_event_proto = out.File
	file_event_proto_rawDesc = nil
	file_event_proto_goTypes = nil
	file_event_proto_depIdxs = nil
}
