package watcher

import (
	"errors"
	"slices"

	"github.com/SergeyCherepiuk/align/internal/resources"
)

type resourceMap map[resources.Resource][]resources.Resource

// TODO: sc: Unit test!
func sortTopologically(rs []resources.Resource) ([][]resources.Resource, error) {
	dependencies := make(resourceMap) // child -> [parents]

	for _, resource := range rs {
		dependencies[resource] = resource.Dependencies()
	}

	layers := make([][]resources.Resource, 0)

	for {
		leaves := removeLeaves(dependencies)
		if len(leaves) == 0 {
			break
		}

		layers = append(layers, leaves)
	}

	if len(dependencies) != 0 {
		return nil, errors.New("cyclic dependency")
	}

	return layers, nil
}

func removeLeaves(dependencies resourceMap) []resources.Resource {
	leaves := make([]resources.Resource, 0)

	for child, parents := range dependencies {
		if len(parents) == 0 {
			leaves = append(leaves, child)
			delete(dependencies, child)
		}
	}

	for child, parents := range dependencies {
		dependencies[child] = slices.DeleteFunc(
			parents,
			func(r resources.Resource) bool {
				return slices.Contains(leaves, r)
			},
		)
	}

	return leaves
}
