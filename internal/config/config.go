package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// AppConfig 全局配置实例
var AppConfig Config

// Config 根配置结构体
type Config struct {
	ServerConfig `mapstructure:"server"`
	KafkaConfig  `mapstructure:"kafka"`
	LLMConfig    `mapstructure:"llm"`
	MySQLConfig  `mapstructure:"mysql"`
	RedisConfig  `mapstructure:"redis"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
}

type LLMConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxLifetime  int    `mapstructure:"max_lifetime"`
}

// DSN 辅助方法：一键生成 GORM 需要的 MySQL 连接字符串
func (m *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.DBName)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Addr 辅助方法：生成 Redis 地址
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// InitConfig 初始化 Viper 配置引擎
func InitConfig() {
	viper.SetConfigName("config")    // 配置文件名称 (不包含扩展名)
	viper.SetConfigType("yaml")      // 配置文件类型
	viper.AddConfigPath("./configs") // 查找配置文件的路径，相对程序执行路径

	// 工业级做法：支持环境变量覆盖 (例如：将 MYSQL_PASSWORD 映射为 mysql.password)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("❌ 无法读取配置文件: %v", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("❌ 无法将配置解析到结构体: %v", err)
	}

	log.Println("✅ 配置文件加载成功")
}
