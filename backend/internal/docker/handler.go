package docker

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"
)

var dockerHost string

func SetDockerHost(host string) {
	dockerHost = host
}

type ContainerInfo struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Image  string   `json:"image"`
	Status string   `json:"status"`
	State  string   `json:"state"`
	Ports  []string `json:"ports"`
	Uptime string   `json:"uptime"`
}

func ListContainers(c *gin.Context) {
	cli, err := client.NewClientWithOpts(
		client.WithHost(dockerHost),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法连接 Docker"})
		return
	}
	defer cli.Close()

	result, err := cli.ContainerList(context.Background(), client.ContainerListOptions{All: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取容器列表失败"})
		return
	}

	list := make([]ContainerInfo, 0, len(result.Items))
	for _, ct := range result.Items {
		name := ""
		if len(ct.Names) > 0 {
			name = strings.TrimPrefix(ct.Names[0], "/")
		}

		ports := make([]string, 0)
		for _, p := range ct.Ports {
			if p.PublicPort > 0 {
				ports = append(ports, fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			} else {
				ports = append(ports, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}

		uptime := ""
		if strings.HasPrefix(ct.Status, "Up ") {
			uptime = strings.TrimPrefix(ct.Status, "Up ")
		}

		id := ct.ID
		if len(id) > 12 {
			id = id[:12]
		}

		list = append(list, ContainerInfo{
			ID:     id,
			Name:   name,
			Image:  ct.Image,
			Status: ct.Status,
			State:  string(ct.State),
			Ports:  ports,
			Uptime: uptime,
		})
	}

	c.JSON(http.StatusOK, list)
}
