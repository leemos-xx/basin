package container

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/leemos-xx/basin/common"
	"github.com/sirupsen/logrus"
)

func ReadContainerLog(containerName string) {
	logFileLocation := fmt.Sprintf(common.ContainerDataUrlFormat, containerName) + common.LogFileName
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		logrus.Errorf("Log container open file %s error %v", logFileLocation, err)
		return
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Errorf("Log container read file %s error %v", logFileLocation, err)
		return
	}
	_, err = fmt.Fprint(os.Stdout, string(content))
	if err != nil {
		logrus.Errorf("Log container Fprint  error %v", err)
		return
	}
}
