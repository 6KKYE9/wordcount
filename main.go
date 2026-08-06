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
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

// countWords 从 reader 里逐行读，把每行拆成词，累计到 map 里。
// 所谓"词"：去掉首尾标点、转小写、按空白和常见标点切分。
func countWords(r *os.File) map[string]int {
	counts := make(map[string]int)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// 转小写，避免 "Go" 和 "go" 被算成两个词。
		line = strings.ToLower(line)
		// FieldsFunc 按"任意非字母数字"切分，比单纯按空格更稳（能吃掉逗号、句号等）。
		words := strings.FieldsFunc(line, func(r rune) bool {
			return !isWordChar(r)
		})
		for _, w := range words {
			if w != "" {
				counts[w]++
			}
		}
	}
	return counts
}

// isWordChar 判断一个字符算不算"词的一部分"（字母或数字）。
func isWordChar(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z'
}

// pair 是排序用的中间结构：一个词 + 它的次数。
type pair struct {
	word  string
	count int
}

func main() {
	file := flag.String("file", "", "要统计的文本文件；不传则读标准输入")
	top := flag.Int("top", 0, "只显示出现最多的前 N 个词（0 表示全部）")
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
		pairs = append(pairs, pair{w, c})
	}
	// 先按次数降序，次数相同再按词升序，结果稳定好看。
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].word < pairs[j].word
	})

	// 限制显示条数。
	limit := len(pairs)
	if *top > 0 && *top < limit {
		limit = *top
	}

	total := 0
	for _, p := range counts {
		total += p
	}
	fmt.Printf("共 %d 个词，%d 个不同词\n", total, len(counts))
	fmt.Println("词频排行：")
	for i := 0; i < limit; i++ {
		fmt.Printf("  %-15s %d\n", pairs[i].word, pairs[i].count)
	}
}
