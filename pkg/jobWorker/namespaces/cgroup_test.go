package namespaces

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

func Test_CGroup(t *testing.T) {
	t.Parallel()

	CPU := 0.5 // half a CPU core
	//IOBytesPerSecond := 100 * MB // 10 MB/s
	MemBytes := 2 * GB // 2 GB

	cgroupName := "fakecgroup" //strings.Replace(uuid.New().String(), "-", "", -1)

	// defer to be sure test cgroup had been removed
	defer func() {
		exist, err := isDirExists(GetCGroupPath(cgroupName))
		if exist && err == nil {
			log.Printf("deleting cgroup directory on defer:%s", GetCGroupPath(cgroupName))
			_ = DeleteCGroup(cgroupName)
		}
	}()

	if err := CreateCGroup(cgroupName); err != nil {
		t.Errorf("could not create cgroup: %v", err)
	}

	if err := AddResourceControl(cgroupName, CpuWeightFile, strconv.Itoa(int(CPU*100))); err != nil {
		t.Errorf("could not add resources into controller:%s, %v", CpuWeightFile, err)
	}
	if err := AddResourceControl(cgroupName, MemoryHighFile, strconv.FormatInt(MemBytes, 10)); err != nil {
		t.Errorf("could not add resources into controller:%s, %v", MemoryHighFile, err)
	}
	//if err := AddResourceControl(cgroupName, IoWeightFile, strconv.FormatInt(IOBytesPerSecond, 10)); err != nil {
	//	t.Errorf("could not add resources into controller:%s, %v", IoWeightFile, err)
	//}

	// assert cgroup exists
	exist, err := isDirExists(GetCGroupPath(cgroupName))
	if !exist || err != nil {
		t.Errorf("expected:%s to exist to represent cgroup", GetCGroupPath(cgroupName))
	}

	// TEST Delete CGroup
	err = DeleteCGroup(cgroupName)
	if err != nil {
		t.Errorf("could not delete cgroup: %v", err)
	}

	//assert file is not there
	exist, err = isDirExists(GetCGroupPath(cgroupName))
	if exist || err != nil {
		t.Errorf("expected cgroup folder: %s NOT to exist to represent cgroup:%s", GetCGroupPath(cgroupName), cgroupName)
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
