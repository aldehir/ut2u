package uz2

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
)

const (
	blockSize         = 32768
	maxCompressedSize = 33096
)

type Writer struct {
	w          io.Writer
	buf        *bytes.Buffer
	compressed *bytes.Buffer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:          w,
		buf:        bytes.NewBuffer(make([]byte, 0, blockSize)),
		compressed: bytes.NewBuffer(make([]byte, 0, maxCompressedSize)),
	}
}

func (u *Writer) Write(p []byte) (int, error) {
	_, err := u.buf.Write(p)
	if err != nil {
		return 0, err
	}

	for u.buf.Len() >= blockSize {
		u.writeChunk(blockSize)
	}

	return len(p), nil
}

func (u *Writer) writeChunk(n int) error {
	if u.buf.Len() == 0 {
		return nil // Nothing to compress
	} else if n < 0 {
		n = u.buf.Len() // Compress all remaining bytes in the buffer
	} else if n > u.buf.Len() {
		panic("not enough bytes to write")
	}

	u.compressed.Reset()
	compressor, _ := zlib.NewWriterLevel(u.compressed, zlib.DefaultCompression)

	_, err := io.CopyN(compressor, u.buf, int64(n))
	if err != nil {
		return err
	}

	err = compressor.Close()
	if err != nil {
		return err
	}

	// Write compressed size as uint32 little-endian
	err = binary.Write(u.w, binary.LittleEndian, uint32(u.compressed.Len()))
	if err != nil {
		return err
	}

	// Write uncompressed size as uint32 little-endian
	err = binary.Write(u.w, binary.LittleEndian, uint32(n))
	if err != nil {
		return err
	}

	// Copy compressed bytes to the underlying writer
	_, err = io.Copy(u.w, u.compressed)
	if err != nil {
		return err
	}

	return nil
}

func (u *Writer) Flush() error {
	for u.buf.Len() >= blockSize {
		u.writeChunk(blockSize)
	}

	return u.writeChunk(-1)
}

func (u *Writer) Close() error {
	return u.Flush()
}
