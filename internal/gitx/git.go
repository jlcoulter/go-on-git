package gitx

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OpError describes a failed git invocation.
type OpError struct {
	Dir    string
	Args   []string
	Stderr string
}

func (e *OpError) Error() string {
	return fmt.Sprintf("git %s failed in %s: %s", strings.Join(e.Args, " "), e.Dir, strings.TrimSpace(e.Stderr))
}

// Run executes git in dir and returns stdout and stderr separately.
func Run(dir string, args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return out.String(), errb.String(), &OpError{Dir: dir, Args: args, Stderr: errb.String()}
	}
	return out.String(), errb.String(), nil
}

// EnsureGit verifies the git binary is available.
func EnsureGit() error {
	if _, err := exec.LookPath("git"); err != nil {
		return errors.New("git not found in PATH")
	}
	return nil
}

// IsRepo reports whether dir contains a .git entry (file or dir).
func IsRepo(dir string) bool {
	_, err := os.Stat(dir + "/.git")
	return err == nil
}

// OriginURL returns the origin remote URL, or "" if there is none.
func OriginURL(dir string) (string, error) {
	out, _, err := Run(dir, "remote", "get-url", "origin")
	if err != nil {
		var op *OpError
		if errors.As(err, &op) {
			return "", nil // no origin remote
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Clone clones url into dest, creating parent directories as needed.
// It clones into a temp sibling then renames, so an existing empty parent
// directory is fine.
func Clone(url, dest string) (string, error) {
	parent := filepath.Dir(dest)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	tmp := dest + ".gog-clone-tmp"
	_ = os.RemoveAll(tmp)
	_, stderr, err := Run(parent, "clone", url, filepath.Base(tmp))
	if err != nil {
		return stderr, err
	}
	if err := os.Rename(tmp, dest); err != nil {
		return "", err
	}
	return "", nil
}

// Pull runs git pull in dir. By default it fast-forwards only; with rebase it
// rebases local commits onto the upstream instead.
func Pull(dir string, rebase bool) (string, error) {
	args := []string{"pull"}
	if rebase {
		args = append(args, "--rebase")
	} else {
		args = append(args, "--ff-only")
	}
	_, stderr, err := Run(dir, args...)
	return stderr, err
}

// Fetch runs git fetch --all --prune in dir.
func Fetch(dir string) (string, error) {
	_, stderr, err := Run(dir, "fetch", "--all", "--prune")
	return stderr, err
}
