package ue2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"io"
	"reflect"
	"strings"
	"unicode/utf16"

	"golang.org/x/text/encoding/charmap"
)

type Index int32

type ColorizedString struct {
	Value       string
	ColorPoints []ColorPoint
}

type ColorPoint struct {
	At    int
	Color color.Color
}

var indexType = reflect.TypeOf(Index(0))
var colorizedStringType = reflect.TypeOf(ColorizedString{})

var ErrInvalidColor = errors.New("invalid color")

func Unmarshal(data []byte, value any) error {
	buf := bytes.NewBuffer(data)
	decoder := NewDecoder(buf)

	err := decoder.Decode(value)
	if err != nil {
		return err
	}

	return nil
}

func Marshal(value any) ([]byte, error) {
	var buf bytes.Buffer
	e := NewEncoder(&buf)
	err := e.Encode(value)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func StripColors(s string) string {
	var builder strings.Builder

	byteString := []byte(s)

	i := 0
	for i < len(byteString) {
		if byteString[i] == 0x1b {
			i += 4
			continue
		}

		builder.Write([]byte{byteString[i]})
		i++
	}

	return builder.String()
}

// ToUTF8 converts a UE2 String to UTF8. If it fails to decode, the original
// string is returned.
func ToUTF8(s string) string {
	decoder := charmap.ISO8859_1.NewDecoder()
	res, err := decoder.Bytes([]byte(s))
	if err != nil {
		return s
	}
	return string(res)
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(value any) (err error) {
	defer catchError(&err)
	v := reflect.Indirect(reflect.ValueOf(value))
	e.value(v)
	return
}

func (e *Encoder) write(val any) {
	err := binary.Write(e.w, binary.LittleEndian, val)
	if err != nil {
		error_(err)
	}
}

func (e *Encoder) uint8(v uint8)   { e.write(v) }
func (e *Encoder) uint16(v uint16) { e.write(v) }
func (e *Encoder) uint32(v uint32) { e.write(v) }
func (e *Encoder) int8(v int8)     { e.uint8(uint8(v)) }
func (e *Encoder) int16(v int16)   { e.uint16(uint16(v)) }
func (e *Encoder) int32(v int32)   { e.uint32(uint32(v)) }

func (e *Encoder) string(v string) {
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

func (e *Encoder) ueIndex(v Index) {
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

func (e *Encoder) value(v reflect.Value) {
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

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode(val any) (err error) {
	defer catchError(&err)

	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	d.value(v)
	return
}

func (d *Decoder) next(val any) {
	err := binary.Read(d.r, binary.LittleEndian, val)
	if err != nil {
		error_(err)
	}
}

func (d *Decoder) uint8() uint8 {
	var val uint8
	d.next(&val)
	return val
}

func (d *Decoder) uint16() uint16 {
	var val uint16
	d.next(&val)
	return val
}

func (d *Decoder) uint32() uint32 {
	var val uint32
	d.next(&val)
	return val
}

func (d *Decoder) int8() int8   { return int8(d.uint8()) }
func (d *Decoder) int16() int16 { return int16(d.uint16()) }
func (d *Decoder) int32() int32 { return int32(d.uint32()) }

func (d *Decoder) string() string {
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

func (d *Decoder) colorizedString() reflect.Value {
	var result ColorizedString

	length := d.ueIndex()
	if length == 0 {
		return reflect.ValueOf(result)
	}

	b := make([]byte, length)
	_, err := d.r.Read(b)
	if err != nil {
		error_(err)
	}

	uncolored := make([]byte, 0, len(b))

	i := 0
	for i < len(b)-1 {
		if b[i] == 0x1b {
			if i+3 > len(b)-1 {
				error_(ErrInvalidColor)
			}

			result.ColorPoints = append(result.ColorPoints, ColorPoint{
				At: len(uncolored),
				Color: color.RGBA{
					R: b[i+1],
					G: b[i+2],
					B: b[i+3],
					A: 255,
				},
			})

			i += 4
			continue
		}

		uncolored = append(uncolored, b[i])
		i++
	}

	decoder := charmap.ISO8859_1.NewDecoder()
	asUTF8, err := decoder.Bytes(uncolored)
	if err != nil {
		error_(err)
	}

	result.Value = string(asUTF8)
	return reflect.ValueOf(result)
}

func (d *Decoder) ueIndex() int32 {
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

func (d *Decoder) value(v reflect.Value) {
	switch v.Kind() {
	case reflect.Array:
		l := v.Len()
		for i := 0; i < l; i++ {
			d.value(v.Index(i))
		}

	case reflect.Struct:
		if v.Type() == colorizedStringType {
			v.Set(d.colorizedString())
		} else {
			l := v.NumField()
			for i := 0; i < l; i++ {
				if v := v.Field(i); v.CanSet() {
					d.value(v)
				}
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
