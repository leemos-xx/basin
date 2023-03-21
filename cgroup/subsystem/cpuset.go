package subsystem

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/leemos-xx/basin/common"
	"github.com/pkg/errors"
)

type CpusetSubSystem struct {
}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}

func (s *CpusetSubSystem) Set(cgroupPath string, config *common.CgroupParam) error {
	if config.CpuSet == "" {
		return nil
	}

	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpuset.cpus"), []byte(config.CpuSet), common.Perm0644); err != nil {
		return errors.Wrapf(err, "set cgroup cpuset failed")
	}

	return nil
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int, config *common.CgroupParam) error {
	if config.CpuSet == "" {
		return nil
	}
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return errors.Wrapf(err, "get cgroup %s", cgroupPath)

	}
	if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), common.Perm0644); err != nil {
		return errors.Wrapf(err, "set cgroup proc failed")
	}
	return nil
}

func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(subsystemCgroupPath)
}
