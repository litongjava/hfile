package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	constant "github.com/litongjava/hfile/const"
)

type Config struct {
	Server       string `toml:"server"`
	Token        string `toml:"token,omitempty"`
	RefreshToken string `toml:"refresh_token,omitempty"`
}

// InitConfig initializes configuration file
func InitConfig(serverURL string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".hfile")
	configPath := filepath.Join(configDir, "config.toml")

	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Set default server URL
	if serverURL == "" {
		serverURL = constant.ServerURL
	}

	config := Config{
		Server: serverURL,
	}

	// Create config file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	// Write config
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	fmt.Printf("✅ Config file created: %s\n", configPath)
	fmt.Printf("Server URL: %s\n", serverURL)
	return nil
}

// LoadConfig loads config file with priority: current dir > ~/.hfile/config.toml > default
func LoadConfig(repoDir string) (string, error) {
	// 1. Check repo directory config
	if serverURL, err := loadFromRepoDir(repoDir); err == nil && serverURL != "" {
		return serverURL, nil
	}

	// 2. Check user home directory config
	if serverURL, err := loadFromHomeDir(); err == nil && serverURL != "" {
		return serverURL, nil
	}

	// 3. Return default URL
	return constant.ServerURL, nil
}

// loadFromRepoDir loads config from current directory
func loadFromRepoDir(repoDir string) (string, error) {
	configPath := filepath.Join(repoDir, ".hfile", "config.toml")
	hlog.Info(configPath)
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("config file not found in current directory")
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return "", fmt.Errorf("failed to parse current directory config: %v", err)
	}

	if cfg.Server == "" {
		return "", fmt.Errorf("server URL not set in current directory config")
	}

	return cfg.Server, nil
}

// loadFromHomeDir loads config from user home directory
func loadFromHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(homeDir, ".hfile", "config.toml")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("config file not found in user home directory")
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return "", fmt.Errorf("failed to parse user home directory config: %v", err)
	}

	if cfg.Server == "" {
		return "", fmt.Errorf("server URL not set in user home directory config")
	}

	return cfg.Server, nil
}

// ListConfigs displays all config information
func ListConfigs(repoDir string) {
	// Display active config
	activeConfig, _ := LoadConfig(repoDir)
	fmt.Printf("activte server: %s\n", activeConfig)

	// Display current directory config
	if cfg, err := getCurrentDirConfig(); err == nil {
		fmt.Printf("current dir config - server: %s, token: %s\n", cfg.Server, maskToken(cfg.Token))
	}

	// Display home directory config
	if cfg, err := getHomeDirConfig(); err == nil {
		fmt.Printf("home dir config - server: %s, token: %s\n", cfg.Server, maskToken(cfg.Token))
	}
}

// getCurrentDirConfig gets current directory config
func getCurrentDirConfig() (Config, error) {
	configPath := filepath.Join(".hfile", "config.toml")
	_, err := os.Stat(configPath)

	if err != nil {
		if os.IsNotExist(err) {
			hlog.Info("not found file:", configPath)
			return Config{}, err
		} else {
			hlog.Error(err.Error())
			return Config{}, err
		}
	}
	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// getHomeDirConfig gets user home directory config
func getHomeDirConfig() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	configPath := filepath.Join(homeDir, ".hfile", "config.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Config{}, err
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// SaveToken saves token to the highest priority config file
func SaveToken(token, refreshToken string) error {
	// First check current directory config
	configPath := filepath.Join(".hfile", "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		return saveTokenToCurrentDir(token, refreshToken)
	}

	// Then check user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	configPath = filepath.Join(homeDir, ".hfile", "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		return saveTokenToHomeDir(token, refreshToken)
	}

	// If no config file exists, create current directory config
	return saveTokenToCurrentDir(token, refreshToken)
}

// saveTokenToCurrentDir saves token to current directory config file
func saveTokenToCurrentDir(token, refreshToken string) error {
	configPath := filepath.Join(".hfile", "config.toml")

	// Read existing config
	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		// If file doesn't exist or parse fails, create new
		cfg = Config{}

		// Ensure .hfile directory exists
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}

	// Update token
	cfg.Token = token
	cfg.RefreshToken = refreshToken

	// Preserve server URL if exists, otherwise set default
	if cfg.Server == "" {
		cfg.Server = constant.ServerURL
	}

	// Write back to file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	fmt.Printf("✅ Token saved to: %s\n", configPath)
	return nil
}

// saveTokenToHomeDir saves token to user home directory config file
func saveTokenToHomeDir(token, refreshToken string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".hfile")
	configPath := filepath.Join(configDir, "config.toml")

	// Read existing config
	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// Update token
	cfg.Token = token
	cfg.RefreshToken = refreshToken

	// Write back to file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	fmt.Printf("✅ Token saved to: %s\n", configPath)
	return nil
}

// LoadToken loads token from config file following priority order
func LoadToken() (string, string, error) {
	// 1. Check current directory config
	if token, refreshToken, err := loadTokenFromCurrentDir(); err == nil && token != "" {
		return token, refreshToken, nil
	}

	// 2. Check user home directory config
	if token, refreshToken, err := loadTokenFromHomeDir(); err == nil && token != "" {
		return token, refreshToken, nil
	}

	return "", "", fmt.Errorf("no valid token configuration found")
}

// loadTokenFromCurrentDir loads token from current directory config file
func loadTokenFromCurrentDir() (string, string, error) {
	configPath := filepath.Join(".hfile", "config.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("config file not found in current directory")
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return "", "", fmt.Errorf("failed to parse current directory config: %v", err)
	}

	return cfg.Token, cfg.RefreshToken, nil
}

// loadTokenFromHomeDir loads token from user home directory config file
func loadTokenFromHomeDir() (string, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	configPath := filepath.Join(homeDir, ".hfile", "config.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("config file not found in user home directory")
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return "", "", fmt.Errorf("failed to parse user home directory config: %v", err)
	}

	return cfg.Token, cfg.RefreshToken, nil
}

// maskToken masks token for display purposes
func maskToken(token string) string {
	if len(token) <= 10 {
		return "****"
	}
	return token[:6] + "****" + token[len(token)-4:]
}
