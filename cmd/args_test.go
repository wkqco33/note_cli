package cmd

import "testing"

func TestRequireMaxArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		max  int
		want bool
	}{
		{name: "within limit", args: []string{"1"}, max: 1, want: true},
		{name: "empty", args: nil, max: 1, want: true},
		{name: "over limit", args: []string{"1", "2"}, max: 1, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireMaxArgs(tt.args, tt.max)
			if (err == nil) != tt.want {
				t.Fatalf("requireMaxArgs(%v, %d) error = %v, want success = %v", tt.args, tt.max, err, tt.want)
			}
		})
	}
}

func TestRequireExactArgs(t *testing.T) {
	if err := requireExactArgs([]string{"key", "value"}, 2); err != nil {
		t.Fatalf("requireExactArgs() unexpected error: %v", err)
	}
	if err := requireExactArgs([]string{"key"}, 2); err == nil {
		t.Fatal("requireExactArgs() expected an error for a missing argument")
	}
}

func TestNormalizeRootArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "global debug before command",
			args: []string{"--debug", "list"},
			want: []string{"list", "--debug"},
		},
		{
			name: "inline global debug before command",
			args: []string{"--debug=false", "list"},
			want: []string{"list", "--debug=false"},
		},
		{
			name: "global yes before command",
			args: []string{"--yes", "delete", "1"},
			want: []string{"delete", "1", "--yes"},
		},
		{
			name: "short yes before command",
			args: []string{"-y", "delete", "1"},
			want: []string{"delete", "1", "-y"},
		},
		{
			name: "global no-input before command",
			args: []string{"--no-input", "add", "-t", "x"},
			want: []string{"add", "-t", "x", "--no-input"},
		},
		{
			name: "multiple global flags before command",
			args: []string{"--no-input", "--yes", "delete", "1"},
			want: []string{"delete", "1", "--no-input", "--yes"},
		},
		{
			name: "other root flag remains unchanged",
			args: []string{"--help", "list"},
			want: []string{"--help", "list"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeRootArgs(tt.args)
			if len(got) != len(tt.want) {
				t.Fatalf("normalizeRootArgs() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("normalizeRootArgs() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
