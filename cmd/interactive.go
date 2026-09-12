package cmd

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
	"os"
)

// ErrInteractionRequired 대화형 입력이 필요하지만 프롬프트를 띄울 수 없을 때 반환하는 센티넬 에러.
// 호출부는 이 에러를 삼키지 말고 그대로 반환해 0이 아닌 종료 코드로 이어지게 해야 한다.
var ErrInteractionRequired = errors.New("대화형 입력이 필요합니다")

// confirmPolicy 확인 프롬프트 처리 방식
type confirmPolicy int

const (
	// confirmPrompt 사람에게 확인을 묻는다
	confirmPrompt confirmPolicy = iota
	// confirmAutoYes --yes로 확인을 자동 승인한다
	confirmAutoYes
	// confirmUnavailable 프롬프트를 띄울 수 없어 실패해야 한다
	confirmUnavailable
)

// canPrompt 순수 함수: 프롬프트 표시 가능 여부.
// --no-input이거나 stdin이 터미널이 아니면(파이프/CI/에이전트) 프롬프트를 띄우지 않는다.
func canPrompt(noInput, stdinTTY bool) bool {
	return !noInput && stdinTTY
}

// resolveConfirmPolicy 순수 함수: 확인 프롬프트를 어떻게 처리할지 결정한다.
// --yes는 프롬프트 가능 여부와 무관하게 자동 승인한다.
func resolveConfirmPolicy(assumeYes, noInput, stdinTTY bool) confirmPolicy {
	if assumeYes {
		return confirmAutoYes
	}
	if !canPrompt(noInput, stdinTTY) {
		return confirmUnavailable
	}

	return confirmPrompt
}

// stdinIsTerminal 표준 입력이 터미널인지 판단한다 (테스트에서 교체 가능한 훅).
// bubbletea는 비TTY에서 /dev/tty를 열어 입력을 기다릴 수 있으므로, 프롬프트 전에 직접 확인한다.
var stdinIsTerminal = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// promptConfirm 확인 프롬프트를 실행한다 (테스트에서 교체 가능한 훅).
// 사용자가 취소하면 huh.ErrUserAborted를 반환한다.
var promptConfirm = func(title, affirmative, negative string) (bool, error) {
	var confirmed bool
	err := huh.NewConfirm().
		Title(title).
		Affirmative(affirmative).
		Negative(negative).
		Value(&confirmed).
		Run()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}

// interactionRequired 프롬프트 대신 사용할 플래그를 안내하는 에러를 만든다.
func interactionRequired(hint string) error {
	if hint == "" {
		return ErrInteractionRequired
	}

	return fmt.Errorf("%w: %s", ErrInteractionRequired, hint)
}

// requirePrompt 프롬프트를 띄울 수 없으면 힌트와 함께 에러를 반환한다.
// TUI 폼/선택을 실행하기 전에 호출한다.
func requirePrompt(hint string) error {
	if !promptAllowed() {
		return interactionRequired(hint)
	}

	return nil
}

// promptAllowed 현재 전역 플래그와 stdin 상태에서 프롬프트를 띄울 수 있는지 판단한다.
func promptAllowed() bool {
	return canPrompt(noInput, stdinIsTerminal())
}

// askConfirm 파괴적 작업 확인.
// --yes면 프롬프트 없이 승인하고, 비대화형이면 hint와 함께 실패한다.
func askConfirm(title, affirmative, negative, hint string) (bool, error) {
	switch resolveConfirmPolicy(assumeYes, noInput, stdinIsTerminal()) {
	case confirmAutoYes:
		return true, nil
	case confirmUnavailable:
		return false, interactionRequired(hint)
	default:
		return promptConfirm(title, affirmative, negative)
	}
}

// handlePromptError 프롬프트 결과를 커맨드 반환값으로 변환한다.
// 프롬프트 불가(ErrInteractionRequired)는 종료 코드를 남기기 위해 그대로 반환하고,
// 사람이 직접 취소한 경우에만 안내 문구를 출력하고 정상 종료한다.
func handlePromptError(err error, cancelMsg string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrInteractionRequired) {
		return err
	}

	fmt.Println(cancelMsg)
	return nil
}
