package ue2

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type TestStruct struct {
	A int32
	B int32
	C Index
	D [3]byte
	E Index
	F Index
	G Index
	H Index

	Name string
}

func TestUnmarshal(t *testing.T) {
	data := []byte{
		0xf6, 0xff, 0xff, 0xff,
		0x0a, 0x00, 0x00, 0x00,
		0x01,
		0x02, 0x03, 0x04,
		0x40 | 0x01, 0x02,
		0x40 | 0x01, 0x80 | 0x02, 0x03,
		0x40 | 0x01, 0x80 | 0x02, 0x80 | 0x03, 0x04,
		0x80 | 0x01,
		0x05, 'T', 'E', 'S', 'T', 0x00,
	}

	var got TestStruct
	err := Unmarshal(data, &got)
	if err != nil {
		t.Error(err)
	}

	want := TestStruct{
		A:    -10,
		B:    10,
		C:    1,
		D:    [3]byte{0x02, 0x03, 0x04},
		E:    0x02<<6 | 0x01,
		F:    0x03<<13 | 0x02<<6 | 0x01,
		G:    0x04<<20 | 0x03<<13 | 0x02<<6 | 0x01,
		H:    -1,
		Name: "TEST",
	}

	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("TestDecode mismatch (-want,+got):\n%s", d)
	}
}

func TestMarshal(t *testing.T) {
	obj := TestStruct{
		A:    -10,
		B:    10,
		C:    1,
		D:    [3]byte{0x02, 0x03, 0x04},
		E:    0x02<<6 | 0x01,
		F:    0x03<<13 | 0x02<<6 | 0x01,
		G:    0x04<<20 | 0x03<<13 | 0x02<<6 | 0x01,
		H:    -1,
		Name: "TEST",
	}

	want := []byte{
		0xf6, 0xff, 0xff, 0xff,
		0x0a, 0x00, 0x00, 0x00,
		0x01,
		0x02, 0x03, 0x04,
		0x40 | 0x01, 0x02,
		0x40 | 0x01, 0x80 | 0x02, 0x03,
		0x40 | 0x01, 0x80 | 0x02, 0x80 | 0x03, 0x04,
		0x80 | 0x01,
		0x05, 'T', 'E', 'S', 'T', 0x00,
	}

	got, err := Marshal(obj)
	if err != nil {
		t.Error(err)
	}

	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("TestEncode mismatch (-want,+got):\n%s", d)
	}
}

func TestStripColors(t *testing.T) {
	s := "\x1b\x00\x01\x02This is a string with \x1b\x00\x00\x00color."

	want := "This is a string with color."
	got := StripColors(s)

	if want != got {
		t.Errorf("want: %v, got: %v", []byte(want), []byte(got))
	}
}
