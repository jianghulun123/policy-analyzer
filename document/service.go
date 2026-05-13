package document

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"policy-analyzer/logger"
	"policy-analyzer/storage"
)

// log 模块级日志记录器
var docLog = logger.WithModule("document")

// Document 文档实体（供前端使用）
type Document struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OriginalName string   `json:"original_name"`
	Type        string    `json:"type"` // pdf/docx/txt/image
	Size        int64     `json:"size"`
	TextContent string    `json:"text_content"`
	Workspace   string    `json:"workspace"`
	Tags        []string  `json:"tags"`

	// 元数据
	Source       string    `json:"source"`
	PublishDate  *time.Time `json:"publish_date"`
	Issuer       string    `json:"issuer"`
	EffectiveDate *time.Time `json:"effective_date"`

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ImportRequest 导入请求
type ImportRequest struct {
	SourcePath  string            // 源文件路径
	Workspace   string            // 目标工作空间
	Tags        []string          // 初始标签
	Metadata    map[string]string // 额外元数据
}

// ListFilter 列表过滤条件
type ListFilter struct {
	Type      string   // 文件类型过滤
	Tags      []string // 标签过滤
	Search    string   // 搜索关键词
	SortBy    string   // 排序字段
	SortDesc  bool     // 是否降序
	Limit     int      // 限制数量
	Offset    int      // 偏移量
}

// Service 文档服务
type Service struct {
	store   *storage.Manager
	parsers map[string]Parser
}

// NewService 创建文档服务
func NewService(store *storage.Manager) *Service {
	svc := &Service{
		store:   store,
		parsers: make(map[string]Parser),
	}

	// 注册内置解析器
	svc.RegisterParser(NewTXTParser())
	svc.RegisterParser(NewMDParser())
	svc.RegisterParser(NewPDFParser())
	svc.RegisterParser(NewDOCXParser())

	return svc
}

// RegisterParser 注册解析器
func (s *Service) RegisterParser(parser Parser) {
	for _, ext := range parser.SupportedExtensions() {
		s.parsers[ext] = parser
	}
}

// Import 导入文档
func (s *Service) Import(ctx context.Context, req ImportRequest) (*Document, error) {
	docLog.Info("开始导入文档",
		logger.F("source_path", req.SourcePath),
		logger.F("workspace", req.Workspace),
		logger.F("tags", req.Tags))

	// 检查源文件
	info, err := os.Stat(req.SourcePath)
	if err != nil {
		docLog.ErrorErr("源文件不存在", err, logger.F("source_path", req.SourcePath))
		return nil, fmt.Errorf("源文件不存在: %w", err)
	}
	docLog.Info("源文件信息",
		logger.F("size", info.Size()),
		logger.F("name", info.Name()))

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(req.SourcePath))
	docLog.Info("文件类型检测", logger.F("extension", ext))

	// 检查是否支持
	parser, ok := s.parsers[ext]
	if !ok {
		docLog.Error("不支持的文件类型",
			logger.F("extension", ext),
			logger.F("supported", s.getSupportedExtensions()))
		return nil, fmt.Errorf("不支持的文件类型: %s（支持的类型: %s）", ext, s.getSupportedExtensions())
	}

	// 计算文件哈希
	hash, err := s.calculateHash(req.SourcePath)
	if err != nil {
		docLog.ErrorErr("计算文件哈希失败", err, logger.F("source_path", req.SourcePath))
		return nil, fmt.Errorf("计算哈希失败: %w", err)
	}
	docLog.Info("文件哈希计算完成", logger.F("hash", hash))

	// 解析文档
	docLog.Info("开始解析文档", logger.F("source_path", req.SourcePath))
	result, err := parser.Parse(ctx, req.SourcePath)
	if err != nil {
		docLog.ErrorErr("文档解析失败", err,
			logger.F("source_path", req.SourcePath),
			logger.F("extension", ext))
		return nil, fmt.Errorf("文档解析失败: %w", err)
	}
	docLog.Info("文档解析完成",
		logger.F("text_length", len(result.Text)),
		logger.F("pages", result.Pages))

	// 创建存储记录
	record := &storage.DocumentRecord{
		ID:           storage.NewDocumentID(),
		Name:         filepath.Base(req.SourcePath),
		OriginalName: filepath.Base(req.SourcePath),
		FileType:     ext,
		FileSize:     info.Size(),
		ContentHash:  hash,
		TextContent:  result.Text,
		Workspace:    req.Workspace,
	}
	docLog.Info("创建文档记录", logger.F("id", record.ID))

	// 复制文件到工作空间
	docDir := s.store.GetDocumentsDir(req.Workspace)
	docFileName := fmt.Sprintf("%s_%s", record.ID, filepath.Base(req.SourcePath))
	destPath := filepath.Join(docDir, docFileName)

	docLog.Info("复制文件到工作空间",
		logger.F("dest_path", destPath),
		logger.F("doc_dir", docDir))

	if err := s.copyFile(req.SourcePath, destPath); err != nil {
		docLog.ErrorErr("复制文件失败", err,
			logger.F("source", req.SourcePath),
			logger.F("dest", destPath))
		return nil, fmt.Errorf("复制文件失败: %w", err)
	}
	record.FilePath = destPath

	// 设置元数据
	if req.Metadata != nil {
		if source, ok := req.Metadata["source"]; ok {
			record.Source = source
		}
		if issuer, ok := req.Metadata["issuer"]; ok {
			record.Issuer = issuer
		}
	}

	// 保存到数据库
	if err := s.store.SaveDocument(record); err != nil {
		docLog.ErrorErr("保存文档记录失败", err, logger.F("id", record.ID))
		return nil, fmt.Errorf("保存文档记录失败: %w", err)
	}
	docLog.Info("文档记录已保存到数据库", logger.F("id", record.ID))

	// 保存标签
	if len(req.Tags) > 0 {
		s.store.SaveDocumentTags(record.ID, req.Tags)
		docLog.Info("文档标签已保存", logger.F("id", record.ID), logger.F("tags", req.Tags))
	}

	// 构建返回文档
	doc := s.recordToDocument(record)
	doc.Tags = req.Tags

	docLog.Info("文档导入完成",
		logger.F("id", doc.ID),
		logger.F("name", doc.Name),
		logger.F("size", doc.Size))

	return &doc, nil
}

// List 获取文档列表
func (s *Service) List(ctx context.Context, workspaceID string) ([]Document, error) {
	records, err := s.store.ListDocuments(workspaceID)
	if err != nil {
		return nil, err
	}

	docs := make([]Document, len(records))
	for i, record := range records {
		tags, _ := s.store.GetDocumentTags(record.ID)
		docs[i] = s.recordToDocument(&record)
		docs[i].Tags = tags
	}

	return docs, nil
}

// Get 获取单个文档详情
func (s *Service) Get(ctx context.Context, docID string) (*Document, error) {
	record, err := s.store.GetDocument(docID)
	if err != nil {
		return nil, err
	}

	tags, _ := s.store.GetDocumentTags(record.ID)
	doc := s.recordToDocument(record)
	doc.Tags = tags

	return &doc, nil
}

// GetContent 获取文档内容
func (s *Service) GetContent(ctx context.Context, docID string) (string, error) {
	record, err := s.store.GetDocument(docID)
	if err != nil {
		return "", err
	}

	// 如果已有文本内容，直接返回
	if record.TextContent != "" {
		return record.TextContent, nil
	}

	// 否则重新解析
	if record.FilePath != "" {
		ext := record.FileType
		parser, ok := s.parsers[ext]
		if ok {
			result, err := parser.Parse(ctx, record.FilePath)
			if err != nil {
				return "", err
			}
			// 更新文本内容
			record.TextContent = result.Text
			s.store.UpdateDocument(record)
			return result.Text, nil
		}
	}

	return "", fmt.Errorf("无法获取文档内容")
}

// Delete 删除文档
func (s *Service) Delete(ctx context.Context, docID string) error {
	return s.store.DeleteDocument(docID)
}

// UpdateMeta 更新文档元数据
func (s *Service) UpdateMeta(ctx context.Context, docID string, meta map[string]interface{}) error {
	record, err := s.store.GetDocument(docID)
	if err != nil {
		return err
	}

	// 更新元数据字段
	if source, ok := meta["source"].(string); ok {
		record.Source = source
	}
	if issuer, ok := meta["issuer"].(string); ok {
		record.Issuer = issuer
	}

	return s.store.UpdateDocument(record)
}

// Search 搜索文档
func (s *Service) Search(ctx context.Context, workspaceID string, query string) ([]Document, error) {
	// 简化实现：在文本内容中搜索
	records, err := s.store.ListDocuments(workspaceID)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	docs := make([]Document, 0)

	for _, record := range records {
		// 搜索名称和内容
		if strings.Contains(strings.ToLower(record.Name), query) ||
			strings.Contains(strings.ToLower(record.TextContent), query) {
			tags, _ := s.store.GetDocumentTags(record.ID)
			doc := s.recordToDocument(&record)
			doc.Tags = tags
			docs = append(docs, doc)
		}
	}

	return docs, nil
}

// ========== 内部方法 ==========

// getSupportedExtensions 获取支持的文件扩展名列表
func (s *Service) getSupportedExtensions() string {
	extensions := make([]string, 0, len(s.parsers))
	for ext := range s.parsers {
		extensions = append(extensions, ext)
	}
	return strings.Join(extensions, ", ")
}

func (s *Service) calculateHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *Service) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *Service) recordToDocument(record *storage.DocumentRecord) Document {
	return Document{
		ID:           record.ID,
		Name:         record.Name,
		OriginalName: record.OriginalName,
		Type:         record.FileType,
		Size:         record.FileSize,
		TextContent:  record.TextContent,
		Workspace:    record.Workspace,
		Source:       record.Source,
		PublishDate:  record.PublishDate,
		Issuer:       record.Issuer,
		EffectiveDate: record.EffectiveDate,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}
}