package container

import (
	"fmt"
	"os"

	"github.com/leemos-xx/basin/common"
	"github.com/sirupsen/logrus"
)

func Remove(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("Get container %s info error %v", containerName, err)
		return
	}

	if containerInfo.Status != common.Stop {
		logrus.Errorf("Couldn't remove running container")
		return
	}
	dirURL := fmt.Sprintf(common.ContainerDataUrlFormat, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		logrus.Errorf("Remove file %s error %v", dirURL, err)
	}
	err = deleteWorkSpace(containerName, containerInfo.Volume)
	if err != nil {
		logrus.Errorf("DeleteWorkSpace error %v", err)
	}
}
