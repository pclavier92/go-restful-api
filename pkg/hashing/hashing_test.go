package hashing

import "testing"

func TestHash(t *testing.T) {
	input := "/src/api/middlewares/recovery.go:ioutil.ReadFile(file):vuln"
	expectedHash := "9dfd930747c3136bda953c9c7a523c93ee35ba44"
	hash, err := Make().SHA1(input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if hash != expectedHash {
		t.Errorf("hash is not what we expected, got:\n %s\n wanted:\n %s", hash, expectedHash)
	}
}
