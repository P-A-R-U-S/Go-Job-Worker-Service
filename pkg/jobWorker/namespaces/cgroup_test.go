package namespaces

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_CGroup(t *testing.T) {
	t.Parallel()
	CPU := 0.5                            // half a CPU core
	IOBytesPerSecond := int64(10_000_000) // 10 MB/s
	MemBytes := int64(1_000_000_000)      // 1 GB

	cgroupName := "fakecgroup" //strings.Replace(uuid.New().String(), "-", "", -1)

	// Set up Cgroup to test with test tmp dir
	// TEST CreateGroup
	cgroupDir := filepath.Join(rootCgroupPath, cgroupName)

	defer func() {
		exist, err := isDirExists(cgroupDir)
		if exist && err == nil {
			log.Printf("deleting cgroup directory on defer:%s", cgroupDir)
			_ = CleanupCGroup(cgroupDir)
		}
	}()

	// defer to be sure test cgroup had been removed

	err := CreateCGroup(cgroupDir, "", CPU, IOBytesPerSecond, MemBytes)
	if err != nil {
		t.Errorf("could not create cgroup: %v", err)
	}
	// assert cgroup exists
	exist, err := isDirExists(cgroupDir)
	if !exist || err != nil {
		t.Errorf("expected:%s to exist to represent cgroup", cgroupDir)
	}

	// TEST DeleteGroup
	err = CleanupCGroup(cgroupName)
	if err != nil {
		t.Errorf("could not delete cgroup: %v", err)
	}

	//assert file is not there
	exist, err = isDirExists(cgroupDir)
	if exist || err != nil {
		t.Errorf("expected cgroup folder: %s NOT to exist to represent cgroup:%s", cgroupDir, cgroupName)
	}

	// just to let system handle cgroup cleanup
	time.Sleep(1000)
}

// exists returns whether the given file or directory exists
func isDirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
