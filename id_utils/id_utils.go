package id_utils

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io"
	"strings"
	"time"

	"github.com/coldstar-507/flatgen"
	"github.com/coldstar-507/utils/utils"
)

// type Iddev = [RAW_NODE_ID_LEN + 4]byte
type Iddev_ = [RAW_IDDEV_LEN_]byte
type Dev = [4]byte
type NodeId = [RAW_NODE_ID_LEN]byte
type MsgId = [RAW_MSG_ID_LEN]byte
type Root = [RAW_ROOT_ID_LEN]byte

const (
	RAW_NODE_ID_LEN        = 1 + 8 + 4                             // 13
	RAW_ROOT_ID_LEN        = 1 + 2*RAW_NODE_ID_LEN + 8 + 4 + 2 + 1 // 42
	RAW_IDDEV_LEN_         = 1 + RAW_NODE_ID_LEN + 4               // 18
	RAW_MSG_ID_LEN         = 1 + RAW_ROOT_ID_LEN + 8 + 4 + 1       // 56
	RAW_PUSH_ID_LEN        = 1 + RAW_NODE_ID_LEN + 4 + 8 + 4       // 30
	RAW_PUSH_ID_PREFIX_LEN = 1 + RAW_NODE_ID_LEN + 4               // 18
	RAW_MEDIA_ID_LEN       = 1 + 8 + 4 + 4 + 2                     // 19
	RAW_MEDIA_REF_LEN      = 1 + 8 + 2 + RAW_MEDIA_ID_LEN + 1      // 31
	RAW_STICKER_ID_LEN     = RAW_MEDIA_REF_LEN + RAW_NODE_ID_LEN   // 44
	RAW_TXID_LEN           = 1 + 32                                // 33
	RAW_PAYMENT_ID_LEN     = 1 + RAW_TXID_LEN                      // 34
	RAW_PAYMENT_REF        = 1 + 8 + RAW_PAYMENT_ID_LEN + 2 + 1    // 46
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

	KIND_IDDEV   = byte(0x10)
	KIND_STICKER = byte(0x11)

	KIND_BOOST = byte(0x70)
)

var NodeIdZero = &flatgen.NodeIdT{
	Prefix:    KIND_NODE,
	Timestamp: 0,
	U32:       0,
}

func MakeRoot(n1, n2 *flatgen.NodeIdT, chatPlace uint16) *flatgen.RootT {
	var primary, secondary *flatgen.NodeIdT
	if n1.Timestamp > n2.Timestamp {
		primary = n1
		secondary = n2
	} else {
		primary = n2
		secondary = n1
	}
	return &flatgen.RootT{
		Prefix:    KIND_ROOT,
		Primary:   primary,
		Secondary: secondary,
		Timestamp: utils.MakeTimestamp(),
		ChatPlace: chatPlace,
	}
}

func RawRoot(r *flatgen.RootT) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, RAW_ROOT_ID_LEN))
	WriteRootT(buf, r)
	return buf.Bytes()
}

func RawRoot2(r *flatgen.RootT) Root {
	_root := Root{}
	buf := bytes.NewBuffer(_root[:0])
	WriteRootT(buf, r)
	return _root
}

func HexRoot(r *flatgen.RootT) string {
	return hex.EncodeToString(RawRoot(r))
}

func RawNodeId(id *flatgen.NodeIdT) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, RAW_NODE_ID_LEN))
	WriteNodeIdT(buf, id)
	return buf.Bytes()
}

func ReadRawMediaId(r io.Reader) *flatgen.MediaIdT {
	mr := &flatgen.MediaIdT{}
	utils.ReadBin(r, &mr.Prefix, &mr.Timestamp, &mr.U32, &mr.AspectRatio, &mr.MediaType)
	return mr
}

func ReadRawMediaRef(r io.Reader) *flatgen.MediaRefT {
	mr := &flatgen.MediaRefT{}
	utils.ReadBin(r, &mr.Prefix)
	mr.MediaId = ReadRawMediaId(r)
	utils.ReadBin(r, &mr.Place, &mr.Permanent)
	return mr
}

func HexNodeId(id *flatgen.NodeIdT) string {
	return hex.EncodeToString(RawNodeId(id))
}

func WriteMediaReference(w io.Writer, r *flatgen.MediaReference) {
	utils.WriteBin(w, r.Prefix(), r.Timestamp(), r.BodyBytes(), r.Place(), r.Perm())
}

func WriteRootT(w io.Writer, r *flatgen.RootT) {
	utils.WriteBin(w, r.Prefix)
	WriteNodeIdT(w, r.Primary)
	WriteNodeIdT(w, r.Secondary)
	utils.WriteBin(w, r.Timestamp, r.U32, r.ChatPlace, r.Confirmed)
}

func WriteChatIdT(w io.Writer, r *flatgen.MessageIdT) {
	utils.WriteBin(w, r.Prefix)
	WriteRootT(w, r.Root)
	utils.WriteBin(w, r.Timestamp, r.U32, r.Suffix)
}

func WriteNodeIdT(w io.Writer, n *flatgen.NodeIdT) {
	utils.WriteBin(w, n.Prefix, n.Timestamp, n.U32)
}

func ParseNodeId(hexr io.Reader) *flatgen.NodeIdT {
	var (
		prefix    byte
		timestamp int64
		u32       uint32
	)
	utils.ReadBin(hexr, &prefix, &timestamp, &u32)
	return &flatgen.NodeIdT{
		Prefix:    prefix,
		Timestamp: timestamp,
		U32:       u32,
	}
}

func ParseMessageId(hx io.Reader) *flatgen.MessageIdT {
	var (
		prefix, suffix byte
		timestamp      int64
		nonce          uint32
		root           *flatgen.RootT
	)
	utils.ReadBin(hx, &prefix)
	root = ParseRoot(hx)
	utils.ReadBin(hx, &timestamp, &nonce, &suffix)
	return &flatgen.MessageIdT{
		Prefix:    prefix,
		Root:      root,
		Timestamp: timestamp,
		U32:       nonce,
	}
}

// func ParseRawMediaRef(raw io.Reader) *flatgen.MediaReferenceT {
// 	r := bufio.NewReader(raw)

// }

func IsMostUpToDateRoot(root *flatgen.RootT, roots []*flatgen.RootT) bool {
	roots = utils.Filter(roots, func(r *flatgen.RootT) bool {
		return IsHomological(root, r) && root.Timestamp > r.Timestamp
	})
	return len(roots) == 0
}

func CouldUpdateRoot(r *flatgen.RootT) bool {
	const oneMonth = time.Hour * 24 * 31 // about 1 month
	now := time.Now()
	oneMonthAgo := now.Add(-oneMonth).UnixMilli()
	couldUpdate := r.Timestamp < oneMonthAgo
	return couldUpdate
}

func MostUpToDateRoot(roots []*flatgen.RootT) *flatgen.RootT {
	if len(roots) == 0 {
		return nil
	}
	return utils.Reduce(roots[1:], roots[0], func(r1, r2 *flatgen.RootT) *flatgen.RootT {
		if r1.Timestamp > r2.Timestamp {
			return r1
		} else {
			return r2
		}
	})
}

func MessageIdFromString(s string) *flatgen.MessageIdT {
	return ParseMessageId(hex.NewDecoder(strings.NewReader(s)))
}

func RootFromString(s string) *flatgen.RootT {
	return ParseRoot(hex.NewDecoder(strings.NewReader(s)))
}

func NodeIdFromString(s string) *flatgen.NodeIdT {
	return ParseNodeId(hex.NewDecoder(strings.NewReader(s)))
}

func ParseRoot(hexr io.Reader) *flatgen.RootT {
	var (
		prefix    byte
		n1, n2    *flatgen.NodeIdT
		timestamp int64
		place     uint16
		confirmed bool
		u32       uint32
	)
	utils.ReadBin(hexr, &prefix)
	n1 = ParseNodeId(hexr)
	n2 = ParseNodeId(hexr)
	utils.ReadBin(hexr, &timestamp, &u32, &place, &confirmed)
	return &flatgen.RootT{
		Prefix:    prefix,
		Timestamp: timestamp,
		Primary:   n1,
		Secondary: n2,
		ChatPlace: place,
		Confirmed: confirmed,
		U32:       u32,
	}
}

func EqualNodeId(id1, id2 *flatgen.NodeIdT) bool {
	return id1.Timestamp == id2.Timestamp && id1.U32 == id2.U32
}

func IsHomological(r1, r2 *flatgen.RootT) bool {
	return EqualNodeId(r1.Primary, r2.Primary) && EqualNodeId(r1.Secondary, r2.Secondary)
}

func RandU32() uint32 {
	b := make([]byte, 4)
	rand.Read(b)
	return binary.BigEndian.Uint32(b)
}

func WriteNodeId(w io.Writer, s *flatgen.NodeId) {
	utils.WriteBin(w, KIND_NODE, s.Timestamp(), s.U32())
}

func WriteRoot(w io.Writer, r *flatgen.Root) {
	utils.WriteBin(w, KIND_ROOT)
	WriteNodeId(w, r.Primary(nil))
	WriteNodeId(w, r.Secondary(nil))
	utils.WriteBin(w, r.Timestamp(), r.U32(), r.ChatPlace(), r.Confirmed())
}

// func WriteMediaId(w io.Writer, id *flatgen.MediaId) {
// 	utils.WriteBin(w, KIND_MEDIA, id.Timestamp(), id.U32(), id.Width(), id.Height(),
// 		id.Squared(), id.Video())
// }

// func MakeRawMediaId(id *flatgen.MediaId) []byte {
// 	buf := bytes.NewBuffer(make([]byte, 0, RAW_MEDIA_ID_LEN))
// 	WriteMediaId(buf, id)
// 	return buf.Bytes()
// }

func MakeRawMediaReference(r *flatgen.MediaReference) []byte {
	buf := new(bytes.Buffer)
	WriteMediaReference(buf, r)
	return buf.Bytes()
}

// func MakeRawMediaRef(id *flatgen.MediaRef) []byte {
// 	var buf *bytes.Buffer
// 	if packId := id.PackId(nil); packId != nil {
// 		buf = bytes.NewBuffer(make([]byte, 0, RAW_STICKER_ID_LEN))
// 	} else {
// 		buf = bytes.NewBuffer(make([]byte, 0, RAW_MEDIA_REF_LEN))
// 	}
// 	WriteMediaRef(buf, id)
// 	return buf.Bytes()
// }

// func WriteMediaRef(w io.Writer, id *flatgen.MediaRef) {
// 	if packId := id.PackId(nil); packId != nil {
// 		utils.WriteBin(w, KIND_STICKER, id.Timestamp())
// 		WriteNodeId(w, packId)
// 	} else {
// 		utils.WriteBin(w, KIND_MEDIA_REF, id.Timestamp())
// 	}
// 	WriteMediaId(w, id.MediaId(nil))
// 	utils.WriteBin(w, id.Place(), id.Permanent())
// }

func WriteMsgId(w io.Writer, id *flatgen.MessageId) {
	utils.WriteBin(w, id.Prefix())
	WriteRoot(w, id.Root(nil))
	utils.WriteBin(w, id.Timestamp(), id.U32(), id.Suffix())
}

func WritePushId(w io.Writer, id *flatgen.PushId) {
	utils.WriteBin(w, KIND_PUSH)
	WriteNodeId(w, id.NodeId(nil))
	utils.WriteBin(w, id.Device(), id.Timestamp(), id.U32())
}

func WritePushIdPrefixId(w io.Writer, id *flatgen.PushId) {
	utils.WriteBin(w, KIND_PUSH)
	WriteNodeId(w, id.NodeId(nil))
	utils.WriteBin(w, id.Device())
}

func MakeRawRoot(r *flatgen.Root) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, RAW_ROOT_ID_LEN))
	WriteRoot(buf, r)
	return buf.Bytes()
}

func MakeRawRoot2(r *flatgen.Root) Root {
	_root := Root{}
	buf := bytes.NewBuffer(_root[:0])
	WriteRoot(buf, r)
	return _root
}

func MakeRawMsgId(id *flatgen.MessageId) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, RAW_MSG_ID_LEN))
	WriteMsgId(buf, id)
	return buf.Bytes()
}

func MakeRawMsgId2(id *flatgen.MessageId) MsgId {
	raw := MsgId{}
	buf := bytes.NewBuffer(raw[:0])
	WriteMsgId(buf, id)
	return raw
}

func MakeRawPushId(id *flatgen.PushId) []byte {
	bw := bytes.NewBuffer(make([]byte, 0, RAW_PUSH_ID_LEN))
	WritePushId(bw, id)
	return bw.Bytes()
}

// prefix is {prefix 1b} | {ts 8b} | {nodeId 13b} -> 1 + 8 + 13 = 22
func BoostIdPrefix(rawId []byte) []byte {
	return rawId[:22]
}

// prefix is {prefix 1b} | {ts 8b} | {nodeId 13b} -> 1 + 8 + 13 = 22
func BoostIdPrefix_(rawId []byte) []byte {
	return rawId[:9]
}

func MsgIdPrefix(idRaw []byte) []byte {
	// Msg prefix byte + root
	return idRaw[:1+RAW_ROOT_ID_LEN]
}

func PushIdPrefix(raw []byte) []byte {
	return raw[:RAW_PUSH_ID_PREFIX_LEN]
}
