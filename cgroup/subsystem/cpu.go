package subsystem

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"

	"github.com/leemos-xx/basin/common"
)

const (
	PeriodDefault = 100000
	Percent       = 100
)

type CpuSubSystem struct {
}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}

func (s *CpuSubSystem) Set(cgroupPath string, config *common.CgroupParam) error {
	if config.CpuCfsQuota == 0 && config.CpuShare == "" {
		return nil
	}
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return err
	}

	if config.CpuShare != "" {
		if err = ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpu.shares"), []byte(config.CpuShare), common.Perm0644); err != nil {
			return errors.Wrapf(err, "set cgroup cpu share failed")
		}
	}

	if config.CpuCfsQuota != 0 {
		if err = ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpu.cfs_period_us"), []byte(string(PeriodDefault)), common.Perm0644); err != nil {
			return errors.Wrapf(err, "set cgroup cpu period us failed")
		}
		if err = ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpu.cfs_quota_us"), []byte(strconv.Itoa(PeriodDefault/Percent*config.CpuCfsQuota)), common.Perm0644); err != nil {
			return errors.Wrapf(err, "set cgroup cpu quoto us failed")
		}
	}
	return nil
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int, config *common.CgroupParam) error {
	if config.CpuCfsQuota == 0 && config.CpuShare == "" {
		return nil
	}

	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return errors.Wrapf(err, "get cgroup %s failed", cgroupPath)
	}
	if err = ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), common.Perm0644); err != nil {
		return errors.Wrapf(err, "set cgroup proc failed")
	}
	return nil
}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	subsystemCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(subsystemCgroupPath)
}
