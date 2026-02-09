package main

import (
	"fmt"
	"time"

	"github.com/yeeaiclub/fasttui/terminal"
)

type SimpleText struct {
	content string
}

func (s *SimpleText) Render(width int) []string {
	return []string{s.content}
}

func (s *SimpleText) HandleInput(data string) {}

func (s *SimpleText) WantsKeyRelease() bool {
	return false
}

func (s *SimpleText) Invalidate() {}

func NewSimpleText(content string) *SimpleText {
	return &SimpleText{content: content}
}

func main() {
	buf := terminal.NewStdinBuffer()

	buf.OnData = func(seq string) {
		fmt.Printf("收到数据: %q\n", seq)
	}

	buf.OnPaste = func(paste string) {
		fmt.Printf("收到粘贴内容: %q\n", paste)
	}

	testCases := []struct {
		name string
		data string
	}{
		{"普通字符", "hello"},
		{"回车键", "\r"},
		{"方向键上", "\x1b[A"},
		{"方向键下", "\x1b[B"},
		{"方向键左", "\x1b[D"},
		{"方向键右", "\x1b[C"},
		{"粘贴内容", "\x1b[200~hello world\x1b[201~"},
		{"Tab键", "\t"},
		{"Delete键", "\x1b[3~"},
		{"Home键", "\x1b[H"},
		{"End键", "\x1b[F"},
	}

	for _, tc := range testCases {
		fmt.Printf("\n测试: %s\n", tc.name)
		buf.Process(tc.data)
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("\n测试完成")
}
