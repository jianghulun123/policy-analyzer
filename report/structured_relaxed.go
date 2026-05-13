package report

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// parseStructuredReportJSONLike 对模型返回的“类 JSON”内容做宽松解析。
// 这里重点兼容全角标点、弯引号、数组括号误用等常见问题，尽量把结构化报告救回来。
func parseStructuredReportJSONLike(content string) (*StructuredReport, error) {
	parser := &jsonLikeParser{input: []rune(strings.TrimSpace(content))}
	value, err := parser.parseValue()
	if err != nil {
		return nil, err
	}

	root, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("结构化报告根节点不是对象")
	}

	data, err := json.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("宽松解析结果序列化失败: %w", err)
	}

	var report StructuredReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("宽松解析结果反序列化失败: %w", err)
	}

	return &report, nil
}

type jsonLikeParser struct {
	input []rune
	pos   int
}

func (p *jsonLikeParser) parseValue() (interface{}, error) {
	p.skipWhitespace()
	if p.eof() {
		return nil, fmt.Errorf("内容为空")
	}

	switch ch := p.peek(); {
	case isObjectStart(ch):
		return p.parseObject()
	case isArrayStart(ch):
		return p.parseArray()
	case isQuoteLike(ch):
		return p.parseString()
	default:
		return p.parseBareToken(), nil
	}
}

func (p *jsonLikeParser) parseObject() (map[string]interface{}, error) {
	if !isObjectStart(p.peek()) {
		return nil, fmt.Errorf("对象必须以 '{' 开始")
	}
	p.pos++

	result := make(map[string]interface{})
	for !p.eof() {
		p.skipWhitespace()
		if p.eof() {
			break
		}
		if isObjectEnd(p.peek()) {
			p.pos++
			return result, nil
		}

		key, err := p.parseObjectKey()
		if err != nil {
			return nil, err
		}
		if key == "" {
			return nil, fmt.Errorf("对象键为空")
		}

		p.skipWhitespace()
		if p.eof() || !isColonLike(p.peek()) {
			return nil, fmt.Errorf("对象键 %q 后缺少分隔符", key)
		}
		p.pos++

		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		result[key] = value

		p.skipWhitespace()
		if p.eof() {
			break
		}
		if isCommaLike(p.peek()) {
			p.pos++
			continue
		}
		if isObjectEnd(p.peek()) {
			p.pos++
			return result, nil
		}
	}

	return result, nil
}

func (p *jsonLikeParser) parseArray() ([]interface{}, error) {
	if !isArrayStart(p.peek()) {
		return nil, fmt.Errorf("数组必须以 '[' 开始")
	}
	p.pos++

	result := make([]interface{}, 0)
	for !p.eof() {
		p.skipWhitespace()
		if p.eof() {
			break
		}
		if isArrayEnd(p.peek()) {
			p.pos++
			return result, nil
		}

		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		result = append(result, value)

		p.skipWhitespace()
		if p.eof() {
			break
		}
		if isCommaLike(p.peek()) {
			p.pos++
			continue
		}
		if isArrayEnd(p.peek()) {
			p.pos++
			return result, nil
		}
	}

	return result, nil
}

func (p *jsonLikeParser) parseObjectKey() (string, error) {
	p.skipWhitespace()
	if p.eof() {
		return "", fmt.Errorf("对象键缺失")
	}

	if isQuoteLike(p.peek()) {
		value, err := p.parseString()
		if err != nil {
			return "", err
		}
		return normalizeLooseJSONKey(value), nil
	}

	start := p.pos
	for !p.eof() {
		ch := p.peek()
		if isColonLike(ch) || isObjectEnd(ch) || isCommaLike(ch) {
			break
		}
		p.pos++
	}
	return normalizeLooseJSONKey(string(p.input[start:p.pos])), nil
}

func (p *jsonLikeParser) parseString() (string, error) {
	if p.eof() || !isQuoteLike(p.peek()) {
		return "", fmt.Errorf("字符串缺少起始引号")
	}
	p.pos++

	var builder strings.Builder
	for !p.eof() {
		ch := p.peek()
		p.pos++

		if ch == '\\' && !p.eof() {
			next := p.peek()
			p.pos++
			switch next {
			case 'n':
				builder.WriteByte('\n')
			case 'r':
				builder.WriteByte('\r')
			case 't':
				builder.WriteByte('\t')
			default:
				builder.WriteRune(next)
			}
			continue
		}

		// 模型经常在字符串内部混用中英文引号，这里只有在“后面跟着结构分隔符”时才判定为收尾。
		if isQuoteLike(ch) && p.shouldCloseString() {
			return strings.TrimSpace(builder.String()), nil
		}

		builder.WriteRune(ch)
	}

	return strings.TrimSpace(builder.String()), nil
}

func (p *jsonLikeParser) parseBareToken() interface{} {
	start := p.pos
	for !p.eof() {
		ch := p.peek()
		if isCommaLike(ch) || isObjectEnd(ch) || isArrayEnd(ch) {
			break
		}
		p.pos++
	}

	token := strings.TrimSpace(string(p.input[start:p.pos]))
	token = normalizeLooseJSONKey(token)
	if token == "" {
		return ""
	}

	switch strings.ToLower(token) {
	case "null":
		return nil
	case "true":
		return true
	case "false":
		return false
	}

	if number, err := strconv.ParseFloat(token, 64); err == nil {
		return number
	}

	return token
}

func (p *jsonLikeParser) shouldCloseString() bool {
	next := p.nextNonSpace()
	if next == 0 {
		return true
	}
	return isCommaLike(next) || isColonLike(next) || isObjectEnd(next) || isArrayEnd(next)
}

func (p *jsonLikeParser) nextNonSpace() rune {
	for i := p.pos; i < len(p.input); i++ {
		if !unicode.IsSpace(p.input[i]) {
			return p.input[i]
		}
	}
	return 0
}

func (p *jsonLikeParser) skipWhitespace() {
	for !p.eof() && unicode.IsSpace(p.peek()) {
		p.pos++
	}
}

func (p *jsonLikeParser) peek() rune {
	return p.input[p.pos]
}

func (p *jsonLikeParser) eof() bool {
	return p.pos >= len(p.input)
}

func normalizeLooseJSONKey(value string) string {
	return strings.Trim(strings.TrimSpace(value), "\"'“”‘’")
}

func isQuoteLike(ch rune) bool {
	switch ch {
	case '"', '\'', '“', '”', '‘', '’':
		return true
	default:
		return false
	}
}

func isColonLike(ch rune) bool {
	return ch == ':' || ch == '：'
}

func isCommaLike(ch rune) bool {
	return ch == ',' || ch == '，'
}

func isObjectStart(ch rune) bool {
	return ch == '{' || ch == '｛'
}

func isObjectEnd(ch rune) bool {
	return ch == '}' || ch == '｝'
}

func isArrayStart(ch rune) bool {
	return ch == '[' || ch == '【'
}

func isArrayEnd(ch rune) bool {
	return ch == ']' || ch == '】'
}
