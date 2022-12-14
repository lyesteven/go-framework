// Code generated by protoc-gen-go. DO NOT EDIT.
// source: topicPubConns.proto

/*
Package topicPubConns is a generated protocol buffer package.

It is generated from these files:
	topicPubConns.proto

It has these top-level messages:
	Conn
	Allconns
	Delaymsg
*/
package topicPubConns

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

type Conn struct {
	ClusterID string `protobuf:"bytes,1,opt,name=clusterID" json:"clusterID,omitempty"`
}

func (m *Conn) Reset()                    { *m = Conn{} }
func (m *Conn) String() string            { return proto.CompactTextString(m) }
func (*Conn) ProtoMessage()               {}
func (*Conn) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Conn) GetClusterID() string {
	if m != nil {
		return m.ClusterID
	}
	return ""
}

type Allconns struct {
	Conns []*Conn `protobuf:"bytes,1,rep,name=conns" json:"conns,omitempty"`
}

func (m *Allconns) Reset()                    { *m = Allconns{} }
func (m *Allconns) String() string            { return proto.CompactTextString(m) }
func (*Allconns) ProtoMessage()               {}
func (*Allconns) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Allconns) GetConns() []*Conn {
	if m != nil {
		return m.Conns
	}
	return nil
}

type Delaymsg struct {
	DcID  int32  `protobuf:"varint,1,opt,name=dcID" json:"dcID,omitempty"`
	Topic string `protobuf:"bytes,2,opt,name=topic" json:"topic,omitempty"`
	Msg   []byte `protobuf:"bytes,3,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (m *Delaymsg) Reset()                    { *m = Delaymsg{} }
func (m *Delaymsg) String() string            { return proto.CompactTextString(m) }
func (*Delaymsg) ProtoMessage()               {}
func (*Delaymsg) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Delaymsg) GetDcID() int32 {
	if m != nil {
		return m.DcID
	}
	return 0
}

func (m *Delaymsg) GetTopic() string {
	if m != nil {
		return m.Topic
	}
	return ""
}

func (m *Delaymsg) GetMsg() []byte {
	if m != nil {
		return m.Msg
	}
	return nil
}

func init() {
	proto.RegisterType((*Conn)(nil), "topicPubConns.conn")
	proto.RegisterType((*Allconns)(nil), "topicPubConns.allconns")
	proto.RegisterType((*Delaymsg)(nil), "topicPubConns.delaymsg")
}

func init() { proto.RegisterFile("topicPubConns.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 168 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2e, 0xc9, 0x2f, 0xc8,
	0x4c, 0x0e, 0x28, 0x4d, 0x72, 0xce, 0xcf, 0xcb, 0x2b, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0xe2, 0x45, 0x11, 0x54, 0x52, 0xe1, 0x62, 0x49, 0xce, 0xcf, 0xcb, 0x13, 0x92, 0xe1, 0xe2, 0x4c,
	0xce, 0x29, 0x2d, 0x2e, 0x49, 0x2d, 0xf2, 0x74, 0x91, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x42,
	0x08, 0x28, 0x99, 0x72, 0x71, 0x24, 0xe6, 0xe4, 0x80, 0x14, 0x16, 0x0b, 0x69, 0x72, 0xb1, 0x82,
	0x19, 0x12, 0x8c, 0x0a, 0xcc, 0x1a, 0xdc, 0x46, 0xc2, 0x7a, 0xa8, 0xb6, 0x80, 0xe4, 0x82, 0x20,
	0x2a, 0x94, 0xdc, 0xb8, 0x38, 0x52, 0x52, 0x73, 0x12, 0x2b, 0x73, 0x8b, 0xd3, 0x85, 0x84, 0xb8,
	0x58, 0x52, 0x92, 0xa1, 0x66, 0xb3, 0x06, 0x81, 0xd9, 0x42, 0x22, 0x5c, 0xac, 0x60, 0xcd, 0x12,
	0x4c, 0x60, 0x0b, 0x21, 0x1c, 0x21, 0x01, 0x2e, 0xe6, 0xdc, 0xe2, 0x74, 0x09, 0x66, 0x05, 0x46,
	0x0d, 0x9e, 0x20, 0x10, 0x33, 0x89, 0x0d, 0xec, 0x74, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x9f, 0x34, 0x71, 0x7d, 0xd1, 0x00, 0x00, 0x00,
}
