package singBox

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"github.com/v2rayA/v2rayA/core/singBox/inbound"
	"github.com/v2rayA/v2rayA/core/singBox/outbound"
)

type (
	jsoniterExtension struct {
		jsoniter.DummyExtension
	}
	jsoniterEmbedded struct {
		reflect2.Type
	}
)

var (
	inboundFormatType  = jsoniterEmbedded{reflect2.TypeOfPtr((*inbound.Format)(nil)).Elem()}
	outboundFormatType = jsoniterEmbedded{reflect2.TypeOfPtr((*outbound.Format)(nil)).Elem()}
)

func init() {
	jsoniter.RegisterExtension(new(jsoniterExtension))
}

func (ex *jsoniterExtension) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	switch typ.RType() {
	case inboundFormatType.RType():
		return inboundFormatType
	case outboundFormatType.RType():
		return outboundFormatType
	}
	return nil
}

func (embed jsoniterEmbedded) IsEmpty(ptr unsafe.Pointer) bool {
	return embed.UnsafeIndirect(ptr) == nil
}

func (embed jsoniterEmbedded) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	dup, release := dupStreamIndention(stream)
	defer release()
	dup.WriteObjectField(embed.Type1().Name())
	offset := dup.Buffered()
	data := stream.Buffer()
	stream.SetBuffer(data[:len(data)-offset])
	dup.WriteObjectStart()
	startOffset := dup.Buffered() - offset
	dup.WriteObjectEnd()
	endOffset := dup.Buffered() - startOffset - offset
	dup.Reset(nil)
	obj := embed.UnsafeIndirect(ptr)
	dup.WriteVal(obj)
	if dup.Error != nil {
		stream.Error = dup.Error
		return
	}
	data = dup.Buffer()
	stream.Write(data[startOffset : len(data)-endOffset])
}

func dupStreamIndention(stream *jsoniter.Stream) (*jsoniter.Stream, func()) {
	pool := stream.Pool()
	dup := pool.BorrowStream(nil)
	typ := reflect2.TypeOfPtr(stream).Elem().(*reflect2.UnsafeStructType)
	field, ok := typ.FieldByName("indention").(*reflect2.UnsafeStructField)
	if !ok {
		return dup, func() { pool.ReturnStream(dup) }
	}
	value := field.UnsafeGet(reflect2.PtrOf(stream))
	indention := *(*int)(value)
	if indention > 0 {
		value = field.UnsafeGet(reflect2.PtrOf(dup))
		*(*int)(value) = indention
	}
	return dup, func() {
		if indention > 0 {
			*(*int)(value) = 0
		}
		pool.ReturnStream(dup)
	}
}
