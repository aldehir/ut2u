package upkg

import (
	"encoding/binary"
	"io"
)

type Decoder struct {
	r   io.ReadSeeker
	pkg *Package
}

type readStep = func() error

func NewDecoder(r io.ReadSeeker) *Decoder {
	return &Decoder{r: r, pkg: &Package{}}
}

func (d *Decoder) Decode() (*Package, error) {
	steps := []readStep{
		d.readHeader,
		d.readNames,
		d.readImports,
	}

	var err error
	for _, step := range steps {
		err = step()
		if err != nil {
			return nil, err
		}
	}

	return d.pkg, nil
}

func (d *Decoder) readHeader() (err error) {
	err = binary.Read(d.r, binary.LittleEndian, &d.pkg.h)
	if err != nil {
		return
	}

	if d.pkg.h.Version >= 68 {
		var genCount uint32

		err = binary.Read(d.r, binary.LittleEndian, &genCount)
		if err != nil {
			return
		}

		var gen generation
		for i := 0; i < int(genCount); i++ {
			err = binary.Read(d.r, binary.LittleEndian, &gen)
			if err != nil {
				return
			}

			d.pkg.gen = append(d.pkg.gen, gen)
		}
	}

	return nil
}

func (d *Decoder) readNames() (err error) {
	// Read names
	_, err = d.r.Seek(int64(d.pkg.h.NameOffset), io.SeekStart)
	if err != nil {
		return
	}

	var str string
	var flags uint32

	d.pkg.names = make([]name, 0, d.pkg.h.NameCount)
	for i := 0; i < int(d.pkg.h.NameCount); i++ {
		str, err = d.decodeName()
		if err != nil {
			return
		}

		err = binary.Read(d.r, binary.LittleEndian, &flags)
		if err != nil {
			return
		}

		d.pkg.names = append(d.pkg.names, name{str, flags})
	}

	return nil
}

func (d *Decoder) readImports() (err error) {
	_, err = d.r.Seek(int64(d.pkg.h.ImportOffset), io.SeekStart)
	if err != nil {
		return
	}

	var imp import_

	d.pkg.imports = make([]import_, 0, d.pkg.h.ImportCount)
	for i := 0; i < int(d.pkg.h.ImportCount); i++ {
		imp.ClassPackageIndex, err = d.decodeIndex()
		if err != nil {
			return
		}

		imp.ClassNameIndex, err = d.decodeIndex()
		if err != nil {
			return
		}

		err = binary.Read(d.r, binary.LittleEndian, &imp.Package)
		if err != nil {
			return
		}

		imp.ObjectNameIndex, err = d.decodeIndex()
		if err != nil {
			return
		}

		d.pkg.imports = append(d.pkg.imports, imp)
	}

	return nil
}

func (d *Decoder) decodeName() (name string, err error) {
	var length uint8

	err = binary.Read(d.r, binary.LittleEndian, &length)
	if err != nil {
		return
	}

	if length == 0 {
		return "", nil
	}

	raw := make([]byte, length)
	_, err = d.r.Read(raw)
	if err != nil {
		return
	}

	return string(raw[:len(raw)-1]), nil
}

// This doesn't look pretty, but it should be straightforward to follow
func (d *Decoder) decodeIndex() (idx index, err error) {
	sign := 1

	var b uint8
	err = binary.Read(d.r, binary.LittleEndian, &b)
	if err != nil {
		return
	}

	var value = int(b & 0x3f)

	if b&0x80 != 0 {
		sign = -1
	}

	if b&0x40 != 0 {
		err = binary.Read(d.r, binary.LittleEndian, &b)
		if err != nil {
			return
		}

		value |= int(b&0x7f) << 6

		if b&0x80 != 0 {
			err = binary.Read(d.r, binary.LittleEndian, &b)
			if err != nil {
				return
			}

			value |= int(b&0x7f) << 13

			if b&0x80 != 0 {
				err = binary.Read(d.r, binary.LittleEndian, &b)
				if err != nil {
					return
				}

				value |= int(b&0x7f) << 20
			}
		}
	}

	return sign * value, nil
}
