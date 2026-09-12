package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveCredentialSource(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		promptOK bool
		want     credentialSource
	}{
		{name: "이메일과 비밀번호 모두 지정", email: "a@b.c", password: "pw", promptOK: false, want: credentialsFromFlags},
		{name: "플래그 없이 프롬프트 가능", email: "", password: "", promptOK: true, want: credentialsFromPrompt},
		{name: "플래그 일부만 + 프롬프트 가능", email: "a@b.c", password: "", promptOK: true, want: credentialsFromPrompt},
		{name: "플래그 없고 프롬프트 불가", email: "", password: "", promptOK: false, want: credentialsUnavailable},
		{name: "일부만 지정했는데 프롬프트 불가", email: "a@b.c", password: "", promptOK: false, want: credentialsUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveCredentialSource(tt.email, tt.password, tt.promptOK)
			if got != tt.want {
				t.Fatalf("resolveCredentialSource(%q, %q, %v) = %v, want %v",
					tt.email, tt.password, tt.promptOK, got, tt.want)
			}
		})
	}
}

func TestReadPasswordStdin(t *testing.T) {
	tests := []struct {
		name    string
		stdin   string
		want    string
		wantErr bool
	}{
		{name: "개행 제거", stdin: "secret\n", want: "secret"},
		{name: "CRLF 제거", stdin: "secret\r\n", want: "secret"},
		{name: "개행 없음", stdin: "secret", want: "secret"},
		{name: "빈 입력은 에러", stdin: "\n", wantErr: true},
		{name: "완전히 빈 입력은 에러", stdin: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readPasswordStdin(strings.NewReader(tt.stdin))
			if (err != nil) != tt.wantErr {
				t.Fatalf("readPasswordStdin() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("readPasswordStdin() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolvePasswordFlag(t *testing.T) {
	dir := t.TempDir()
	pwFile := filepath.Join(dir, "pw.txt")
	if err := os.WriteFile(pwFile, []byte("file-secret\n"), 0o600); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}
	emptyFile := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte("\n"), 0o600); err != nil {
		t.Fatalf("빈 테스트 파일 생성 실패: %v", err)
	}

	tests := []struct {
		name      string
		password  string
		fromStdin bool
		filePath  string
		stdin     string
		want      string
		wantErr   bool
	}{
		{name: "플래그 값 사용", password: "pw", want: "pw"},
		{name: "stdin에서 읽기", fromStdin: true, stdin: "pw\n", want: "pw"},
		{name: "파일에서 읽기", filePath: pwFile, want: "file-secret"},
		{name: "아무것도 없으면 빈 문자열", want: ""},
		{name: "stdin 비어 있으면 에러", fromStdin: true, stdin: "", wantErr: true},
		{name: "빈 파일은 에러", filePath: emptyFile, wantErr: true},
		{name: "없는 파일은 에러", filePath: filepath.Join(dir, "missing.txt"), wantErr: true},
		{name: "password와 stdin 동시 지정은 에러", password: "pw", fromStdin: true, stdin: "x\n", wantErr: true},
		{name: "password와 file 동시 지정은 에러", password: "pw", filePath: pwFile, wantErr: true},
		{name: "stdin과 file 동시 지정은 에러", fromStdin: true, filePath: pwFile, stdin: "x\n", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePasswordFlag(tt.password, tt.fromStdin, tt.filePath, strings.NewReader(tt.stdin))
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolvePasswordFlag() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("resolvePasswordFlag() = %q, want %q", got, tt.want)
			}
		})
	}
}
