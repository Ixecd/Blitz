package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasLetterAndDigit(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"纯字母", "abcdefgh", false},
		{"纯数字", "12345678", false},
		{"字母+数字", "abc12345", true},
		{"大写+数字", "ABCD1234", true},
		{"混合大小写+数字", "AbcDef12", true},
		{"纯符号", "!@#$%^&*", false},
		{"空字符串", "", false},
		{"短密码字母+数字", "a1", true},
		{"含空格字母+数字", "abc 123 45", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasLetterAndDigit(tt.password))
		})
	}
}
