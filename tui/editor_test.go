package tui

import (
	"runtime"
	"testing"
)

func TestResolveEditorCommandDefault(t *testing.T) {
	t.Setenv("EDITOR", "")

	name, args := resolveEditorCommand()
	want := "vim"
	if runtime.GOOS == "windows" {
		want = "notepad"
	}
	if name != want {
		t.Fatalf("unexpected default editor: %q", name)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
}

func TestResolveEditorCommandSplitsArgs(t *testing.T) {
	// PATH에 없는 명령이면 공백 기준으로 명령과 인자를 분리해야 한다
	t.Setenv("EDITOR", "definitely-not-a-real-editor --wait --new-window")

	name, args := resolveEditorCommand()
	if name != "definitely-not-a-real-editor" {
		t.Fatalf("unexpected editor name: %q", name)
	}
	if len(args) != 2 || args[0] != "--wait" || args[1] != "--new-window" {
		t.Fatalf("unexpected editor args: %v", args)
	}
}
