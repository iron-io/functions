package docker

import (
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func CheckImageLocally(image string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	// Have to use ctx.Background() for now because the docker official client
	// package uses the old ctx package (golang.org/x/net/context so it conflicts
	// with our current context package
	_, _, err = cli.ImageInspectWithRaw(context.Background(), image)

	return err
}
