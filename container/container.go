package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/leemos-xx/basin/common"
	"github.com/pkg/errors"
)

func getContainerInfoByName(containerName string) (*common.BaseConfig, error) {
	dirURL := fmt.Sprintf(common.ContainerDataUrlFormat, containerName)
	configFilePath := dirURL + common.ConfigFileName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "read file %s", configFilePath)
	}
	var containerInfo common.BaseConfig
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}
