package resources

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"syscall"

	"github.com/SergeyCherepiuk/align/internal/logger"
	"github.com/SergeyCherepiuk/align/internal/types"
	"github.com/fsnotify/fsnotify"
)

// TODO: sc: Add asynchronous structural logging.
// TODO: sc: Allow specifying the content. Figure out a way to restore it
// without incurring a huge performance penalty (e.g., storing it in RAM.)
// TODO: sc: Watch function fires check twice: once when fsnotify emits an
// event due to change, and for the second time when correction is applied.

type File struct {
	path  string
	mode  types.Optional[os.FileMode]
	owner types.Optional[string]
	group types.Optional[string]
}

func NewFile(path string, opts ...FileOption) *File {
	file := &File{path: path}

	for _, opt := range opts {
		opt(file)
	}

	return file
}

type FileOption func(file *File)

func WithMode(mode os.FileMode) FileOption {
	return func(file *File) {
		logger.Global().Info("specifying file mode", "path", file.path, "mode", mode)
		file.mode = types.NewOptional(mode)
	}
}

func WithOwner(owner string) FileOption {
	return func(file *File) {
		logger.Global().Info("specifying file owner", "path", file.path, "owner", owner)
		file.owner = types.NewOptional(owner)
	}
}

func WithGroup(group string) FileOption {
	return func(file *File) {
		logger.Global().Info("specifying file group", "path", file.path, "group", group)
		file.group = types.NewOptional(group)
	}
}

func (f *File) Id() string {
	return f.path
}

func (f *File) Check() ([]Correction, error) {
	stat, err := os.Stat(f.path)

	if errors.Is(err, os.ErrNotExist) {
		corrections := []Correction{
			f.create,
			f.changeMode,
			f.changeOwner,
			f.changeGroup,
		}
		return corrections, ErrUnalignedResource
	}

	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	corrections := make([]Correction, 0)

	if f.mode.Ok() && stat.Mode() != f.mode.Value() {
		corrections = append(corrections, f.changeMode)
	}

	// TODO: sc: Getting linux-specific file info. All resources should be cross-platform.
	linuxFileInfo, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		panic("failed to get system-specific file info: not running on linux")
	}

	owner, _ := lookupUid(int(linuxFileInfo.Uid))
	if f.owner.Ok() && owner != f.owner.Value() {
		corrections = append(corrections, f.changeOwner)
	}

	group, _ := lookupGid(int(linuxFileInfo.Gid))
	if f.group.Ok() && group != f.group.Value() {
		corrections = append(corrections, f.changeGroup)
	}

	if len(corrections) > 0 {
		return corrections, ErrUnalignedResource
	}

	return nil, nil
}

func (f *File) Watch(ctx context.Context, errCh chan<- error) {
	err := checkAndCorrect(f)
	if err != nil {
		errCh <- err
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		errCh <- err
		return
	}
	defer watcher.Close()

	err = watcher.Add(f.path)
	if err != nil {
		errCh <- err
		return
	}

	targetOps := []fsnotify.Op{
		fsnotify.Create,
		fsnotify.Write,
		fsnotify.Remove,
		fsnotify.Rename,
		fsnotify.Chmod,
	}

	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return

		case event := <-watcher.Events:
			logger.Global().Debug("got fsnotify event", "event", event.String())

			if slices.Contains(targetOps, event.Op) {
				err := checkAndCorrect(f)
				if err != nil {
					errCh <- err
					return
				}
			}
		}
	}
}

func (f *File) create() error {
	file, err := os.Create(f.path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

func (f *File) changeMode() error {
	if !f.mode.Ok() {
		return nil
	}

	err := os.Chmod(f.path, f.mode.Value())
	if err != nil {
		return fmt.Errorf("failed to change file's mode: %w", err)
	}

	return nil
}

func (f *File) changeOwner() error {
	if !f.owner.Ok() {
		return nil
	}

	uid, err := lookupUser(f.owner.Value())
	if err != nil {
		return fmt.Errorf("failed to lookup user: %w", err)
	}

	err = os.Chown(f.path, uid, -1)
	if err != nil {
		return fmt.Errorf("failed to change file's owner: %w", err)
	}

	return nil
}

func (f *File) changeGroup() error {
	if !f.group.Ok() {
		return nil
	}

	gid, err := lookupGroup(f.group.Value())
	if err != nil {
		return fmt.Errorf("failed to lookup group: %w", err)
	}

	err = os.Chown(f.path, -1, gid)
	if err != nil {
		return fmt.Errorf("failed to change file's group: %w", err)
	}

	return nil
}
