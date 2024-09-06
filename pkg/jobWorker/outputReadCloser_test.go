package jobWorker

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
	"testing"
	"time"
)

func Test_OutputReadCloser_Expecting_Read_returns_full_content(t *testing.T) {
	t.Parallel()

	output := NewCommandOutput()

	_, err := output.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Expected no error writing content, got %v", err)
	}

	outputReadCloser := NewOutputReadCloser(output)

	// read returns content from the start
	buffer := make([]byte, 4)

	bytesRead, err := outputReadCloser.Read(buffer)
	if err != nil {
		t.Fatalf("Expected no error reading initial content, got %v", err)
	}

	expectedBytesRead := 4
	if bytesRead != expectedBytesRead {
		t.Errorf("Expected %d bytes read, got %d", expectedBytesRead, bytesRead)
	}

	expectedContent := "hell"
	if string(buffer) != expectedContent {
		t.Errorf("Expected %q, got %q", expectedContent, string(buffer))
	}

	// invoking read again returns the next chunk of data
	_, err = output.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Expected no error writing more content, got %v", err)
	}

	_, err = outputReadCloser.Read(buffer)
	if err != nil {
		t.Fatalf("Expected no error reading again, got %v", err)
	}

	if string(buffer) != "ohel" {
		t.Errorf("Expected %q since Read should continue reading from next byte, got %q", "ohel", string(buffer))
	}

	// reading the rest of the content
	err = output.Close()
	if err != nil {
		t.Fatalf("Expected no error closing output, got %v", err)
	}

	bytesRead, err = outputReadCloser.Read(buffer)
	if err == nil {
		t.Fatalf("Expected an error when end of content is reached, but got nil")
	}

	if !errors.Is(err, io.EOF) {
		t.Errorf("Expected %q when end of content is reached, got %q", io.EOF, err)
	}

	if string(buffer[0:bytesRead]) != "lo" {
		t.Errorf("Expected %q since Read should return rest of the content, got %q", "lo", string(buffer))
	}
}

func Test_OutputReadCloser_GetsNewContentAsWritten(t *testing.T) {
	t.Parallel()

	output := NewCommandOutput()

	var waitGroup sync.WaitGroup

	// multiple readers should get the same content
	for range 3 {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			outputReadCloser := NewOutputReadCloser(output)

			var chunksRead []string

			for {
				buffer := make([]byte, 10)

				bytesRead, err := outputReadCloser.Read(buffer)
				if err != nil && !errors.Is(err, io.EOF) {
					t.Errorf("Expected no error reading content, got %v", err)
				}

				if bytesRead == 0 && !errors.Is(err, io.EOF) {
					// bytesRead should only be zero if no bytes are available to read
					// and output is closed
					t.Error("Expected EOF when zero bytes read. Read should not return with zero bytes read otherwise.")
				}

				chunksRead = append(chunksRead, string(buffer[0:bytesRead]))

				if errors.Is(err, io.EOF) {
					break
				}
			}

			// validate 4 chunks are returned by read
			// hello
			// world
			// test
			// (empty) EOF
			if len(chunksRead) == 4 {
				if chunksRead[0] != "hello" || chunksRead[1] != "world" || chunksRead[2] != "test" || chunksRead[3] != "" {
					t.Errorf("Expected %q, %q, %q, %q, got %q, %q, %q, %q", "hello", "world", "test", "", chunksRead[0], chunksRead[1], chunksRead[2], chunksRead[3])
				}
			} else {
				t.Errorf("Expected 4 chunks read, got %d", len(chunksRead))
			}
		}()
	}

	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()

		for _, content := range []string{"hello", "world", "test"} {
			_, err := output.Write([]byte(content))
			if err != nil {
				t.Errorf("Expected no error writing content, got %v", err)
			}

			// sleep to prove readers get content as it is written
			time.Sleep(10 * time.Millisecond)
		}

		err := output.Close()
		if err != nil {
			t.Errorf("Expected no error closing output, got %v", err)
		}
	}()

	waitGroup.Wait()
}

func Test_OutputReadCloser_Expecting_calling_Close_from_different_goroutine_close_Reader(t *testing.T) {
	t.Parallel()

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	output := NewCommandOutput()

	outputReadCloser := NewOutputReadCloser(output)

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()

		for {
			log.Printf("write data")
			_, err := output.Write([]byte("hello world"))
			if err != nil {
				if errors.Is(err, ErrClosedOutput) {
					break
				} else {
					t.Errorf("Expected no error writing content, got %v", err)
				}
			}
			// sleep to prove readers get content as it is written
			time.Sleep(10 * time.Millisecond)
		}
	}()

	buffer := make([]byte, 4)

	go func() {
		select {
		case <-cancelCtx.Done():
			log.Printf("receving cancel")
			if err := outputReadCloser.Close(); err != nil {
				t.Errorf("Expected no error closing reader, got %v", err)
			}
		}
	}()

	// cancel reader from another process
	go func() {
		time.Sleep(20 * time.Millisecond)
		log.Printf("triggering cancel")
		cancel()
	}()

	for {
		log.Printf("read data")
		if _, err := outputReadCloser.Read(buffer); err != nil {
			if errors.Is(err, ErrReaderClosed) {
				break
			} else {
				t.Errorf("Expected no error reading content, got %v", err)
			}
		}
		time.Sleep(20 * time.Millisecond)
	}

	if err := output.Close(); err != nil {
		t.Errorf("Expected no error closing output, got %v", err)
	}

	waitGroup.Wait()
}
