package cmd

import "testing"

func TestResolvePagerCommand(t *testing.T) {
	tests := []struct {
		name     string
		pagerEnv string
		wantName string
		wantArgs []string
		wantOK   bool
	}{
		{name: "미설정이면 기본 less -FIRX", pagerEnv: "", wantName: "less", wantArgs: []string{"-FIRX"}, wantOK: true},
		{name: "공백만 있으면 기본값", pagerEnv: "   ", wantName: "less", wantArgs: []string{"-FIRX"}, wantOK: true},
		{name: "사용자 지정", pagerEnv: "more -f", wantName: "more", wantArgs: []string{"-f"}, wantOK: true},
		{name: "cat은 페이징 안 함", pagerEnv: "cat", wantName: "", wantArgs: nil, wantOK: false},
		{name: "none은 페이징 안 함", pagerEnv: "none", wantName: "", wantArgs: nil, wantOK: false},
		{name: "옵션 포함 cat도 페이징 안 함", pagerEnv: "cat -n", wantName: "", wantArgs: nil, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, args, ok := resolvePagerCommand(tt.pagerEnv)
			if ok != tt.wantOK {
				t.Fatalf("resolvePagerCommand(%q) ok = %v, want %v", tt.pagerEnv, ok, tt.wantOK)
			}
			if name != tt.wantName {
				t.Fatalf("resolvePagerCommand(%q) name = %q, want %q", tt.pagerEnv, name, tt.wantName)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("resolvePagerCommand(%q) args = %v, want %v", tt.pagerEnv, args, tt.wantArgs)
			}
			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Fatalf("resolvePagerCommand(%q) args = %v, want %v", tt.pagerEnv, args, tt.wantArgs)
				}
			}
		})
	}
}

func TestShouldPage(t *testing.T) {
	tests := []struct {
		name      string
		stdoutTTY bool
		quiet     bool
		noPager   bool
		pagerEnv  string
		want      bool
	}{
		{name: "TTY + 페이저", stdoutTTY: true, pagerEnv: "", want: true},
		{name: "비TTY(파이프)는 페이징 안 함", stdoutTTY: false, pagerEnv: "", want: false},
		{name: "quiet이면 안 함", stdoutTTY: true, quiet: true, pagerEnv: "", want: false},
		{name: "--no-pager면 안 함", stdoutTTY: true, noPager: true, pagerEnv: "", want: false},
		{name: "PAGER=cat이면 안 함", stdoutTTY: true, pagerEnv: "cat", want: false},
		{name: "PAGER=none이면 안 함", stdoutTTY: true, pagerEnv: "none", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldPage(tt.stdoutTTY, tt.quiet, tt.noPager, tt.pagerEnv); got != tt.want {
				t.Fatalf("shouldPage(%v, %v, %v, %q) = %v, want %v",
					tt.stdoutTTY, tt.quiet, tt.noPager, tt.pagerEnv, got, tt.want)
			}
		})
	}
}
