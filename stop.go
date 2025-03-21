package main

import (
	"encoding/json"
	"fmt"
	"mydocker/network"
	"os"
	"path"
	"strconv"
	"syscall"

	"mydocker/constant"
	"mydocker/container"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

func stopContainer(containerId string) {
    // 1. 根据容器Id查询容器信息
    containerInfo, err := getInfoByContainerId(containerId)
    if err != nil {
        log.Errorf("Get container %s info error %v", containerId, err)
        return
    }

    // 检查容器进程是否存在
    pidInt, err := strconv.Atoi(containerInfo.Pid)
    if err != nil {
        log.Errorf("Convert pid from string to int error %v", err)
        return
    }

    if !isProcessRunning(pidInt) {
        log.Infof("Container %s process is already stopped or doesn't exist", containerId)
        // 如果进程已经停止，直接更新容器状态为 STOP
        containerInfo.Status = container.STOP
        containerInfo.Pid = " "
        newContentBytes, err := json.Marshal(containerInfo)
        if err != nil {
            log.Errorf("Json marshal %s error %v", containerId, err)
            return
        }

        // 重新写回容器信息文件
        dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
        configFilePath := path.Join(dirPath, container.ConfigName)
        if err = os.WriteFile(configFilePath, newContentBytes, constant.Perm0622); err != nil {
            log.Errorf("Write file %s error:%v", configFilePath, err)
        }
        return
    }

    // 2. 进程存在，发送 SIGTERM 信号
    if err = syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
        log.Errorf("Stop container %s error %v", containerId, err)
        return
    }

    // 3. 修改容器信息，将容器置为 STOP 状态，并清空 PID
    containerInfo.Status = container.STOP
    containerInfo.Pid = " "
    newContentBytes, err := json.Marshal(containerInfo)
    if err != nil {
        log.Errorf("Json marshal %s error %v", containerId, err)
        return
    }

    // 4. 重新写回存储容器信息的文件
    dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
    configFilePath := path.Join(dirPath, container.ConfigName)
    if err = os.WriteFile(configFilePath, newContentBytes, constant.Perm0622); err != nil {
        log.Errorf("Write file %s error:%v", configFilePath, err)
    }
}

// 检查进程是否存在
func isProcessRunning(pid int) bool {
    // 通过发送 0 信号检查进程是否存在
    err := syscall.Kill(pid, 0)
    return err == nil
}

func getInfoByContainerId(containerId string) (*container.Info, error) {
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	contentBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "read file %s", configFilePath)
	}
	var containerInfo container.Info
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}

func removeContainer(containerId string, force bool) {
	containerInfo, err := getInfoByContainerId(containerId)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerId, err)
		return
	}

	switch containerInfo.Status {
	case container.STOP: // STOP 状态容器直接删除即可
		// 先删除配置目录，再删除rootfs 目录
		if err = container.DeleteContainerInfo(containerId); err != nil {
			log.Errorf("Remove container [%s]'s config failed, detail: %v", containerId, err)
			return
		}
		container.DeleteWorkSpace(containerId, containerInfo.Volume)
		if containerInfo.NetworkName != "" { // 清理网络资源
			if err = network.Disconnect(containerInfo.NetworkName, containerInfo); err != nil {
				log.Errorf("Remove container [%s]'s config failed, detail: %v", containerId, err)
				return
			}
		}
	case container.RUNNING: // RUNNING 状态容器如果指定了 force 则先 stop 然后再删除
		if !force {
			log.Errorf("Couldn't remove running container [%s], Stop the container before attempting removal or"+
				" force remove", containerId)
			return
		}
		log.Infof("force delete running container [%s]", containerId)
		stopContainer(containerId)
		removeContainer(containerId, force)
	default:
		log.Errorf("Couldn't remove container,invalid status %s", containerInfo.Status)
		return
	}
}
