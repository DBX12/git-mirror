package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

func mirror(cfg config, r repo) error {
	repoPath := path.Join(cfg.BasePath, r.Name)
	var cmdEnv []string
	if *flags.onDemand {
		cmdEnv = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
	}
	if _, err := os.Stat(repoPath); err == nil {
		// Directory exists, update.
		cmd := exec.Command("git", "remote", "update")
		cmd.Dir = repoPath
		cmd.Env = cmdEnv
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to update remote in %s, %s", repoPath, err)
		}
	} else if os.IsNotExist(err) {
		// Clone
		parent := path.Dir(repoPath)
		if err = os.MkdirAll(parent, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for cloning %s, %s", repoPath, err)
		}
		cmd := exec.Command("git", "clone", "--mirror", r.Origin, repoPath)
		cmd.Env = cmdEnv
		cmd.Dir = parent
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone %s, %s", r.Origin, err)
		}
	} else {
		return fmt.Errorf("failed to stat %s, %s", repoPath, err)
	}
	cmd := exec.Command("git", "update-server-info")
	cmd.Dir = repoPath
	cmd.Env = cmdEnv
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update-server-info for %s, %s", repoPath, err)
	}
	return nil
}
