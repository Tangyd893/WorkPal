package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Services ServicesConfig
	Gateway  GatewayConfig
	Registry RegistryConfig
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
	InternalToken  string
	JWTExpiryHours int
}

type ServicesConfig struct {
	GatewayPort      int
	UserPort         int
	IMPort           int
	FilePort         int
	SearchPort       int
	WorkspacePort    int
	ProjectPort      int
	DocsPort         int
	CalendarPort     int
	ApprovalPort     int
	AIPort           int
	NotificationPort int
	UserURL          string
	IMURL            string
	FileURL          string
	SearchURL        string
	WorkspaceURL     string
	ProjectURL       string
	DocsURL          string
	CalendarURL      string
	ApprovalURL      string
	AIURL            string
	NotificationURL  string
}

func (s ServicesConfig) BaseURLFor(service string) (string, error) {
	switch service {
	case "gateway":
		return fmt.Sprintf("http://localhost:%d", s.GatewayPort), nil
	case "user-service":
		return s.UserURL, nil
	case "im-service":
		return s.IMURL, nil
	case "file-service":
		return s.FileURL, nil
	case "search-service":
		return s.SearchURL, nil
	case "workspace-service":
		return s.WorkspaceURL, nil
	case "project-service":
		return s.ProjectURL, nil
	case "docs-service":
		return s.DocsURL, nil
	case "calendar-service":
		return s.CalendarURL, nil
	case "approval-service":
		return s.ApprovalURL, nil
	case "ai-service":
		return s.AIURL, nil
	case "notification-service":
		return s.NotificationURL, nil
	default:
		return "", fmt.Errorf("service %q does not have a configured base URL", service)
	}
}

func (s ServicesConfig) PortFor(service string) (int, error) {
	switch service {
	case "gateway":
		return s.GatewayPort, nil
	case "user-service":
		return s.UserPort, nil
	case "im-service":
		return s.IMPort, nil
	case "file-service":
		return s.FilePort, nil
	case "search-service":
		return s.SearchPort, nil
	case "workspace-service":
		return s.WorkspacePort, nil
	case "project-service":
		return s.ProjectPort, nil
	case "docs-service":
		return s.DocsPort, nil
	case "calendar-service":
		return s.CalendarPort, nil
	case "approval-service":
		return s.ApprovalPort, nil
	case "ai-service":
		return s.AIPort, nil
	case "notification-service":
		return s.NotificationPort, nil
	default:
		return 0, fmt.Errorf("service %q does not have a configured port", service)
	}
}

type GatewayConfig struct {
	HealthTimeoutMS int                          `mapstructure:"healthTimeoutMs"`
	RateLimit       GatewayRateLimitConfig       `mapstructure:"rateLimit"`
	Retry           GatewayRetryConfig           `mapstructure:"retry"`
	CircuitBreaker  GatewayCircuitBreakerConfig  `mapstructure:"circuitBreaker"`
	Timeouts        GatewayServiceTimeoutsConfig `mapstructure:"timeouts"`
}

type GatewayRateLimitConfig struct {
	Requests int `mapstructure:"requests"`
	WindowMS int `mapstructure:"windowMs"`
}

type GatewayRetryConfig struct {
	MaxAttempts int `mapstructure:"maxAttempts"`
	BackoffMS   int `mapstructure:"backoffMs"`
}

type GatewayCircuitBreakerConfig struct {
	FailureThreshold int `mapstructure:"failureThreshold"`
	CoolDownMS       int `mapstructure:"coolDownMs"`
}

type GatewayServiceTimeoutsConfig struct {
	DefaultMS   int `mapstructure:"defaultMs"`
	UserMS      int `mapstructure:"userMs"`
	IMMS        int `mapstructure:"imMs"`
	FileMS      int `mapstructure:"fileMs"`
	SearchMS    int `mapstructure:"searchMs"`
	WorkspaceMS int `mapstructure:"workspaceMs"`
	ProjectMS   int `mapstructure:"projectMs"`
}

type RegistryConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Namespace   string `mapstructure:"namespace"`
	TTLMS       int    `mapstructure:"ttlMs"`
	HeartbeatMS int    `mapstructure:"heartbeatMs"`
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
	User         string `mapstructure:"user"`
	IM           string `mapstructure:"im"`
	File         string `mapstructure:"file"`
	Workspace    string `mapstructure:"workspace"`
	Project      string `mapstructure:"project"`
	Docs         string `mapstructure:"docs"`
	Calendar     string `mapstructure:"calendar"`
	Approval     string `mapstructure:"approval"`
	Notification string `mapstructure:"notification"`
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
	case "project-service":
		return d.Names.Project, nil
	case "docs-service":
		return d.Names.Docs, nil
	case "calendar-service":
		return d.Names.Calendar, nil
	case "approval-service":
		return d.Names.Approval, nil
	case "notification-service":
		return d.Names.Notification, nil
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
	viper.SetDefault("server.internalToken", "dev-internal-token")
	viper.SetDefault("services.gatewayPort", 8080)
	viper.SetDefault("services.userPort", 8081)
	viper.SetDefault("services.imPort", 8082)
	viper.SetDefault("services.filePort", 8083)
	viper.SetDefault("services.searchPort", 8084)
	viper.SetDefault("services.workspacePort", 8085)
	viper.SetDefault("services.notificationPort", 8090)
	viper.SetDefault("services.projectPort", 8086)
	viper.SetDefault("services.docsPort", 8087)
	viper.SetDefault("services.calendarPort", 8088)
	viper.SetDefault("services.approvalPort", 8089)
	viper.SetDefault("services.aiPort", 8091)
	viper.SetDefault("services.userURL", "http://localhost:8081")
	viper.SetDefault("services.imURL", "http://localhost:8082")
	viper.SetDefault("services.fileURL", "http://localhost:8083")
	viper.SetDefault("services.searchURL", "http://localhost:8084")
	viper.SetDefault("services.workspaceURL", "http://localhost:8085")
	viper.SetDefault("services.notificationURL", "http://localhost:8090")
	viper.SetDefault("services.projectURL", "http://localhost:8086")
	viper.SetDefault("services.docsURL", "http://localhost:8087")
	viper.SetDefault("services.calendarURL", "http://localhost:8088")
	viper.SetDefault("services.approvalURL", "http://localhost:8089")
	viper.SetDefault("services.aiURL", "http://localhost:8091")
	viper.SetDefault("gateway.healthTimeoutMs", 3000)
	viper.SetDefault("gateway.rateLimit.requests", 180)
	viper.SetDefault("gateway.rateLimit.windowMs", 60000)
	viper.SetDefault("gateway.retry.maxAttempts", 2)
	viper.SetDefault("gateway.retry.backoffMs", 75)
	viper.SetDefault("gateway.circuitBreaker.failureThreshold", 5)
	viper.SetDefault("gateway.circuitBreaker.coolDownMs", 10000)
	viper.SetDefault("gateway.timeouts.defaultMs", 8000)
	viper.SetDefault("gateway.timeouts.userMs", 5000)
	viper.SetDefault("gateway.timeouts.imMs", 15000)
	viper.SetDefault("gateway.timeouts.fileMs", 20000)
	viper.SetDefault("gateway.timeouts.searchMs", 5000)
	viper.SetDefault("gateway.timeouts.workspaceMs", 8000)
	viper.SetDefault("gateway.timeouts.projectMs", 8000)
	viper.SetDefault("registry.enabled", true)
	viper.SetDefault("registry.namespace", "workpal:registry")
	viper.SetDefault("registry.ttlMs", 15000)
	viper.SetDefault("registry.heartbeatMs", 5000)
	viper.SetDefault("database.adminDBName", "postgres")
	viper.SetDefault("database.maxOpenConns", 25)
	viper.SetDefault("database.maxIdleConns", 5)
	viper.SetDefault("database.names.user", "workpal_user")
	viper.SetDefault("database.names.im", "workpal_im")
	viper.SetDefault("database.names.file", "workpal_file")
	viper.SetDefault("database.names.workspace", "workpal_workspace")
	viper.SetDefault("database.names.notification", "workpal_notification")
	viper.SetDefault("database.names.project", "workpal_project")
	viper.SetDefault("database.names.docs", "workpal_docs")
	viper.SetDefault("database.names.calendar", "workpal_calendar")
	viper.SetDefault("database.names.approval", "workpal_approval")
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
