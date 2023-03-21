package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"syscall"

	"github.com/leemos-xx/basin/common"
	"github.com/sirupsen/logrus"
)

func Stop(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("Get container %s info error %v", containerName, err)
		return
	}
	pidInt, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		logrus.Errorf("Conver pid from string to int error %v", err)
		return
	}

	if err = syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		logrus.Errorf("Stop container %s error %v", containerName, err)
		return
	}

	containerInfo.Status = common.Stop
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("Json marshal %s error %v", containerName, err)
		return
	}

	dirURL := fmt.Sprintf(common.ContainerDataUrlFormat, containerName)
	configFilePath := dirURL + common.ConfigFileName
	if err := ioutil.WriteFile(configFilePath, newContentBytes, common.Perm0622); err != nil {
		logrus.Errorf("Write file %s error:%v", configFilePath, err)
	}
}
