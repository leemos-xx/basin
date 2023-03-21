package cgroup

import (
	"github.com/leemos-xx/basin/cgroup/subsystem"
	"github.com/leemos-xx/basin/common"
	"github.com/pkg/errors"
)

type Manager struct {
	Path   string
	Config *common.CgroupParam
}

func NewCgroupManager(path string) *Manager {
	return &Manager{
		Path: path,
	}
}

func (c *Manager) Set(config *common.CgroupParam) error {
	for _, supportedSubSystem := range subsystem.SupportedSubSystems {
		err := supportedSubSystem.Set(c.Path, config)
		if err != nil {
			return errors.Wrapf(err, "set subsystem[%s] failed", supportedSubSystem.Name())
		}
	}
	return nil
}

func (c *Manager) Apply(pid int, config *common.CgroupParam) error {
	for _, supportedSubSystem := range subsystem.SupportedSubSystems {
		err := supportedSubSystem.Apply(c.Path, pid, config)
		if err != nil {
			return errors.Wrapf(err, "apply subsystem[%s] failed", supportedSubSystem.Name())
		}
	}
	return nil
}

func (c *Manager) Destroy() error {
	for _, supportedSubSystem := range subsystem.SupportedSubSystems {
		if err := supportedSubSystem.Remove(c.Path); err != nil {
			return errors.Wrapf(err, "remove cgroup %s failed", supportedSubSystem.Name())
		}
	}
	return nil
}
