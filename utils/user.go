package utils

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// SwitchUserGroup switches the current process to the specified user and group
// This should be called after binding to privileged ports (if started as root)
func SwitchUserGroup(username, groupname string) error {
	var uid, gid uint32

	// Get current UID/GID
	currentUID := uint32(os.Getuid())
	currentGID := uint32(os.Getgid())

	// Look up user if specified
	if username != "" {
		u, err := user.Lookup(username)
		if err != nil {
			return fmt.Errorf("error looking up user %s: %w", username, err)
		}
		uidInt, err := strconv.Atoi(u.Uid)
		if err != nil {
			return fmt.Errorf("error converting UID: %w", err)
		}
		uid = uint32(uidInt)
	} else {
		uid = currentUID
	}

	// Look up group if specified
	if groupname != "" {
		g, err := user.LookupGroup(groupname)
		if err != nil {
			return fmt.Errorf("error looking up group %s: %w", groupname, err)
		}
		gidInt, err := strconv.Atoi(g.Gid)
		if err != nil {
			return fmt.Errorf("error converting GID: %w", err)
		}
		gid = uint32(gidInt)
	} else {
		gid = currentGID
	}

	// Only switch if we're not already running as that user/group
	if uid != currentUID || gid != currentGID {
		// Set group first (requires root if changing)
		if gid != currentGID {
			if err := syscall.Setgid(int(gid)); err != nil {
				return fmt.Errorf("error setting GID: %w", err)
			}
		}

		// Set user (requires root if changing)
		if uid != currentUID {
			if err := syscall.Setuid(int(uid)); err != nil {
				return fmt.Errorf("error setting UID: %w", err)
			}
		}

		// Set supplementary groups (optional, but good practice)
		// This ensures the process has access to the user's groups
		if username != "" {
			u, err := user.Lookup(username)
			if err == nil {
				groupIds, err := u.GroupIds()
				if err == nil {
					var gids []int
					for _, gidStr := range groupIds {
						if gidInt, err := strconv.Atoi(gidStr); err == nil {
							gids = append(gids, gidInt)
						}
					}
					if len(gids) > 0 {
						syscall.Setgroups(gids)
					}
				}
			}
		}
	}

	return nil
}

// GetCurrentUser returns the current effective user and group
func GetCurrentUser() (string, string, error) {
	uid := os.Getuid()
	gid := os.Getgid()

	u, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return "", "", err
	}

	g, err := user.LookupGroupId(strconv.Itoa(gid))
	if err != nil {
		return u.Username, "", err
	}

	return u.Username, g.Name, nil
}
