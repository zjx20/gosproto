package sproto

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

const (
	EncodeBufferSize = 4096
	MaxEmbeddedInt   = 0x7fff - 1
	MaxInt32         = 0x7fffffff
	MinInt32         = -0x80000000
	DOUBLE_SZ        = int(unsafe.Sizeof(float64(0.0)))
)

// little endian
// put int into []byte
func writeUint16(buf []byte, v uint16) {
	buf[0] = uint8(v & 0xff)
	buf[1] = uint8(v >> 8)
}

func writeUint32(buf []byte, v uint32) {
	buf[0] = uint8(v & 0xff)
	buf[1] = uint8((v >> 8) & 0xff)
	buf[2] = uint8((v >> 16) & 0xff)
	buf[3] = uint8((v >> 24) & 0xff)
}

func writeUint64(buf []byte, v uint64) {
	buf[0] = uint8(v & 0xff)
	buf[1] = uint8((v >> 8) & 0xff)
	buf[2] = uint8((v >> 16) & 0xff)
	buf[3] = uint8((v >> 24) & 0xff)
	buf[4] = uint8((v >> 32) & 0xff)
	buf[5] = uint8((v >> 40) & 0xff)
	buf[6] = uint8((v >> 48) & 0xff)
	buf[7] = uint8((v >> 56) & 0xff)
}

func headerEncodeDefault(sf *SprotoField, v reflect.Value) (uint16, bool) {
	if !v.IsValid() {
		return 0, true
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice:
		if v.IsNil() {
			return 0, true
		}
	}
	return 0, false
}

func headerEncodeBool(sf *SprotoField, v reflect.Value) (uint16, bool) {
	if !v.IsValid() {
		return 0, true
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice:
		if v.IsNil() {
			return 0, true
		}
	}
	var n uint16 = 0
	if v.Elem().Bool() {
		n = 1
	}
	return 2 * (n + 1), false
}

func headerEncodeInt(sf *SprotoField, v reflect.Value) (uint16, bool) {
	if !v.IsValid() {
		return 0, true
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice:
		if v.IsNil() {
			return 0, true
		}
	}

	var n uint64
	switch e := v.Elem(); e.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		n = uint64(e.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		n = e.Uint()
	}
	if n <= MaxEmbeddedInt {
		return uint16(2 * (n + 1)), false
	}
	return 0, false
}

func extractInt(v reflect.Value) (n uint64, sz int) {
	switch v.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32:
		n = uint64(v.Int())
		sz = 4
	case reflect.Int, reflect.Int64:
		n1 := v.Int()
		n = uint64(n1)
		if n1 >= -0x80000000 && n1 <= 0x7fffffff {
			sz = 4
		} else {
			sz = 8
		}
	case reflect.Uint8, reflect.Uint16:
		n = v.Uint()
		sz = 4
	default: //case reflect.Uint32, reflect.Uint64, reflect.Uint:
		n := v.Uint()
		if n <= MaxInt32 {
			sz = 4
		} else {
			sz = 8
		}
	}
	return
}

func encodeInt(sf *SprotoField, v reflect.Value) []byte {
	n, sz := extractInt(v.Elem())
	if n <= MaxEmbeddedInt {
		return nil
	}

	buf := make([]byte, sz)
	if sz == 4 {
		writeUint32(buf, uint32(n))
	} else {
		writeUint64(buf, n)
	}
	return buf
}

func encodeDouble(sf *SprotoField, v reflect.Value) []byte {
	n := math.Float64bits(v.Elem().Float())
	buf := make([]byte, DOUBLE_SZ)
	writeUint64(buf, n)
	return buf
}

func encodeString(sf *SprotoField, v reflect.Value) []byte {
	str := v.Elem().String()
	buf := make([]byte, len(str))
	copy(buf, str)
	return buf
}

func encodeBytes(sf *SprotoField, v reflect.Value) []byte {
	bytes := v.Bytes()
	buf := make([]byte, len(bytes))
	copy(buf, bytes)
	return buf
}

func encodeStruct(sf *SprotoField, v reflect.Value) []byte {
	return encodeMessage(sf.st, v)
}

func encodeBoolSlice(sf *SprotoField, v reflect.Value) []byte {
	sz := v.Len()
	buf := make([]byte, sz)
	offset := 0
	for i := 0; i < sz; i++ {
		if v.Index(i).Bool() {
			buf[offset+i] = 1
		} else {
			buf[offset+i] = 0
		}
	}
	return buf
}

func encodeBytesSlice(sf *SprotoField, v reflect.Value) []byte {
	var sz int
	for i := 0; i < v.Len(); i++ {
		bs := v.Index(i).Bytes()
		sz += 4 + len(bs)
	}
	buf := make([]byte, sz)
	offset := 0
	for i := 0; i < v.Len(); i++ {
		bs := v.Index(i).Bytes()
		strLen := len(bs)
		writeUint32(buf[offset:], uint32(strLen))
		copy(buf[offset+4:], bs)
		offset += 4 + strLen
	}
	return buf
}

func encodeStringSlice(sf *SprotoField, v reflect.Value) []byte {
	var sz int
	for i := 0; i < v.Len(); i++ {
		str := v.Index(i).String()
		sz += 4 + len(str)
	}
	buf := make([]byte, sz)
	offset := 0
	for i := 0; i < v.Len(); i++ {
		str := v.Index(i).String()
		strLen := len(str)
		writeUint32(buf[offset:], uint32(strLen))
		copy(buf[offset+4:], str)
		offset += 4 + strLen
	}
	return buf
}

func encodeIntSlice(sf *SprotoField, v reflect.Value) []byte {
	vals := make([]uint64, v.Len())
	var intLen int = 4 // could be 4 and 8
	for i := 0; i < v.Len(); i++ {
		n, tmp := extractInt(v.Index(i))
		if tmp > intLen {
			intLen = tmp
		}
		vals[i] = n
	}

	buf := make([]byte, 1+intLen*v.Len())
	buf[0] = uint8(intLen) // put intLen
	offset := 1
	for _, val := range vals {
		if intLen == 4 {
			writeUint32(buf[offset:], uint32(val))
		} else {
			writeUint64(buf[offset:], val)
		}
		offset += intLen
	}
	return buf
}

func encodeDoubleSlice(sf *SprotoField, v reflect.Value) []byte {
	buf := make([]byte, 1+DOUBLE_SZ*v.Len())
	buf[0] = uint8(DOUBLE_SZ)
	offset := 1
	for i := 0; i < v.Len(); i++ {
		dv := v.Index(i).Float()
		writeUint64(buf[offset:], math.Float64bits(dv))
		offset += DOUBLE_SZ
	}
	return buf
}

func encodeStructSlice(sf *SprotoField, v reflect.Value) []byte {
	sz := 0
	vals := make([][]byte, v.Len())
	for i := 0; i < v.Len(); i++ {
		val := encodeMessage(sf.st, v.Index(i))
		vals[i] = val
		sz += len(val) + 4
	}

	buf := make([]byte, sz)
	offset := 0
	for _, val := range vals {
		valLen := len(val)
		writeUint32(buf[offset:], uint32(valLen))
		copy(buf[offset+4:], val)
		offset += valLen + 4
	}
	return buf
}

// v is a map
func encodeMap(sf *SprotoField, v reflect.Value) []byte {
	st := sf.st

	// map convert to slice
	vals := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(st.Type)), 0, v.Len())
	iter := v.MapRange()
	for iter.Next() {
		if sf.ValueTag == -1 {
			// normal map, slice element = map's value
			vals = reflect.Append(vals, iter.Value())
		} else {
			// simple map, construct slice element by map's key and value
			keySprotoField := st.FieldByTag(sf.KeyTag)
			valueSprotoField := st.FieldByTag(sf.ValueTag)

			val := reflect.New(st.Type)
			elem := val.Elem()
			// 处理值赋值到指针的情况；比如map key是值类型，但是slice元素字段定义为指针类型
			setValue(elem.FieldByIndex(keySprotoField.field.Index), iter.Key())
			setValue(elem.FieldByIndex(valueSprotoField.field.Index), iter.Value())
			vals = reflect.Append(vals, val)
		}
	}
	return encodeStructSlice(sf, vals)
}

func skipTag(tag, nextTag int) uint16 {
	if nextTag > tag+1 {
		span := nextTag - tag - 1
		return uint16((span-1)*2 + 1)
	}
	return 0
}

func encodeHeaders(headers []uint16, reserved int) []byte {
	buf := make([]byte, (len(headers)+1)*2+reserved)
	i := 0
	writeUint16(buf, uint16(len(headers)))
	i = i + 2
	for _, v := range headers {
		writeUint16(buf[i:], v)
		i = i + 2
	}
	return buf[:i]
}

func encodeMessage(st *SprotoType, v reflect.Value) []byte {
	headers := make([]uint16, len(st.Fields)*2)   // max header len is fieldNum * 2
	buffer := make([]byte, EncodeBufferSize)[0:0] // pre-allocate 4k buffer

	tag, offset := -1, 0
	dataLen := make([]byte, 4, 4)
	if !v.IsNil() { // struct could be nil in struct array
		for _, i := range st.order {
			sf := st.Fields[i]
			v1 := v.Elem().FieldByIndex(sf.field.Index)
			nextTag := sf.Tag
			if nextTag < 0 {
				continue
			}
			if v1.Kind() != reflect.Ptr &&
				v1.Kind() != reflect.Slice &&
				v1.Kind() != reflect.Array &&
				v1.Kind() != reflect.Struct &&
				v1.Kind() != reflect.Map {
				// 替内部处理取地址
				v1 = v1.Addr()
			}
			if header, isNil := sf.headerEnc(sf, v1); !isNil {
				if skip := skipTag(tag, nextTag); skip > 0 {
					headers[offset] = skip
					offset++
				}
				headers[offset] = header
				offset++
				tag = nextTag
				if sf.enc != nil {
					if data := sf.enc(sf, v1); data != nil {
						writeUint32(dataLen, uint32(len(data)))
						buffer = Append(buffer, dataLen)
						buffer = Append(buffer, data)
					}
				}
			}
		}
	}
	return Append(encodeHeaders(headers[:offset], len(buffer)), buffer)
}

func Encode(sp interface{}) (_ []byte, err error) {
	defer func() {
		if obj := recover(); obj != nil {
			err = fmt.Errorf("sproto: Encode recovered from panic, err: %v", obj)
		}
	}()
	t, v, err := getbase(sp)
	if err != nil {
		return nil, err
	}

	st, err := GetSprotoType(t.Elem())
	if err != nil {
		return nil, err
	}
	return encodeMessage(st, v), nil
}

func MustEncode(sp interface{}) []byte {
	d, err := Encode(sp)
	if err != nil {
		panic(err)
	}
	return d
}
