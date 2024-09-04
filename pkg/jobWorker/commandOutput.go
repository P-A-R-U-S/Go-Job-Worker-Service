package jobWorker

import (
	"errors"
	"io"
	"sync"
)

// CommandOutput implements io.Closer and io.Writer and implement a buffer
//				that can write and read from multiple goroutines concurrently.
// Note should not be created directly, but instead by calling NewOutput.

var (
	// ErrClosedOutput returned when attempting to write to a closed Output.
	ErrClosedOutput = errors.New("cannot write to closed Output")
	// ErrOffsetOutsideContentBounds is returned when the offset is greater than the length of the content.
	ErrOffsetOutsideContentBounds = errors.New("offset is greater than the length of the content")
)

type CommandOutput struct {
	// content store the written data
	content []byte
	// isClosed is true once Close() is called to prevent further writes
	isClosed bool
	// mutex is used to handle concurrent writes and reads
	mutex sync.RWMutex
	// waitCondition is used so goroutines can wait for content to be written or output to be closed
	waitCondition *sync.Cond
}

// NewCommandOutput returns a new instance of CommandOutput.
func NewCommandOutput() *CommandOutput {
	output := CommandOutput{}
	output.waitCondition = sync.NewCond(&output.mutex)

	return &output
}

// Write appends newContent to the content of the Output.
func (output *CommandOutput) Write(newContent []byte) (int, error) {
	output.mutex.Lock()
	defer output.mutex.Unlock()

	if output.isClosed {
		return 0, ErrClosedOutput
	}

	output.content = append(output.content, newContent...)

	output.waitCondition.Broadcast()

	return len(newContent), nil
}

// ReadPartial copies content from the CommandOutput to buffer starting at the given offset
//
//	and not block if less bytes are available than requested.
func (output *CommandOutput) ReadPartial(buffer []byte, off int64) (int, error) {
	output.mutex.RLock()
	defer output.mutex.RUnlock()

	content := output.content

	if off > int64(len(content)) {
		return 0, ErrOffsetOutsideContentBounds
	}

	bytesCopied := copy(buffer, content[off:])

	if bytesCopied+int(off) == len(content) && output.isClosed {
		return bytesCopied, io.EOF
	}

	return bytesCopied, nil
}

// Wait blocks until new content is written to the CommandOutput or the CommandOutput is closed.
func (output *CommandOutput) Wait(nextByteIndex int64) {
	output.mutex.RLock()

	closed := output.isClosed
	contentLength := int64(len(output.content))

	output.mutex.RUnlock()

	// only wait for changes if the output is open or the content contains the next byte to read already
	if closed || contentLength > nextByteIndex {
		return
	}

	output.mutex.Lock()
	for !output.isClosed && nextByteIndex >= int64(len(output.content)) {
		output.waitCondition.Wait()
	}
	output.mutex.Unlock()
}

// Close closes the CommandOutput preventing any further writes.
func (output *CommandOutput) Close() error {
	output.mutex.Lock()
	defer output.mutex.Unlock()

	output.isClosed = true

	output.waitCondition.Broadcast()

	return nil
}
