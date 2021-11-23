// UNCHECKED

package client

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
)

// Stash will store the data for the client, this will speed up booting
// the QQ client but some data may not sync with the server. So it is
// recommended to use this in development mode only.

//go:generate stringer -type=syncMarker -trimprefix=syncMarker
type syncMarker int8

const (
	syncMarkerNone syncMarker = iota
	syncMarkerFriendList
	syncMarkerFriendInfo
	syncMarkerGroupList
	syncMarkerGroupInfo
	syncMarkerGroupMemberList
	syncMarkerGroupMemberInfo
)

type StashSyncMarkerError struct {
	marker   syncMarker
	expected syncMarker
}

func (e *StashSyncMarkerError) Error() string {
	return "stash sync marker error: expected " + e.expected.String() + ", got " + e.marker.String()
}

// WriteStash will write the stash to the given writer.
func WriteStash(client *QQClient, writer io.Writer) {
	var out intWriter
	w := stashWriter{
		stringIndex: make(map[string]uint64),
	}

	w.friendList(client.FriendList)
	w.groupList(client.GroupList)

	out.uvarint(uint64(w.strings.Len()))
	out.uvarint(uint64(w.data.Len()))
	_, _ = io.Copy(&out, &w.strings)
	_, _ = io.Copy(&out, &w.data)
	_, _ = io.Copy(writer, &out)
}

type stashWriter struct {
	data        intWriter
	strings     intWriter
	stringIndex map[string]uint64
}

func (w *stashWriter) string(s string) {
	off, ok := w.stringIndex[s]
	if !ok {
		off = uint64(w.strings.Len())
		w.strings.uvarint(uint64(len(s)))
		_, _ = w.strings.WriteString(s)
		w.stringIndex[s] = off
	}
	w.uvarint(off)
}

func (w *stashWriter) sync(marker syncMarker) { w.data.uvarint(uint64(marker)) }
func (w *stashWriter) uvarint(v uint64)       { w.data.uvarint(v) }
func (w *stashWriter) svarint(v int64)        { w.data.svarint(v) }

func (w *stashWriter) int64(v int64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(v))
	_, _ = w.data.Write(buf[:])
}

func (w *stashWriter) friendList(list []*FriendInfo) {
	w.sync(syncMarkerFriendList)
	w.uvarint(uint64(len(list)))
	for _, friend := range list {
		w.sync(syncMarkerFriendInfo)
		w.int64(friend.Uin)
		w.string(friend.Nickname)
		w.string(friend.Remark)
		w.svarint(int64(friend.FaceId))
	}
}

func (w *stashWriter) groupList(list []*GroupInfo) {
	w.sync(syncMarkerGroupList)
	w.uvarint(uint64(len(list)))
	for _, group := range list {
		w.sync(syncMarkerGroupInfo)
		w.int64(group.Uin)
		w.int64(group.Code)
		w.string(group.Name)
		w.string(group.Memo)
		w.int64(group.OwnerUin)
		w.uvarint(uint64(group.GroupCreateTime))
		w.uvarint(uint64(group.MemberCount))
		w.uvarint(uint64(group.MaxMemberCount))
		w.svarint(group.LastMsgSeq)

		w.groupMemberList(group.Members)
	}
}

func (w *stashWriter) groupMemberList(list []*GroupMemberInfo) {
	w.sync(syncMarkerGroupMemberList)
	w.uvarint(uint64(len(list)))
	for _, member := range list {
		w.sync(syncMarkerGroupMemberInfo)
		w.int64(member.Uin)
		w.uvarint(uint64(member.Gender))
		w.string(member.Nickname)
		w.string(member.CardName)
		w.uvarint(uint64(member.Level))
		w.int64(member.JoinTime)
		w.int64(member.LastSpeakTime)
		w.string(member.SpecialTitle)
		w.svarint(member.SpecialTitleExpireTime)
		w.svarint(member.ShutUpTimestamp)
		w.uvarint(uint64(member.Permission))
	}
}

type intWriter struct {
	bytes.Buffer
}

func (w *intWriter) svarint(x int64) {
	w.uvarint(uint64(x)<<1 ^ uint64(x>>63))
}

func (w *intWriter) uvarint(x uint64) {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], x)
	_, _ = w.Write(buf[:n])
}

// ReadStash will read the stash from the given reader and apply to the given QQClient.
func ReadStash(client *QQClient, data string) (err error) {
	in := newIntReader(data)
	sl := int64(in.uvarint())
	dl := int64(in.uvarint())
	whence, err := in.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	sData := data[whence : whence+sl]
	dData := data[whence+sl : whence+sl+dl]

	r := stashReader{
		data:        newIntReader(dData),
		strings:     newIntReader(sData),
		stringIndex: make(map[uint64]string),
	}
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
				return
			}
			panic(r)
		}
	}()

	client.FriendList = r.friendList()
	client.GroupList = r.groupList(client)
	return nil
}

type stashReader struct {
	data        intReader
	strings     intReader
	stringIndex map[uint64]string
}

func (r *stashReader) string() string {
	off := r.data.uvarint()
	if off == 0 {
		return ""
	}
	if s, ok := r.stringIndex[off]; ok {
		return s
	}
	_, _ = r.strings.Seek(int64(off), io.SeekStart)
	l := int64(r.strings.uvarint())
	whence, _ := r.strings.Seek(0, io.SeekCurrent)
	s := r.strings.data[whence : whence+l]
	r.stringIndex[off] = s
	return s
}

func (r *stashReader) sync(marker syncMarker) {
	got := syncMarker(r.data.uvarint())
	if got != marker {
		panic(&StashSyncMarkerError{
			marker:   got,
			expected: marker,
		})
	}
}

func (r *stashReader) friendList() []*FriendInfo {
	r.sync(syncMarkerFriendList)
	l := r.uvarint()
	list := make([]*FriendInfo, l)
	for i := uint64(0); i < l; i++ {
		r.sync(syncMarkerFriendInfo)
		list[i] = &FriendInfo{
			Uin:      r.int64(),
			Nickname: r.string(),
			Remark:   r.string(),
			FaceId:   int16(r.svarint()),
		}
	}
	return list
}

func (r *stashReader) groupList(client *QQClient) []*GroupInfo {
	r.sync(syncMarkerGroupList)
	l := r.uvarint()
	list := make([]*GroupInfo, l)
	for i := uint64(0); i < l; i++ {
		r.sync(syncMarkerGroupInfo)
		list[i] = &GroupInfo{
			Uin:             r.int64(),
			Code:            r.int64(),
			Name:            r.string(),
			Memo:            r.string(),
			OwnerUin:        r.int64(),
			GroupCreateTime: uint32(r.uvarint()),
			GroupLevel:      uint32(r.uvarint()),
			MemberCount:     uint16(r.uvarint()),
			MaxMemberCount:  uint16(r.uvarint()),
			client:          client,
		}
		list[i].Members = r.groupMemberList(list[i])
		list[i].LastMsgSeq = r.svarint()
	}
	return list
}

func (r *stashReader) groupMemberList(group *GroupInfo) []*GroupMemberInfo {
	r.sync(syncMarkerGroupMemberList)
	l := r.uvarint()
	list := make([]*GroupMemberInfo, l)
	for i := uint64(0); i < l; i++ {
		r.sync(syncMarkerGroupMemberInfo)
		list[i] = &GroupMemberInfo{
			Group:                  group,
			Uin:                    r.int64(),
			Gender:                 byte(r.uvarint()),
			Nickname:               r.string(),
			CardName:               r.string(),
			Level:                  uint16(r.uvarint()),
			JoinTime:               r.int64(),
			LastSpeakTime:          r.int64(),
			SpecialTitle:           r.string(),
			SpecialTitleExpireTime: r.svarint(),
			ShutUpTimestamp:        r.svarint(),
			Permission:             MemberPermission(r.uvarint()),
		}
	}
	return list
}

func (r *stashReader) uvarint() uint64 { return r.data.uvarint() }
func (r *stashReader) svarint() int64  { return r.data.svarint() }

func (r *stashReader) int64() int64 {
	var buf [8]byte
	_, _ = r.data.Read(buf[:])
	return int64(binary.LittleEndian.Uint64(buf[:]))
}

type intReader struct {
	data string
	*strings.Reader
}

func newIntReader(s string) intReader {
	return intReader{
		data:   s,
		Reader: strings.NewReader(s),
	}
}

func (r *intReader) svarint() int64 {
	i, _ := binary.ReadVarint(r)
	return i
}

func (r *intReader) uvarint() uint64 {
	i, _ := binary.ReadUvarint(r)
	return i
}
