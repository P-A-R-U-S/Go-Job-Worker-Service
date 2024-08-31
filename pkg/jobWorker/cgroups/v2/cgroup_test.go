package v2

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func validateCgroupController(cgroupName, controller, value string) (error error) {
	// Set up Cgroup to test with test tmp dir
	// TEST CreateGroup
	testcgroupPath := groupPath(cgroupName)
	//defer func() {
	//	exist, err := exists(testcgroupPath)
	//	if exist || err != nil {
	//		DeleteGroup(cgroupName)
	//	}
	//}()

	err := CreateGroup(cgroupName)
	if err != nil {
		return fmt.Errorf("could not create cgroup: %v", err)
	}
	// assert cgroup exists
	exist, err := exists(testcgroupPath)
	if !exist || err != nil {
		return fmt.Errorf("expected:%s to exist to represent cgroup", testcgroupPath)
	}

	// TEST AddResourceControl
	err = AddResourceControl(cgroupName, controller, value)
	if err != nil {
		return fmt.Errorf("could not add resource controls to cgroup (%s) controller: %v", controller, err)
	}
	// assert cgroup controller files are updated
	controllerValue, err := os.ReadFile(filepath.Join(testcgroupPath, controller))
	if err != nil {
		return fmt.Errorf("could not read cgroup (%s) controller: %v", controller, err)
	}
	if string(controllerValue) != value {
		return fmt.Errorf("controller:(%s) is incorrect: %v", controller, err)
	}
	// TEST DeleteGroup
	err = DeleteGroup(cgroupName)
	if err != nil {
		return fmt.Errorf("could not delete cgroup: %v", err)
	}
	// assert file is not there
	//exist, err = exists(testcgroupPath)
	//if exist || err != nil {
	//	return fmt.Errorf("expected cgroup folder: %s NOT to exist to represent cgroup:%s", testcgroupPath, cgroupName)
	//}
	return nil
}

func Test_CGroup(t *testing.T) {
	cgroupName := "fakeCGroupName"

	t.Parallel()

	testCases := []struct {
		cgroup     string
		controller string
		value      string
	}{
		{cgroupName, CPU_WEIGHT_FILE, strconv.Itoa(50)},
		//{cgroupName, MEMORY_HIGH_FILE, strconv.Itoa(10 * 1024 * 1024 * 1024)},
		//{cgroupName, IO_WEIGHT_FILE, strconv.Itoa(1_000_000)},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf("run for controller:%s, cgroup:%s", tc.controller, tc.cgroup)
		t.Run(testName, func(t *testing.T) {
			err := validateCgroupController(tc.cgroup, tc.controller, tc.value)
			if err != nil {
				t.Errorf("failed: %v", err)
			}
		})
	}
}
