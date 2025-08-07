package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	constant "github.com/litongjava/hftp/const"
)

type Config struct {
	Server string `toml:"server"`
}

// InitConfig 初始化配置文件
func InitConfig(serverURL string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取用户主目录: %v", err)
	}

	configDir := filepath.Join(homeDir, ".hftp")
	configPath := filepath.Join(configDir, "config.toml")

	// 创建配置目录
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 设置默认服务器地址
	if serverURL == "" {
		serverURL = constant.ServerURL
	}

	config := Config{
		Server: serverURL,
	}

	// 创建配置文件
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("创建配置文件失败: %v", err)
	}
	defer file.Close()

	// 写入配置
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	fmt.Printf("✅ 配置文件已创建: %s\n", configPath)
	fmt.Printf("服务器地址: %s\n", serverURL)
	return nil
}

// LoadConfig 读取配置文件，优先级：当前目录 > ~/.hftp/config.toml > 默认地址
func LoadConfig() (string, error) {
	// 1. 检查当前目录下的 config.toml
	if serverURL, err := loadFromCurrentDir(); err == nil && serverURL != "" {
		return serverURL, nil
	}

	// 2. 检查用户主目录下的配置
	if serverURL, err := loadFromHomeDir(); err == nil && serverURL != "" {
		return serverURL, nil
	}

	// 3. 返回默认地址
	return constant.ServerURL, nil
}

// loadFromCurrentDir 从当前目录加载配置
func loadFromCurrentDir() (string, error) {
	configPath := "config.toml"

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("当前目录下不存在 config.toml")
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return "", fmt.Errorf("解析当前目录配置文件失败: %v", err)
	}

	if cfg.Server == "" {
		return "", fmt.Errorf("当前目录配置文件中未设置服务器地址")
	}

	return cfg.Server, nil
}

// loadFromHomeDir 从用户主目录加载配置
func loadFromHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(homeDir, ".hftp", "config.toml")

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("用户主目录下不存在配置文件")
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return "", fmt.Errorf("解析用户主目录配置文件失败: %v", err)
	}

	if cfg.Server == "" {
		return "", fmt.Errorf("用户主目录配置文件中未设置服务器地址")
	}

	return cfg.Server, nil
}

// ListConfigs 显示所有配置信息
func ListConfigs() {
	// . 显示实际生效的配置
	activeConfig, _ := LoadConfig()
	fmt.Printf("server: %s\n", activeConfig)
}

// getCurrentDirConfig 获取当前目录配置（不返回错误，用于显示）
func getCurrentDirConfig() (Config, error) {
	configPath := "config.toml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Config{}, err
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// getHomeDirConfig 获取用户主目录配置（不返回错误，用于显示）
func getHomeDirConfig() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	configPath := filepath.Join(homeDir, ".hftp", "config.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Config{}, err
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
