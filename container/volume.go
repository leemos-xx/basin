package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/leemos-xx/basin/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewWorkspace(containerName, imageName, volume string) error {
	// 创建lower层
	err := createLower(containerName, imageName)
	if err != nil {
		return err
	}

	// 创建upper&work层
	err = createUpperWork(containerName)
	if err != nil {
		return err
	}

	// 通过overlayfs进行联合挂载
	err = mountOverlayFS(containerName)
	if err != nil {
		logrus.Errorf("mount overlay fs err: %v", err)
	}

	// 如果指定了其他卷，则在此处处理挂载
	if volume != "" {
		urls := strings.Split(volume, ":")
		if len(urls) == 2 && urls[0] != "" && urls[1] != "" {
			err = mountVolume(containerName, urls[0], urls[1])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func deleteWorkSpace(containerName, volume string) error {
	if volume != "" {
		urls := strings.Split(volume, ":")
		if len(urls) == 2 && urls[0] != "" && urls[1] != "" {
			err := umountVolume(containerName, urls[0], urls[1])
			if err != nil {
				return err
			}
		}
	}

	err := removeDirs(containerName)
	if err != nil {
		return errors.Wrap(err, "remove dirs")
	}

	err = umountOverlayFS(containerName)
	if err != nil {
		return errors.Wrap(err, "umount overlayfs")
	}

	root := common.RootUrl + containerName
	if err = os.RemoveAll(root); err != nil {
		return errors.Wrap(err, "remove root")
	}

	return nil
}

func createLower(containerName, imageName string) error {
	imageUrl := common.RootUrl + imageName + ".tar"
	lowerUrl := fmt.Sprintf(common.LowerDirFormat, containerName)

	if err := os.MkdirAll(lowerUrl, common.Perm0622); err != nil {
		return errors.Wrapf(err, "mkdir[%s] failed", lowerUrl)
	}
	if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", lowerUrl).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "untar dir[%s] failed", lowerUrl)
	}

	return nil
}

func createUpperWork(containerName string) error {
	upperUrl := fmt.Sprintf(common.UpperDirFormat, containerName)
	if err := os.MkdirAll(upperUrl, common.Perm0777); err != nil {
		return errors.Wrapf(err, "mkdir[%s] dir failed", upperUrl)
	}

	workUrl := fmt.Sprintf(common.WorkDirFormat, containerName)
	if err := os.MkdirAll(workUrl, common.Perm0777); err != nil {
		return errors.Wrapf(err, "mkdir[%s] dir failed", workUrl)
	}

	return nil
}

func mountOverlayFS(containerName string) error {
	mntUrl := fmt.Sprintf(common.MergedDirFormat, containerName)
	if err := os.MkdirAll(mntUrl, common.Perm0777); err != nil {
		return errors.Wrapf(err, "mkdir dir[%s] failed", mntUrl)
	}

	var (
		lowerUrl  = fmt.Sprintf(common.LowerDirFormat, containerName)
		upperUrl  = fmt.Sprintf(common.UpperDirFormat, containerName)
		workerUrl = fmt.Sprintf(common.WorkDirFormat, containerName)
		mergedUrl = fmt.Sprintf(common.MergedDirFormat, containerName)
		dirs      = fmt.Sprintf(common.OverlayFsFormat, lowerUrl, upperUrl, workerUrl)
	)

	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mergedUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	return errors.Wrapf(err, "mount dir[%s] failed", mntUrl)
}

func mountVolume(containerName string, hostUrl, containerUrl string) error {
	if err := os.Mkdir(hostUrl, common.Perm0777); err != nil && !os.IsExist(err) {
		return errors.Wrapf(err, "mkdir host dir[%s] failed", hostUrl)
	}

	containerActualUrl := fmt.Sprintf(common.MergedDirFormat, containerName) + "/" + containerUrl
	if err := os.Mkdir(containerActualUrl, common.Perm0777); err != nil && !os.IsExist(err) {
		return errors.Wrapf(err, "mkdir container dir[%s] failed", containerActualUrl)
	}

	cmd := exec.Command("mount", "-o", "bind", hostUrl, containerActualUrl)
	if _, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "bind mount %s to %s failed", hostUrl, containerUrl)
	}

	return nil
}

func umountVolume(containerName string, hostUrl, containerUrl string) error {
	containerActualUrl := fmt.Sprintf(common.MergedDirFormat, containerName) + "/" + containerUrl
	cmd := exec.Command("umount", containerActualUrl)
	if _, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "umount %s", containerActualUrl)
	}

	return nil
}

func umountOverlayFS(containerName string) error {
	mergedUrl := fmt.Sprintf(common.MergedDirFormat, containerName)
	cmd := exec.Command("umount", mergedUrl)
	if _, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "umount mountpoint %s", mergedUrl)
	}

	if err := os.RemoveAll(mergedUrl); err != nil {
		return errors.Wrapf(err, "remove mountpoint dir %s", mergedUrl)
	}
	return nil
}

func removeDirs(containerName string) error {
	lower := fmt.Sprintf(common.LowerDirFormat, containerName)
	upper := fmt.Sprintf(common.UpperDirFormat, containerName)
	work := fmt.Sprintf(common.WorkDirFormat, containerName)

	if err := os.RemoveAll(lower); err != nil {
		return errors.Wrapf(err, "remove dir %s", lower)
	}
	if err := os.RemoveAll(upper); err != nil {
		return errors.Wrapf(err, "remove dir %s", upper)
	}
	if err := os.RemoveAll(work); err != nil {
		return errors.Wrapf(err, "remove dir %s", work)
	}
	return nil
}
