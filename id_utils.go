package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"

	// "encoding/hex"

	"github.com/coldstar-507/flatgen"
)

type Iddev = [RAW_NODE_ID_LEN + 4]byte
type Dev = [4]byte
type NodeId = [RAW_NODE_ID_LEN]byte
type Root = [RAW_ROOT_ID_LEN]byte

const (
	RAW_NODE_ID_LEN = 1 + 8 + 4
	RAW_ROOT_ID_LEN = 1 + 2*RAW_NODE_ID_LEN + 8 + 2
	RAW_IDDEV_LEN   = RAW_NODE_ID_LEN + 4

	RAW_MSG_ID_LEN = 1 + RAW_ROOT_ID_LEN + 8 + 4 + 1

	RAW_PUSH_ID_LEN        = 1 + RAW_NODE_ID_LEN + 4 + 8 + 4
	RAW_PUSH_ID_PREFIX_LEN = 1 + RAW_NODE_ID_LEN + 4

	RAW_MEDIA_ID_LEN  = 1 + 8 + 4 + 2 + 2 + 1 + 1
	RAW_MEDIA_REF_LEN = 1 + 8 + 2 + RAW_MEDIA_ID_LEN + 1

	RAW_TXID_LEN    = 32
	RAW_PAYMENT_REF = 1 + 8 + 2 + RAW_TXID_LEN
)

const (
	Unsent byte = iota
	Reaction
	Increment
	Chat
	Snip
	Placeholder
)

const (
	KIND_MESSAGE byte = iota
	KIND_ROOT
	KIND_MEDIA
	KIND_MEDIA_REF
	KIND_TX_REF
	KIND_NODE
	KIND_PUSH
	KIND_SNIP

	KIND_BOOST = 0x70
)

func RandU32() uint32 {
	b := make([]byte, 4)
	rand.Read(b)
	return binary.BigEndian.Uint32(b)
}

func writeNodeId(buf *bytes.Buffer, s *flatgen.NodeId) {
	buf.WriteByte(KIND_NODE)
	binary.Write(buf, binary.BigEndian, s.Timestamp())
	binary.Write(buf, binary.BigEndian, s.U32())
}

func WriteRoot(buf *bytes.Buffer, r *flatgen.Root) {
	buf.WriteByte(KIND_ROOT)
	writeNodeId(buf, r.Primary(nil))
	writeNodeId(buf, r.Secondary(nil))
	binary.Write(buf, binary.BigEndian, r.Timestamp())
	binary.Write(buf, binary.BigEndian, r.ChatPlace())
}

func WriteMediaId(buf *bytes.Buffer, id *flatgen.MediaId) int {
	binary.Write(buf, binary.BigEndian, KIND_MEDIA)
	binary.Write(buf, binary.BigEndian, id.Timestamp())
	binary.Write(buf, binary.BigEndian, id.U32())
	binary.Write(buf, binary.BigEndian, id.Width())
	binary.Write(buf, binary.BigEndian, id.Height())
	binary.Write(buf, binary.BigEndian, id.Squared())
	binary.Write(buf, binary.BigEndian, id.Video())
	return buf.Len()
}

func MakeRawMediaId(id *flatgen.MediaId) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, RAW_MEDIA_ID_LEN))
	WriteMediaId(buf, id)
	return buf.Bytes()
}

func MakeRawMediaRef(id *flatgen.MediaRef) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, RAW_MEDIA_REF_LEN))
	WriteMediaRef(buf, id)
	return buf.Bytes()
}

func WriteMediaRef(buf *bytes.Buffer, id *flatgen.MediaRef) int {
	binary.Write(buf, binary.BigEndian, KIND_MEDIA_REF)
	binary.Write(buf, binary.BigEndian, id.Timestamp())
	binary.Write(buf, binary.BigEndian, id.Place())
	WriteMediaId(buf, id.MediaId(nil))
	binary.Write(buf, binary.BigEndian, id.Permanent())
	return buf.Len()
}

func WriteMsgId(buf *bytes.Buffer, id *flatgen.MessageId) int {
	buf.WriteByte(id.Prefix())
	WriteRoot(buf, id.Root(nil))
	binary.Write(buf, binary.BigEndian, id.Timestamp())
	binary.Write(buf, binary.BigEndian, id.U32())
	buf.WriteByte(id.Suffix())
	return buf.Len()
}

// func WritePushRef(buf *bytes.Buffer, id *flatgen.PushRef) int {
// 	binary.Write(buf, binary.BigEndian, KIND_PUSH_REF)
// 	WritePushId(buf, id.PushId(nil))
// 	binary.Write(buf, binary.BigEndian, id.Timestamp())
// 	binary.Write(buf, binary.BigEndian, id.Place())
// 	return buf.Len()
// }

func WritePushId(buf *bytes.Buffer, id *flatgen.PushId) int {
	binary.Write(buf, binary.BigEndian, KIND_PUSH)
	writeNodeId(buf, id.NodeId(nil))
	binary.Write(buf, binary.BigEndian, id.Device())
	binary.Write(buf, binary.BigEndian, id.Timestamp())
	binary.Write(buf, binary.BigEndian, id.U32())
	return buf.Len()
}

func WritePushIdPrefixId(buf *bytes.Buffer, id *flatgen.PushId) int {
	binary.Write(buf, binary.BigEndian, KIND_PUSH)
	writeNodeId(buf, id.NodeId(nil))
	binary.Write(buf, binary.BigEndian, id.Device())
	return buf.Len()
}

func MakeRawRoot(r *flatgen.Root) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, RAW_ROOT_ID_LEN))
	WriteRoot(buf, r)
	return buf.Bytes()
}

func MakeRawMsgId(id *flatgen.MessageId) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, RAW_MSG_ID_LEN))
	WriteMsgId(buf, id)
	return buf.Bytes()
}

func MakeRawPushId(id *flatgen.PushId) []byte {
	bw := bytes.NewBuffer(make([]byte, 0, RAW_PUSH_ID_LEN))
	WritePushId(bw, id)
	return bw.Bytes()
}

func MsgIdPrefix(idRaw []byte) []byte {
	// Msg prefix byte + root
	return idRaw[:1+RAW_ROOT_ID_LEN]
}

func PushIdPrefix(raw []byte) []byte {
	return raw[:RAW_PUSH_ID_PREFIX_LEN]
}
