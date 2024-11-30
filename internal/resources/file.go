package resources

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"slices"
	"strconv"
	"syscall"

	"github.com/SergeyCherepiuk/align/internal/logger"
	"github.com/fsnotify/fsnotify"
)

// TODO: sc: Add asynchronous structural logging.
// TODO: sc: Allow specifying the content. Figure out a way to restore it
// without incurring a huge performance penalty (e.g., storing it in RAM.)
// TODO: sc: Watch function fires check twice: once when fsnotify emits an
// event due to change, and for the second time when correction is applied.

type File struct {
	path  string
	mode  *os.FileMode
	owner *string
	group *string
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
		file.mode = &mode
	}
}

func WithOwner(owner string) FileOption {
	return func(file *File) {
		logger.Global().Info("specifying file owner", "path", file.path, "owner", owner)
		file.owner = &owner
	}
}

func WithGroup(group string) FileOption {
	return func(file *File) {
		logger.Global().Info("specifying file group", "path", file.path, "group", group)
		file.group = &group
	}
}

func (f *File) Id() string {
	return f.path
}

func (f *File) Check() ([]Correction, error) {
	stat, err := os.Stat(f.path)

	if errors.Is(err, os.ErrNotExist) {
		corrections := []Correction{
			func() error { return f.create() },
			func() error { return f.changeMode() },
			func() error { return f.changeOwner() },
			func() error { return f.changeGroup() },
		}
		return corrections, ErrUnalignedResource
	}

	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	corrections := make([]Correction, 0)

	if f.mode != nil && stat.Mode() != *f.mode {
		correction := func() error { return f.changeMode() }
		corrections = append(corrections, correction)
	}

	// TODO: Getting linux-specific file info. All resources should be cross-platform.
	linuxFileInfo, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		panic("failed to get system-specific file info: not running on linux")
	}

	owner, _ := lookupUid(int(linuxFileInfo.Uid))
	if f.owner != nil && owner != *f.owner {
		correction := func() error { return f.changeOwner() }
		corrections = append(corrections, correction)
	}

	group, _ := lookupGid(int(linuxFileInfo.Gid))
	if f.group != nil && group != *f.group {
		correction := func() error { return f.changeGroup() }
		corrections = append(corrections, correction)
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
	if f.mode == nil {
		return nil
	}

	err := os.Chmod(f.path, *f.mode)
	if err != nil {
		return fmt.Errorf("failed to change file's mode: %w", err)
	}

	return nil
}

func (f *File) changeOwner() error {
	if f.owner == nil {
		return nil
	}

	uid, err := lookupUser(*f.owner)
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
	if f.group == nil {
		return nil
	}

	gid, err := lookupGroup(*f.group)
	if err != nil {
		return fmt.Errorf("failed to lookup group: %w", err)
	}

	err = os.Chown(f.path, -1, gid)
	if err != nil {
		return fmt.Errorf("failed to change file's group: %w", err)
	}

	return nil
}

func lookupUser(name string) (int, error) {
	user, err := user.Lookup(name)
	if err != nil {
		return 0, fmt.Errorf("failed to lookup user by name: %w", err)
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uid: %w", err)
	}

	return uid, nil
}

func lookupUid(uid int) (string, error) {
	user, err := user.LookupId(fmt.Sprint(uid))
	if err != nil {
		return "", fmt.Errorf("failed to lookup user by id: %w", err)
	}

	return user.Username, nil
}

func lookupGroup(name string) (int, error) {
	group, err := user.LookupGroup(name)
	if err != nil {
		return 0, fmt.Errorf("failed to lookup group by name: %w", err)
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return 0, fmt.Errorf("failed to parse gid: %w", err)
	}

	return gid, nil
}

func lookupGid(gid int) (string, error) {
	group, err := user.LookupGroupId(fmt.Sprint(gid))
	if err != nil {
		return "", fmt.Errorf("failed to lookup group by id: %w", err)
	}

	return group.Name, nil
}
