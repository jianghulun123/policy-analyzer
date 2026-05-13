package textutil

import "testing"

func TestSanitizeModelOutput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "standard think block",
			input:  "前言\n<think>这里是思考过程</think>\n结论",
			expect: "前言\n\n结论",
		},
		{
			name:   "html escaped think block",
			input:  "摘要\n&lt;think&gt;内部推理&lt;/think&gt;\n结果",
			expect: "摘要\n\n结果",
		},
		{
			name:   "reasoning block with fullwidth close",
			input:  "A<reasoning＞hidden</reasoning＞B",
			expect: "AB",
		},
		{
			name:   "unterminated think block",
			input:  "可见内容<think>后续都应移除",
			expect: "可见内容",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeModelOutput(tt.input)
			if got != tt.expect {
				t.Fatalf("SanitizeModelOutput() = %q, want %q", got, tt.expect)
			}
		})
	}
}

func TestNormalizeSummaryText(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "plain heading line",
			input:  "执行摘要\n这是正文",
			expect: "这是正文",
		},
		{
			name:   "markdown heading line",
			input:  "## 执行摘要：\n这是正文",
			expect: "这是正文",
		},
		{
			name:   "keep normal content",
			input:  "这是正文",
			expect: "这是正文",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSummaryText(tt.input)
			if got != tt.expect {
				t.Fatalf("NormalizeSummaryText() = %q, want %q", got, tt.expect)
			}
		})
	}
}

func TestNormalizeSectionBody(t *testing.T) {
	tests := []struct {
		name        string
		sectionName string
		input       string
		expect      string
	}{
		{
			name:        "markdown heading duplicated",
			sectionName: "总体变化概览",
			input:       "## 总体变化概览\n这里是正文",
			expect:      "这里是正文",
		},
		{
			name:        "plain heading duplicated",
			sectionName: "影响评估",
			input:       "影响评估：\n这里是正文",
			expect:      "这里是正文",
		},
		{
			name:        "keep plain body",
			sectionName: "关键变化拆解",
			input:       "这里是正文",
			expect:      "这里是正文",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSectionBody(tt.sectionName, tt.input)
			if got != tt.expect {
				t.Fatalf("NormalizeSectionBody() = %q, want %q", got, tt.expect)
			}
		})
	}
}
