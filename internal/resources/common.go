package resources

import (
	"errors"
	"fmt"
	"os/user"
	"strconv"

	"github.com/SergeyCherepiuk/align/internal/logger"
)

func checkAndCorrect(resource Resource) error {
	logger.Global().Info("checking resource", "resource", resource.Id())

	corrections, err := resource.Check()

	if errors.Is(err, ErrUnalignedResource) {
		logger.Global().Warn("resource is not aligned, executing corrections", "resource", resource.Id(), "count", len(corrections))

		err := executeCorrections(corrections)
		if err != nil {
			logger.Global().Error("correction failed", "resource", resource.Id(), "error", err.Error())
			return err
		}

		logger.Global().Info("corrections executed succesfully", "resource", resource.Id())

		return nil
	}

	if err != nil {
		return err
	}

	logger.Global().Info("resource is aligned", "resource", resource.Id())

	return nil
}

func executeCorrections(corrections []Correction) error {
	for _, correction := range corrections {
		err := correction()
		if err != nil {
			return err
		}
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

func lookupUserDetails(name string) (int, int, []int, error) {
	user, err := user.Lookup(name)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to lookup user by name: %w", err)
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to parse uid: %w", err)
	}

	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to parse gid: %w", err)
	}

	groupIdsStr, err := user.GroupIds()
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to get group ids: %w", err)
	}

	groupIds := make([]int, 0, len(groupIdsStr))
	for _, groupIdStr := range groupIdsStr {
		groupId, err := strconv.Atoi(groupIdStr)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("failed to parse group's id: %w", err)
		}

		groupIds = append(groupIds, groupId)
	}

	return uid, gid, groupIds, nil
}
