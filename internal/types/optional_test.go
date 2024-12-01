package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionalUnit(t *testing.T) {
	t.Run("optional string is present", func(t *testing.T) {
		optional := NewOptional("test")

		assert.Equal(t, true, optional.Ok())
		assert.Equal(t, "test", optional.Value())
	})

	t.Run("optional string is not present", func(t *testing.T) {
		var optional Optional[string]

		assert.Equal(t, false, optional.Ok())
		assert.Equal(t, "", optional.Value())
	})

	t.Run("optional integer pointer is present", func(t *testing.T) {
		value := 69420

		optional := NewOptional(&value)

		assert.Equal(t, true, optional.Ok())
		assert.Equal(t, &value, optional.Value())
		assert.Equal(t, 69420, *optional.Value())
	})

	t.Run("optional integer pointer is not present", func(t *testing.T) {
		var optional Optional[*int]

		assert.Equal(t, false, optional.Ok())
		assert.Nil(t, optional.Value())
	})

	t.Run("optional struct is present", func(t *testing.T) {
		type point struct{ x, y int }

		optional := NewOptional(point{69, 420})

		assert.Equal(t, true, optional.Ok())
		assert.Equal(t, point{69, 420}, optional.Value())
	})

	t.Run("optional structure is not present", func(t *testing.T) {
		type point struct{ x, y int }

		var optional Optional[point]

		assert.Equal(t, false, optional.Ok())
		assert.Equal(t, point{0, 0}, optional.Value())
	})
}
