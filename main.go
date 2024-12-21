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
		fatal("failed to create resource watcher", "error", err)
	}

	err = watcher.Watch(ctx)
	if err != nil {
		fatal("failed to watch resources", "error", err)
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

func fatal(msg string, args ...any) {
	logger.Global().Error(msg, args...)
	os.Exit(1)
}
