package report

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateMarkdown(t *testing.T) {
	generator := NewGenerator("Policy Analyzer")

	data := ReportData{
		Title:      "政策对比分析报告",
		WorkflowID: "standard-compare",
		Status:     "completed",
		Summary:    "这是测试摘要",
		Content:    "这是测试内容",
		Metrics: Metrics{
			Duration:     120,
			TokensInput:  1000,
			TokensOutput: 2000,
			APICalls:     5,
		},
		CreatedAt: time.Now(),
	}

	output, err := generator.GenerateMarkdown(data)
	if err != nil {
		t.Fatalf("GenerateMarkdown failed: %v", err)
	}

	result := string(output)

	// 验证关键内容
	if !strings.Contains(result, "政策对比分析报告") {
		t.Error("markdown should contain title")
	}
	if !strings.Contains(result, "standard-compare") {
		t.Error("markdown should contain workflow id")
	}
	if !strings.Contains(result, "执行摘要") {
		t.Error("markdown should contain summary section")
	}
	if !strings.Contains(result, "这是测试摘要") {
		t.Error("markdown should contain summary content")
	}
	if !strings.Contains(result, "分析报告") {
		t.Error("markdown should contain report section")
	}
	if !strings.Contains(result, "执行指标") {
		t.Error("markdown should contain metrics section")
	}
	if !strings.Contains(result, "120 秒") {
		t.Error("markdown should contain duration")
	}
}

func TestGenerateMarkdownFiltersThinkTags(t *testing.T) {
	generator := NewGenerator("Policy Analyzer")

	data := ReportData{
		Title:      "政策对比分析报告",
		WorkflowID: "standard-compare",
		Status:     "completed",
		Summary:    "摘要前<think>不该出现</think>摘要后",
		Content:    "正文前\n<think>隐藏推理</think>\n正文后",
		CreatedAt:  time.Now(),
	}

	output, err := generator.GenerateMarkdown(data)
	if err != nil {
		t.Fatalf("GenerateMarkdown failed: %v", err)
	}

	result := string(output)
	if strings.Contains(result, "<think>") || strings.Contains(result, "</think>") {
		t.Fatal("markdown should not contain think tags")
	}
	if strings.Contains(result, "不该出现") || strings.Contains(result, "隐藏推理") {
		t.Fatal("markdown should not contain reasoning content")
	}
	if !strings.Contains(result, "摘要前摘要后") {
		t.Fatal("markdown should keep visible summary content")
	}
	if !strings.Contains(result, "正文前\n\n正文后") {
		t.Fatal("markdown should keep visible report content")
	}
}

func TestGeneratePDF(t *testing.T) {
	generator := NewGenerator("Policy Analyzer")

	data := ReportData{
		Title:      "政策对比分析报告",
		WorkflowID: "standard-compare",
		Status:     "completed",
		Summary:    "这是测试摘要",
		Content:    "这是测试内容",
		Metrics: Metrics{
			Duration:     120,
			TokensInput:  1000,
			TokensOutput: 2000,
			APICalls:     5,
		},
		CreatedAt: time.Now(),
	}

	output, err := generator.GeneratePDF(data)
	if err != nil {
		t.Fatalf("GeneratePDF failed: %v", err)
	}

	// PDF 文件以 %PDF- 开头
	if len(output) < 4 {
		t.Fatal("PDF output too short")
	}
	if string(output[:4]) != "%PDF" {
		t.Error("output should be a valid PDF file")
	}
}

func TestGeneratePDFReturnsBinaryBytes(t *testing.T) {
	generator := NewGenerator("Policy Analyzer")
	output, err := generator.GeneratePDF(ReportData{
		Title:      "政策对比分析报告",
		WorkflowID: "wf",
		Status:     "completed",
		Summary:    "摘要",
		Content:    "正文",
		CreatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("GeneratePDF failed: %v", err)
	}
	if len(output) == 0 {
		t.Fatal("expected non-empty pdf bytes")
	}
	if string(output[:4]) != "%PDF" {
		t.Fatalf("expected pdf header, got %q", string(output[:4]))
	}
}

func TestGenerateMarkdownEmpty(t *testing.T) {
	generator := NewGenerator("Policy Analyzer")

	data := ReportData{
		Title:      "测试报告",
		WorkflowID: "test",
		Status:     "completed",
		CreatedAt:  time.Now(),
	}

	output, err := generator.GenerateMarkdown(data)
	if err != nil {
		t.Fatalf("GenerateMarkdown failed: %v", err)
	}

	// 即使没有内容也应该生成报告
	if len(output) == 0 {
		t.Error("markdown should not be empty")
	}
}

func TestProcessMarkdownForPDF(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			input:  "# 标题",
			expect: "标题",
		},
		{
			input:  "**粗体**",
			expect: "粗体",
		},
		{
			input:  "`代码`",
			expect: "代码",
		},
		{
			input:  "```代码块```",
			expect: "代码块",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := processMarkdownForPDF(tt.input)
			if got != tt.expect {
				t.Errorf("processMarkdownForPDF(%s) = %s, want %s", tt.input, got, tt.expect)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		expect string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 10, "this is..."},
		{"exact", 5, "exact"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.expect {
				t.Errorf("truncateString(%s, %d) = %s, want %s", tt.input, tt.maxLen, got, tt.expect)
			}
		})
	}
}

func TestGenerateMarkdownWithStructuredReport(t *testing.T) {
	generator := NewGenerator("Policy Analyzer")

	data := ReportData{
		Title:      "政策对比分析报告",
		WorkflowID: "standard-compare",
		Status:     "completed",
		Summary:    "后备摘要",
		Content:    "后备正文",
		Structured: &StructuredReport{
			ExecutiveSummary: "结构化摘要",
			HeadlineMetrics: []HeadlineMetric{
				{Title: "经济目标", Value: "4.5-5%", Delta: "+弹性", Insight: "区间目标更务实"},
			},
			ComparisonTable: []ComparisonRow{
				{Dimension: "增速目标", Old: "5%", New: "4.5-5%", Impact: "目标更灵活"},
			},
			TrendPoints: []TrendPoint{
				{Name: "改革", Old: 6, New: 9, Delta: 3},
			},
			Sections: SectionBodies{
				Overview: "总体概述正文",
			},
		},
		CreatedAt: time.Now(),
	}

	output, err := generator.GenerateMarkdown(data)
	if err != nil {
		t.Fatalf("GenerateMarkdown failed: %v", err)
	}

	result := string(output)
	if !strings.Contains(result, "结构化摘要") {
		t.Fatal("markdown should contain structured summary")
	}
	if !strings.Contains(result, "| 指标 | 数值 | 变化 | 说明 |") {
		t.Fatal("markdown should contain metrics table")
	}
	if !strings.Contains(result, "| 维度 | 旧版 | 新版 | 影响 |") {
		t.Fatal("markdown should contain comparison table")
	}
	if !strings.Contains(result, "图表解读") {
		t.Fatal("markdown should contain chart description")
	}
}
