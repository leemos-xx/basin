package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/leemos-xx/basin/common"
	"github.com/sirupsen/logrus"
)

func ListContainers() {
	files, err := ioutil.ReadDir(common.ContainerDataUrl)
	if err != nil {
		logrus.Errorf("read dir %s error %v", common.ContainerDataUrl, err)
		return
	}
	containers := make([]*common.BaseConfig, 0, len(files))
	for _, file := range files {
		if file.Name() == "network" {
			continue
		}
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			logrus.Errorf("get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, err = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	if err != nil {
		logrus.Errorf("Fprint error %v", err)
	}
	for _, item := range containers {
		_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
		if err != nil {
			logrus.Errorf("Fprint error %v", err)
		}
	}
	if err = w.Flush(); err != nil {
		logrus.Errorf("Flush error %v", err)
	}
}

func getContainerInfo(file os.FileInfo) (*common.BaseConfig, error) {
	containerName := file.Name()
	configFileDir := fmt.Sprintf(common.ContainerDataUrlFormat, containerName)
	configFileDir = configFileDir + common.ConfigFileName

	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		logrus.Errorf("read file %s error %v", configFileDir, err)
		return nil, err
	}
	info := new(common.BaseConfig)
	if err = json.Unmarshal(content, info); err != nil {
		logrus.Errorf("json unmarshal error %v", err)
		return nil, err
	}

	return info, nil
}
