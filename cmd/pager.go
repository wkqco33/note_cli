package cmd

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

// defaultPagerCommand 기본 페이저 명령.
// clig.dev 권장 옵션(-FIRX): 한 화면이면 페이징하지 않고, 대소문자 무시 검색,
// 색상/포맷 유지, 종료 후 화면을 지우지 않는다.
const defaultPagerCommand = "less -FIRX"

// resolvePagerCommand PAGER 환경변수를 실행 커맨드와 인자로 분해한다.
// 설정이 없으면 기본값을 사용하고, "cat"/"none"이면 페이징하지 않는다.
func resolvePagerCommand(pagerEnv string) (name string, args []string, ok bool) {
	fields := strings.Fields(strings.TrimSpace(pagerEnv))
	if len(fields) == 0 {
		fields = strings.Fields(defaultPagerCommand)
	}

	switch fields[0] {
	case "cat", "none":
		return "", nil, false
	}

	return fields[0], fields[1:], true
}

// shouldPage 사람이 읽는 긴 텍스트를 페이저로 넘길지 결정한다.
// stdout이 터미널이고, quiet/--no-pager가 아니며, 유효한 페이저가 있을 때만 페이징한다.
// (stdout이 파이프/리다이렉트면 절대 페이징하지 않는다.)
func shouldPage(stdoutTTY, quiet, noPager bool, pagerEnv string) bool {
	if quiet || noPager || !stdoutTTY {
		return false
	}

	_, _, ok := resolvePagerCommand(pagerEnv)

	return ok
}

// writePaged 사람이 읽는 긴 텍스트를 페이저로 출력한다.
// 페이징 조건이 아니면 stdout에 직접 쓴다. 페이저 실행에 실패해도 직접 출력으로 폴백한다.
func writePaged(text string) error {
	if !shouldPage(term.IsTerminal(int(os.Stdout.Fd())), quietMode, noPagerMode, os.Getenv("PAGER")) {
		_, err := io.WriteString(os.Stdout, text)
		return err
	}

	name, args, _ := resolvePagerCommand(os.Getenv("PAGER"))
	// PAGER는 사용자가 자신의 환경에서 지정하는 값이다($EDITOR와 동일한 신뢰 수준이므로 주입 위험이 없다).
	cmd := exec.Command(name, args...) // nosemgrep
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// 페이저가 없거나 실패해도 결과는 보여준다.
		_, writeErr := io.WriteString(os.Stdout, text)
		return writeErr
	}

	return nil
}
