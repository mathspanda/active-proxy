// Code generated by protoc-gen-go.
// source: HAZKInfo.proto
// DO NOT EDIT!

/*
Package hadoop_hdfs is a generated protocol buffer package.

It is generated from these files:
	HAZKInfo.proto

It has these top-level messages:
	ActiveNodeInfo
*/
package hadoop_hdfs

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ActiveNodeInfo struct {
	NameserviceId    *string `protobuf:"bytes,1,req,name=nameserviceId" json:"nameserviceId,omitempty"`
	NamenodeId       *string `protobuf:"bytes,2,req,name=namenodeId" json:"namenodeId,omitempty"`
	Hostname         *string `protobuf:"bytes,3,req,name=hostname" json:"hostname,omitempty"`
	Port             *int32  `protobuf:"varint,4,req,name=port" json:"port,omitempty"`
	ZkfcPort         *int32  `protobuf:"varint,5,req,name=zkfcPort" json:"zkfcPort,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *ActiveNodeInfo) Reset()                    { *m = ActiveNodeInfo{} }
func (m *ActiveNodeInfo) String() string            { return proto.CompactTextString(m) }
func (*ActiveNodeInfo) ProtoMessage()               {}
func (*ActiveNodeInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ActiveNodeInfo) GetNameserviceId() string {
	if m != nil && m.NameserviceId != nil {
		return *m.NameserviceId
	}
	return ""
}

func (m *ActiveNodeInfo) GetNamenodeId() string {
	if m != nil && m.NamenodeId != nil {
		return *m.NamenodeId
	}
	return ""
}

func (m *ActiveNodeInfo) GetHostname() string {
	if m != nil && m.Hostname != nil {
		return *m.Hostname
	}
	return ""
}

func (m *ActiveNodeInfo) GetPort() int32 {
	if m != nil && m.Port != nil {
		return *m.Port
	}
	return 0
}

func (m *ActiveNodeInfo) GetZkfcPort() int32 {
	if m != nil && m.ZkfcPort != nil {
		return *m.ZkfcPort
	}
	return 0
}

func init() {
	proto.RegisterType((*ActiveNodeInfo)(nil), "hadoop.hdfs.ActiveNodeInfo")
}

func init() { proto.RegisterFile("HAZKInfo.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 196 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0xf3, 0x70, 0x8c, 0xf2,
	0xf6, 0xcc, 0x4b, 0xcb, 0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0xce, 0x48, 0x4c, 0xc9,
	0xcf, 0x2f, 0xd0, 0xcb, 0x48, 0x49, 0x2b, 0x56, 0x5a, 0xc4, 0xc8, 0xc5, 0xe7, 0x98, 0x5c, 0x92,
	0x59, 0x96, 0xea, 0x97, 0x9f, 0x92, 0x0a, 0x52, 0x25, 0xa4, 0xc2, 0xc5, 0x9b, 0x97, 0x98, 0x9b,
	0x5a, 0x9c, 0x5a, 0x54, 0x96, 0x99, 0x9c, 0xea, 0x99, 0x22, 0xc1, 0xa8, 0xc0, 0xa4, 0xc1, 0x19,
	0x84, 0x2a, 0x28, 0x24, 0xc7, 0xc5, 0x05, 0x12, 0xc8, 0x03, 0xe9, 0x4a, 0x91, 0x60, 0x02, 0x2b,
	0x41, 0x12, 0x11, 0x92, 0xe2, 0xe2, 0xc8, 0xc8, 0x2f, 0x2e, 0x01, 0x89, 0x48, 0x30, 0x83, 0x65,
	0xe1, 0x7c, 0x21, 0x21, 0x2e, 0x96, 0x82, 0xfc, 0xa2, 0x12, 0x09, 0x16, 0x05, 0x26, 0x0d, 0xd6,
	0x20, 0x30, 0x1b, 0xa4, 0xbe, 0x2a, 0x3b, 0x2d, 0x39, 0x00, 0x24, 0xce, 0x0a, 0x16, 0x87, 0xf3,
	0x9d, 0x1c, 0xb9, 0xf4, 0xf3, 0x8b, 0xd2, 0xf5, 0x12, 0x0b, 0x12, 0x93, 0x33, 0x52, 0xf5, 0x90,
	0x9c, 0xaf, 0x07, 0x72, 0x4f, 0x6a, 0x91, 0x1e, 0xcc, 0x5e, 0xbd, 0x8c, 0x44, 0x88, 0x27, 0x9d,
	0xe0, 0x9e, 0x0e, 0x00, 0x71, 0x8b, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x40, 0x9f, 0xf6, 0xb4,
	0x05, 0x01, 0x00, 0x00,
}
