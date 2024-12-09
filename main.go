package main

import (
	"context"
	"log"
	"os"

	"github.com/scherepiuk/align/internal/logger"
	"github.com/scherepiuk/align/internal/resources"
	"github.com/scherepiuk/align/internal/watcher"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Setup(ctx)
	defer logger.Global().Close()

	resources := expectedResources()
	watcher, err := watcher.NewResourceWatcher(resources...)
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Watch(ctx)
	if err != nil {
		log.Fatal(err)
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
