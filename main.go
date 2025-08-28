package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/hfile/client"
	"github.com/litongjava/hfile/config"
	constant "github.com/litongjava/hfile/const"
	"github.com/litongjava/hfile/utils"
	"os"
	"path/filepath"
)

const (
	RegisterPath = "/api/v1/register"
	LoginPath    = "/api/v1/login"
	ProfilePath  = "/api/v1/user/profile"
	RepoListPath = "/repo/list"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	// 处理 config 子命令
	if cmd == "config" {
		if len(os.Args) < 3 {
			fmt.Println("❌Missing config subcommand")
			printConfigUsage()
			os.Exit(1)
		}

		subCmd := os.Args[2]
		switch subCmd {
		case "list":
			repoDir := os.Args[2]
			if repoDir == "" {
				repoDir = "."
			}
			config.ListConfigs(repoDir)
			return
		default:
			fmt.Printf("❌ Invalid config subcommand: %s\n", subCmd)
			printConfigUsage()
			os.Exit(1)
		}
	}

	if cmd == "repo" {
		if len(os.Args) < 3 {
			fmt.Println("❌ Missing repo subcommand")
			printConfigUsage()
			os.Exit(1)
		}

		subCmd := os.Args[2]
		switch subCmd {
		case "list":
			var repoDir = "."
			if len(os.Args) > 3 {
				repoDir = os.Args[3]
			}
			handleListRepos(repoDir)
			return
		default:
			fmt.Printf("❌ Invalid config subcommand: %s\n", subCmd)
			printConfigUsage()
			os.Exit(1)
		}
	}

	// 处理其他主命令
	repoDir := "."

	switch cmd {
	case "init":
		handleInit()
	case "init-local":
		if len(os.Args) > 2 {
			repoDir = os.Args[2]
		}
		handleInitLocal(repoDir)
	case "register":
		if len(os.Args) > 4 {
			repoDir = os.Args[4]
		}

		handleRegister(repoDir)
	case "login":
		if len(os.Args) < 5 {
			repoDir = "."
		} else {
			repoDir := os.Args[4]
			if repoDir == "" {
				repoDir = "."
			}
		}
		handleLogin(repoDir)
	case "profile":
		if len(os.Args) > 2 {
			repoDir = os.Args[2]
		}
		handleProfile(repoDir)
	case "push":
		if len(os.Args) > 2 {
			repoDir = os.Args[2]
		}
		handlePush(repoDir)
	case "pull":
		if len(os.Args) > 2 {
			repoDir = os.Args[2]
		}
		handlePull(repoDir)
	case "status":
		if len(os.Args) > 3 {
			repoDir = os.Args[2]
		}
		handleStatus(repoDir)
	default:
		fmt.Println("❌ 无效命令:", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  hfile init [server_url]          # 初始化用户主目录配置文件")
	fmt.Println("  hfile init-local [server_url]    # 初始化当前目录配置文件")
	fmt.Println("  hfile config list                # 显示所有配置信息")
	fmt.Println("  hfile repo list                # show all repository")
	fmt.Println("  hfile register <email> <password>       # 注册用户")
	fmt.Println("  hfile login <email> <password>          # 用户登录")
	fmt.Println("  hfile push                       # 推送本地变更到远程")
	fmt.Println("  hfile pull                       # 拉取远程变更到本地")
	fmt.Println("  hfile status                     # 显示待上传/下载的文件")
	fmt.Printf("  默认服务器地址: %s\n", constant.ServerURL)
}

func printConfigUsage() {
	fmt.Println("Usage:")
	fmt.Println("  hfile config list                # 显示所有配置信息")
}

func handleInit() {
	var serverURL string
	if len(os.Args) > 2 {
		serverURL = os.Args[2]
	}

	if err := config.InitConfig(serverURL); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	err := os.Mkdir(".hfile", 0755)
	if err != nil {
		hlog.Error("Failed:", err.Error())
	}
}

func handleInitLocal(repoDir string) {
	var serverURL string
	if len(os.Args) > 2 {
		serverURL = os.Args[2]
	}

	// 设置默认服务器地址
	if serverURL == "" {
		serverURL = constant.ServerURL
	}

	localConfig := config.Config{
		Server: serverURL,
	}

	// 创建当前目录的配置文件
	configDir := filepath.Join(repoDir, ".hfile")
	os.Mkdir(configDir, 0755)
	configFilePath := filepath.Join(configDir, "config.toml")
	file, err := os.Create(configFilePath)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// 写入配置
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(localConfig); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ created: %s\n", configFilePath)
	fmt.Printf("server url: %s\n", serverURL)
}

func handleRegister(repoDir string) {
	if len(os.Args) < 4 {
		fmt.Println("❌ 缺少参数。用法: hfile register <username> <password>")
		os.Exit(1)
	}

	username := os.Args[2]
	password := os.Args[3]

	serverURL, err := config.LoadConfig(repoDir)
	if err != nil {
		fmt.Println("❌ 加载配置失败:", err)
		os.Exit(1)
	}

	fmt.Printf("🔧 server url: %s\n", serverURL)
	client.Register(serverURL+RegisterPath, username, password)
}

func handleLogin(repoDir string) {
	if len(os.Args) < 4 {
		fmt.Println("❌ 缺少参数。用法: hfile login <email> <password>")
		os.Exit(1)
	}

	username := os.Args[2]
	password := os.Args[3]

	serverURL, err := config.LoadConfig(repoDir)
	if err != nil {
		fmt.Println("❌ Failed:", err)
		os.Exit(1)
	}

	fmt.Printf("🔧 server url: %s\n", serverURL)
	client.Login(serverURL+LoginPath, username, password)
}

func handleProfile(repoDir string) {
	serverURL, err := config.LoadConfig(repoDir)
	if err != nil {
		fmt.Println("❌ Failed:", err)
		os.Exit(1)
	}
	token, _, err := config.LoadToken()
	if err != nil {
		fmt.Println("❌ not found token，please login first")
		os.Exit(1)
	}
	client.Profile(serverURL+ProfilePath, token)
}

func handleListRepos(repoDir string) {
	serverURL, err := config.LoadConfig(repoDir)
	if err != nil {
		fmt.Println("❌ Failed:", err)
		os.Exit(1)
	}
	token, _, err := config.LoadToken()
	if err != nil {
		fmt.Println("❌ not found token，please login first")
		os.Exit(1)
	}
	client.RepoList(serverURL+RepoListPath, token)
}

func handlePush(repoDir string) {
	repo, err := utils.GetRepoName(repoDir)
	if err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}

	serverURL, err := config.LoadConfig(repoDir)
	if err != nil {
		fmt.Println("❌ Failed to load config:", err)
		os.Exit(1)
	}

	token, _, err := config.LoadToken()
	if err != nil {
		fmt.Println("❌ Not logged in. Please login first.")
		os.Exit(1)
	}

	remoteFiles, err := client.FetchRemoteFiles(serverURL, token, repo)
	if err != nil {
		fmt.Println("❌ Failed to fetch remote files:", err.Error())
		os.Exit(1)
	}

	localFiles, err := utils.ScanLocalFiles(repoDir)
	if err != nil {
		fmt.Println("❌ Failed to scan local files:", err)
		os.Exit(1)
	}

	uploadList := client.CompareForUpload(localFiles, remoteFiles)

	for _, file := range uploadList {
		relpath := file.Path
		fmt.Printf("📤 Uploading: %s\n", relpath)
		localFilePath := filepath.Join(repoDir, relpath)

		err := client.UploadFile(serverURL, token, repo, relpath, localFilePath, file.ModTime)
		if err != nil {
			fmt.Printf("❌ Upload failed for %v\n", err)
		} else {
			fmt.Printf("✅ Uploaded: %s\n", relpath)
		}
	}
}

func handlePull(repoDir string) {
	repo, err := utils.GetRepoName(repoDir)
	if err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}

	serverURL, err := config.LoadConfig(repoDir)
	if err != nil {
		fmt.Println("❌ Failed to load config:", err)
		os.Exit(1)
	}

	token, _, err := config.LoadToken()
	if err != nil {
		fmt.Println("❌ Not logged in. Please login first.")
		os.Exit(1)
	}

	remoteFiles, err := client.FetchRemoteFiles(serverURL, token, repo)
	if err != nil {
		fmt.Println("❌ Failed to fetch remote files:", err)
		os.Exit(1)
	}

	localFiles, err := utils.ScanLocalFiles(".")
	if err != nil {
		fmt.Println("❌ Failed to scan local files:", err)
		os.Exit(1)
	}

	downloadList := client.CompareForDownload(localFiles, remoteFiles)

	for _, file := range downloadList {
		relPath := file.Path
		fmt.Printf("📥 Downloading: %s\n", relPath)
		err := client.DownloadFile(serverURL, token, repo, relPath)
		if err != nil {
			fmt.Printf("❌ Download failed for %s: %v\n", relPath, err)
		} else {
			fmt.Printf("✅ Downloaded: %s\n", relPath)
		}
	}
}

func handleStatus(repoDir string) {
	repo, err := utils.GetRepoName(repoDir)
	if err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}

	serverURL, err := config.LoadConfig(repoDir)
	if err != nil {
		fmt.Println("❌ Failed to load config:", err)
		os.Exit(1)
	}

	token, _, err := config.LoadToken()
	if err != nil {
		fmt.Println("❌ Not logged in. Please login first.")
		os.Exit(1)
	}

	remoteFiles, err := client.FetchRemoteFiles(serverURL, token, repo)
	if err != nil {
		fmt.Println("❌ Failed to fetch remote files:", err)
		os.Exit(1)
	}

	localFiles, err := utils.ScanLocalFiles(repoDir)
	if err != nil {
		fmt.Println("❌ Failed to scan local files:", err)
		os.Exit(1)
	}

	toUpload := client.CompareForUpload(localFiles, remoteFiles)
	toDownload := client.CompareForDownload(localFiles, remoteFiles)

	if len(toUpload) > 0 {
		fmt.Println("🟢 Files to upload:")
		for _, f := range toUpload {
			fmt.Println("  +", f.Path)
		}
	} else {
		fmt.Println("🟢 No files need to be uploaded.")
	}

	if len(toDownload) > 0 {
		fmt.Println("🔵 Files to download:")
		for _, f := range toDownload {
			fmt.Println("  -", f.Path)
		}
	} else {
		fmt.Println("🔵 No files need to be download.")
	}

}
