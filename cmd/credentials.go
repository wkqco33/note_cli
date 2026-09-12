package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// credentialSource 인증 정보(이메일/비밀번호) 획득 경로
type credentialSource int

const (
	// credentialsFromFlags 플래그로 모두 지정됨
	credentialsFromFlags credentialSource = iota
	// credentialsFromPrompt 대화형 폼으로 입력받아야 함
	credentialsFromPrompt
	// credentialsUnavailable 프롬프트를 띄울 수 없어 실패해야 함
	credentialsUnavailable
)

// resolveCredentialSource 순수 함수: 플래그 값과 프롬프트 가능 여부로 획득 경로를 결정한다.
// 이메일과 비밀번호가 모두 있으면 프롬프트 없이 진행하고, 아니면 프롬프트 가능할 때만 폼을 띄운다.
func resolveCredentialSource(email, password string, promptOK bool) credentialSource {
	if email != "" && password != "" {
		return credentialsFromFlags
	}
	if promptOK {
		return credentialsFromPrompt
	}

	return credentialsUnavailable
}

// readPasswordStdin stdin에서 비밀번호를 읽고 끝 개행을 제거한다.
// 에이전트/CI가 셸 히스토리에 비밀번호를 남기지 않고 전달할 수 있게 한다.
func readPasswordStdin(stdin io.Reader) (string, error) {
	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("stdin을 읽지 못했습니다: %w", err)
	}

	password := strings.TrimRight(string(data), "\r\n")
	if password == "" {
		return "", fmt.Errorf("stdin으로 전달된 비밀번호가 비어 있습니다")
	}

	return password, nil
}

// readPasswordFile 파일에서 비밀번호를 읽고 끝 개행을 제거한다.
// 에이전트/CI가 셸 히스토리나 프로세스 목록에 비밀번호를 노출하지 않고 전달할 수 있게 한다.
func readPasswordFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("비밀번호 파일을 읽지 못했습니다 (%s): %w", path, err)
	}

	password := strings.TrimRight(string(data), "\r\n")
	if password == "" {
		return "", fmt.Errorf("비밀번호 파일이 비어 있습니다: %s", path)
	}

	return password, nil
}

// resolvePasswordFlag --password/--password-stdin/--password-file을 해석한다.
// 두 가지 이상 동시 지정은 에러. 어느 것도 지정하지 않으면 빈 문자열을 반환한다.
func resolvePasswordFlag(password string, fromStdin bool, filePath string, stdin io.Reader) (string, error) {
	sources := 0
	if password != "" {
		sources++
	}
	if fromStdin {
		sources++
	}
	if filePath != "" {
		sources++
	}
	if sources > 1 {
		return "", fmt.Errorf("--password, --password-stdin, --password-file은 동시에 지정할 수 없습니다")
	}

	if fromStdin {
		return readPasswordStdin(stdin)
	}
	if filePath != "" {
		return readPasswordFile(filePath)
	}

	return password, nil
}

// warnInsecurePassword 플래그로 비밀번호를 직접 전달할 때 노출 위험을 안내한다.
func warnInsecurePassword(password string) {
	if password == "" {
		return
	}
	statusln("경고: --password는 프로세스 목록(ps)과 셸 히스토리에 노출될 수 있습니다. --password-stdin 또는 --password-file을 사용하세요.")
}

// credentialsRequiredHint 프롬프트 없이 인증 정보를 받기 위한 플래그 안내
func credentialsRequiredHint() string {
	return "--email과 --password (또는 --password-stdin) 플래그를 지정하세요"
}
