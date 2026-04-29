package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Services ServicesConfig
	Database DatabaseConfig
	Redis    RedisConfig
	File     FileConfig
	Search   SearchConfig
	WS       WSConfig
}

type ServerConfig struct {
	Port           int
	Mode           string
	JWTSecret      string
	JWTExpiryHours int
}

type ServicesConfig struct {
	GatewayPort   int
	UserPort      int
	IMPort        int
	FilePort      int
	SearchPort    int
	WorkspacePort int
	UserURL       string
	IMURL         string
	FileURL       string
	SearchURL     string
	WorkspaceURL  string
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	AdminDBName  string
	MaxOpenConns int
	MaxIdleConns int
	Names        DatabaseNames `mapstructure:"names"`
}

type DatabaseNames struct {
	User      string `mapstructure:"user"`
	IM        string `mapstructure:"im"`
	File      string `mapstructure:"file"`
	Workspace string `mapstructure:"workspace"`
}

func (d DatabaseConfig) DSN(dbName string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, dbName,
	)
}

func (d DatabaseConfig) ServiceDBName(service string) (string, error) {
	switch service {
	case "user-service":
		return d.Names.User, nil
	case "im-service":
		return d.Names.IM, nil
	case "file-service":
		return d.Names.File, nil
	case "workspace-service":
		return d.Names.Workspace, nil
	default:
		return "", fmt.Errorf("service %q does not own a database", service)
	}
}

type RedisConfig struct {
	Host       string
	Port       int
	Password   string
	DB         int
	StreamsKey string
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type FileConfig struct {
	StoreType     string
	LocalBaseDir  string
	MaxFileSizeMB int
	MinIO         MinIOConfig
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type SearchConfig struct {
	Engine        string
	Bleve         BleveConfig
	Elasticsearch ElasticsearchConfig
}

type BleveConfig struct {
	IndexPath string
}

type ElasticsearchConfig struct {
	Addresses []string
	Index     string
}

type WSConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	PingIntervalSec int
	PongWaitSec     int
	MaxMessageSize  int
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.jwtExpiryHours", 72)
	viper.SetDefault("services.gatewayPort", 8080)
	viper.SetDefault("services.userPort", 8081)
	viper.SetDefault("services.imPort", 8082)
	viper.SetDefault("services.filePort", 8083)
	viper.SetDefault("services.searchPort", 8084)
	viper.SetDefault("services.workspacePort", 8085)
	viper.SetDefault("services.userURL", "http://localhost:8081")
	viper.SetDefault("services.imURL", "http://localhost:8082")
	viper.SetDefault("services.fileURL", "http://localhost:8083")
	viper.SetDefault("services.searchURL", "http://localhost:8084")
	viper.SetDefault("services.workspaceURL", "http://localhost:8085")
	viper.SetDefault("database.adminDBName", "postgres")
	viper.SetDefault("database.maxOpenConns", 25)
	viper.SetDefault("database.maxIdleConns", 5)
	viper.SetDefault("database.names.user", "workpal_user")
	viper.SetDefault("database.names.im", "workpal_im")
	viper.SetDefault("database.names.file", "workpal_file")
	viper.SetDefault("database.names.workspace", "workpal_workspace")
	viper.SetDefault("redis.streamsKey", "workpal:streams:messages")
	viper.SetDefault("file.storeType", "local")
	viper.SetDefault("file.localBaseDir", "./uploads")
	viper.SetDefault("file.maxFileSizeMB", 50)
	viper.SetDefault("search.bleve.indexPath", "./data/search")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
