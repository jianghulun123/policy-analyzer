package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"policy-analyzer/textutil"
)

// StructuredReport 结构化报告数据，供详情页展示与导出复用。
type StructuredReport struct {
	ExecutiveSummary  string            `json:"executive_summary"`
	HeadlineMetrics   []HeadlineMetric  `json:"headline_metrics"`
	ComparisonTable   []ComparisonRow   `json:"comparison_table"`
	TrendPoints       []TrendPoint      `json:"trend_points"`
	SectionHighlights SectionHighlights `json:"section_highlights"`
	Sections          SectionBodies     `json:"sections"`
	// 五法解读字段
	WordingChanges     []WordingChange     `json:"wording_changes,omitempty"`
	RankingChanges     *RankingChanges     `json:"ranking_changes,omitempty"`
	NewPhrases         []NewPhrase         `json:"new_phrases,omitempty"`
	DisappearedPhrases []DisappearedPhrase `json:"disappeared_phrases,omitempty"`
	FocusShift         *FocusShift         `json:"focus_shift,omitempty"`
}

type HeadlineMetric struct {
	Title   string `json:"title"`
	Value   string `json:"value"`
	Delta   string `json:"delta"`
	Insight string `json:"insight"`
}

type ComparisonRow struct {
	Dimension string `json:"dimension"`
	Old       string `json:"old"`
	New       string `json:"new"`
	Impact    string `json:"impact"`
}

type TrendPoint struct {
	Name  string  `json:"name"`
	Old   float64 `json:"old"`
	New   float64 `json:"new"`
	Delta float64 `json:"delta"`
}

type SectionHighlights struct {
	Overview   string `json:"overview"`
	KeyChanges string `json:"key_changes"`
	Impacts    string `json:"impacts"`
	Actions    string `json:"actions"`
}

type SectionBodies struct {
	Overview   string `json:"overview"`
	KeyChanges string `json:"key_changes"`
	Impacts    string `json:"impacts"`
	Actions    string `json:"actions"`
}

// 五法解读类型定义

type WordingChange struct {
	Topic  string `json:"topic"`
	Old    string `json:"old"`
	New    string `json:"new"`
	Signal string `json:"signal"`
}

type RankingChanges struct {
	Rising      []string `json:"rising,omitempty"`
	Falling     []string `json:"falling,omitempty"`
	NewItems    []string `json:"new,omitempty"`
	Disappeared []string `json:"disappeared,omitempty"`
}

type NewPhrase struct {
	Phrase  string `json:"phrase"`
	Meaning string `json:"meaning"`
	Impact  string `json:"impact"`
}

type DisappearedPhrase struct {
	Phrase  string `json:"phrase"`
	Context string `json:"context"`
	Signal  string `json:"signal"`
}

type FocusShift struct {
	MacroTone          string   `json:"macro_tone,omitempty"`
	PriorityDirections []string `json:"priority_directions,omitempty"`
	RiskAreas          []string `json:"risk_areas,omitempty"`
}

// ParseStructuredReport 将任意输出值解析为结构化报告对象。
func ParseStructuredReport(value interface{}) (*StructuredReport, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case *StructuredReport:
		return NormalizeStructuredReport(v), nil
	case StructuredReport:
		return NormalizeStructuredReport(&v), nil
	case string:
		return parseStructuredReportString(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("序列化结构化报告失败: %w", err)
		}
		var report StructuredReport
		if err := json.Unmarshal(data, &report); err != nil {
			return nil, fmt.Errorf("解析结构化报告失败: %w", err)
		}
		return NormalizeStructuredReport(&report), nil
	}
}

func parseStructuredReportString(content string) (*StructuredReport, error) {
	content = strings.TrimSpace(textutil.SanitizeModelOutput(content))
	if content == "" {
		return nil, nil
	}

	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		content = content[start : end+1]
	}

	var report StructuredReport
	if err := json.Unmarshal([]byte(content), &report); err != nil {
		relaxed, relaxedErr := parseStructuredReportJSONLike(content)
		if relaxedErr != nil {
			return nil, fmt.Errorf("解析结构化报告JSON失败: %w", err)
		}
		return NormalizeStructuredReport(relaxed), nil
	}
	return NormalizeStructuredReport(&report), nil
}

// NormalizeStructuredReport 对结构化报告做统一清洗、字段补齐与标题去重。
func NormalizeStructuredReport(report *StructuredReport) *StructuredReport {
	if report == nil {
		return nil
	}

	report.ExecutiveSummary = textutil.NormalizeSummaryText(report.ExecutiveSummary)

	for idx := range report.HeadlineMetrics {
		report.HeadlineMetrics[idx].Title = textutil.SanitizeModelOutput(report.HeadlineMetrics[idx].Title)
		report.HeadlineMetrics[idx].Value = textutil.SanitizeModelOutput(report.HeadlineMetrics[idx].Value)
		report.HeadlineMetrics[idx].Delta = textutil.SanitizeModelOutput(report.HeadlineMetrics[idx].Delta)
		report.HeadlineMetrics[idx].Insight = textutil.SanitizeModelOutput(report.HeadlineMetrics[idx].Insight)
	}

	for idx := range report.ComparisonTable {
		report.ComparisonTable[idx].Dimension = textutil.SanitizeModelOutput(report.ComparisonTable[idx].Dimension)
		report.ComparisonTable[idx].Old = textutil.SanitizeModelOutput(report.ComparisonTable[idx].Old)
		report.ComparisonTable[idx].New = textutil.SanitizeModelOutput(report.ComparisonTable[idx].New)
		report.ComparisonTable[idx].Impact = textutil.SanitizeModelOutput(report.ComparisonTable[idx].Impact)
	}

	// 五法解读字段清洗
	for idx := range report.WordingChanges {
		report.WordingChanges[idx].Topic = textutil.SanitizeModelOutput(report.WordingChanges[idx].Topic)
		report.WordingChanges[idx].Old = textutil.SanitizeModelOutput(report.WordingChanges[idx].Old)
		report.WordingChanges[idx].New = textutil.SanitizeModelOutput(report.WordingChanges[idx].New)
		report.WordingChanges[idx].Signal = textutil.SanitizeModelOutput(report.WordingChanges[idx].Signal)
	}

	for idx := range report.NewPhrases {
		report.NewPhrases[idx].Phrase = textutil.SanitizeModelOutput(report.NewPhrases[idx].Phrase)
		report.NewPhrases[idx].Meaning = textutil.SanitizeModelOutput(report.NewPhrases[idx].Meaning)
		report.NewPhrases[idx].Impact = textutil.SanitizeModelOutput(report.NewPhrases[idx].Impact)
	}

	for idx := range report.DisappearedPhrases {
		report.DisappearedPhrases[idx].Phrase = textutil.SanitizeModelOutput(report.DisappearedPhrases[idx].Phrase)
		report.DisappearedPhrases[idx].Context = textutil.SanitizeModelOutput(report.DisappearedPhrases[idx].Context)
		report.DisappearedPhrases[idx].Signal = textutil.SanitizeModelOutput(report.DisappearedPhrases[idx].Signal)
	}

	if report.RankingChanges != nil {
		report.RankingChanges.Rising = sanitizeStringSlice(report.RankingChanges.Rising)
		report.RankingChanges.Falling = sanitizeStringSlice(report.RankingChanges.Falling)
		report.RankingChanges.NewItems = sanitizeStringSlice(report.RankingChanges.NewItems)
		report.RankingChanges.Disappeared = sanitizeStringSlice(report.RankingChanges.Disappeared)
	}

	if report.FocusShift != nil {
		report.FocusShift.MacroTone = textutil.SanitizeModelOutput(report.FocusShift.MacroTone)
		report.FocusShift.PriorityDirections = sanitizeStringSlice(report.FocusShift.PriorityDirections)
		report.FocusShift.RiskAreas = sanitizeStringSlice(report.FocusShift.RiskAreas)
	}

	report.SectionHighlights.Overview = textutil.NormalizeSectionBody("总体变化概览", report.SectionHighlights.Overview)
	report.SectionHighlights.KeyChanges = textutil.NormalizeSectionBody("关键变化拆解", report.SectionHighlights.KeyChanges)
	report.SectionHighlights.Impacts = textutil.NormalizeSectionBody("影响评估", report.SectionHighlights.Impacts)
	report.SectionHighlights.Actions = textutil.NormalizeSectionBody("应对建议", report.SectionHighlights.Actions)

	report.Sections.Overview = textutil.NormalizeSectionBody("总体变化概览", report.Sections.Overview)
	report.Sections.KeyChanges = textutil.NormalizeSectionBody("关键变化拆解", report.Sections.KeyChanges)
	report.Sections.Impacts = textutil.NormalizeSectionBody("影响评估", report.Sections.Impacts)
	report.Sections.Actions = textutil.NormalizeSectionBody("应对建议", report.Sections.Actions)

	// 趋势点去重
	report.TrendPoints = deduplicateTrendPoints(report.TrendPoints)

	return report
}

func sanitizeStringSlice(items []string) []string {
	if items == nil {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(textutil.SanitizeModelOutput(item)); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func deduplicateTrendPoints(points []TrendPoint) []TrendPoint {
	if len(points) == 0 {
		return points
	}
	seen := make(map[string]bool)
	result := make([]TrendPoint, 0, len(points))
	for _, p := range points {
		if !seen[p.Name] {
			seen[p.Name] = true
			result = append(result, p)
		}
	}
	return result
}

// RenderStructuredContentMarkdown 将结构化报告渲染为 Markdown 内容（不含摘要）。
func RenderStructuredContentMarkdown(structured *StructuredReport, fallbackContent string) string {
	structured = NormalizeStructuredReport(structured)
	if structured == nil {
		return strings.TrimSpace(textutil.SanitizeModelOutput(fallbackContent))
	}

	var buf bytes.Buffer

	if len(structured.HeadlineMetrics) > 0 {
		buf.WriteString("## 关键指标\n\n")
		buf.WriteString("| 指标 | 数值 | 变化 | 说明 |\n")
		buf.WriteString("| --- | --- | --- | --- |\n")
		for _, metric := range structured.HeadlineMetrics {
			buf.WriteString(fmt.Sprintf(
				"| %s | %s | %s | %s |\n",
				escapeMarkdownTable(metric.Title),
				escapeMarkdownTable(metric.Value),
				escapeMarkdownTable(metric.Delta),
				escapeMarkdownTable(metric.Insight),
			))
		}
		buf.WriteString("\n")
	}

	if len(structured.ComparisonTable) > 0 {
		buf.WriteString("## 关键变化对比\n\n")
		buf.WriteString("| 维度 | 旧版 | 新版 | 影响 |\n")
		buf.WriteString("| --- | --- | --- | --- |\n")
		for _, row := range structured.ComparisonTable {
			buf.WriteString(fmt.Sprintf(
				"| %s | %s | %s | %s |\n",
				escapeMarkdownTable(row.Dimension),
				escapeMarkdownTable(row.Old),
				escapeMarkdownTable(row.New),
				escapeMarkdownTable(row.Impact),
			))
		}
		buf.WriteString("\n")
	}

	if len(structured.TrendPoints) > 0 {
		buf.WriteString("## 政策重心趋势\n\n")
		buf.WriteString("| 维度 | 旧版分值 | 新版分值 | 变化 |\n")
		buf.WriteString("| --- | --- | --- | --- |\n")
		for _, point := range structured.TrendPoints {
			buf.WriteString(fmt.Sprintf(
				"| %s | %s | %s | %s |\n",
				escapeMarkdownTable(point.Name),
				formatTrendValue(point.Old),
				formatTrendValue(point.New),
				formatTrendValue(point.Delta),
			))
		}
		buf.WriteString("\n")

		description := DescribeTrendPoints(structured.TrendPoints)
		if description != "" {
			buf.WriteString("### 图表解读\n\n")
			buf.WriteString(description)
			buf.WriteString("\n\n")
		}
	}

	highlightLines := []struct {
		title   string
		content string
	}{
		{"总体变化概览", structured.SectionHighlights.Overview},
		{"关键变化拆解", structured.SectionHighlights.KeyChanges},
		{"影响评估", structured.SectionHighlights.Impacts},
		{"应对建议", structured.SectionHighlights.Actions},
	}
	if hasNonEmptyContent(highlightLines) {
		buf.WriteString("## 核心结论\n\n")
		for _, item := range highlightLines {
			if strings.TrimSpace(item.content) == "" {
				continue
			}
			buf.WriteString(fmt.Sprintf("- **%s**：%s\n", item.title, item.content))
		}
		buf.WriteString("\n")
	}

	actionsContent := firstNonEmptyContent(structured.SectionHighlights.Actions, structured.Sections.Actions)
	if strings.TrimSpace(actionsContent) != "" {
		buf.WriteString("## 行动建议\n\n")
		buf.WriteString(actionsContent)
		buf.WriteString("\n\n")
	}

	detailSections := []struct {
		title   string
		content string
	}{
		{"总体变化概览", structured.Sections.Overview},
		{"关键变化拆解", structured.Sections.KeyChanges},
		{"影响评估", structured.Sections.Impacts},
	}
	if hasNonEmptyContent(detailSections) {
		buf.WriteString("## 详细解读\n\n")
		for _, item := range detailSections {
			if strings.TrimSpace(item.content) == "" {
				continue
			}
			buf.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", item.title, item.content))
		}
	}

	supplemental := renderSupplementalObservations(structured)
	if supplemental != "" {
		buf.WriteString("## 补充观察\n\n")
		buf.WriteString(supplemental)
		buf.WriteString("\n\n")
	}

	rendered := strings.TrimSpace(buf.String())
	if rendered != "" {
		return rendered
	}

	return strings.TrimSpace(textutil.SanitizeModelOutput(fallbackContent))
}

// RenderStructuredMarkdown 将结构化报告渲染为完整 Markdown，用于文本导出。
func RenderStructuredMarkdown(structured *StructuredReport, fallbackContent string) string {
	structured = NormalizeStructuredReport(structured)
	if structured == nil {
		return strings.TrimSpace(textutil.SanitizeModelOutput(fallbackContent))
	}

	var buf bytes.Buffer
	if structured.ExecutiveSummary != "" {
		buf.WriteString("## 执行摘要\n\n")
		buf.WriteString(structured.ExecutiveSummary)
		buf.WriteString("\n\n")
	}

	content := RenderStructuredContentMarkdown(structured, fallbackContent)
	if content != "" {
		buf.WriteString(content)
	}

	return strings.TrimSpace(buf.String())
}

// DescribeTrendPoints 生成图表数据的文字解读，用于 Markdown/PDF 导出。
func DescribeTrendPoints(points []TrendPoint) string {
	if len(points) == 0 {
		return ""
	}

	sorted := append([]TrendPoint(nil), points...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Delta > sorted[j].Delta
	})

	topRise := sorted[0]
	topDrop := sorted[len(sorted)-1]

	parts := make([]string, 0, 2)
	parts = append(parts, fmt.Sprintf("新版政策中，`%s` 的重视度提升最明显（%s → %s，变化 %s）。",
		topRise.Name, formatTrendValue(topRise.Old), formatTrendValue(topRise.New), formatTrendValue(topRise.Delta)))
	if len(sorted) > 1 && topDrop.Delta < 0 {
		parts = append(parts, fmt.Sprintf("相较之下，`%s` 的权重有所回落（%s → %s，变化 %s）。",
			topDrop.Name, formatTrendValue(topDrop.Old), formatTrendValue(topDrop.New), formatTrendValue(topDrop.Delta)))
	}

	return strings.Join(parts, " ")
}

func escapeMarkdownTable(value string) string {
	value = strings.ReplaceAll(textutil.SanitizeModelOutput(value), "\n", "<br>")
	return strings.ReplaceAll(value, "|", "\\|")
}

func formatTrendValue(value float64) string {
	if value == float64(int(value)) {
		return strconv.Itoa(int(value))
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

func firstNonEmptyContent(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func hasNonEmptyContent(items []struct {
	title   string
	content string
}) bool {
	for _, item := range items {
		if strings.TrimSpace(item.content) != "" {
			return true
		}
	}
	return false
}

func renderSupplementalObservations(structured *StructuredReport) string {
	if structured == nil {
		return ""
	}

	var buf bytes.Buffer

	if len(structured.WordingChanges) > 0 {
		buf.WriteString("### 措辞变化\n\n")
		buf.WriteString("| 议题 | 旧版表述 | 新版表述 | 信号解读 |\n")
		buf.WriteString("| --- | --- | --- | --- |\n")
		for _, row := range structured.WordingChanges {
			buf.WriteString(fmt.Sprintf(
				"| %s | %s | %s | %s |\n",
				escapeMarkdownTable(row.Topic),
				escapeMarkdownTable(row.Old),
				escapeMarkdownTable(row.New),
				escapeMarkdownTable(row.Signal),
			))
		}
		buf.WriteString("\n")
	}

	if structured.RankingChanges != nil && (len(structured.RankingChanges.Rising) > 0 || len(structured.RankingChanges.Falling) > 0 || len(structured.RankingChanges.NewItems) > 0 || len(structured.RankingChanges.Disappeared) > 0) {
		buf.WriteString("### 重点任务排序变化\n\n")
		if len(structured.RankingChanges.Rising) > 0 {
			buf.WriteString(fmt.Sprintf("- **上升**：%s\n", strings.Join(structured.RankingChanges.Rising, "、")))
		}
		if len(structured.RankingChanges.Falling) > 0 {
			buf.WriteString(fmt.Sprintf("- **下降**：%s\n", strings.Join(structured.RankingChanges.Falling, "、")))
		}
		if len(structured.RankingChanges.NewItems) > 0 {
			buf.WriteString(fmt.Sprintf("- **新增**：%s\n", strings.Join(structured.RankingChanges.NewItems, "、")))
		}
		if len(structured.RankingChanges.Disappeared) > 0 {
			buf.WriteString(fmt.Sprintf("- **消失**：%s\n", strings.Join(structured.RankingChanges.Disappeared, "、")))
		}
		buf.WriteString("\n")
	}

	if len(structured.NewPhrases) > 0 {
		buf.WriteString("### 新提法识别\n\n")
		buf.WriteString("| 新提法 | 政策含义 | 潜在影响 |\n")
		buf.WriteString("| --- | --- | --- |\n")
		for _, row := range structured.NewPhrases {
			buf.WriteString(fmt.Sprintf(
				"| %s | %s | %s |\n",
				escapeMarkdownTable(row.Phrase),
				escapeMarkdownTable(row.Meaning),
				escapeMarkdownTable(row.Impact),
			))
		}
		buf.WriteString("\n")
	}

	if len(structured.DisappearedPhrases) > 0 {
		buf.WriteString("### 消失提法识别\n\n")
		buf.WriteString("| 消失提法 | 旧版语境 | 信号解读 |\n")
		buf.WriteString("| --- | --- | --- |\n")
		for _, row := range structured.DisappearedPhrases {
			buf.WriteString(fmt.Sprintf(
				"| %s | %s | %s |\n",
				escapeMarkdownTable(row.Phrase),
				escapeMarkdownTable(row.Context),
				escapeMarkdownTable(row.Signal),
			))
		}
		buf.WriteString("\n")
	}

	if structured.FocusShift != nil {
		buf.WriteString("### 政策重心迁移\n\n")
		if structured.FocusShift.MacroTone != "" {
			buf.WriteString(fmt.Sprintf("- **宏观政策基调**：%s\n", structured.FocusShift.MacroTone))
		}
		if len(structured.FocusShift.PriorityDirections) > 0 {
			buf.WriteString(fmt.Sprintf("- **重点发展方向**：%s\n", strings.Join(structured.FocusShift.PriorityDirections, "、")))
		}
		if len(structured.FocusShift.RiskAreas) > 0 {
			buf.WriteString(fmt.Sprintf("- **风险防范领域**：%s\n", strings.Join(structured.FocusShift.RiskAreas, "、")))
		}
		buf.WriteString("\n")
	}

	return strings.TrimSpace(buf.String())
}
