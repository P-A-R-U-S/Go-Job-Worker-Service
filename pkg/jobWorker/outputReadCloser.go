package jobWorker

import (
	"errors"
)

var (
	ErrOutputMissing = errors.New("OutputReader's CommandOutput is nil")
	ErrReaderClosed  = errors.New("OutputReader is closed")
)

// OutputReadCloser implements io.ReadCloser interface to read from the provided CommandOutput
//
//	and close output if it is no longer need it .
type OutputReadCloser struct {
	output *CommandOutput
	// readIndex is the index of the next byte to read from the Output
	readIndex int64
	isClosed  bool
}

func NewOutputReadCloser(output *CommandOutput) *OutputReadCloser {
	return &OutputReadCloser{output: output, readIndex: 0}
}

// Read reads from the Output and returns the number of bytes read and an error if any.
//
//	Wait for changes to the CommandOutput if no content is available to read.
//	Returns EOF if the CommandOutput is closed and all the content has been read.
func (outputReadCloser *OutputReadCloser) Read(buffer []byte) (n int, err error) {
	if outputReadCloser.isClosed {
		return 0, ErrReaderClosed
	}

	if outputReadCloser.output == nil {
		return 0, ErrOutputMissing
	}

	if len(buffer) == 0 {
		return 0, nil
	}

	bytesRead, err := outputReadCloser.output.ReadPartial(buffer, outputReadCloser.readIndex)
	if err != nil {
		return bytesRead, err
	}

	// If bytesRead is zero then wait for changes to the Output and read again.
	if bytesRead == 0 {
		outputReadCloser.output.Wait(outputReadCloser.readIndex)

		bytesRead, err = outputReadCloser.output.ReadPartial(buffer, outputReadCloser.readIndex)
		if err != nil {
			return bytesRead, err
		}
	}

	outputReadCloser.readIndex += int64(bytesRead)

	return bytesRead, nil
}

func (outputReadCloser *OutputReadCloser) Close() error {
	if outputReadCloser.output == nil {
		return ErrOutputMissing
	}

	if outputReadCloser.isClosed {
		return ErrReaderClosed
	}

	outputReadCloser.isClosed = true

	return nil
}
