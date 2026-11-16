package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "# comment\n" +
		"DOTENV_T_PLAIN=value\n" +
		"DOTENV_T_INLINE=168          # trailing comment\n" +
		"DOTENV_T_EMPTY=               # empty with comment\n" +
		"DOTENV_T_HASH=abc#def\n" +
		"DOTENV_T_QUOTED=\"pa ss # not a comment\"\n" +
		"export DOTENV_T_EXPORT=x\n" +
		"DOTENV_T_PRESET=from-file\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DOTENV_T_PRESET", "from-env")
	for _, k := range []string{"PLAIN", "INLINE", "EMPTY", "HASH", "QUOTED", "EXPORT"} {
		os.Unsetenv("DOTENV_T_" + k)
		t.Cleanup(func() { os.Unsetenv("DOTENV_T_" + k) })
	}

	LoadDotEnv(path)

	want := map[string]string{
		"DOTENV_T_PLAIN":  "value",
		"DOTENV_T_INLINE": "168",
		"DOTENV_T_EMPTY":  "",
		"DOTENV_T_HASH":   "abc#def",
		"DOTENV_T_QUOTED": "pa ss # not a comment",
		"DOTENV_T_EXPORT": "x",
		"DOTENV_T_PRESET": "from-env", // real environment wins
	}
	for k, v := range want {
		if got := os.Getenv(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}
