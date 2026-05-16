package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/Ixecd/blitz/internal/wallet/core"
	"golang.org/x/term"
)

// encrypt-seed 用 passphrase 加密 HD 种子文件
//
// 用法:
//
//	encrypt-seed <种子hex> <输出文件路径>
//
// 示例:
//
//	encrypt-seed "001bf04eb9b836141765..." configs/secrets/hd-seed.enc
func main() {
	if len(os.Args) < 3 {
		fmt.Println("用法: encrypt-seed <种子hex> <输出文件路径>")
		fmt.Println("示例: encrypt-seed \"001bf04eb9b8...\" configs/secrets/hd-seed.enc")
		os.Exit(1)
	}

	seedHex := os.Args[1]
	outPath := os.Args[2]

	fmt.Print("设置加密密码: ")
	passBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取密码失败: %v\n", err)
		os.Exit(1)
	}
	if len(passBytes) < 8 {
		fmt.Fprintln(os.Stderr, "密码至少 8 个字符")
		os.Exit(1)
	}

	encrypted, err := core.EncryptSeed(seedHex, string(passBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "加密失败: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, encrypted, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "写入文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 种子已加密写入 %s\n", outPath)
	fmt.Printf("⛔  原始 hex 种子请从 .env / 环境变量中删除\n")
	fmt.Printf("🔑 启动时需提供密码（环境变量 SEED_PASSPHRASE 或交互输入）\n")
}
