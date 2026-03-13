// Package types 定义 Mesh 的公共数据类型
package types

// Insight 表示一条认知记录
type Insight struct {
	ID        string // 唯一标识（UUID）
	Source    string // 来源
	Content   string // 原始内容
	Summary   string // 摘要
	Tags      string // 标签（逗号分隔）
	AgentID   string // Agent ID
	CreatedAt int64  // 创建时间戳
	UpdatedAt int64  // 更新时间戳
	IsShared  int    // 是否团队共享（0/1）
}

// Agent 表示一个 Agent
type Agent struct {
	ID        string            // Agent ID
	Name      string            // Agent 名称
	Type      string            // Agent 类型
	Device    string            // 设备标识
	LastSeen  int64             // 最后活跃时间
	Metadata  map[string]string // 元数据（扩展信息）
}

// Config 表示 Mesh 配置
type Config struct {
	SyncSpace string `json:"sync_space"` // 同步空间路径
	DBPath    string `json:"db_path"`    // 本地数据库路径
}

// Message 表示 Agent 协议消息
type Message struct {
	Version     string                 `json:"version"`
	MessageType string                 `json:"message_type"` // register/collect/query/heartbeat
	AgentID     string                 `json:"agent_id"`
	Timestamp   int64                  `json:"timestamp"`
	Payload     map[string]interface{} `json:"payload"`
}

// QueryOptions 查询选项
type QueryOptions struct {
	Query  string
	Source string
	Tags   string
	Limit  int
}

// CollectOptions 采集选项
type CollectOptions struct {
	Source  string
	Content string
	Summary string
	Tags    string
}
