package uz2

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
)

type reader struct {
	r   io.Reader
	buf *bytes.Buffer
}

func NewReader(r io.Reader) io.Reader {
	return &reader{
		r:   r,
		buf: bytes.NewBuffer(make([]byte, 0, blockSize)),
	}
}

func (u *reader) Read(p []byte) (int, error) {
	for u.buf.Len() < len(p) {
		err := u.readChunk()
		if err == io.EOF {
			break
		} else {
			return 0, err
		}
	}

	return u.buf.Read(p)
}

func (u *reader) readChunk() error {
	var compSize, uncompSize uint32

	err := binary.Read(u.r, binary.LittleEndian, &compSize)
	if err != nil {
		return err
	}

	err = binary.Read(u.r, binary.LittleEndian, &uncompSize)
	if err != nil {
		return err
	}

	compressed := bytes.NewBuffer(make([]byte, 0, maxCompressedSize))
	_, err = io.CopyN(compressed, u.r, int64(compSize))
	if err != nil {
		return err
	}

	decompressor, err := zlib.NewReader(compressed)
	if err != nil {
		return err
	}

	defer decompressor.Close()

	_, err = io.Copy(u.buf, decompressor)
	if err != nil {
		return err
	}

	return nil
}
