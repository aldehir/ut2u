package ue2

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

func Decode(r io.Reader, data any) (err error) {
	defer catchError(&err)

	d := &decoder{r: r}
	v := reflect.ValueOf(data)
	d.value(v)
	return nil
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
	length := d.uint8()

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
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

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
