package main

import (
	"context"
	"fmt"
	"os"

	"github.com/SergeyCherepiuk/align/internal/logger"
	"github.com/SergeyCherepiuk/align/internal/resources"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Setup(ctx)
	defer logger.Global().Close()

	errCh := make(chan error)
	for _, resource := range expectedResources() {
		go resource.Watch(ctx, errCh)
	}

	fmt.Println(<-errCh) // ?
}

func expectedResources() []resources.ResourceWatcher {
	return []resources.ResourceWatcher{
		resources.NewUser(
			"align-testing-user", 42069, 1000,
			resources.WithGroups("root", "wheel"),
		),
		resources.NewFile(
			"/tmp/align-testing-file",
			resources.WithMode(os.FileMode(0o664)),
			resources.WithOwner("scherepiuk"),
			resources.WithGroup("scherepiuk"),
		),
	}
}
