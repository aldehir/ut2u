package ue2

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"unicode/utf16"
)

type Index int32

var indexType = reflect.TypeOf(Index(0))

func Decode(r io.Reader, data any) (err error) {
	defer catchError(&err)

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	d := &decoder{r: r}
	d.value(v)
	return
}

func Encode(w io.Writer, data any) (err error) {
	defer catchError(&err)

	e := &encoder{w: w}
	v := reflect.Indirect(reflect.ValueOf(data))
	e.value(v)
	return
}

type encoder struct {
	w io.Writer
}

func (e *encoder) write(val any) {
	err := binary.Write(e.w, binary.LittleEndian, val)
	if err != nil {
		error_(err)
	}
}

func (e *encoder) uint8(v uint8)   { e.write(v) }
func (e *encoder) uint16(v uint16) { e.write(v) }
func (e *encoder) uint32(v uint32) { e.write(v) }
func (e *encoder) int8(v int8)     { e.uint8(uint8(v)) }
func (e *encoder) int16(v int16)   { e.uint16(uint16(v)) }
func (e *encoder) int32(v int32)   { e.uint32(uint32(v)) }

func (e *encoder) string(v string) {
	if len(v) == 0 {
		e.ueIndex(Index(0))
		return
	}

	e.ueIndex(Index(len(v) + 1))

	b := []byte(v)
	b = append(b, 0x00)

	_, err := e.w.Write(b)
	if err != nil {
		error_(err)
	}
}

func (e *encoder) ueIndex(v Index) {
	var negative bool

	if v < 0 {
		negative = true
		v = -1 * v
	}

	n1 := uint8(v & 0x3f)
	n2 := uint8((v >> 6) & 0x7f)
	n3 := uint8((v >> 13) & 0x7f)
	n4 := uint8((v >> 20) & 0x7f)

	b := n1
	if negative {
		b |= 0x80
	}

	if n2 != 0 || n3 != 0 || n4 != 0 {
		b |= 0x40
	}

	e.uint8(b)

	if n2 == 0 && n3 == 0 && n4 == 0 {
		return
	}

	b = n2
	if n3 != 0 || n4 != 0 {
		b |= 0x80
	}

	e.uint8(b)

	if n3 == 0 && n4 == 0 {
		return
	}

	b = n3
	if n4 != 0 {
		b |= 0x80
	}

	e.uint8(b)

	if n4 == 0 {
		return
	}

	e.uint8(n4)
}

func (e *encoder) value(v reflect.Value) {
	switch v.Kind() {
	case reflect.Array:
		l := v.Len()
		for i := 0; i < l; i++ {
			e.value(v.Index(i))
		}

	case reflect.Struct:
		l := v.NumField()
		for i := 0; i < l; i++ {
			e.value(v.Field(i))
		}

	case reflect.Slice:
		l := v.Len()
		for i := 0; i < l; i++ {
			e.value(v.Index(i))
		}

	case reflect.Int8:
		e.int8(int8(v.Int()))
	case reflect.Uint8:
		e.uint8(uint8(v.Uint()))
	case reflect.Int16:
		e.int16(int16(v.Int()))
	case reflect.Uint16:
		e.uint16(uint16(v.Uint()))
	case reflect.Int32:
		if v.Type() == indexType {
			e.ueIndex(Index(v.Int()))
		} else {
			e.int32(int32(v.Int()))
		}
	case reflect.Uint32:
		e.uint32(uint32(v.Uint()))
	case reflect.String:
		e.string(v.String())
	}
}

type decoder struct {
	r io.Reader
}

func (d *decoder) next(val any) {
	err := binary.Read(d.r, binary.LittleEndian, val)
	if err != nil {
		error_(err)
	}
}

func (d *decoder) uint8() uint8 {
	var val uint8
	d.next(&val)
	return val
}

func (d *decoder) uint16() uint16 {
	var val uint16
	d.next(&val)
	return val
}

func (d *decoder) uint32() uint32 {
	var val uint32
	d.next(&val)
	return val
}

func (d *decoder) int8() int8   { return int8(d.uint8()) }
func (d *decoder) int16() int16 { return int16(d.uint16()) }
func (d *decoder) int32() int32 { return int32(d.uint32()) }

func (d *decoder) string() string {
	length := d.ueIndex()

	if length == 0 {
		return ""
	}

	if length < 0 {
		// Handle unicode strings
		b := make([]uint16, -1*length)
		err := binary.Read(d.r, binary.LittleEndian, b)
		if err != nil {
			error_(err)
		}

		runes := utf16.Decode(b)
		return string(runes[:len(runes)-1])
	}

	b := make([]byte, length)
	_, err := d.r.Read(b)
	if err != nil {
		error_(err)
	}

	return string(b[:len(b)-1])
}

func (d *decoder) ueIndex() int32 {
	sign := int32(1)

	b := d.uint8()
	value := int32(b & 0x3f)

	if b&0x80 != 0 {
		sign = -1
	}

	if b&0x40 != 0 {
		b = d.uint8()
		value |= int32(b&0x7f) << 6

		if b&0x80 != 0 {
			b = d.uint8()
			value |= int32(b&0x7f) << 13

			if b&0x80 != 0 {
				b = d.uint8()
				value |= int32(b&0x7f) << 20
			}
		}
	}

	return sign * value
}

func (d *decoder) value(v reflect.Value) {
	switch v.Kind() {
	case reflect.Array:
		l := v.Len()
		for i := 0; i < l; i++ {
			d.value(v.Index(i))
		}

	case reflect.Struct:
		l := v.NumField()
		for i := 0; i < l; i++ {
			if v := v.Field(i); v.CanSet() {
				d.value(v)
			}
		}

	case reflect.Slice:
		l := v.Len()
		for i := 0; i < l; i++ {
			d.value(v.Index(i))
		}

	case reflect.Int8:
		v.SetInt(int64(d.int8()))
	case reflect.Uint8:
		v.SetUint(uint64(d.uint8()))
	case reflect.Int16:
		v.SetInt(int64(d.int16()))
	case reflect.Uint16:
		v.SetUint(uint64(d.uint16()))
	case reflect.Int32:
		if v.Type() == indexType {
			v.SetInt(int64(d.ueIndex()))
		} else {
			v.SetInt(int64(d.int32()))
		}
	case reflect.Uint32:
		v.SetUint(uint64(d.uint32()))
	case reflect.String:
		v.SetString(d.string())
	}

}

type ue2Error struct {
	err error
}

func errorf(format string, args ...any) {
	error_(fmt.Errorf(format, args...))
}

func error_(err error) {
	panic(ue2Error{err})
}

func catchError(err *error) {
	if e := recover(); e != nil {
		ue, ok := e.(ue2Error)
		if !ok {
			panic(e)
		}
		*err = ue.err
	}
}
