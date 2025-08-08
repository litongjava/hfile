package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/litongjava/hfile/model"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func GetRepoName(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", startDir, err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".hfile")); err == nil {
			return filepath.Base(dir), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("not a hfile repository (or any of the parent directories): .hfile not found")
}

func ScanLocalFiles(repoDir string) (map[string]model.FileMeta, error) {
	result := make(map[string]model.FileMeta)

	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(repoDir, path)
		// Skip the ignore file
		if relPath == ".hfileignore" {
			return nil
		}
		if strings.HasPrefix(relPath, ".hfile") {
			return nil
		}

		standardizedRelPath := filepath.ToSlash(relPath)

		hash, err := calculateQuickHash(path)
		if err != nil {
			return err
		}

		result[standardizedRelPath] = model.FileMeta{
			Path:    standardizedRelPath,
			Hash:    hash,
			ModTime: info.ModTime().Unix(),
		}
		return nil
	})

	return result, err
}

func calculateQuickHash(filePath string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}

	fileSize := fileInfo.Size()
	modTime := fileInfo.ModTime().Unix()

	// 创建 SHA256 hasher
	hasher := sha256.New()

	// 更新文件大小信息
	hasher.Write([]byte(fmt.Sprintf("%d", fileSize)))

	// 更新文件修改时间信息
	hasher.Write([]byte(fmt.Sprintf("%d", modTime)))

	// 如果文件较小，计算完整哈希
	if fileSize < 1024*1024 { // 1MB以下计算完整哈希
		file, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer file.Close()

		_, err = io.Copy(hasher, file)
		if err != nil {
			return "", err
		}
	} else {
		// 大文件只取头部和尾部数据
		file, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer file.Close()

		// 读取头部
		headSize := int64(4096)
		if fileSize < headSize {
			headSize = fileSize
		}
		head := make([]byte, headSize)
		_, err = file.Read(head)
		if err != nil && err != io.EOF {
			return "", err
		}
		hasher.Write(head)

		// 读取尾部
		if fileSize > 8192 {
			_, err = file.Seek(fileSize-4096, 0)
			if err != nil {
				return "", err
			}
			tail := make([]byte, 4096)
			_, err = file.Read(tail)
			if err != nil && err != io.EOF {
				return "", err
			}
			hasher.Write(tail)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
