package jce

import (
	goBinary "encoding/binary"
	"math"

	"github.com/Mrs4s/MiraiGo/utils"
)

type JceReader struct {
	buf []byte
	off int
}

type HeadData struct {
	Type byte
	Tag  int
}

func NewJceReader(data []byte) *JceReader {
	return &JceReader{buf: data}
}

func (r *JceReader) readHead() (hd HeadData, l int) {
	hd, l = r.peakHead()
	r.off += l
	return
}

func (r *JceReader) peakHead() (hd HeadData, l int) {
	b := r.buf[r.off]
	hd.Type = b & 0xF
	hd.Tag = int(uint(b) >> 4)
	l = 1
	if hd.Tag == 0xF {
		b = r.buf[r.off+1]
		hd.Tag = int(uint(b))
		l = 2
	}
	return
}

func (r *JceReader) skipHead() {
	l := 1
	if int(uint(r.buf[r.off])>>4) == 0xF {
		l = 2
	}
	r.off += l
}

func (r *JceReader) skip(l int) {
	r.off += l
}

func (r *JceReader) skipField(t byte) {
	switch t {
	case 0:
		r.skip(1)
	case 1:
		r.skip(2)
	case 2, 4:
		r.skip(4)
	case 3, 5:
		r.skip(8)
	case 6:
		r.skip(int(r.readByte()))
	case 7:
		r.skip(int(r.readUInt32()))
	case 8:
		s := r.ReadInt32(0)
		for i := 0; i < int(s)*2; i++ {
			r.skipNextField()
		}
	case 9:
		s := r.ReadInt32(0)
		for i := 0; i < int(s); i++ {
			r.skipNextField()
		}
	case 13:
		r.skipHead()
		s := r.ReadInt32(0)
		r.skip(int(s))
	case 10:
		r.skipToStructEnd()
	}
}

func (r *JceReader) skipNextField() {
	hd, _ := r.readHead()
	r.skipField(hd.Type)
}

func (r *JceReader) SkipField(c int) {
	for i := 0; i < c; i++ {
		r.skipNextField()
	}
}

func (r *JceReader) readBytes(n int) []byte {
	if r.off+n > len(r.buf) {
		panic("readBytes: EOF")
	}
	b := make([]byte, n)
	r.off += copy(b, r.buf[r.off:])
	return b
}

func (r *JceReader) readByte() byte {
	if r.off >= len(r.buf) {
		panic("readByte: EOF")
	}
	b := r.buf[r.off]
	r.off++
	return b
}

func (r *JceReader) readUInt16() uint16 {
	b := make([]byte, 2)
	r.off += copy(b, r.buf[r.off:])
	return goBinary.BigEndian.Uint16(b)
}

func (r *JceReader) readUInt32() uint32 {
	b := make([]byte, 4)
	r.off += copy(b, r.buf[r.off:])
	return goBinary.BigEndian.Uint32(b)
}

func (r *JceReader) readUInt64() uint64 {
	b := make([]byte, 8)
	r.off += copy(b, r.buf[r.off:])
	return goBinary.BigEndian.Uint64(b)
}

func (r *JceReader) readFloat32() float32 {
	return math.Float32frombits(r.readUInt32())
}

func (r *JceReader) readFloat64() float64 {
	return math.Float64frombits(r.readUInt64())
}

func (r *JceReader) skipToTag(tag int) bool {
	hd, l := r.peakHead()
	for tag > hd.Tag && hd.Type != 11 {
		r.skip(l)
		r.skipField(hd.Type)
		hd, l = r.peakHead()
	}
	return tag == hd.Tag
}

func (r *JceReader) skipToStructEnd() {
	hd, _ := r.readHead()
	for hd.Type != 11 {
		r.skipField(hd.Type)
		hd, _ = r.readHead()
	}
}

func (r *JceReader) ReadByte(tag int) byte {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 0:
		return r.readByte()
	default:
		return 0
	}
}

func (r *JceReader) ReadBool(tag int) bool {
	return r.ReadByte(tag) != 0
}

func (r *JceReader) ReadInt16(tag int) int16 {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 0:
		return int16(r.readByte())
	case 1:
		return int16(r.readUInt16())
	default:
		return 0
	}
}

func (r *JceReader) ReadInt32(tag int) int32 {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 0:
		return int32(r.readByte())
	case 1:
		return int32(r.readUInt16())
	case 2:
		return int32(r.readUInt32())
	default:
		return 0
	}
}

func (r *JceReader) ReadInt64(tag int) int64 {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 0:
		return int64(r.readByte())
	case 1:
		return int64(int16(r.readUInt16()))
	case 2:
		return int64(r.readUInt32())
	case 3:
		return int64(r.readUInt64())
	default:
		return 0
	}
}

func (r *JceReader) ReadFloat32(tag int) float32 {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 4:
		return r.readFloat32()
	default:
		return 0
	}
}

func (r *JceReader) ReadFloat64(tag int) float64 {
	if !r.skipToTag(tag) {
		return 0
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 12:
		return 0
	case 4:
		return float64(r.readFloat32())
	case 5:
		return r.readFloat64()
	default:
		return 0
	}
}

func (r *JceReader) ReadString(tag int) string {
	if !r.skipToTag(tag) {
		return ""
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 6:
		return utils.B2S(r.readBytes(int(r.readByte())))
	case 7:
		return utils.B2S(r.readBytes(int(r.readUInt32())))
	default:
		return ""
	}
}

func (r *JceReader) ReadBytes(tag int) []byte {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		b := make([]byte, s)
		for i := 0; i < int(s); i++ {
			b[i] = r.ReadByte(0)
		}
		return b
	case 13:
		r.skipHead()
		return r.readBytes(int(r.ReadInt32(0)))
	default:
		return nil
	}
}

func (r *JceReader) ReadByteArrArr(tag int) (baa [][]byte) {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		baa = make([][]byte, s)
		for i := 0; i < int(s); i++ {
			baa[i] = r.ReadBytes(0)
		}
		return baa
	default:
		return nil
	}
}

/*
// ReadAny Read any type via tag, unsupported JceStruct
func (r *JceReader) ReadAny(tag int) interface{} {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 0:
		return r.readByte()
	case 1:
		return r.readUInt16()
	case 2:
		return r.readUInt32()
	case 3:
		return r.readUInt64()
	case 4:
		return r.readFloat32()
	case 5:
		return r.readFloat64()
	case 6:
		return utils.B2S(r.readBytes(int(r.readByte())))
	case 7:
		return utils.B2S(r.readBytes(int(r.readUInt32())))
	case 8:
		s := r.ReadInt32(0)
		m := make(map[interface{}]interface{})
		for i := 0; i < int(s); i++ {
			m[r.ReadAny(0)] = r.ReadAny(1)
		}
		return m
	case 9:
		var sl []interface{}
		s := r.ReadInt32(0)
		for i := 0; i < int(s); i++ {
			sl = append(sl, r.ReadAny(0))
		}
		return sl
	case 12:
		return 0
	case 13:
		r.skipHead()
		return r.readBytes(int(r.ReadInt32(0)))
	default:
		return nil
	}
}
*/

func (r *JceReader) ReadJceStruct(obj IJceStruct, tag int) {
	if !r.skipToTag(tag) {
		return
	}
	hd, _ := r.readHead()
	if hd.Type != 10 {
		return
	}
	obj.ReadFrom(r)
	r.skipToStructEnd()
}

func (r *JceReader) ReadMapStrStr(tag int) map[string]string {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 8:
		s := r.ReadInt32(0)
		m := make(map[string]string, s)
		for i := 0; i < int(s); i++ {
			m[r.ReadString(0)] = r.ReadString(1)
		}
		return m
	default:
		return nil
	}
}

func (r *JceReader) ReadMapStrByte(tag int) map[string][]byte {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 8:
		s := r.ReadInt32(0)
		m := make(map[string][]byte, s)
		for i := 0; i < int(s); i++ {
			m[r.ReadString(0)] = r.ReadBytes(1)
		}
		return m
	default:
		return nil
	}
}

func (r *JceReader) ReadMapStrMapStrByte(tag int) map[string]map[string][]byte {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 8:
		s := r.ReadInt32(0)
		m := make(map[string]map[string][]byte, s)
		for i := 0; i < int(s); i++ {
			m[r.ReadString(0)] = r.ReadMapStrByte(1)
		}
		return m
	default:
		return nil
	}
}

/*
func (r *JceReader) ReadMap(i interface{}, tag int) {
	r.readMap(reflect.ValueOf(i), tag)
}

func (r *JceReader) readMap(v reflect.Value, tag int) {
	if v.Kind() != reflect.Map || !r.skipToTag(tag) {
		return
	}
	t := v.Type()

	kt := t.Key()
	vt := t.Elem()
	r.skipHead()
	s := r.ReadInt32(0)

	// map with string key or string value is very common.
	// specialize for string
	if kt.Kind() == reflect.String && vt.Kind() == reflect.String {
		for i := 0; i < int(s); i++ {
			kv := reflect.ValueOf(r.ReadString(0))
			vv := reflect.ValueOf(r.ReadString(1))
			v.SetMapIndex(kv, vv)
		}
		return
	}

	if kt.Kind() == reflect.String {
		vv := reflect.New(vt)
		for i := 0; i < int(s); i++ {
			kv := reflect.ValueOf(r.ReadString(0))
			r.readObject(vv, 1)
			v.SetMapIndex(kv, vv.Elem())
		}
		return
	}

	kv := reflect.New(kt)
	vv := reflect.New(vt)
	for i := 0; i < int(s); i++ {
		r.readObject(kv, 0)
		r.readObject(vv, 1)
		v.SetMapIndex(kv.Elem(), vv.Elem())
	}
}
*/

func (r *JceReader) ReadFileStorageServerInfos(tag int) []FileStorageServerInfo {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		sl := make([]FileStorageServerInfo, s)
		for i := 0; i < int(s); i++ {
			r.skipHead()
			sl[i].ReadFrom(r)
			r.skipToStructEnd()
		}
		return sl
	default:
		return nil
	}
}

func (r *JceReader) ReadBigDataIPLists(tag int) []BigDataIPList {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		sl := make([]BigDataIPList, s)
		for i := 0; i < int(s); i++ {
			r.skipHead()
			sl[i].ReadFrom(r)
			r.skipToStructEnd()
		}
		return sl
	default:
		return nil
	}
}

func (r *JceReader) ReadBigDataIPInfos(tag int) []BigDataIPInfo {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		sl := make([]BigDataIPInfo, s)
		for i := 0; i < int(s); i++ {
			r.skipHead()
			sl[i].ReadFrom(r)
			r.skipToStructEnd()
		}
		return sl
	default:
		return nil
	}
}

func (r *JceReader) ReadOnlineInfos(tag int) []OnlineInfo {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		sl := make([]OnlineInfo, s)
		for i := 0; i < int(s); i++ {
			r.skipHead()
			sl[i].ReadFrom(r)
			r.skipToStructEnd()
		}
		return sl
	default:
		return nil
	}
}

func (r *JceReader) ReadInstanceInfos(tag int) []InstanceInfo {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		sl := make([]InstanceInfo, s)
		for i := 0; i < int(s); i++ {
			r.skipHead()
			sl[i].ReadFrom(r)
			r.skipToStructEnd()
		}
		return sl
	default:
		return nil
	}
}

func (r *JceReader) ReadSsoServerInfos(tag int) []SsoServerInfo {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		sl := make([]SsoServerInfo, s)
		for i := 0; i < int(s); i++ {
			r.skipHead()
			sl[i].ReadFrom(r)
			r.skipToStructEnd()
		}
		return sl
	default:
		return nil
	}
}

func (r *JceReader) ReadFriendInfos(tag int) []FriendInfo {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		sl := make([]FriendInfo, s)
		for i := 0; i < int(s); i++ {
			r.skipHead()
			sl[i].ReadFrom(r)
			r.skipToStructEnd()
		}
		return sl
	default:
		return nil
	}
}

func (r *JceReader) ReadTroopNumbers(tag int) []TroopNumber {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		sl := make([]TroopNumber, s)
		for i := 0; i < int(s); i++ {
			r.skipHead()
			sl[i].ReadFrom(r)
			r.skipToStructEnd()
		}
		return sl
	default:
		return nil
	}
}

func (r *JceReader) ReadTroopMemberInfos(tag int) []TroopMemberInfo {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		sl := make([]TroopMemberInfo, s)
		for i := 0; i < int(s); i++ {
			r.skipHead()
			sl[i].ReadFrom(r)
			r.skipToStructEnd()
		}
		return sl
	default:
		return nil
	}
}

func (r *JceReader) ReadPushMessageInfos(tag int) []PushMessageInfo {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		sl := make([]PushMessageInfo, s)
		for i := 0; i < int(s); i++ {
			r.skipHead()
			sl[i].ReadFrom(r)
			r.skipToStructEnd()
		}
		return sl
	default:
		return nil
	}
}

func (r *JceReader) ReadSvcDevLoginInfos(tag int) []SvcDevLoginInfo {
	if !r.skipToTag(tag) {
		return nil
	}
	hd, _ := r.readHead()
	switch hd.Type {
	case 9:
		s := r.ReadInt32(0)
		sl := make([]SvcDevLoginInfo, s)
		for i := 0; i < int(s); i++ {
			r.skipHead()
			sl[i].ReadFrom(r)
			r.skipToStructEnd()
		}
		return sl
	default:
		return nil
	}
}

/*
func (r *JceReader) ReadSlice(i interface{}, tag int) {
	r.readSlice(reflect.ValueOf(i), tag)
}

func (r *JceReader) readSlice(v reflect.Value, tag int) {
	t := v.Type()
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice || !r.skipToTag(tag) {
		return
	}
	v = v.Elem()
	t = t.Elem()
	hd, _ := r.readHead()
	if hd.Type == 9 {
		s := r.ReadInt32(0)
		sv := reflect.MakeSlice(t, int(s), int(s))
		t = t.Elem()
		val := reflect.New(t)
		for i := 0; i < int(s); i++ {
			r.readObject(val, 0)
			sv.Index(i).Set(val.Elem())
		}
		v.Set(sv)
	}
	if hd.Type == 13 && t.Elem().Kind() == reflect.Uint8 {
		r.skipHead()
		arr := r.readBytes(int(r.ReadInt32(0)))
		v.SetBytes(arr)
	}
}

func (r *JceReader) ReadObject(i interface{}, tag int) {
	r.readObject(reflect.ValueOf(i), tag)
}

func (r *JceReader) readObject(v reflect.Value, tag int) {
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return
	}
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Map {
		elem := v.Elem()
		elem.Set(reflect.MakeMap(elem.Type()))
		r.readMap(elem, tag)
		return
	} else if elemType.Kind() == reflect.Slice && // *[]byte
		elemType.Elem().Kind() == reflect.Uint8 {
		elem := v.Elem()
		elem.SetBytes(r.ReadBytes(tag))
		return
	}

	switch elemType.Kind() {
	case reflect.Uint8, reflect.Int8:
		*(*uint8)(pointerOf(v)) = r.ReadByte(tag)
	case reflect.Bool:
		*(*bool)(pointerOf(v)) = r.ReadBool(tag)
	case reflect.Uint16, reflect.Int16:
		*(*int16)(pointerOf(v)) = r.ReadInt16(tag)
	case reflect.Uint32, reflect.Int32:
		*(*int32)(pointerOf(v)) = r.ReadInt32(tag)
	case reflect.Uint64, reflect.Int64:
		*(*int64)(pointerOf(v)) = r.ReadInt64(tag)
	case reflect.String:
		*(*string)(pointerOf(v)) = r.ReadString(tag)

	default:
		// other cases
		switch o := v.Interface().(type) {
		case IJceStruct:
			r.skipHead()
			o.ReadFrom(r)
			r.skipToStructEnd()
		case *float32:
			*o = r.ReadFloat32(tag)
		case *float64:
			*o = r.ReadFloat64(tag)
		}
	}
}
*/
