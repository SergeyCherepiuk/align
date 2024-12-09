package resources

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"slices"
	"strings"
	"time"

	"github.com/scherepiuk/align/internal/logger"
	"github.com/scherepiuk/align/internal/types"
)

// TODO: sc: Add asynchronous structural logging.

type User struct {
	BaseDependant
	name   string
	uid    int
	gid    int
	groups types.Optional[[]string]
}

func NewUser(name string, uid, gid int, opts ...UserOption) *User {
	user := &User{name: name, uid: uid, gid: gid}

	for _, opt := range opts {
		opt(user)
	}

	return user
}

type UserOption func(user *User)

func WithGroups(groups ...string) UserOption {
	return func(user *User) {
		logger.Global().Info("specifying user groups", "name", user.name, "groups", groups)
		user.groups = types.NewOptional(groups)
	}
}

func (u *User) Id() string {
	return u.name
}

func (u *User) Check() ([]Correction, error) {
	uid, gid, groupIds, err := lookupUserDetails(u.name)

	if errors.Is(err, user.UnknownUserError(u.name)) {
		corrections := []Correction{
			u.create,
			u.changeUid,
			u.changeGid,
			u.setGroups,
		}

		return corrections, ErrUnalignedResource
	}

	if err != nil {
		return nil, fmt.Errorf("failed to lookup user details: %w", err)
	}

	corrections := make([]Correction, 0)

	if uid != u.uid {
		corrections = append(corrections, u.changeUid)
	}

	if gid != u.gid {
		corrections = append(corrections, u.changeGid)
	}

	if u.groups.Ok() {
		for _, group := range u.groups.Value() {
			gid, err := lookupGroup(group)
			if err != nil {
				return nil, fmt.Errorf("failed to lookup group: %w", err)
			}

			if !slices.Contains(groupIds, gid) {
				corrections = append(corrections, u.setGroups)
				break
			}
		}
	}

	if len(corrections) > 0 {
		return corrections, ErrUnalignedResource
	}

	return nil, nil
}

func (u *User) Watch(
	ctx context.Context,
	correctionsCh chan<- []Correction,
	errCh chan<- error,
) {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return

		case <-ticker.C:
			corrections, err := u.Check()

			if errors.Is(err, ErrUnalignedResource) {
				correctionsCh <- corrections
				continue
			}

			if err != nil {
				errCh <- err
				return
			}
		}
	}
}

func (u *User) create() error {
	cmd := exec.Command(
		"useradd",
		"-u", fmt.Sprint(u.uid),
		"-g", fmt.Sprint(u.gid),
		"-G", strings.Join(u.groups.Value(), ","),
		u.name,
	)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (u *User) changeUid() error {
	cmd := exec.Command("usermod", "-u", fmt.Sprint(u.uid), u.name)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to change user's uid: %w", err)
	}

	return nil
}

func (u *User) changeGid() error {
	cmd := exec.Command("usermod", "-g", fmt.Sprint(u.gid), u.name)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to change user's gid: %w", err)
	}

	return nil
}

func (u *User) setGroups() error {
	if !u.groups.Ok() {
		return nil
	}

	cmd := exec.Command("usermod", "-G", strings.Join(u.groups.Value(), ","), u.name)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to set user's groups: %w", err)
	}

	return nil
}
