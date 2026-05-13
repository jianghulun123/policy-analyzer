package textutil

import (
	"regexp"
	"strings"
)

var markdownHeadingPattern = regexp.MustCompile(`^\s{0,3}#{1,6}\s*`)

var reasoningTagPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?is)<think[>＞].*?</think[>＞]`),
	regexp.MustCompile(`(?is)<reasoning[>＞].*?</reasoning[>＞]`),
	regexp.MustCompile(`(?is)<think[>＞].*$`),
	regexp.MustCompile(`(?is)<reasoning[>＞].*$`),
	regexp.MustCompile(`(?i)</?think[>＞]`),
	regexp.MustCompile(`(?i)</?reasoning[>＞]`),
}

// SanitizeModelOutput 移除思考模型可能泄漏到最终展示内容中的推理标签与包裹内容。
// 这里做展示层与导出层共用的统一净化，既处理标准 <think> 标签，也兼容部分模型返回的
// HTML 转义标签和全角结束符，避免报告、摘要、导出文本出现思考链原文。
func SanitizeModelOutput(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}

	replacer := strings.NewReplacer(
		"&lt;think&gt;", "<think>",
		"&lt;/think&gt;", "</think>",
		"&lt;reasoning&gt;", "<reasoning>",
		"&lt;/reasoning&gt;", "</reasoning>",
	)
	content = replacer.Replace(content)

	for _, pattern := range reasoningTagPatterns {
		content = pattern.ReplaceAllString(content, "")
	}

	lines := strings.Split(content, "\n")
	cleaned := make([]string, 0, len(lines))
	prevEmpty := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if !prevEmpty {
				cleaned = append(cleaned, "")
			}
			prevEmpty = true
			continue
		}

		cleaned = append(cleaned, line)
		prevEmpty = false
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

// NormalizeSummaryText 将模型摘要规范为可直接展示的正文。
// 它会先做通用清洗，再去掉模型偶尔自行补上的“执行摘要”标题，
// 避免详情页标题和摘要正文中的标题重复出现。
func NormalizeSummaryText(content string) string {
	content = SanitizeModelOutput(content)
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	start := 0
	for start < len(lines) {
		line := strings.TrimSpace(lines[start])
		line = strings.TrimLeft(line, "#")
		line = strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(line, ":"), "："))

		if line == "" {
			start++
			continue
		}
		if line == "执行摘要" {
			start++
			continue
		}
		break
	}

	return strings.TrimSpace(strings.Join(lines[start:], "\n"))
}

// NormalizeSectionBody 清洗章节正文，并去掉与模板重复的标题行。
func NormalizeSectionBody(sectionTitle, content string) string {
	content = SanitizeModelOutput(content)
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	start := 0
	normalizedTitle := normalizeLooseHeading(sectionTitle)

	for start < len(lines) {
		current := normalizeLooseHeading(lines[start])
		if current == "" {
			start++
			continue
		}
		if current == normalizedTitle {
			start++
			continue
		}
		break
	}

	return strings.TrimSpace(strings.Join(lines[start:], "\n"))
}

func normalizeLooseHeading(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = markdownHeadingPattern.ReplaceAllString(value, "")
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(strings.TrimSuffix(value, ":"), "：")
	return strings.TrimSpace(value)
}
