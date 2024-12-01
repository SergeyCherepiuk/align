package resources

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/SergeyCherepiuk/align/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewFileUnit(t *testing.T) {
	t.Run("new file without options", func(t *testing.T) {
		file := NewFile("/tmp/testing")

		assert.Equal(t, "/tmp/testing", file.path)
		assert.Equal(t, "/tmp/testing", file.Id())
		assert.Equal(t, types.Optional[os.FileMode]{}, file.mode)
		assert.Equal(t, types.Optional[string]{}, file.owner)
		assert.Equal(t, types.Optional[string]{}, file.group)
	})

	t.Run("new file with mode option", func(t *testing.T) {
		file := NewFile("/tmp/testing", WithMode(0o777))

		assert.Equal(t, "/tmp/testing", file.path)
		assert.Equal(t, "/tmp/testing", file.Id())
		assert.Equal(t, types.NewOptional[os.FileMode](0o777), file.mode)
		assert.Equal(t, types.Optional[string]{}, file.owner)
		assert.Equal(t, types.Optional[string]{}, file.group)
	})

	t.Run("new file with owner option", func(t *testing.T) {
		file := NewFile("/tmp/testing", WithOwner("owner"))

		assert.Equal(t, "/tmp/testing", file.path)
		assert.Equal(t, "/tmp/testing", file.Id())
		assert.Equal(t, types.Optional[os.FileMode]{}, file.mode)
		assert.Equal(t, types.NewOptional("owner"), file.owner)
		assert.Equal(t, types.Optional[string]{}, file.group)
	})

	t.Run("new file with group option", func(t *testing.T) {
		file := NewFile("/tmp/testing", WithGroup("group"))

		assert.Equal(t, "/tmp/testing", file.path)
		assert.Equal(t, "/tmp/testing", file.Id())
		assert.Equal(t, types.Optional[os.FileMode]{}, file.mode)
		assert.Equal(t, types.Optional[string]{}, file.owner)
		assert.Equal(t, types.NewOptional("group"), file.group)
	})

	t.Run("new file with multiple options", func(t *testing.T) {
		file := NewFile(
			"/tmp/testing",
			WithOwner("owner"),
			WithGroup("group"),
		)

		assert.Equal(t, "/tmp/testing", file.path)
		assert.Equal(t, "/tmp/testing", file.Id())
		assert.Equal(t, types.Optional[os.FileMode]{}, file.mode)
		assert.Equal(t, types.NewOptional("owner"), file.owner)
		assert.Equal(t, types.NewOptional("group"), file.group)
	})

	t.Run("new file with multiple same options", func(t *testing.T) {
		file := NewFile(
			"/tmp/testing",
			WithOwner("owner1"),
			WithOwner("owner2"),
			WithOwner("owner3"),
		)

		assert.Equal(t, "/tmp/testing", file.path)
		assert.Equal(t, "/tmp/testing", file.Id())
		assert.Equal(t, types.Optional[os.FileMode]{}, file.mode)
		assert.Equal(t, types.NewOptional("owner3"), file.owner)
		assert.Equal(t, types.Optional[string]{}, file.group)
	})
}

// TODO: sc: Figure out how to avoid the use of reflection.

func TestFileCheckIntegration(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		path := fmt.Sprintf("/tmp/check-testing-file-%s", uuid.NewString())

		file := NewFile(path)

		expected := []Correction{
			file.create,
			file.changeMode,
			file.changeOwner,
			file.changeGroup,
		}

		corrections, err := file.Check()
		if assert.Equal(t, 4, len(corrections)) {
			for i := range 4 {
				assert.Equal(
					t,
					reflect.ValueOf(expected[i]).Pointer(),
					reflect.ValueOf(corrections[i]).Pointer(),
				)
			}
		}
		assert.ErrorIs(t, err, ErrUnalignedResource)
	})

	t.Run("file has wrong mode", func(t *testing.T) {
		path := fmt.Sprintf("/tmp/check-testing-file-%s", uuid.NewString())

		f, err := os.OpenFile(path, os.O_CREATE, 0o664)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close(); os.Remove(path) })

		file := NewFile(path, WithMode(0o777))

		corrections, err := file.Check()
		if assert.Equal(t, 1, len(corrections)) {
			assert.Equal(
				t,
				reflect.ValueOf(file.changeMode).Pointer(),
				reflect.ValueOf(corrections[0]).Pointer(),
			)
		}
		assert.ErrorIs(t, err, ErrUnalignedResource)
	})

	t.Run("file has wrong owner", func(t *testing.T) {
		path := fmt.Sprintf("/tmp/check-testing-file-%s", uuid.NewString())
		owner := fmt.Sprintf("check-testing-owner-%s", uuid.NewString())

		f, err := os.OpenFile(path, os.O_CREATE, 0o664)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close(); os.Remove(path) })

		file := NewFile(path, WithOwner(owner))

		corrections, err := file.Check()
		if assert.Equal(t, 1, len(corrections)) {
			assert.Equal(
				t,
				reflect.ValueOf(file.changeOwner).Pointer(),
				reflect.ValueOf(corrections[0]).Pointer(),
			)
		}
		assert.ErrorIs(t, err, ErrUnalignedResource)
	})

	t.Run("file has wrong group", func(t *testing.T) {
		path := fmt.Sprintf("/tmp/check-testing-file-%s", uuid.NewString())
		group := fmt.Sprintf("check-testing-group-%s", uuid.NewString())

		f, err := os.OpenFile(path, os.O_CREATE, 0o664)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close(); os.Remove(path) })

		file := NewFile(path, WithGroup(group))

		corrections, err := file.Check()
		if assert.Equal(t, 1, len(corrections)) {
			assert.Equal(
				t,
				reflect.ValueOf(file.changeGroup).Pointer(),
				reflect.ValueOf(corrections[0]).Pointer(),
			)
		}
		assert.ErrorIs(t, err, ErrUnalignedResource)
	})

	t.Run("file is aligned", func(t *testing.T) {
		path := fmt.Sprintf("/tmp/check-testing-file-%s", uuid.NewString())

		f, err := os.OpenFile(path, os.O_CREATE, 0o664)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close(); os.Remove(path) })

		file := NewFile(path)

		corrections, err := file.Check()
		assert.Empty(t, corrections)
		assert.NoError(t, err)
	})
}
