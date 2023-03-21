package container

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/leemos-xx/basin/cgroup"
	"github.com/leemos-xx/basin/common"
	"github.com/leemos-xx/basin/network"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func Run(param *common.RunParam) {
	// 随机生成容器的id
	containerId := randStringBytes(common.IdLength)
	if len(param.ContainerName) == 0 {
		param.ContainerName = containerId
	}

	// TODO 创建子进程，即实际的容器进程
	subprocess, writePipe, err := newSubprocess(param.ContainerName, param.Volume, param.ImageName, param.Envs, param.TTY)
	if err != nil {
		logrus.Errorf("new subprocess err: %v", err)
		return
	}
	if err := subprocess.Start(); err != nil {
		logrus.Errorf("run subprocess.Start err: %v", err)
		return
	}

	// 将容器运行信息记录到配置文件中
	err = recordConfig(subprocess.Process.Pid, containerId, param.ContainerName, param.Volume, param.ContainerCommands)
	if err != nil {
		logrus.Errorf("record container config err: %v", err)
		return
	}

	// 根据参数信息进行资源限制，并将子进程（容器进程）加入该资源组
	cgroupManager := cgroup.NewCgroupManager(common.CgroupName)
	defer cgroupManager.Destroy()
	_ = cgroupManager.Set(param.CgroupConfig)
	_ = cgroupManager.Apply(subprocess.Process.Pid, param.CgroupConfig)

	// 如果有指定网络，则尝试将容器接入该网络
	if param.Network != "" {
		network.Init()
		containerInfo := &common.BaseConfig{
			Id:          containerId,
			Pid:         strconv.Itoa(subprocess.Process.Pid),
			Name:        param.ContainerName,
			PortMapping: param.PortMapping,
		}
		if err = network.Connect(param.Network, containerInfo); err != nil {
			logrus.Errorf("connect network err: %v", err)
			return
		}
	}

	// 当子进程状态就绪后，将容器命令发送给子进程
	sendContainerCommand(param.ContainerCommands, writePipe)

	if param.TTY {
		_ = subprocess.Wait()
		deleteContainerInfo(param.ContainerName)
		deleteWorkSpace(param.ContainerName, param.Volume)
	}
}

func newSubprocess(containerName, volume, imageName string, envSlice []string, tty bool) (*exec.Cmd, *os.File, error) {
	// 在子进程会通过readPipe监听命令信息，当父进程（本进程）为子进程分配好cgroup、network等资源后，在执行实际的逻辑
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "new pipe error")
	}

	// `/proc/self/exe`为当前进程的运行信息，通过ReadLink可以获得当前程序的绝对路径
	initCmd, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return nil, nil, errors.Wrap(err, "readLink /proc/self/exe failed")
	}

	subprocessCmd := exec.Command(initCmd, "init")
	subprocessCmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	// 开启TTY时，使用系统输入输出，不开启TTY时，输出到日志文件
	if tty {
		subprocessCmd.Stdin = os.Stdin
		subprocessCmd.Stdout = os.Stdout
		subprocessCmd.Stderr = os.Stderr
	} else {
		containerDataUrl := fmt.Sprintf(common.ContainerDataUrlFormat, containerName)

		if err := os.MkdirAll(containerDataUrl, common.Perm0622); err != nil {
			return nil, nil, errors.Wrapf(err, "mkdir[%s] err while new subprocess", containerDataUrl)
		}

		containerLogFileUrl := containerDataUrl + common.LogFileName
		containerLogFile, err := os.Create(containerLogFileUrl)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "create file[%s] failed while new subprocess", containerLogFileUrl)
		}
		subprocessCmd.Stdout = containerLogFile
	}

	// 将readPipe以ExtraFiles的形式传递给子进程
	subprocessCmd.ExtraFiles = []*os.File{readPipe}
	// 设置环境变量
	subprocessCmd.Env = append(os.Environ(), envSlice...)
	// TODO 将overlayfs联合挂载后的目录作为子进程的默认目录
	subprocessCmd.Dir = fmt.Sprintf(common.MergedDirFormat, containerName)

	// 实际处理子进程的workspace
	err = NewWorkspace(containerName, imageName, volume)
	if err != nil {
		return nil, nil, err
	}

	return subprocessCmd, writePipe, nil
}

func recordConfig(containerPid int, containerId, containerName, volume string, containerCommands []string) error {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(containerCommands, " ")
	config := &common.BaseConfig{
		Pid:         strconv.Itoa(containerPid),
		Id:          containerId,
		Name:        containerName,
		Volume:      volume,
		Command:     command,
		CreatedTime: createTime,
		Status:      common.Running,
	}

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "record container info err")
	}

	containerDataUrl := fmt.Sprintf(common.ContainerDataUrlFormat, containerName)
	if err = os.MkdirAll(containerDataUrl, common.Perm0622); err != nil {
		return errors.Wrapf(err, "mkdir[%s] failed", containerDataUrl)
	}

	containerConfigFileUrl := containerDataUrl + "/" + common.ConfigFileName
	containerConfigFile, err := os.Create(containerConfigFileUrl)
	defer containerConfigFile.Close()
	if err != nil {
		return errors.Wrapf(err, "create containerConfigFile[%s] failed", containerConfigFileUrl)
	}

	if _, err = containerConfigFile.WriteString(string(jsonBytes)); err != nil {
		return errors.Wrapf(err, "write container config file[%s] failed", containerConfigFileUrl)
	}

	return nil
}

func sendContainerCommand(containerCommand []string, writePipe *os.File) {
	commandLine := strings.Join(containerCommand, " ")
	_, _ = writePipe.WriteString(commandLine)
	_ = writePipe.Close()
}

func randStringBytes(n int) string {
	letters := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func deleteContainerInfo(containerName string) {
	dirURL := fmt.Sprintf(common.ContainerDataUrlFormat, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		logrus.Errorf("remove dir %s failed: %v", dirURL, err)
	}
}
