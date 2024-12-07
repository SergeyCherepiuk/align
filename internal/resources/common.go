package resources

import (
	"fmt"
	"os/user"
	"strconv"
)

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
