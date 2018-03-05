package buffer

import (
	"io"
	"errors"
)

const MinRead = 512

var (
	ErrTooLarge          = errors.New("io buffer: too large")
	ErrNegativeCount     = errors.New("io buffer: negative count")
	ErrInvalidWriteCount = errors.New("io buffer: invalid write count")
)

// IoBuffer
type IoBuffer struct {
	buf []byte // contents: buf[off : len(buf)]
	off int    // read from &buf[off], write to &buf[len(buf)]
}

func (b *IoBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	if b.off >= len(b.buf) {
		b.Reset()
	}

	if free := cap(b.buf) - len(b.buf); free < MinRead {
		// not enough space at end
		newBuf := b.buf
		if b.off+free < MinRead {
			// not enough space using beginning of buffer;
			// double buffer capacity
			newBuf = makeSlice(2*cap(b.buf) + MinRead)
		}
		copy(newBuf, b.buf[b.off:])
		b.buf = newBuf[:len(b.buf)-b.off]
		b.off = 0
	}

	m, err := r.Read(b.buf[len(b.buf):len(b.buf)+MinRead])

	b.buf = b.buf[0: len(b.buf)+m]
	n += int64(m)

	return
}

func (b *IoBuffer) WriteTo(w io.Writer) (n int64, err error) {
	for b.off < len(b.buf) {
		nBytes := b.Len()
		m, e := w.Write(b.buf[b.off:])

		if m > nBytes {
			panic(ErrInvalidWriteCount)
		}

		b.off += m
		n += int64(m)

		if e != nil {
			return n, e
		}
	}

	return
}

func (b *IoBuffer) Append(data []byte) error {
	if b.off >= len(b.buf) {
		b.Reset()
	}

	dataLen := len(data)

	if free := cap(b.buf) - len(b.buf); free < dataLen {
		// not enough space at end
		newBuf := b.buf
		if b.off+free < dataLen {
			// not enough space using beginning of buffer;
			// double buffer capacity
			newBuf = makeSlice(2*cap(b.buf) + dataLen)
		}
		copy(newBuf, b.buf[b.off:])
		b.buf = newBuf[:len(b.buf)-b.off]
		b.off = 0
	}

	m := copy(b.buf[len(b.buf):len(b.buf)+dataLen], data)
	b.buf = b.buf[0: len(b.buf)+m]

	return nil
}

func (b *IoBuffer) Peek(n int) []byte {
	if len(b.buf)-b.off < n {
		return nil
	}

	return b.buf[b.off:b.off+n]
}

func (b *IoBuffer) Bytes() []byte {
	return b.buf[b.off:]
}

func (b *IoBuffer) String() string {
	return string(b.buf[b.off:])
}

func (b *IoBuffer) Len() int {
	return len(b.buf) - b.off
}

func (b *IoBuffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
}

func (b *IoBuffer) available() int {
	return len(b.buf) - b.off
}

func makeSlice(n int) []byte {
	// TODO: handle large buffer
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]byte, n)
}

func NewIoBuffer(bufSize int) *IoBuffer {
	buf := make([]byte, 0, bufSize)

	return &IoBuffer{buf: buf}
}

func NewIoBufferString(s string) *IoBuffer {
	return &IoBuffer{buf: []byte(s)}
}
