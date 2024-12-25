package main

import (
	"context"
	"os"

	"github.com/scherepiuk/align/internal/logger"
	"github.com/scherepiuk/align/internal/resources"
	"github.com/scherepiuk/align/internal/watcher"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Setup(ctx, logger.LevelDebug)
	defer logger.Global().Close()

	resources := expectedResources()
	watcher, err := watcher.NewResourceWatcher(resources...)
	if err != nil {
		logger.Global().Error("failed to create resource watcher", "error", err)
		return
	}

	err = watcher.Watch(ctx)
	if err != nil {
		logger.Global().Error("failed to watch resources", "error", err)
		return
	}
}

func expectedResources() []resources.Resource {
	alignUser := resources.NewUser(
		"align-testing-user", 42069, 1000,
		resources.WithGroups("root", "wheel"),
	)

	alignFile := resources.NewFile(
		"/tmp/align-testing-file",
		resources.WithMode(os.FileMode(0o664)),
		resources.WithOwner("align-testing-user"),
		resources.WithGroup("scherepiuk"),
	)

	alignFile.SetDependencies(alignUser)

	return []resources.Resource{alignFile, alignUser}
}
