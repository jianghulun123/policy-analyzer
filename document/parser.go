package document

import (
	"context"
	"os"
	"strings"
)

// ParseResult 解析结果
type ParseResult struct {
	Text       string            // 纯文本内容
	Pages      int               // 页数
	Meta       map[string]string // 文档元数据
	Structured *DocumentStructure // 结构化数据
}

// DocumentStructure 文档结构
type DocumentStructure struct {
	Title    string
	Sections []Section
	Tables   []Table
}

// Section 文档章节
type Section struct {
	Title   string
	Content string
	Level   int
}

// Table 文档表格
type Table struct {
	Headers []string
	Rows    [][]string
}

// Parser 文档解析接口
type Parser interface {
	// 支持的文件扩展名
	SupportedExtensions() []string

	// 解析文档
	Parse(ctx context.Context, path string) (*ParseResult, error)

	// 提取元数据（可选）
	ExtractMeta(ctx context.Context, path string) (map[string]string, error)
}

// ========== TXT解析器 ==========

// TXTParser 纯文本文件解析器
type TXTParser struct{}

func NewTXTParser() *TXTParser {
	return &TXTParser{}
}

func (p *TXTParser) SupportedExtensions() []string {
	return []string{".txt"}
}

func (p *TXTParser) Parse(ctx context.Context, path string) (*ParseResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	text := string(data)

	// 尝试识别结构
	structured := p.parseStructure(text)

	return &ParseResult{
		Text:       text,
		Pages:      1,
		Structured: structured,
	}, nil
}

func (p *TXTParser) ExtractMeta(ctx context.Context, path string) (map[string]string, error) {
	return nil, nil
}

func (p *TXTParser) parseStructure(text string) *DocumentStructure {
	lines := strings.Split(text, "\n")
	sections := make([]Section, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 尝试识别标题行（以数字开头或特殊格式）
		level := 0
		if strings.HasPrefix(line, "第") && strings.Contains(line, "章") {
			level = 1
		} else if strings.HasPrefix(line, "第") && strings.Contains(line, "条") {
			level = 2
		} else if strings.HasPrefix(line, "一、") ||
			strings.HasPrefix(line, "二、") ||
			strings.HasPrefix(line, "三、") ||
			strings.HasPrefix(line, "四、") ||
			strings.HasPrefix(line, "五、") {
			level = 2
		} else if strings.HasPrefix(line, "（一）") ||
			strings.HasPrefix(line, "（二）") ||
			strings.HasPrefix(line, "（三）") {
			level = 3
		}

		if level > 0 {
			sections = append(sections, Section{
				Title:   line,
				Level:   level,
				Content: "",
			})
		}
	}

	if len(sections) > 0 {
		return &DocumentStructure{Sections: sections}
	}
	return nil
}

// ========== Markdown解析器 ==========

// MDParser Markdown文件解析器
type MDParser struct{}

func NewMDParser() *MDParser {
	return &MDParser{}
}

func (p *MDParser) SupportedExtensions() []string {
	return []string{".md", ".markdown"}
}

func (p *MDParser) Parse(ctx context.Context, path string) (*ParseResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	text := string(data)

	// 解析Markdown结构
	structured := p.parseMarkdown(text)

	return &ParseResult{
		Text:       text,
		Pages:      1,
		Structured: structured,
	}, nil
}

func (p *MDParser) ExtractMeta(ctx context.Context, path string) (map[string]string, error) {
	// 可以从Markdown的front matter提取元数据
	return nil, nil
}

func (p *MDParser) parseMarkdown(text string) *DocumentStructure {
	lines := strings.Split(text, "\n")
	sections := make([]Section, 0)
	title := ""

	for _, line := range lines {
		// 提取标题
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			title = strings.TrimSpace(title)
			continue
		}

		// 提取章节
		level := 0
		if strings.HasPrefix(line, "## ") {
			level = 2
		} else if strings.HasPrefix(line, "### ") {
			level = 3
		} else if strings.HasPrefix(line, "#### ") {
			level = 4
		}

		if level > 0 {
			titleText := strings.TrimPrefix(line, strings.Repeat("#", level)+" ")
			sections = append(sections, Section{
				Title:   strings.TrimSpace(titleText),
				Level:   level,
				Content: "",
			})
		}
	}

	return &DocumentStructure{
		Title:    title,
		Sections: sections,
	}
}