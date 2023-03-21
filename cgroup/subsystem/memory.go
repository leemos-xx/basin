package subsystem

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/leemos-xx/basin/common"
	"github.com/pkg/errors"
)

type MemorySubSystem struct {
}

// Name 返回cgroup名字
func (s *MemorySubSystem) Name() string {
	return "memory"
}

func (s *MemorySubSystem) Set(cgroupPath string, config *common.CgroupParam) error {
	if config.MemoryLimit == "" {
		return nil
	}
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "memory.limit_in_bytes"), []byte(config.MemoryLimit), common.Perm0644); err != nil {
		return errors.Wrapf(err, "set cgroup memory failed")
	}
	return nil
}

func (s *MemorySubSystem) Apply(cgroupPath string, pid int, config *common.CgroupParam) error {
	if config.MemoryLimit == "" {
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

func (s *MemorySubSystem) Remove(cgroupPath string) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(subsystemCgroupPath)
}
