package report

import (
	"strings"
	"testing"
)

func TestParseStructuredReport(t *testing.T) {
	raw := `{
	  "executive_summary": "执行摘要",
	  "headline_metrics": [{"title":"经济目标","value":"4.5-5%","delta":"+弹性","insight":"更务实"}],
	  "comparison_table": [{"dimension":"增速目标","old":"5%","new":"4.5-5%","impact":"更灵活"}],
	  "trend_points": [{"name":"改革","old":6,"new":9,"delta":3}],
	  "section_highlights": {"overview":"导语","key_changes":"","impacts":"","actions":""},
	  "sections": {"overview":"## 总体变化概览\n正文","key_changes":"","impacts":"","actions":""}
	}`

	report, err := ParseStructuredReport(raw)
	if err != nil {
		t.Fatalf("ParseStructuredReport failed: %v", err)
	}
	if report == nil {
		t.Fatal("expected structured report")
	}
	if report.Sections.Overview != "正文" {
		t.Fatalf("expected deduped overview body, got %q", report.Sections.Overview)
	}
}

func TestParseStructuredReportWithLooseJSON(t *testing.T) {
	raw := `{
	  "executive_summary"："第一段**加粗**内容。\n\n- 要点一\n- 要点二",
	  "headline_metrics":[{"title"："GDP增速目标","value"："4.5%-5%","delta"："区间化","insight"："首次采用区间表述"}],
	  "comparison_table":[{"dimension":"经济增速目标","old":"5%左右","new":"4.5%-5%","impact":"目标表述更灵活"}],
	  "trend_points":[{"name":"科技创新","old":60,"new":82,"delta":22}],
	  "section_highlights":{"overview":"概览亮点","key_changes":"","impacts":"","actions":""},
	  "sections":{"overview":"### 总体变化概览\n正文内容","key_changes":"","impacts":"","actions":""}
	}`

	report, err := ParseStructuredReport(raw)
	if err != nil {
		t.Fatalf("ParseStructuredReport loose json failed: %v", err)
	}
	if report == nil {
		t.Fatal("expected structured report")
	}
	if report.ExecutiveSummary == "" || !strings.Contains(report.ExecutiveSummary, "**加粗**") {
		t.Fatalf("expected executive summary markdown preserved, got %q", report.ExecutiveSummary)
	}
	if got := len(report.HeadlineMetrics); got != 1 {
		t.Fatalf("expected 1 headline metric, got %d", got)
	}
	if report.Sections.Overview != "正文内容" {
		t.Fatalf("expected normalized overview body, got %q", report.Sections.Overview)
	}
}

func TestRenderStructuredMarkdown(t *testing.T) {
	report := &StructuredReport{
		ExecutiveSummary: "结构化摘要",
		HeadlineMetrics: []HeadlineMetric{
			{Title: "经济目标", Value: "4.5-5%", Delta: "+弹性", Insight: "更务实"},
		},
		ComparisonTable: []ComparisonRow{
			{Dimension: "增速目标", Old: "5%", New: "4.5-5%", Impact: "更灵活"},
		},
		TrendPoints: []TrendPoint{
			{Name: "改革", Old: 6, New: 9, Delta: 3},
		},
		Sections: SectionBodies{
			Overview: "正文",
		},
	}

	rendered := RenderStructuredMarkdown(report, "")
	if rendered == "" {
		t.Fatal("expected rendered markdown")
	}
	if !containsAll(rendered, []string{"## 执行摘要", "## 关键指标", "## 关键变化对比", "## 详细解读"}) {
		t.Fatalf("unexpected rendered markdown: %s", rendered)
	}
}

func containsAll(content string, fragments []string) bool {
	for _, fragment := range fragments {
		if !strings.Contains(content, fragment) {
			return false
		}
	}
	return true
}
