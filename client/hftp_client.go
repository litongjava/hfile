package client

import (
  "bytes"
  "encoding/json"
  "fmt"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "github.com/litongjava/hfile/config"
  "github.com/litongjava/hfile/model"
  "io"
  "mime/multipart"
  "net/http"
  "os"
  "path/filepath"
)

const ChunkSize = 10 * 1024 * 1024 // 10 MB

func Register(url, username, password string) {
  reqBody := model.RegisterRequest{
    Username:         username,
    Password:         password,
    UserType:         1,
    VerificationType: 0, // 不验证邮箱
  }

  jsonData, _ := json.Marshal(reqBody)

  resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
  if err != nil {
    fmt.Println("❌ Failed:", err)
    os.Exit(1)
  }
  defer resp.Body.Close()

  body, _ := io.ReadAll(resp.Body)
  var apiResp model.APIResponse
  json.Unmarshal(body, &apiResp)

  if apiResp.Ok {
    fmt.Println("✅ Successfully!")
  } else {
    fmt.Printf("❌ Failed: %s\n", string(body))
    if data, ok := apiResp.Data.([]interface{}); ok {
      for _, item := range data {
        if fieldMap, ok := item.(map[string]interface{}); ok {
          field := fieldMap["field"]
          messages := fieldMap["messages"]
          fmt.Println("error:", field, " ", messages)

        }
      }
    }
  }
}

func Login(url, username, password string) {
  reqBody := model.LoginRequest{
    Username: username,
    Password: password,
  }

  jsonData, _ := json.Marshal(reqBody)

  resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
  if err != nil {
    fmt.Println("❌ Failed:", err)
    os.Exit(1)
  }
  defer resp.Body.Close()

  body, _ := io.ReadAll(resp.Body)
  var apiResp model.APIResponse
  json.Unmarshal(body, &apiResp)

  if apiResp.Ok {
    fmt.Println("✅ Successfully!")
    // 解析 data 字段
    data, ok := apiResp.Data.(map[string]interface{})
    if !ok {
      fmt.Println("❌ Failed to parse data field")
      os.Exit(1)
    }

    token, _ := data["token"].(string)
    refreshToken, _ := data["refresh_token"].(string)

    // 保存 token 到配置文件
    if err := config.SaveToken(token, refreshToken); err != nil {
      fmt.Println("❌ Failed to save token:", err)
      os.Exit(1)
    }
  } else {
    fmt.Printf("❌ Failed: %s\n", string(body))
  }
}

func Profile(url string, token string) {
  client := &http.Client{}
  req, _ := http.NewRequest("GET", url, nil)
  req.Header.Set("Authorization", "Bearer "+token)

  resp, err := client.Do(req)
  if err != nil {
    fmt.Println("❌ Failed:", err)
    return
  }
  defer resp.Body.Close()

  body, _ := io.ReadAll(resp.Body)
  var apiResp model.APIResponse
  json.Unmarshal(body, &apiResp)

  if apiResp.Ok {
    fmt.Println("✅ Successfully!")
    fmt.Println(string(body))
  } else {
    fmt.Println("❌ Failed")
    hlog.Errorf(string(body))
  }

}

func FetchRemoteFiles(serverURL, token, repo string) (map[string]model.FileMeta, error) {
  url := fmt.Sprintf("%s/file/list?repo=%s", serverURL, repo)
  req, _ := http.NewRequest("GET", url, nil)
  req.Header.Set("Authorization", "Bearer "+token)

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  body, _ := io.ReadAll(resp.Body)
  if resp.StatusCode != 200 {
    return nil, fmt.Errorf("server error: %s", string(body))
  }

  var apiResp model.APIResponse
  json.Unmarshal(body, &apiResp)

  if !apiResp.Ok {
    return nil, fmt.Errorf("API error: %s", *apiResp.Msg)
  }

  remoteMap := make(map[string]model.FileMeta)
  data, ok := apiResp.Data.([]interface{})
  if !ok {
    return nil, fmt.Errorf("invalid data format")
  }

  for _, item := range data {
    m := item.(map[string]interface{})
    path := m["path"].(string)
    hash := m["hash"].(string)
    modTime := int64(m["mod_time"].(float64))
    remoteMap[path] = model.FileMeta{
      Path:    path,
      Hash:    hash,
      ModTime: modTime,
    }
  }

  return remoteMap, nil
}

func UploadFile(serverURL, token, repo, filePath string) error {
  fileInfo, err := os.Stat(filePath)
  if err != nil {
    return err
  }

  if fileInfo.Size() > 100*1024*1024 {
    return uploadInChunks(serverURL, token, repo, filePath)
  }

  url := fmt.Sprintf("%s/file/upload?repo=%s", serverURL, repo)
  file, err := os.Open(filePath)
  if err != nil {
    return err
  }
  defer file.Close()

  body := &bytes.Buffer{}
  writer := multipart.NewWriter(body)
  part, _ := writer.CreateFormFile("file", filePath)
  io.Copy(part, file)
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

  if resp.StatusCode != 200 {
    body, _ := io.ReadAll(resp.Body)
    return fmt.Errorf("upload failed: %s", string(body))
  }
  return nil
}

func uploadInChunks(serverURL, token, repo, filePath string) error {
  // 实现分片上传逻辑，调用 /file/upload/slice 接口
  // 略
  return nil
}

func DownloadFile(serverURL, token, repo, remotePath string) error {
  localPath := remotePath
  dir := filepath.Dir(localPath)
  if err := os.MkdirAll(dir, 0755); err != nil {
    return err
  }

  url := fmt.Sprintf("%s/file/download?repo=%s&file=%s", serverURL, repo, remotePath)

  var start int64 = 0
  if stat, err := os.Stat(localPath); err == nil {
    start = stat.Size()
  }

  req, _ := http.NewRequest("GET", url, nil)
  req.Header.Set("Authorization", "Bearer "+token)
  if start > 0 {
    req.Header.Set("Range", fmt.Sprintf("bytes=%d-", start))
  }

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  if resp.StatusCode == 416 {
    // 已下载完毕
    return nil
  }

  mode := os.O_CREATE | os.O_WRONLY
  if start > 0 {
    mode |= os.O_APPEND
  }

  file, err := os.OpenFile(localPath, mode, 0644)
  if err != nil {
    return err
  }
  defer file.Close()

  _, err = io.Copy(file, resp.Body)
  return err
}

func CompareForUpload(local, remote map[string]model.FileMeta) []model.FileMeta {
  var result []model.FileMeta
  for path, l := range local {
    if r, ok := remote[path]; ok {
      if l.Hash != r.Hash && l.ModTime > r.ModTime {
        result = append(result, l)
      }
    } else {
      result = append(result, l)
    }
  }
  return result
}

func CompareForDownload(local, remote map[string]model.FileMeta) []model.FileMeta {
  var result []model.FileMeta
  for path, r := range remote {
    if l, ok := local[path]; ok {
      if r.Hash != l.Hash && r.ModTime > l.ModTime {
        result = append(result, r)
      }
    } else {
      result = append(result, r)
    }
  }
  return result
}
