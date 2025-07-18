package nginx

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

func startContainer(ctx context.Context) (*nginxContainer, error) {
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
		fmt.Println("Container logs:")
		fmt.Print(err)

	}
	logs, err := container.Logs(ctx)
	if err == nil {
		defer logs.Close()
		fmt.Println("Container logs:")
	}
	var nginxC *nginxContainer
	if container != nil {
		nginxC = &nginxContainer{Container: container}
	}
	if err != nil {
		return nginxC, err
	}

	port, err := container.MappedPort(ctx, "80")
	if err != nil {
		return nil, err
	}

	fmt.Println("Mapped port:", port.Port())

	endpoint, err := container.PortEndpoint(ctx, "80", "http")
	if err != nil {
		return nginxC, err
	}
	fmt.Println("Endpoint:", endpoint)

	ip, err := nginxC.Host(ctx)
	if err != nil {
		log.Printf("failed to create container: %s", err)
		return nginxC, err
	}
	fmt.Println("IP port:", ip)

	nginxC.URI = endpoint
	return nginxC, nil
}
