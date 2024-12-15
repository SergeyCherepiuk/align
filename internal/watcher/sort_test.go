package watcher

import (
	"errors"
	"slices"
	"testing"

	"github.com/scherepiuk/align/internal/resources"
	"github.com/stretchr/testify/assert"
)

func TestSortTopologicallyUnit(t *testing.T) {
	type testCase struct {
		name             string
		resources        func() []resources.Resource
		expectedTopology [][]resources.Resource
		expectedErr      error
	}

	testCases := []testCase{
		{
			name:             "empty topology",
			resources:        emptyTopology,
			expectedTopology: [][]resources.Resource{},
			expectedErr:      nil,
		},
		{
			name:      "flat topology",
			resources: flatTopology,
			expectedTopology: [][]resources.Resource{
				{
					resources.NewFile("/tmp/first"),
					resources.NewFile("/tmp/second"),
					resources.NewFile("/tmp/third"),
				},
			},
			expectedErr: nil,
		},
		{
			name:      "sequential topology",
			resources: sequentialTopology,
			expectedTopology: [][]resources.Resource{
				{
					resources.NewFile("/tmp/first"),
				},
				{
					resources.NewFile("/tmp/second"),
				},
				{
					resources.NewFile("/tmp/third"),
				},
			},
			expectedErr: nil,
		},
		{
			name:      "graph topology",
			resources: graphTopology,
			expectedTopology: [][]resources.Resource{
				{
					resources.NewFile("/tmp/first"),
					resources.NewFile("/tmp/second"),
					resources.NewFile("/tmp/third"),
				},
				{
					resources.NewFile("/tmp/fourth"),
				},
				{
					resources.NewFile("/tmp/fifth"),
					resources.NewFile("/tmp/sixth"),
				},
			},
			expectedErr: nil,
		},
		{
			name:             "cyclic topology",
			resources:        cyclicTopology,
			expectedTopology: nil,
			expectedErr:      errors.New("cyclic dependency"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resources := tc.resources()
			actualTopology, actualErr := sortTopologically(resources)

			assertTopology(t, tc.expectedTopology, actualTopology)
			assert.Equal(t, tc.expectedErr, actualErr)
		})
	}
}

func assertTopology(t *testing.T, expected, actual [][]resources.Resource) {
	if assert.Len(t, actual, len(expected)) {
		for i, layer := range actual {
			var (
				actualLayer   = copyAndSortResources(layer)
				expectedLayer = copyAndSortResources(expected[i])
			)

			for j, actualResource := range actualLayer {
				expectedResource := expectedLayer[j]
				assert.True(t, expectedResource.Id() == actualResource.Id())
			}
		}
	}
}

func copyAndSortResources(rs []resources.Resource) []resources.Resource {
	rsCopy := make([]resources.Resource, len(rs))
	copy(rsCopy, rs)
	slices.SortFunc(rsCopy, cmpResources)
	return rsCopy
}

func cmpResources(r1, r2 resources.Resource) int {
	if r1.Id() < r2.Id() {
		return -1
	} else if r1.Id() > r2.Id() {
		return 1
	}
	return 0
}

func emptyTopology() []resources.Resource {
	return []resources.Resource{}
}

func flatTopology() []resources.Resource {
	return []resources.Resource{
		resources.NewFile("/tmp/first"),
		resources.NewFile("/tmp/second"),
		resources.NewFile("/tmp/third"),
	}
}

func sequentialTopology() []resources.Resource {
	f1 := resources.NewFile("/tmp/first")
	f2 := resources.NewFile("/tmp/second")
	f3 := resources.NewFile("/tmp/third")

	f2.SetDependencies(f1)
	f3.SetDependencies(f2)

	return []resources.Resource{f1, f2, f3}
}

func graphTopology() []resources.Resource {
	f1 := resources.NewFile("/tmp/first")
	f2 := resources.NewFile("/tmp/second")
	f3 := resources.NewFile("/tmp/third")
	f4 := resources.NewFile("/tmp/fourth")
	f5 := resources.NewFile("/tmp/fifth")
	f6 := resources.NewFile("/tmp/sixth")

	f4.SetDependencies(f1, f2, f3)
	f5.SetDependencies(f4)
	f6.SetDependencies(f4)

	return []resources.Resource{f1, f2, f3, f4, f5, f6}
}

func cyclicTopology() []resources.Resource {
	f1 := resources.NewFile("/tmp/first")
	f2 := resources.NewFile("/tmp/second")
	f3 := resources.NewFile("/tmp/third")

	f1.SetDependencies(f2)
	f2.SetDependencies(f3)
	f3.SetDependencies(f1)

	return []resources.Resource{f1, f2, f3}
}
