package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/litongjava/hftp/config"
	constant "github.com/litongjava/hftp/const"
	"github.com/litongjava/hftp/model"
)

const (
	RegisterPath = "/api/v1/register"
	LoginPath    = "/api/v1/login"
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
			fmt.Println("❌ 缺少 config 子命令")
			printConfigUsage()
			os.Exit(1)
		}

		subCmd := os.Args[2]
		switch subCmd {
		case "list":
			config.ListConfigs()
			return
		default:
			fmt.Printf("❌ 无效的 config 子命令: %s\n", subCmd)
			printConfigUsage()
			os.Exit(1)
		}
	}

	// 处理其他主命令
	switch cmd {
	case "init":
		handleInit()
	case "init-local":
		handleInitLocal()
	case "register":
		handleRegister()
	case "login":
		handleLogin()
	default:
		fmt.Println("❌ 无效命令:", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  hftp init [server_url]          # 初始化用户主目录配置文件")
	fmt.Println("  hftp init-local [server_url]    # 初始化当前目录配置文件")
	fmt.Println("  hftp config list                # 显示所有配置信息")
	fmt.Println("  hftp register <email> <password>       # 注册用户")
	fmt.Println("  hftp login <email> <password>          # 用户登录")
	fmt.Printf("  默认服务器地址: %s\n", constant.ServerURL)
}

func printConfigUsage() {
	fmt.Println("Usage:")
	fmt.Println("  hftp config list                # 显示所有配置信息")
}

func handleInit() {
	var serverURL string
	if len(os.Args) > 2 {
		serverURL = os.Args[2]
	}

	if err := config.InitConfig(serverURL); err != nil {
		fmt.Printf("❌ 初始化配置失败: %v\n", err)
		os.Exit(1)
	}
}

func handleInitLocal() {
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
	file, err := os.Create("config.toml")
	if err != nil {
		fmt.Printf("❌ 创建本地配置文件失败: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// 写入配置
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(localConfig); err != nil {
		fmt.Printf("❌ 写入本地配置文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 本地配置文件已创建: %s\n", "config.toml")
	fmt.Printf("服务器地址: %s\n", serverURL)
}

func handleRegister() {
	if len(os.Args) < 4 {
		fmt.Println("❌ 缺少参数。用法: hftp register <email> <password>")
		os.Exit(1)
	}

	email := os.Args[2]
	password := os.Args[3]

	serverURL, err := config.LoadConfig()
	if err != nil {
		fmt.Println("❌ 加载配置失败:", err)
		os.Exit(1)
	}

	fmt.Printf("🔧 使用服务器地址: %s\n", serverURL)
	register(serverURL+RegisterPath, email, password)
}

func handleLogin() {
	if len(os.Args) < 4 {
		fmt.Println("❌ 缺少参数。用法: hftp login <email> <password>")
		os.Exit(1)
	}

	email := os.Args[2]
	password := os.Args[3]

	serverURL, err := config.LoadConfig()
	if err != nil {
		fmt.Println("❌ 加载配置失败:", err)
		os.Exit(1)
	}

	fmt.Printf("🔧 使用服务器地址: %s\n", serverURL)
	login(serverURL+LoginPath, email, password)
}

func register(url, email, password string) {
	reqBody := model.RegisterRequest{
		Email:            email,
		Password:         password,
		UserType:         1,
		VerificationType: 0, // 不验证邮箱
	}

	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("❌ 注册请求失败:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var apiResp model.APIResponse
	json.Unmarshal(body, &apiResp)

	if apiResp.Ok {
		fmt.Println("✅ 注册成功!")
	} else {
		fmt.Printf("❌ 注册失败: %s\n", string(body))
		if data, ok := apiResp.Data.([]interface{}); ok {
			for _, item := range data {
				if fieldMap, ok := item.(map[string]interface{}); ok {
					field := fieldMap["field"]
					messages := fieldMap["messages"]
					if field == "password" {
						fmt.Println("密码错误:", messages)
					}
				}
			}
		}
	}
}

func login(url, email, password string) {
	reqBody := model.LoginRequest{
		Email:    email,
		Password: password,
	}

	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("❌ 登录请求失败:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var apiResp model.APIResponse
	json.Unmarshal(body, &apiResp)

	if apiResp.Ok {
		fmt.Println("✅ 登录成功!")
		// 可以打印 token 等信息
		fmt.Println("响应:", string(body))
	} else {
		fmt.Printf("❌ 登录失败: %s\n", string(body))
	}
}
