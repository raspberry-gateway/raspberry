package gateway

import (
	"context"

	cli "github.com/IBlogLab/raspberry/cli"
)

func Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli.Init()
}
