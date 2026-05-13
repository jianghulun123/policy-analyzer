package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"policy-analyzer/logger"
)

const defaultLogRetentionDays = 7

func (a *App) getLogDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".policy-analyzer", "logs")
}

func (a *App) GetLogDirectory() (string, error) {
	logDir := a.getLogDir()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", err
	}
	return logDir, nil
}

func (a *App) ExportLogs() (string, error) {
	logDir, err := a.GetLogDirectory()
	if err != nil {
		return "", err
	}

	logFiles, err := listLogFiles(logDir)
	if err != nil {
		return "", err
	}
	if len(logFiles) == 0 {
		return "", fmt.Errorf("暂无日志文件可导出")
	}

	defaultName := fmt.Sprintf("policy-analyzer-logs-%s.zip", time.Now().Format("20060102-150405"))
	outputPath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "导出日志文件",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "ZIP 压缩包",
				Pattern:     "*.zip",
			},
		},
	})
	if err != nil {
		return "", err
	}
	if outputPath == "" {
		return "", nil
	}
	if !strings.HasSuffix(strings.ToLower(outputPath), ".zip") {
		outputPath += ".zip"
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	for _, logFile := range logFiles {
		source, err := os.Open(logFile)
		if err != nil {
			zipWriter.Close()
			return "", err
		}

		entry, err := zipWriter.Create(filepath.Base(logFile))
		if err != nil {
			source.Close()
			zipWriter.Close()
			return "", err
		}

		if _, err := io.Copy(entry, source); err != nil {
			source.Close()
			zipWriter.Close()
			return "", err
		}
		source.Close()
	}
	if err := zipWriter.Close(); err != nil {
		return "", err
	}

	log.Info("日志导出完成",
		logger.F("path", outputPath),
		logger.F("file_count", len(logFiles)))

	return outputPath, nil
}

func (a *App) DeleteHistoricalLogs() (int, error) {
	logDir, err := a.GetLogDirectory()
	if err != nil {
		return 0, err
	}

	deletedCount, err := deleteHistoricalLogs(logDir)
	if err != nil {
		return deletedCount, err
	}

	log.Info("历史日志删除完成",
		logger.F("log_dir", logDir),
		logger.F("deleted_count", deletedCount))

	return deletedCount, nil
}

func (a *App) CleanupLogs(retentionDays int) (int, error) {
	logDir, err := a.GetLogDirectory()
	if err != nil {
		return 0, err
	}

	deletedCount, err := cleanupLogs(logDir, retentionDays)
	if err != nil {
		return deletedCount, err
	}

	log.Info("日志清理完成",
		logger.F("log_dir", logDir),
		logger.F("retention_days", normalizeRetentionDays(retentionDays)),
		logger.F("deleted_count", deletedCount))

	return deletedCount, nil
}

func (a *App) applyConfiguredLogCleanup() {
	if a.config == nil || !a.config.Storage.LogAutoCleanup {
		return
	}

	deletedCount, err := cleanupLogs(a.getLogDir(), a.config.Storage.LogRetentionDays)
	if err != nil {
		log.Warn("自动清理日志失败", logger.F("error", err.Error()))
		return
	}

	log.Info("自动清理日志完成",
		logger.F("retention_days", normalizeRetentionDays(a.config.Storage.LogRetentionDays)),
		logger.F("deleted_count", deletedCount))
}

func listLogFiles(logDir string) ([]string, error) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "app-") || !strings.HasSuffix(strings.ToLower(name), ".log") {
			continue
		}
		files = append(files, filepath.Join(logDir, name))
	}

	sort.Strings(files)
	return files, nil
}

func deleteHistoricalLogs(logDir string) (int, error) {
	logFiles, err := listLogFiles(logDir)
	if err != nil {
		return 0, err
	}

	todayName := currentLogFileName()
	deletedCount := 0
	for _, logFile := range logFiles {
		if filepath.Base(logFile) == todayName {
			continue
		}
		if err := os.Remove(logFile); err != nil {
			return deletedCount, err
		}
		deletedCount++
	}

	return deletedCount, nil
}

func cleanupLogs(logDir string, retentionDays int) (int, error) {
	logFiles, err := listLogFiles(logDir)
	if err != nil {
		return 0, err
	}

	retentionDays = normalizeRetentionDays(retentionDays)
	todayName := currentLogFileName()
	cutoff := time.Now().AddDate(0, 0, -(retentionDays - 1))
	cutoffDate := time.Date(cutoff.Year(), cutoff.Month(), cutoff.Day(), 0, 0, 0, 0, cutoff.Location())

	deletedCount := 0
	for _, logFile := range logFiles {
		name := filepath.Base(logFile)
		if name == todayName {
			continue
		}

		logDate, ok := parseLogDate(name)
		if !ok {
			continue
		}
		if logDate.Before(cutoffDate) {
			if err := os.Remove(logFile); err != nil {
				return deletedCount, err
			}
			deletedCount++
		}
	}

	return deletedCount, nil
}

func normalizeRetentionDays(days int) int {
	if days <= 0 {
		return defaultLogRetentionDays
	}
	return days
}

func currentLogFileName() string {
	return fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02"))
}

func parseLogDate(filename string) (time.Time, bool) {
	datePart := strings.TrimSuffix(strings.TrimPrefix(filename, "app-"), ".log")
	parsed, err := time.Parse("2006-01-02", datePart)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}
