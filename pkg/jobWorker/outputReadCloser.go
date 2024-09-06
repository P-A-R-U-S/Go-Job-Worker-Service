package jobWorker

import (
	"errors"
	"sync"
)

var (
	ErrOutputMissing = errors.New("OutputReader's CommandOutput is nil")
	ErrReaderClosed  = errors.New("OutputReader is closed")
)

// OutputReadCloser implements io.ReadCloser interface to read from the provided CommandOutput
//
//	and close output if it is no longer need it .
type OutputReadCloser struct {
	output  *CommandOutput
	rwmutex sync.RWMutex
	// readIndex is the index of the next byte to read from the Output
	readIndex int64
	//isClosed is true if was reader closed
	isClosed bool
}

func NewOutputReadCloser(output *CommandOutput) *OutputReadCloser {
	return &OutputReadCloser{output: output, readIndex: 0}
}

// Read reads from the Output and returns the number of bytes read and an error if any.
//
//	Wait for changes to the CommandOutput if no content is available to read.
//	Returns EOF if the CommandOutput is closed and all the content has been read.
func (orc *OutputReadCloser) Read(buffer []byte) (n int, err error) {
	orc.rwmutex.RLock()
	defer orc.rwmutex.RUnlock()

	if orc.isClosed {
		return 0, ErrReaderClosed
	}

	if orc.output == nil {
		return 0, ErrOutputMissing
	}

	if len(buffer) == 0 {
		return 0, nil
	}

	bytesRead, err := orc.output.ReadPartial(buffer, orc.readIndex)
	if err != nil {
		return bytesRead, err
	}

	// If bytesRead is zero then wait for changes to the Output and read again.
	if bytesRead == 0 {
		orc.output.Wait(orc.readIndex)

		bytesRead, err = orc.output.ReadPartial(buffer, orc.readIndex)
		if err != nil {
			return bytesRead, err
		}
	}

	orc.readIndex += int64(bytesRead)

	return bytesRead, nil
}

func (orc *OutputReadCloser) Close() error {
	if orc.isClosed {
		return ErrReaderClosed
	}

	if orc.output == nil {
		return ErrOutputMissing
	}

	orc.rwmutex.RLock()
	defer orc.rwmutex.RUnlock()

	orc.isClosed = true
	orc.output = nil

	return nil
}
