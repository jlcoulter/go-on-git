package gitx

import "testing"

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		want  RepoStatus
		state string
	}{
		{
			name:  "clean with upstream",
			in:    "## main...origin/main\n",
			want:  RepoStatus{Branch: "main", Upstream: "origin/main"},
			state: "clean",
		},
		{
			name:  "ahead and behind",
			in:    "## main...origin/main [ahead 1, behind 2]\n",
			want:  RepoStatus{Branch: "main", Upstream: "origin/main", Ahead: 1, Behind: 2},
			state: "diverged",
		},
		{
			name:  "ahead only",
			in:    "## main...origin/main [ahead 3]\n",
			want:  RepoStatus{Branch: "main", Upstream: "origin/main", Ahead: 3},
			state: "ahead",
		},
		{
			name:  "behind only",
			in:    "## main...origin/main [behind 4]\n",
			want:  RepoStatus{Branch: "main", Upstream: "origin/main", Behind: 4},
			state: "behind",
		},
		{
			name:  "no upstream",
			in:    "## main\n",
			want:  RepoStatus{Branch: "main"},
			state: "clean",
		},
		{
			name:  "detached",
			in:    "## HEAD (no branch)\n",
			want:  RepoStatus{Detached: true},
			state: "detached",
		},
		{
			name:  "unborn",
			in:    "## No commits yet on main\n",
			want:  RepoStatus{Unborn: true, Branch: "main"},
			state: "unborn",
		},
		{
			name:  "dirty and untracked",
			in:    "## main...origin/main\n M file.go\nA  staged.go\n?? new.go\nR  old.go -> new.go\n",
			want:  RepoStatus{Branch: "main", Upstream: "origin/main", Dirty: 3, Untracked: 1},
			state: "dirty",
		},
		{
			name:  "empty",
			in:    "",
			want:  RepoStatus{},
			state: "clean",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseStatus(tt.in)
			if got.Branch != tt.want.Branch || got.Upstream != tt.want.Upstream ||
				got.Ahead != tt.want.Ahead || got.Behind != tt.want.Behind ||
				got.Dirty != tt.want.Dirty || got.Untracked != tt.want.Untracked ||
				got.Detached != tt.want.Detached || got.Unborn != tt.want.Unborn {
				t.Errorf("ParseStatus(%q) = %+v, want %+v", tt.in, got, tt.want)
			}
			if got.State() != tt.state {
				t.Errorf("State() = %q, want %q", got.State(), tt.state)
			}
		})
	}
}
