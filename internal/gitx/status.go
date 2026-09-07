package gitx

import (
	"strconv"
	"strings"
)

// RepoStatus is the parsed result of `git status --porcelain=v1 --branch`.
type RepoStatus struct {
	Branch    string
	Upstream  string
	Ahead     int
	Behind    int
	Dirty     int
	Untracked int
	Detached  bool
	Unborn    bool
}

// State derives a single-word summary. Precedence:
// detached > unborn > dirty > ahead > behind > diverged > clean.
func (s *RepoStatus) State() string {
	switch {
	case s.Detached:
		return "detached"
	case s.Unborn:
		return "unborn"
	case s.Dirty > 0:
		return "dirty"
	case s.Ahead > 0 && s.Behind > 0:
		return "diverged"
	case s.Ahead > 0:
		return "ahead"
	case s.Behind > 0:
		return "behind"
	default:
		return "clean"
	}
}

// Status runs git status in dir and parses the result.
func Status(dir string) (*RepoStatus, error) {
	out, _, err := Run(dir, "status", "--porcelain=v1", "--branch")
	if err != nil {
		return nil, err
	}
	return ParseStatus(out), nil
}

// ParseStatus parses porcelain=v1 --branch output. Pure function, unit-tested.
func ParseStatus(out string) *RepoStatus {
	s := &RepoStatus{}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return s
	}
	head := lines[0]
	if !strings.HasPrefix(head, "## ") {
		return s
	}
	head = strings.TrimPrefix(head, "## ")

	switch {
	case head == "HEAD (no branch)":
		s.Detached = true
		return s
	case strings.HasPrefix(head, "No commits yet on "):
		s.Unborn = true
		s.Branch = strings.TrimPrefix(head, "No commits yet on ")
		return s
	}

	// <local>...<upstream> [ahead N, behind M]
	local := head
	rest := ""
	if i := strings.Index(head, "..."); i >= 0 {
		local = head[:i]
		rest = head[i+3:]
	}
	s.Branch = local
	if rest != "" {
		if i := strings.Index(rest, " ["); i >= 0 {
			s.Upstream = rest[:i]
			parseCounts(s, rest[i+2:])
		} else {
			s.Upstream = rest
		}
	}

	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			s.Untracked++
			continue
		}
		if len(line) >= 2 && (line[0] != ' ' || line[1] != ' ') {
			s.Dirty++
		}
	}
	return s
}

func parseCounts(s *RepoStatus, bracket string) {
	// bracket like "ahead 1, behind 2]"
	bracket = strings.TrimSuffix(bracket, "]")
	for _, part := range strings.Split(bracket, ",") {
		part = strings.TrimSpace(part)
		switch {
		case strings.HasPrefix(part, "ahead "):
			s.Ahead, _ = strconv.Atoi(strings.TrimPrefix(part, "ahead "))
		case strings.HasPrefix(part, "behind "):
			s.Behind, _ = strconv.Atoi(strings.TrimPrefix(part, "behind "))
		}
	}
}
