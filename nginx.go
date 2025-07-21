package poc_testcontainers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type nginxContainer struct {
	testcontainers.Container
	URI string
}

func StartContainer(ctx context.Context) (*nginxContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "nginx:latest",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForListeningPort("80/tcp").WithStartupTimeout(90 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err == nil {
		fmt.Print(err)

	}
	logs, err := container.Logs(ctx)
	if err == nil {
		defer logs.Close()
	}
	var nginxC *nginxContainer
	if container != nil {
		nginxC = &nginxContainer{Container: container}
	}
	if err != nil {
		return nginxC, err
	}

	_, err = container.MappedPort(ctx, "80")
	if err != nil {
		return nil, err
	}

	endpoint, err := container.PortEndpoint(ctx, "80", "http")
	if err != nil {
		return nginxC, err
	}

	_, err = nginxC.Host(ctx)
	if err != nil {
		log.Printf("failed to create container: %s", err)
		return nginxC, err
	}

	nginxC.URI = endpoint
	return nginxC, nil
}
