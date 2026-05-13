package document

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	"policy-analyzer/logger"
)

// log 模块级日志记录器
var pdfLog = logger.WithModule("pdf-parser")

// PDFParser PDF文档解析器
// 使用pdfcpu库提取PDF文本内容，支持中文文档
type PDFParser struct{}

// NewPDFParser 创建PDF解析器
func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

// SupportedExtensions 返回支持的文件扩展名
func (p *PDFParser) SupportedExtensions() []string {
	return []string{".pdf"}
}

// Parse 解析PDF文档
// 提取文本内容并尝试识别文档结构
func (p *PDFParser) Parse(ctx context.Context, path string) (*ParseResult, error) {
	pdfLog.Info("开始解析PDF文档", logger.F("path", path))

	// 检查文件是否存在
	fileInfo, err := os.Stat(path)
	if err != nil {
		pdfLog.ErrorErr("PDF文件不存在或无法访问", err, logger.F("path", path))
		return nil, fmt.Errorf("PDF文件不存在或无法访问: %w", err)
	}
	pdfLog.Info("PDF文件信息", logger.F("size", fileInfo.Size()), logger.F("path", path))

	// 获取PDF信息
	info, err := p.getPDFInfo(path)
	if err != nil {
		pdfLog.ErrorErr("获取PDF信息失败", err, logger.F("path", path))
		return nil, fmt.Errorf("获取PDF信息失败: %w", err)
	}
	pdfLog.Info("PDF信息获取成功", logger.F("pages", info.PageCount), logger.F("path", path))

	// 提取文本内容
	text, err := p.extractText(path)
	if err != nil {
		pdfLog.ErrorErr("提取PDF文本失败", err, logger.F("path", path))
		return nil, fmt.Errorf("提取PDF文本失败: %w", err)
	}

	// 检查提取的文本内容
	if len(text) == 0 {
		pdfLog.Warn("PDF文本内容为空，可能是扫描版PDF", logger.F("path", path))
	} else {
		pdfLog.Info("PDF文本提取成功",
			logger.F("text_length", len(text)),
			logger.F("preview", truncateString(text, 100)),
			logger.F("path", path))
	}

	// 解析文档结构
	structured := p.parseStructure(text)
	pdfLog.Info("PDF文档结构解析完成",
		logger.F("sections", len(structured.Sections)),
		logger.F("title", structured.Title),
		logger.F("path", path))

	return &ParseResult{
		Text:       text,
		Pages:      info.PageCount,
		Meta:       info.Meta,
		Structured: structured,
	}, nil
}

// ExtractMeta 提取PDF元数据
func (p *PDFParser) ExtractMeta(ctx context.Context, path string) (map[string]string, error) {
	info, err := p.getPDFInfo(path)
	if err != nil {
		return nil, err
	}
	return info.Meta, nil
}

// PDFInfo PDF文档信息
type PDFInfo struct {
	PageCount int
	Meta      map[string]string
}

// getPDFInfo 获取PDF基本信息
func (p *PDFParser) getPDFInfo(path string) (*PDFInfo, error) {
	pdfLog.Debug("读取PDF上下文", logger.F("path", path))

	// 使用pdfcpu获取PDF上下文
	ctx, err := api.ReadContextFile(path)
	if err != nil {
		pdfLog.ErrorErr("读取PDF上下文失败", err, logger.F("path", path))
		return nil, fmt.Errorf("读取PDF上下文失败: %w", err)
	}

	info := &PDFInfo{
		PageCount: ctx.PageCount,
		Meta:      make(map[string]string),
	}

	// 添加基本统计信息
	info.Meta["page_count"] = fmt.Sprintf("%d", info.PageCount)

	pdfLog.Debug("PDF基本信息",
		logger.F("page_count", info.PageCount),
		logger.F("path", path))

	return info, nil
}

// extractText 提取PDF文本内容
func (p *PDFParser) extractText(path string) (string, error) {
	pdfLog.Debug("创建临时目录用于提取PDF内容", logger.F("path", path))

	// 创建临时目录存放提取的内容
	tmpDir, err := os.MkdirTemp("", "pdfcpu-extract-")
	if err != nil {
		pdfLog.ErrorErr("创建临时目录失败", err)
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pdfLog.Debug("使用pdfcpu提取PDF内容",
		logger.F("tmp_dir", tmpDir),
		logger.F("path", path))

	// 使用pdfcpu提取内容
	err = api.ExtractContentFile(path, tmpDir, nil, nil)
	if err != nil {
		pdfLog.Warn("pdfcpu ExtractContentFile失败，尝试备用方法",
			logger.F("error", err.Error()),
			logger.F("path", path))
		return p.extractTextAlternative(path)
	}

	// 读取提取的文本文件
	var output strings.Builder
	files, _ := os.ReadDir(tmpDir)
	pdfLog.Debug("读取提取的文本文件",
		logger.F("file_count", len(files)),
		logger.F("tmp_dir", tmpDir))

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".txt") {
			filePath := tmpDir + "/" + f.Name()
			data, err := os.ReadFile(filePath)
			if err == nil {
				output.WriteString(string(data))
				output.WriteString("\n")
				pdfLog.Debug("读取文本文件成功",
					logger.F("file", f.Name()),
					logger.F("size", len(data)))
			} else {
				pdfLog.Warn("读取文本文件失败",
					logger.F("file", f.Name()),
					logger.F("error", err.Error()))
			}
		}
	}

	result := output.String()
	if len(result) == 0 {
		pdfLog.Warn("PDF内容提取结果为空", logger.F("path", path))
	}

	return result, nil
}

// extractTextAlternative 备用文本提取方法
func (p *PDFParser) extractTextAlternative(path string) (string, error) {
	pdfLog.Info("使用备用方法提取PDF文本", logger.F("path", path))

	ctx, err := api.ReadContextFile(path)
	if err != nil {
		pdfLog.ErrorErr("备用方法读取PDF上下文失败", err, logger.F("path", path))
		return "", fmt.Errorf("备用方法读取PDF上下文失败: %w", err)
	}

	var output strings.Builder
	for pageNum := 1; pageNum <= ctx.PageCount; pageNum++ {
		output.WriteString(fmt.Sprintf("[Page %d]\n", pageNum))
	}

	pdfLog.Warn("PDF可能是扫描版或加密文档，无法提取文本内容",
		logger.F("pages", ctx.PageCount),
		logger.F("path", path))

	return output.String(), nil
}

// truncateString 截断字符串用于日志显示
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// parseStructure 解析文档结构
func (p *PDFParser) parseStructure(text string) *DocumentStructure {
	lines := strings.Split(text, "\n")
	sections := make([]Section, 0)
	title := ""

	patterns := []struct {
		prefix string
		level  int
	}{
		{"第一章", 1}, {"第二章", 1}, {"第三章", 1}, {"第四章", 1},
		{"第五章", 1}, {"第六章", 1}, {"第七章", 1}, {"第八章", 1},
		{"第一条", 2}, {"第二条", 2}, {"第三条", 2}, {"第四条", 2},
		{"一、", 2}, {"二、", 2}, {"三、", 2}, {"四、", 2}, {"五、", 2},
		{"（一）", 3}, {"（二）", 3}, {"（三）", 3}, {"（四）", 3},
		{"1.", 4}, {"2.", 4}, {"3.", 4}, {"4.", 4},
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		for _, pattern := range patterns {
			if strings.HasPrefix(line, pattern.prefix) {
				sections = append(sections, Section{
					Title: line,
					Level: pattern.level,
				})
				break
			}
		}

		if title == "" && len(line) > 2 && len(line) < 50 {
			if !strings.ContainsAny(line, "，。、；：\"\"''（）") {
				title = line
			}
		}
	}

	structure := &DocumentStructure{
		Sections: sections,
	}
	if title != "" {
		structure.Title = title
	}

	return structure
}
