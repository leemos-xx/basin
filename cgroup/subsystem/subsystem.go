package subsystem

import (
	"bufio"
	"os"
	"path"
	"strings"

	"github.com/leemos-xx/basin/common"
	"github.com/sirupsen/logrus"
)

type Subsystem interface {
	Name() string
	Set(path string, config *common.CgroupParam) error
	Apply(path string, pid int, config *common.CgroupParam) error
	Remove(path string) error
}

var SupportedSubSystems = []Subsystem{
	&CpusetSubSystem{},
	&MemorySubSystem{},
	&CpuSubSystem{},
}

func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := findCgroupMountpoint(subsystem)
	absPath := path.Join(cgroupRoot, cgroupPath)
	if !autoCreate {
		return absPath, nil
	}

	_, err := os.Stat(absPath)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(absPath, common.Perm0755)
		return absPath, err
	}

	return absPath, nil
}

func findCgroupMountpoint(subsystem string) string {
	mountinfoFile, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer mountinfoFile.Close()

	scanner := bufio.NewScanner(mountinfoFile)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " ")

		subsystems := strings.Split(fields[len(fields)-1], ",")
		for _, item := range subsystems {
			if item == subsystem {
				return fields[common.MountPointIndex]
			}
		}
	}

	if err = scanner.Err(); err != nil {
		logrus.Errorf("read mountinfo err: %v", err)
		return ""
	}
	return ""
}
