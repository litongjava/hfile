package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/hfile/model"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

// 支持上传ID的分片上传
func UploadInChunks(serverURL, token, repo, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := fileInfo.Size()
	totalParts := int((fileSize + ChunkSize - 1) / ChunkSize)
	modTime := fileInfo.ModTime().Unix()

	hlog.Infof("Start chunk upload: file=%s, size=%d, chunks=%d", filePath, fileSize, totalParts)

	// 1. 初始化分片上传，获取upload_id
	uploadID, err := initChunkedUpload(serverURL, token, repo, filePath, fileSize, totalParts, modTime)
	if err != nil {
		return fmt.Errorf("failed to init chunked upload: %w", err)
	}

	// 2. 逐个上传分片
	for partIndex := 0; partIndex < totalParts; partIndex++ {
		start := int64(partIndex) * ChunkSize
		end := start + ChunkSize
		if end > fileSize {
			end = fileSize
		}

		chunk := make([]byte, end-start)
		_, err := file.ReadAt(chunk, start)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read chunk %d: %w", partIndex, err)
		}

		err = uploadChunk(serverURL, token, repo, uploadID, partIndex, chunk, filePath)
		if err != nil {
			return fmt.Errorf("failed to upload chunk %d: %w", partIndex, err)
		}

		hlog.Infof("Chunk %d/%d uploaded successfully", partIndex+1, totalParts)
	}

	// 3. 完成分片上传
	err = completeChunkedUpload(serverURL, token, repo, uploadID)
	if err != nil {
		return fmt.Errorf("failed to complete chunked upload: %w", err)
	}

	hlog.Infof("All chunks uploaded and merged successfully for file: %s", filePath)
	return nil
}

// 初始化分片上传
func initChunkedUpload(serverURL, token, repo, fileName string, fileSize int64, totalParts int, modTime int64) (string, error) {
	url := fmt.Sprintf("%s/file/upload/init?repo=%s", serverURL, repo)

	reqBody := map[string]interface{}{
		"repo":              repo,
		"file_name":         fileName,
		"file_size":         fileSize,
		"total_parts":       totalParts,
		"original_mod_time": modTime,
	}

	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("init failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp model.APIResponse
	json.Unmarshal(body, &apiResp)

	if !apiResp.Ok {
		return "", fmt.Errorf("init failed: %s", *apiResp.Msg)
	}

	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid init response format")
	}

	uploadID, ok := data["upload_id"].(string)
	if !ok {
		return "", fmt.Errorf("upload_id not found in response")
	}

	return uploadID, nil
}

// 上传单个分片
func uploadChunk(serverURL, token, repo, uploadID string, partIndex int, chunk []byte, fileName string) error {
	url := fmt.Sprintf("%s/file/upload/chunk?repo=%s", serverURL, repo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}
	_, err = part.Write(chunk)
	if err != nil {
		return err
	}

	_ = writer.WriteField("upload_id", uploadID)
	_ = writer.WriteField("part_index", strconv.Itoa(partIndex))
	writer.Close()

	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("chunk upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp model.APIResponse
	json.Unmarshal(respBody, &apiResp)

	if !apiResp.Ok {
		return fmt.Errorf("chunk upload failed: %s", *apiResp.Msg)
	}

	return nil
}

// 完成分片上传
func completeChunkedUpload(serverURL, token, repo, uploadID string) error {
	url := fmt.Sprintf("%s/file/upload/complete?repo=%s", serverURL, repo)

	reqBody := map[string]interface{}{
		"upload_id": uploadID,
	}

	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("complete failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp model.APIResponse
	json.Unmarshal(body, &apiResp)

	if !apiResp.Ok {
		return fmt.Errorf("complete failed: %s", *apiResp.Msg)
	}

	return nil
}
