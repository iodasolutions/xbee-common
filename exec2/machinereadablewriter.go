package exec2

import (
	"bytes"
	"io"
	"os"
)

// A Writer to accumulate data in memory and eventually to
// a file and a secondary Writer (possibly stdout or stderr)
type MachineReadableWriter struct {
	buf     *bytes.Buffer
	writers []io.Writer
}

func NewMachineOnlyReadableWriter() *MachineReadableWriter {
	return &MachineReadableWriter{
		buf: &bytes.Buffer{},
	}
}

func NewStdOutMachineReadableWriter() *MachineReadableWriter {
	return &MachineReadableWriter{
		buf:     &bytes.Buffer{},
		writers: []io.Writer{os.Stdout},
	}
}

func NewStdErrMachineReadableWriter() *MachineReadableWriter {
	return &MachineReadableWriter{
		buf:     &bytes.Buffer{},
		writers: []io.Writer{os.Stderr},
	}
}

func (mrw *MachineReadableWriter) Write(p []byte) (n int, err error) {
	if mrw.buf != nil {
		n, err = mrw.buf.Write(p)
		if err != nil {
			return
		}
	}
	for _, writer := range mrw.writers {
		n, err = writer.Write(p)
		if err != nil {
			return
		}
	}
	return
}

func (mrw *MachineReadableWriter) String() string {
	if mrw.buf == nil {
		return ""
	}
	return mrw.buf.String()
}
