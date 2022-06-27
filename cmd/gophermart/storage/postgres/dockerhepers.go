package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func stopDB(id string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	err = cli.ContainerKill(context.Background(), id, "")
	if err != nil {
		return err
	}
	return nil
}

func startDB(name, image string) (dburl string, id string, err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", "", err
	}

	imgFilter := filters.NewArgs()
	imgFilter.Add("reference", image)
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{Filters: imgFilter})
	if err != nil {
		return "", "", err
	}
	if len(images) == 0 {
		_, err = cli.ImagePull(context.Background(), "docker.io/library/"+image, types.ImagePullOptions{})
		if err != nil {
			return "", "", err
		}
		var count = 0
		fmt.Print("Pulling")
		for count == 0 {
			images, err = cli.ImageList(context.Background(), types.ImageListOptions{Filters: imgFilter})
			if err != nil {
				return "", "", err
			}
			count = len(images)
			time.Sleep(time.Second)
			fmt.Print("*")
		}
		fmt.Println("Done")

	}
	health := &container.HealthConfig{
		Interval: time.Second,
		Test:     []string{"CMD-SHELL", "pg_isready"},
	}
	containerConfig := &container.Config{
		Image: image,
		Env: []string{
			"POSTGRES_DB=market",
			"POSTGRES_USER=hihi",
			"POSTGRES_PASSWORD=werySTRONGmethods36",
		},
		Healthcheck: health,
	}
	hostConfig := &container.HostConfig{
		AutoRemove:      true,
		PublishAllPorts: true,
	}
	resp, err := cli.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, name)
	if err != nil {
		return "", "", err
	}
	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", "", err
	}
	filter := filters.NewArgs()
	filter.Add("id", resp.ID)
	var C types.Container
	for {
		containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{Filters: filter})
		if err != nil {
			return "", "", err
		}
		if strings.Contains(containers[0].Status, "healthy") {
			C = containers[0]
			time.Sleep(time.Second)
			break
		}
	}
	return fmt.Sprintf("postgres://hihi:werySTRONGmethods36@localhost:%v/market?sslmode=disable", C.Ports[0].PublicPort), resp.ID, nil
}

// DBConnect .
// func DBConnect() (dockerId string, client *mongo.Client, cancel context.CancelFunc, err error) {
// 	var dburl string
// 	dburl, dockerId, err = startDB("postgres:14")
// 	if err != nil {
// 		return "", nil, nil, err
// 	}
// 	client, cancel, err = Connect(dburl)
// 	if err != nil {
// 		return "", nil, nil, err
// 	}
// 	return dockerId, client, cancel, nil
// }

// DBClose .
// func DBClose(id string, cancel context.CancelFunc) error {
// 	cancel()
// 	return stopDB(id)
// }

// Connect .
// func Connect(dburl string) (*mongo.Client, context.CancelFunc, error) {
// 	client, err := mongo.NewClient(options.Client().ApplyURI(dburl))
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	ctx, cancel := context.WithCancel(context.Background())
// 	err = client.Connect(ctx)
// 	if err != nil {
// 		cancel()
// 		return nil, nil, err
// 	}
// 	log.Info("Database is connected")
// 	return client, cancel, nil
// }
