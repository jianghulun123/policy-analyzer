package document

import (
	"context"
	"fmt"
	"strings"

	"github.com/unidoc/unioffice/document"
)

// DOCXParser DOCX文档解析器
type DOCXParser struct{}

// NewDOCXParser 创建DOCX解析器
func NewDOCXParser() *DOCXParser {
	return &DOCXParser{}
}

// SupportedExtensions 返回支持的文件扩展名
func (p *DOCXParser) SupportedExtensions() []string {
	return []string{".docx"}
}

// Parse 解析DOCX文档
func (p *DOCXParser) Parse(ctx context.Context, path string) (*ParseResult, error) {
	doc, err := document.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开DOCX文件失败: %w", err)
	}
	defer doc.Close()

	text, sections := p.extractContent(doc)
	structured := p.buildStructure(text, sections)
	pageCount := p.estimatePageCount(text)

	return &ParseResult{
		Text:       text,
		Pages:      pageCount,
		Meta:       p.extractMeta(doc),
		Structured: structured,
	}, nil
}

// ExtractMeta 提取元数据
func (p *DOCXParser) ExtractMeta(ctx context.Context, path string) (map[string]string, error) {
	doc, err := document.Open(path)
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	return p.extractMeta(doc), nil
}

// DocumentSection 文档节信息
type DocumentSection struct {
	Text  string
	Style string
}

// extractContent 提取文档内容
func (p *DOCXParser) extractContent(doc *document.Document) (string, []DocumentSection) {
	var output strings.Builder
	sections := make([]DocumentSection, 0)

	for _, para := range doc.Paragraphs() {
		text := p.extractParagraphText(para)
		if text == "" {
			continue
		}

		output.WriteString(text)
		output.WriteString("\n")

		style := p.getParagraphStyle(para)
		sections = append(sections, DocumentSection{
			Text:  text,
			Style: style,
		})
	}

	for _, table := range doc.Tables() {
		tableText := p.extractTableText(table)
		if tableText != "" {
			output.WriteString("\n[表格]\n")
			output.WriteString(tableText)
			output.WriteString("\n")
		}
	}

	return output.String(), sections
}

// extractParagraphText 提取段落文本
func (p *DOCXParser) extractParagraphText(para document.Paragraph) string {
	var text strings.Builder
	for _, run := range para.Runs() {
		text.WriteString(run.Text())
	}
	return strings.TrimSpace(text.String())
}

// getParagraphStyle 获取段落样式
func (p *DOCXParser) getParagraphStyle(para document.Paragraph) string {
	props := para.Properties()
	if props.X() == nil {
		return "Normal"
	}

	styleVal := props.X().PStyle
	if styleVal != nil && styleVal.ValAttr != "" {
		return styleVal.ValAttr
	}

	return "Normal"
}

// extractTableText 提取表格文本
func (p *DOCXParser) extractTableText(table document.Table) string {
	var output strings.Builder

	for _, row := range table.Rows() {
		cells := row.Cells()
		for i, cell := range cells {
			for _, para := range cell.Paragraphs() {
				text := p.extractParagraphText(para)
				output.WriteString(text)
				if i < len(cells)-1 {
					output.WriteString("\t")
				}
			}
		}
		output.WriteString("\n")
	}

	return output.String()
}

// extractMeta 提取文档元数据
func (p *DOCXParser) extractMeta(doc *document.Document) map[string]string {
	meta := make(map[string]string)

	coreProps := doc.CoreProperties

	if coreProps.Title() != "" {
		meta["title"] = coreProps.Title()
	}
	if coreProps.Author() != "" {
		meta["author"] = coreProps.Author()
	}
	if !coreProps.Created().IsZero() {
		meta["created"] = coreProps.Created().Format("2006-01-02")
	}
	if !coreProps.Modified().IsZero() {
		meta["modified"] = coreProps.Modified().Format("2006-01-02")
	}

	return meta
}

// buildStructure 根据样式构建文档结构
func (p *DOCXParser) buildStructure(text string, sections []DocumentSection) *DocumentStructure {
	structure := &DocumentStructure{
		Sections: make([]Section, 0),
		Tables:   make([]Table, 0),
	}

	styleToLevel := map[string]int{
		"Heading1": 1,
		"Heading2": 2,
		"Heading3": 3,
		"Heading4": 4,
		"Heading5": 5,
		"Title":    0,
	}

	for _, section := range sections {
		if level, ok := styleToLevel[section.Style]; ok {
			if level == 0 {
				structure.Title = section.Text
			} else {
				structure.Sections = append(structure.Sections, Section{
					Title: section.Text,
					Level: level,
				})
			}
			continue
		}

		level := p.detectSectionLevel(section.Text)
		if level > 0 {
			structure.Sections = append(structure.Sections, Section{
				Title: section.Text,
				Level: level,
			})
		}
	}

	if structure.Title == "" && len(sections) > 0 {
		firstLine := sections[0].Text
		if len(firstLine) > 2 && len(firstLine) < 50 {
			structure.Title = firstLine
		}
	}

	return structure
}

// detectSectionLevel 基于文本模式检测章节级别
func (p *DOCXParser) detectSectionLevel(text string) int {
	if strings.HasPrefix(text, "第一章") || strings.HasPrefix(text, "第二章") ||
		strings.HasPrefix(text, "第三章") || strings.HasPrefix(text, "第四章") ||
		strings.HasPrefix(text, "第五章") || strings.HasPrefix(text, "第六章") ||
		strings.HasPrefix(text, "第七章") || strings.HasPrefix(text, "第八章") {
		return 1
	}

	if strings.HasPrefix(text, "第一条") || strings.HasPrefix(text, "第二条") ||
		strings.HasPrefix(text, "第三条") || strings.HasPrefix(text, "第四条") {
		return 2
	}

	if len(text) > 2 && (strings.HasPrefix(text, "一、") ||
		strings.HasPrefix(text, "二、") || strings.HasPrefix(text, "三、") ||
		strings.HasPrefix(text, "四、") || strings.HasPrefix(text, "五、")) {
		return 2
	}

	if len(text) > 3 && (strings.HasPrefix(text, "（一）") ||
		strings.HasPrefix(text, "（二）") || strings.HasPrefix(text, "（三）")) {
		return 3
	}

	return 0
}

// estimatePageCount 估算页数
func (p *DOCXParser) estimatePageCount(text string) int {
	charCount := len([]rune(text))
	pages := charCount / 2000
	if pages < 1 {
		pages = 1
	}
	return pages
}

// parseTables 解析文档中的表格
func (p *DOCXParser) parseTables(doc *document.Document) []Table {
	tables := make([]Table, 0)

	for _, table := range doc.Tables() {
		t := Table{
			Headers: make([]string, 0),
			Rows:    make([][]string, 0),
		}

		rows := table.Rows()
		if len(rows) == 0 {
			continue
		}

		for _, cell := range rows[0].Cells() {
			header := ""
			for _, para := range cell.Paragraphs() {
				header += p.extractParagraphText(para)
			}
			t.Headers = append(t.Headers, strings.TrimSpace(header))
		}

		for i := 1; i < len(rows); i++ {
			row := make([]string, 0)
			for _, cell := range rows[i].Cells() {
				cellText := ""
				for _, para := range cell.Paragraphs() {
					cellText += p.extractParagraphText(para)
				}
				row = append(row, strings.TrimSpace(cellText))
			}
			t.Rows = append(t.Rows, row)
		}

		tables = append(tables, t)
	}

	return tables
}
