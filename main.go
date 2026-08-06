// wordcount：统计文本里每个词出现了多少次（词频统计）
//
// 这是个很常见的文本分析小工具：读一段文字，把每个"单词"拆出来，数次数，再按多少排个序。
// 全程只用标准库：os（读文件/管道）、bufio（高效按行读）、strings（切词/清洗）、sort（排序）。
//
// 用法：
//   wordcount -file 文章.txt          # 统计某个文件
//   type 文章.txt | wordcount -top 10 # 从管道读，只看出现最多的 10 个词
//   wordcount                          # 不传 -file 就等你粘贴文本，按 Ctrl+Z 结束（Windows）
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
)

// isWordChar 判断一个字符算不算"词的一部分"：字母（含中文等 Unicode 字母）、
// 数字、或连接符（连字符、下划线、撇号，用来保留 don't / 状态-词 这类）。
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) ||
		r == '\'' || r == '-' || r == '_'
}

// countWords 从 reader 里逐行读，把每行拆成词，累计到 map 里。
// 所谓"词"：转小写、按非词字符切分；支持中文等多语言（不在 ASCII 范围也照样算词）。
func countWords(r io.Reader) map[string]int {
	counts := make(map[string]int)
	scanner := bufio.NewScanner(r)
	// 放大单行上限，避免长行（比如整段没换行）被截断。
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		// 转小写，避免 "Go" 和 "go" 被算成两个词；中文没大小写概念，转换无副作用。
		line = strings.ToLower(line)
		words := strings.FieldsFunc(line, func(r rune) bool {
			return !isWordChar(r)
		})
		for _, w := range words {
			// 去掉首尾的连接符（避免 "don't" 切出 "'" 或尾随 "-"）。
			w = strings.Trim(w, "'-_")
			if w != "" {
				counts[w]++
			}
		}
	}
	return counts
}

// pair 是排序用的中间结构：一个词 + 它的次数。
type pair struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

func main() {
	file := flag.String("file", "", "要统计的文本文件；不传则读标准输入")
	top := flag.Int("top", 0, "只显示出现最多的前 N 个词（0 表示全部）")
	min := flag.Int("min", 1, "只显示出现次数 >= 该值的词（过滤罕见词）")
	jsonOut := flag.Bool("json", false, "以 JSON 数组输出结果（便于管道给其他程序）")
	flag.Parse()

	var src *os.File
	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "打开文件失败:", err)
			os.Exit(1)
		}
		defer f.Close()
		src = f
	} else {
		src = os.Stdin
	}

	counts := countWords(src)

	// 转成切片才能排序（map 本身无序）。
	pairs := make([]pair, 0, len(counts))
	for w, c := range counts {
		if c >= *min {
			pairs = append(pairs, pair{w, c})
		}
	}
	// 先按次数降序，次数相同再按词升序，结果稳定好看。
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Count != pairs[j].Count {
			return pairs[i].Count > pairs[j].Count
		}
		return pairs[i].Word < pairs[j].Word
	})

	// 限制显示条数。
	limit := len(pairs)
	if *top > 0 && *top < limit {
		limit = *top
	}

	if *jsonOut {
		out, err := json.MarshalIndent(pairs[:limit], "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "JSON 编码失败:", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
		return
	}

	total := 0
	for _, p := range counts {
		total += p
	}
	fmt.Printf("共 %d 个词，%d 个不同词\n", total, len(counts))
	fmt.Println("词频排行：")
	for i := 0; i < limit; i++ {
		fmt.Printf("  %-15s %d\n", pairs[i].Word, pairs[i].Count)
	}
}
