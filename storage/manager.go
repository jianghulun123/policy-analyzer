package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"policy-analyzer/logger"
)

// storageLog 模块级日志记录器
var storageLog = logger.WithModule("storage")

// Workspace 工作空间实体
type Workspace struct {
	ID          string    `gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// DocumentRecord 文档记录（数据库实体）
type DocumentRecord struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	OriginalName string    `json:"original_name"`
	FilePath     string    `json:"file_path"`
	FileType     string    `json:"file_type"` // pdf/docx/txt/image
	FileSize     int64     `json:"file_size"`
	ContentHash  string    `json:"content_hash"`
	TextContent  string    `json:"text_content"`
	Workspace    string    `gorm:"index" json:"workspace"`

	// 元数据
	Source       string    `json:"source"`       // 来源
	PublishDate  *time.Time `json:"publish_date"` // 发布日期
	Issuer       string    `json:"issuer"`       // 发布机构
	EffectiveDate *time.Time `json:"effective_date"` // 生效日期

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DocumentTag 文档标签
type DocumentTag struct {
	DocumentID string `gorm:"primaryKey;autoIncrement:false"`
	Tag        string `gorm:"primaryKey;autoIncrement:false"`
}

// AnalysisRecord 分析记录
type AnalysisRecord struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	WorkflowID   string    `json:"workflow_id"`
	Workspace    string    `gorm:"index" json:"workspace"`
	Status       string    `json:"status"` // pending/running/completed/failed

	DocumentIDs  string    `json:"document_ids"` // JSON数组
	Config       string    `json:"config"`       // JSON

	ResultPath   string    `json:"result_path"`
	Summary      string    `json:"summary"`

	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// Manager 存储管理器
type Manager struct {
	baseDir string
	db      *gorm.DB
}

// NewManager 创建存储管理器
func NewManager(baseDir string) *Manager {
	return &Manager{baseDir: baseDir}
}

// Init 初始化存储
func (m *Manager) Init() error {
	storageLog.Info("开始初始化存储层", logger.F("base_dir", m.baseDir))

	// 创建目录结构
	dirs := []string{
		m.baseDir,
		filepath.Join(m.baseDir, "workspaces"),
		filepath.Join(m.baseDir, "config"),
		filepath.Join(m.baseDir, "templates"),
		filepath.Join(m.baseDir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			storageLog.ErrorErr("创建目录失败", err, logger.F("dir", dir))
			return fmt.Errorf("创建目录失败: %w", err)
		}
		storageLog.Debug("目录已创建", logger.F("dir", dir))
	}

	// 初始化数据库
	dbPath := filepath.Join(m.baseDir, "database.sqlite")
	storageLog.Info("连接数据库", logger.F("db_path", dbPath))

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		storageLog.ErrorErr("数据库连接失败", err, logger.F("db_path", dbPath))
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	m.db = db
	storageLog.Info("数据库连接成功")

	// 自动迁移表结构
	if err := m.db.AutoMigrate(
		&Workspace{},
		&DocumentRecord{},
		&DocumentTag{},
		&AnalysisRecord{},
	); err != nil {
		storageLog.ErrorErr("数据库迁移失败", err)
		return fmt.Errorf("数据库迁移失败: %w", err)
	}
	storageLog.Info("数据库迁移完成")

	// 创建默认工作空间
	if err := m.ensureDefaultWorkspace(); err != nil {
		return err
	}

	storageLog.Info("存储层初始化完成")
	return nil
}

// ensureDefaultWorkspace 确保存在默认工作空间
func (m *Manager) ensureDefaultWorkspace() error {
	var count int64
	m.db.Model(&Workspace{}).Count(&count)
	if count == 0 {
		defaultWS := &Workspace{
			ID:          "default",
			Name:        "默认工作空间",
			Description: "系统默认工作空间",
			CreatedAt:   time.Now(),
		}
		if err := m.db.Create(defaultWS).Error; err != nil {
			return err
		}
	}

	// 无论默认工作空间记录是否已存在，都确保目录结构存在。
	wsDir := filepath.Join(m.baseDir, "workspaces", "default")
	subDirs := []string{
		wsDir,
		filepath.Join(wsDir, "documents"),
		filepath.Join(wsDir, "analysis"),
		filepath.Join(wsDir, "cache"),
	}
	for _, dir := range subDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			storageLog.ErrorErr("创建默认工作空间目录失败", err, logger.F("dir", dir))
			return err
		}
	}
	storageLog.Info("默认工作空间已就绪")

	return nil
}

// Close 关闭存储
func (m *Manager) Close() error {
	if m.db != nil {
		sqlDB, _ := m.db.DB()
		return sqlDB.Close()
	}
	return nil
}

// ========== 工作空间操作 ==========

// ListWorkspaces 获取工作空间列表
func (m *Manager) ListWorkspaces() ([]Workspace, error) {
	var workspaces []Workspace
	err := m.db.Find(&workspaces).Error
	return workspaces, err
}

// CreateWorkspace 创建工作空间
func (m *Manager) CreateWorkspace(ws *Workspace) error {
	// 创建工作空间目录
	wsDir := filepath.Join(m.baseDir, "workspaces", ws.ID)
	subDirs := []string{
		wsDir,
		filepath.Join(wsDir, "documents"),
		filepath.Join(wsDir, "analysis"),
		filepath.Join(wsDir, "cache"),
	}

	for _, dir := range subDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	ws.CreatedAt = time.Now()
	return m.db.Create(ws).Error
}

// DeleteWorkspace 删除工作空间
func (m *Manager) DeleteWorkspace(id string) error {
	// 不允许删除默认工作空间
	if id == "default" {
		return fmt.Errorf("不能删除默认工作空间")
	}

	// 删除目录
	wsDir := filepath.Join(m.baseDir, "workspaces", id)
	if err := os.RemoveAll(wsDir); err != nil {
		return err
	}

	return m.db.Delete(&Workspace{}, "id = ?", id).Error
}

// ========== 文档操作 ==========

// NewDocumentID 生成文档ID
func NewDocumentID() string {
	return uuid.New().String()
}

// SaveDocument 保存文档记录
func (m *Manager) SaveDocument(doc *DocumentRecord) error {
	if doc.ID == "" {
		doc.ID = NewDocumentID()
	}
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	// 确保文档目录存在
	docDir := m.GetDocumentsDir(doc.Workspace)
	if err := os.MkdirAll(docDir, 0755); err != nil {
		storageLog.ErrorErr("创建文档目录失败", err, logger.F("dir", docDir))
		return fmt.Errorf("创建文档目录失败: %w", err)
	}

	storageLog.Info("保存文档记录",
		logger.F("id", doc.ID),
		logger.F("name", doc.Name),
		logger.F("workspace", doc.Workspace))

	return m.db.Create(doc).Error
}

// GetDocument 获取文档
func (m *Manager) GetDocument(id string) (*DocumentRecord, error) {
	var doc DocumentRecord
	err := m.db.First(&doc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// ListDocuments 获取文档列表
func (m *Manager) ListDocuments(workspaceID string) ([]DocumentRecord, error) {
	var docs []DocumentRecord
	err := m.db.Where("workspace = ?", workspaceID).Order("created_at DESC").Find(&docs).Error
	return docs, err
}

// UpdateDocument 更新文档
func (m *Manager) UpdateDocument(doc *DocumentRecord) error {
	doc.UpdatedAt = time.Now()
	return m.db.Save(doc).Error
}

// DeleteDocument 删除文档
func (m *Manager) DeleteDocument(id string) error {
	doc, err := m.GetDocument(id)
	if err != nil {
		return err
	}

	// 删除文件
	if doc.FilePath != "" {
		os.Remove(doc.FilePath)
	}

	// 删除标签
	m.db.Delete(&DocumentTag{}, "document_id = ?", id)

	// 删除记录
	return m.db.Delete(&DocumentRecord{}, "id = ?", id).Error
}

// SaveDocumentTags 保存文档标签
func (m *Manager) SaveDocumentTags(docID string, tags []string) error {
	// 先删除旧标签
	m.db.Delete(&DocumentTag{}, "document_id = ?", docID)

	// 保存新标签
	for _, tag := range tags {
		dt := &DocumentTag{DocumentID: docID, Tag: tag}
		if err := m.db.Create(dt).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetDocumentTags 获取文档标签
func (m *Manager) GetDocumentTags(docID string) ([]string, error) {
	var tags []DocumentTag
	err := m.db.Where("document_id = ?", docID).Find(&tags).Error
	if err != nil {
		return nil, err
	}

	result := make([]string, len(tags))
	for i, t := range tags {
		result[i] = t.Tag
	}
	return result, nil
}

// ========== 分析记录操作 ==========

// SaveAnalysisRecord 保存分析记录
func (m *Manager) SaveAnalysisRecord(record *AnalysisRecord) error {
	if strings.TrimSpace(record.ID) == "" {
		record.ID = uuid.New().String()
	}
	record.CreatedAt = time.Now()
	return m.db.Create(record).Error
}

// UpdateAnalysisRecord 更新分析记录
func (m *Manager) UpdateAnalysisRecord(record *AnalysisRecord) error {
	return m.db.Save(record).Error
}

// GetAnalysisRecord 获取分析记录
func (m *Manager) GetAnalysisRecord(id string) (*AnalysisRecord, error) {
	var record AnalysisRecord
	err := m.db.First(&record, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListAnalysisRecords 获取分析记录列表
func (m *Manager) ListAnalysisRecords(workspaceID string) ([]AnalysisRecord, error) {
	var records []AnalysisRecord
	err := m.db.Where("workspace = ?", workspaceID).Order("created_at DESC").Find(&records).Error
	return records, err
}

// DeleteAnalysisRecord 删除分析记录
func (m *Manager) DeleteAnalysisRecord(id string) error {
	// 先获取记录以删除关联文件
	record, err := m.GetAnalysisRecord(id)
	if err != nil {
		return err
	}

	// 删除结果文件
	if record.ResultPath != "" && m.FileExists(record.ResultPath) {
		if err := os.Remove(record.ResultPath); err != nil {
			storageLog.Warn("删除分析结果文件失败", logger.F("path", record.ResultPath), logger.F("error", err.Error()))
		}
	}

	// 删除数据库记录
	return m.db.Delete(&AnalysisRecord{}, "id = ?", id).Error
}

// ========== 文件存储 ==========

// GetWorkspaceDir 获取工作空间目录
func (m *Manager) GetWorkspaceDir(workspaceID string) string {
	return filepath.Join(m.baseDir, "workspaces", workspaceID)
}

// GetDocumentsDir 获取文档目录
func (m *Manager) GetDocumentsDir(workspaceID string) string {
	return filepath.Join(m.baseDir, "workspaces", workspaceID, "documents")
}

// GetAnalysisDir 获取分析结果目录
func (m *Manager) GetAnalysisDir(workspaceID string) string {
	return filepath.Join(m.baseDir, "workspaces", workspaceID, "analysis")
}

// SaveFile 保存文件到指定目录
func (m *Manager) SaveFile(dir string, filename string, data []byte) (string, error) {
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}
	return filePath, nil
}

// ReadFile 读取文件
func (m *Manager) ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// FileExists 检查文件是否存在
func (m *Manager) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}