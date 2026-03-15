// Package config 提供 mesh 配置读写能力。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mesh-cli/pkg/types"
)

const defaultConfigFile = ".mesh/config.json"

// ResolveConfigPath 返回配置文件路径（支持 MESH_CONFIG_PATH 覆盖）。
func ResolveConfigPath() (string, error) {
	if fromEnv := strings.TrimSpace(os.Getenv("MESH_CONFIG_PATH")); fromEnv != "" {
		return fromEnv, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(homeDir, defaultConfigFile), nil
}

// Save 保存配置到文件。
func Save(cfg types.Config) (string, error) {
	path, err := ResolveConfigPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}
	return path, nil
}

// Load 加载配置。
func Load() (types.Config, error) {
	var cfg types.Config
	path, err := ResolveConfigPath()
	if err != nil {
		return cfg, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
