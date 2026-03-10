package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	PodName      string
	NodeID       string
	GRPCPort     int
	RaftPort     int
	SeedNodeID   string
	SeedNodeAddr string
}

type CLIConfig struct {
	ServerAddr  string
	Command     string
	Key         string
	Value       string
	NodeID      string
	RaftAddress string
}

func LoadServerConfig() *ServerConfig {
	_ = godotenv.Load()

	return &ServerConfig{
		PodName:      getEnv("POD_NAME", "db-0"),
		NodeID:       getEnv("NODE_ID", "db-0"),
		GRPCPort:     getEnvAsInt("GRPC_PORT", 50051),
		RaftPort:     getEnvAsInt("RAFT_PORT", 60051),
		SeedNodeID:   getEnv("SEED_NODE_ID", "db-0"),
		SeedNodeAddr: getEnv("SEED_NODE_ADDR", "localhost:50051"),
	}
}

func LoadCLIConfig() *CLIConfig {
	_ = godotenv.Load()

	return &CLIConfig{
		ServerAddr:  getEnv("SERVER_ADDR", "localhost:50051"),
		Command:     getEnv("CMD", "get"),
		Key:         getEnv("KEY", ""),
		Value:       getEnv("VAL", ""),
		NodeID:      getEnv("JOIN_NODE_ID", ""),
		RaftAddress: getEnv("JOIN_RAFT_ADDR", ""),
	}
}

func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return fallback
}
