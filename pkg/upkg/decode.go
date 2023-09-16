package upkg

import (
	"io"

	"github.com/aldehir/ut2u/pkg/encoding/ue2"
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
	err = ue2.Decode(d.r, &d.pkg.h)
	if err != nil {
		return
	}

	if d.pkg.h.Version >= 68 {
		var genCount uint32

		err = ue2.Decode(d.r, &genCount)
		if err != nil {
			return
		}

		var gen generation
		for i := 0; i < int(genCount); i++ {
			err = ue2.Decode(d.r, &gen)
			if err != nil {
				return
			}

			d.pkg.gen = append(d.pkg.gen, gen)
		}
	}

	return nil
}

func (d *Decoder) readNames() (err error) {
	_, err = d.r.Seek(int64(d.pkg.h.NameOffset), io.SeekStart)
	if err != nil {
		return
	}

	var n name

	d.pkg.names = make([]name, 0, d.pkg.h.NameCount)
	for i := 0; i < int(d.pkg.h.NameCount); i++ {
		err = ue2.Decode(d.r, &n)
		if err != nil {
			return
		}

		d.pkg.names = append(d.pkg.names, n)
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
		err = ue2.Decode(d.r, &imp)
		if err != nil {
			return
		}

		d.pkg.imports = append(d.pkg.imports, imp)
	}

	return nil
}
