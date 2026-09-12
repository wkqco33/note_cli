package cmd

import (
	"errors"
	"strings"
	"testing"
)

// withInteractiveEnv 전역 플래그/훅을 테스트 값으로 바꾸고 원복한다.
func withInteractiveEnv(t *testing.T, noInputFlag, assumeYesFlag bool, tty bool) {
	t.Helper()
	prevNoInput, prevAssumeYes, prevTTY := noInput, assumeYes, stdinIsTerminal
	noInput, assumeYes = noInputFlag, assumeYesFlag
	stdinIsTerminal = func() bool { return tty }
	t.Cleanup(func() {
		noInput, assumeYes, stdinIsTerminal = prevNoInput, prevAssumeYes, prevTTY
	})
}

// stubConfirm promptConfirm을 대체하고 호출 여부를 기록한다.
func stubConfirm(t *testing.T, result bool, err error) *bool {
	t.Helper()
	called := false
	prev := promptConfirm
	promptConfirm = func(_, _, _ string) (bool, error) {
		called = true
		return result, err
	}
	t.Cleanup(func() { promptConfirm = prev })
	return &called
}

func TestCanPrompt(t *testing.T) {
	tests := []struct {
		name     string
		noInput  bool
		stdinTTY bool
		want     bool
	}{
		{name: "TTY이고 --no-input 없음", noInput: false, stdinTTY: true, want: true},
		{name: "--no-input이면 TTY여도 금지", noInput: true, stdinTTY: true, want: false},
		{name: "비TTY면 금지", noInput: false, stdinTTY: false, want: false},
		{name: "둘 다면 금지", noInput: true, stdinTTY: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canPrompt(tt.noInput, tt.stdinTTY); got != tt.want {
				t.Fatalf("canPrompt(%v, %v) = %v, want %v", tt.noInput, tt.stdinTTY, got, tt.want)
			}
		})
	}
}

func TestResolveConfirmPolicy(t *testing.T) {
	tests := []struct {
		name      string
		assumeYes bool
		noInput   bool
		stdinTTY  bool
		want      confirmPolicy
	}{
		{name: "--yes면 프롬프트 가능 여부와 무관하게 자동 승인", assumeYes: true, noInput: false, stdinTTY: true, want: confirmAutoYes},
		{name: "--yes + --no-input도 자동 승인", assumeYes: true, noInput: true, stdinTTY: false, want: confirmAutoYes},
		{name: "비TTY에서 --yes 없으면 사용 불가", noInput: false, stdinTTY: false, want: confirmUnavailable},
		{name: "--no-input이면 사용 불가", noInput: true, stdinTTY: true, want: confirmUnavailable},
		{name: "TTY면 프롬프트", noInput: false, stdinTTY: true, want: confirmPrompt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveConfirmPolicy(tt.assumeYes, tt.noInput, tt.stdinTTY); got != tt.want {
				t.Fatalf("resolveConfirmPolicy(%v, %v, %v) = %v, want %v",
					tt.assumeYes, tt.noInput, tt.stdinTTY, got, tt.want)
			}
		})
	}
}

func TestAskConfirmAutoApprovesWithYes(t *testing.T) {
	withInteractiveEnv(t, false, true, false)
	called := stubConfirm(t, false, errors.New("프롬프트가 호출되면 안 됨"))

	got, err := askConfirm("정말 삭제할까요?", "예", "아니오", "힌트")
	if err != nil {
		t.Fatalf("askConfirm() unexpected error: %v", err)
	}
	if !got {
		t.Fatal("askConfirm() = false, want true with --yes")
	}
	if *called {
		t.Fatal("--yes인데 확인 프롬프트가 호출되었습니다")
	}
}

func TestAskConfirmFailsWithoutTTY(t *testing.T) {
	withInteractiveEnv(t, false, false, false)
	called := stubConfirm(t, true, nil)

	_, err := askConfirm("정말 삭제할까요?", "예", "아니오", "삭제하려면 --yes를 지정하세요")
	if !errors.Is(err, ErrInteractionRequired) {
		t.Fatalf("askConfirm() error = %v, want ErrInteractionRequired", err)
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("에러 메시지에 플래그 안내가 없습니다: %v", err)
	}
	if *called {
		t.Fatal("비대화형인데 확인 프롬프트가 호출되었습니다")
	}
}

func TestAskConfirmPromptsInteractive(t *testing.T) {
	withInteractiveEnv(t, false, false, true)
	stubConfirm(t, true, nil)

	got, err := askConfirm("정말 삭제할까요?", "예", "아니오", "힌트")
	if err != nil {
		t.Fatalf("askConfirm() unexpected error: %v", err)
	}
	if !got {
		t.Fatal("askConfirm() = false, want true")
	}
}

func TestAskConfirmRespectsDecline(t *testing.T) {
	withInteractiveEnv(t, false, false, true)
	stubConfirm(t, false, nil)

	got, err := askConfirm("정말 삭제할까요?", "예", "아니오", "힌트")
	if err != nil {
		t.Fatalf("askConfirm() unexpected error: %v", err)
	}
	if got {
		t.Fatal("askConfirm() = true, want false when user declines")
	}
}

func TestRequirePrompt(t *testing.T) {
	withInteractiveEnv(t, false, false, false)

	err := requirePrompt("ID를 인자로 지정하세요")
	if !errors.Is(err, ErrInteractionRequired) {
		t.Fatalf("requirePrompt() error = %v, want ErrInteractionRequired", err)
	}
	if !strings.Contains(err.Error(), "ID를 인자로 지정하세요") {
		t.Fatalf("힌트가 에러 메시지에 없습니다: %v", err)
	}
}

func TestRequirePromptPassesWithTTY(t *testing.T) {
	withInteractiveEnv(t, false, false, true)

	if err := requirePrompt("힌트"); err != nil {
		t.Fatalf("requirePrompt() with TTY error = %v, want nil", err)
	}
}

func TestHandlePromptErrorPropagatesInteractionRequired(t *testing.T) {
	err := handlePromptError(interactionRequired("--title을 지정하세요"), "작성이 취소되었습니다.")
	if !errors.Is(err, ErrInteractionRequired) {
		t.Fatalf("handlePromptError() = %v, want ErrInteractionRequired", err)
	}
}

func TestHandlePromptErrorSwallowsUserAbort(t *testing.T) {
	if err := handlePromptError(errors.New("user aborted"), "작성이 취소되었습니다."); err != nil {
		t.Fatalf("handlePromptError() = %v, want nil for user abort", err)
	}
	if err := handlePromptError(nil, "취소"); err != nil {
		t.Fatalf("handlePromptError(nil) = %v, want nil", err)
	}
}
