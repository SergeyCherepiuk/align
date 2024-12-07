package resources

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

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
		path := testFilePath()

		file := NewFile(path)
		expected := []Correction{
			file.create,
			file.changeMode,
			file.changeOwner,
			file.changeGroup,
		}

		actual, err := file.Check()
		assertCorrections(t, expected, actual)
		assert.ErrorIs(t, err, ErrUnalignedResource)
	})

	t.Run("file has wrong mode", func(t *testing.T) {
		path := testFilePath()

		f, err := os.OpenFile(path, os.O_CREATE, 0o664)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close(); os.Remove(path) })

		file := NewFile(path, WithMode(0o777))
		expected := []Correction{file.changeMode}

		actual, err := file.Check()
		assertCorrections(t, expected, actual)
		assert.ErrorIs(t, err, ErrUnalignedResource)
	})

	t.Run("file has wrong owner", func(t *testing.T) {
		path, owner := testFilePath(), testOwnerName()

		f, err := os.OpenFile(path, os.O_CREATE, 0o664)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close(); os.Remove(path) })

		file := NewFile(path, WithOwner(owner))
		expected := []Correction{file.changeOwner}

		actual, err := file.Check()
		assertCorrections(t, expected, actual)
		assert.ErrorIs(t, err, ErrUnalignedResource)
	})

	t.Run("file has wrong group", func(t *testing.T) {
		path, group := testFilePath(), testGroupName()

		f, err := os.OpenFile(path, os.O_CREATE, 0o664)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close(); os.Remove(path) })

		file := NewFile(path, WithGroup(group))
		expected := []Correction{file.changeGroup}

		actual, err := file.Check()
		assertCorrections(t, expected, actual)
		assert.ErrorIs(t, err, ErrUnalignedResource)
	})

	t.Run("file is aligned", func(t *testing.T) {
		path := testFilePath()

		f, err := os.OpenFile(path, os.O_CREATE, 0o664)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close(); os.Remove(path) })

		file := NewFile(path)

		corrections, err := file.Check()
		assert.Nil(t, corrections)
		assert.NoError(t, err)
	})
}

func TestFileWatchIntegration(t *testing.T) {
	t.Run("file is aligned", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(cancel)

		path := testFilePath()
		correctionsCh, errCh := make(chan []Correction), make(chan error)

		f, err := os.OpenFile(path, os.O_CREATE, 0o664)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close(); os.Remove(path) })

		file := NewFile(path)

		go file.Watch(ctx, correctionsCh, errCh)
		time.Sleep(500 * time.Millisecond)

		err = <-errCh
		assert.ErrorContains(t, err, "context deadline exceeded")
	})

	t.Run("file has been removed", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		t.Cleanup(cancel)

		path := testFilePath()
		correctionsCh, errCh := make(chan []Correction), make(chan error)

		f, err := os.OpenFile(path, os.O_CREATE, 0o664)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close(); os.Remove(path) })

		file := NewFile(path)
		expected := []Correction{
			file.create,
			file.changeMode,
			file.changeOwner,
			file.changeGroup,
		}

		go file.Watch(ctx, correctionsCh, errCh)
		time.Sleep(time.Second)

		err = os.Remove(path)
		if err != nil {
			t.Fatal(err)
		}

		actual := <-correctionsCh
		assertCorrections(t, expected, actual)
	})

	t.Run("file's mode has been changed", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		t.Cleanup(cancel)

		path := testFilePath()
		correctionsCh, errCh := make(chan []Correction), make(chan error)

		f, err := os.OpenFile(path, os.O_CREATE, 0o664)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close(); os.Remove(path) })

		file := NewFile(path, WithMode(0o664))
		expected := []Correction{file.changeMode}

		go file.Watch(ctx, correctionsCh, errCh)
		time.Sleep(time.Second)

		err = f.Chmod(0o777)
		if err != nil {
			t.Fatal(err)
		}

		actual := <-correctionsCh
		assertCorrections(t, expected, actual)
	})
}

func assertCorrections(t *testing.T, expected []Correction, actual []Correction) {
	if assert.Equal(t, len(expected), len(actual)) {
		for i := range len(expected) {
			assert.Equal(
				t,
				reflect.ValueOf(expected[i]).Pointer(),
				reflect.ValueOf(actual[i]).Pointer(),
			)
		}
	}
}

func testFilePath() string {
	return fmt.Sprintf("/tmp/check-testing-file-%s", uuid.NewString())
}

func testOwnerName() string {
	return fmt.Sprintf("check-testing-owner-%s", uuid.NewString())
}

func testGroupName() string {
	return fmt.Sprintf("check-testing-group-%s", uuid.NewString())
}
