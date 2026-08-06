package main

import (
	"os"
	"strings"
	"testing"
)

func countFromString(s string) map[string]int {
	return countWords(strings.NewReader(s))
}

func TestCountWordsBasic(t *testing.T) {
	got := countFromString("Go go GO the the language")
	if got["go"] != 3 {
		t.Fatalf("go 应为 3，实际 %d", got["go"])
	}
	if got["the"] != 2 {
		t.Fatalf("the 应为 2，实际 %d", got["the"])
	}
	if got["language"] != 1 {
		t.Fatalf("language 应为 1，实际 %d", got["language"])
	}
}

func TestCountWordsChinese(t *testing.T) {
	// 中文词应被当作整体，而不是拆成单字。
	got := countFromString("语言 语言 标准库 中文 中文 中文")
	if got["语言"] != 2 {
		t.Fatalf("语言 应为 2，实际 %d", got["语言"])
	}
	if got["标准库"] != 1 {
		t.Fatalf("标准库 应为 1，实际 %d", got["标准库"])
	}
	if got["中文"] != 3 {
		t.Fatalf("中文 应为 3，实际 %d", got["中文"])
	}
	// 不应出现单字碎片。
	if _, ok := got["语"]; ok {
		t.Fatalf("中文不应被拆成单字 语")
	}
}

func TestCountWordsPunctuation(t *testing.T) {
	got := countFromString("don't, hello-world! foo_bar: 123")
	if got["don't"] != 1 {
		t.Fatalf("don't 应为 1，实际 %d (%v)", got["don't"], got)
	}
	if got["hello-world"] != 1 {
		t.Fatalf("hello-world 应为 1，实际 %d (%v)", got["hello-world"], got)
	}
	if got["foo_bar"] != 1 {
		t.Fatalf("foo_bar 应为 1，实际 %d (%v)", got["foo_bar"], got)
	}
	if got["123"] != 1 {
		t.Fatalf("123 应为 1，实际 %d", got["123"])
	}
}

func TestCountWordsEmpty(t *testing.T) {
	got := countFromString("   ,,  ... \n\n")
	if len(got) != 0 {
		t.Fatalf("纯标点应为空 map，实际 %v", got)
	}
}

func TestIsWordChar(t *testing.T) {
	for _, c := range "abcXYZ0129中文" {
		if !isWordChar(c) {
			t.Fatalf("%q 应算词字符", c)
		}
	}
	for _, c := range " ,.!?" {
		if isWordChar(c) {
			t.Fatalf("%q 不应算词字符", c)
		}
	}
}

func TestCountFromFile(t *testing.T) {
	f, err := os.CreateTemp("", "wc-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("alpha beta alpha gamma beta beta")
	f.Close()

	ff, err := os.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer ff.Close()
	got := countWords(ff)
	if got["alpha"] != 2 || got["beta"] != 3 || got["gamma"] != 1 {
		t.Fatalf("文件统计错误: %v", got)
	}
}
