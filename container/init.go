package container

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/leemos-xx/basin/common"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const fdIndex = 3

func RunContainerInitProcess() error {
	containerCommand := readContainerCommand()
	if len(containerCommand) == 0 {
		return errors.New("run container get user command error, containerCommand is nil")
	}

	err := setupMount()
	if err != nil {
		logrus.Errorf("setup mount failed: %v", err)
		return err
	}

	path, err := exec.LookPath(containerCommand[0])
	if err != nil {
		logrus.Errorf("Exec loop path error %v", err)
		return err
	}

	if err = syscall.Exec(path, containerCommand[0:], os.Environ()); err != nil {
		logrus.Errorf("RunContainerInitProcess exec :" + err.Error())
	}
	return nil
}

func readContainerCommand() []string {
	pipe := os.NewFile(uintptr(fdIndex), "pipe")
	defer pipe.Close()
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		logrus.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func setupMount() error {
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrapf(err, "get current location failed")
	}

	if err = pivotRoot(pwd); err != nil {
		return errors.Wrapf(err, "pivot root failed")
	}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		return errors.Wrapf(err, "mount proc failed")
	}
	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		return errors.Wrapf(err, "mount tmpfs failed")
	}

	return nil
}

func pivotRoot(root string) error {
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		if !os.IsExist(err) {
			return errors.Wrap(err, "mount /")
		}
	}

	if err := syscall.Mount(root, root, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return errors.Wrap(err, "mount rootfs to itself")
	}

	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, common.Perm0777); err != nil {
		return errors.Wrapf(err, "mkdir[%s] failed", pivotDir)
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return errors.Wrapf(err, "pivot_root")
	}

	if err := syscall.Chdir("/"); err != nil {
		return errors.Wrapf(err, "chdir /")
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return errors.Wrapf(err, "unmount pivot_root dir")
	}

	return os.Remove(pivotDir)
}
