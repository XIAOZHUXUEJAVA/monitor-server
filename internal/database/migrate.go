package database

import (
	"fmt"
	"monitor-server/internal/model"
)

// AutoMigrate 自动迁移所有模型
func (db *DB) AutoMigrate() error {
	models := []interface{}{
		&model.SystemMetrics{},
		&model.SystemInfoDB{},
		&model.AlertRule{},
		&model.Alert{},
		&model.MonitoringConfig{},
		// 主机管理相关模型
		&model.Host{},
		&model.HostConfig{},
		&model.HostGroup{},
		&model.HostGroupMember{},
	}

	for _, m := range models {
		if err := db.DB.AutoMigrate(m); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", m, err)
		}
	}

	return nil
}

// InitializeDefaultConfigs 初始化默认配置
func (db *DB) InitializeDefaultConfigs() error {
	defaultConfigs := []model.MonitoringConfig{
		{
			Key:         "refresh_interval",
			Value:       "60",
			Type:        "int",
			Category:    "system",
			Description: "数据刷新间隔（秒）",
			Editable:    true,
		},
		{
			Key:         "history_points",
			Value:       "20",
			Type:        "int",
			Category:    "dashboard",
			Description: "历史数据点数量",
			Editable:    true,
		},
		{
			Key:         "auto_refresh",
			Value:       "true",
			Type:        "bool",
			Category:    "system",
			Description: "启用自动刷新",
			Editable:    true,
		},
		{
			Key:         "cpu_alert_threshold",
			Value:       "80.0",
			Type:        "float",
			Category:    "alert",
			Description: "CPU使用率告警阈值（%）",
			Editable:    true,
		},
		{
			Key:         "memory_alert_threshold",
			Value:       "85.0",
			Type:        "float",
			Category:    "alert",
			Description: "内存使用率告警阈值（%）",
			Editable:    true,
		},
		{
			Key:         "disk_alert_threshold",
			Value:       "90.0",
			Type:        "float",
			Category:    "alert",
			Description: "磁盘使用率告警阈值（%）",
			Editable:    true,
		},
	}

	for _, config := range defaultConfigs {
		// 使用 FirstOrCreate 避免重复插入
		var existingConfig model.MonitoringConfig
		if err := db.Where("key = ?", config.Key).FirstOrCreate(&existingConfig, config).Error; err != nil {
			return fmt.Errorf("failed to create default config %s: %w", config.Key, err)
		}
	}

	return nil
}

// InitializeDefaultAlertRules 初始化默认告警规则
func (db *DB) InitializeDefaultAlertRules() error {
	defaultRules := []model.AlertRule{
		{
			Name:        "CPU使用率过高",
			MetricType:  "cpu",
			Operator:    ">",
			Threshold:   80.0,
			Duration:    300, // 5分钟
			Severity:    "warning",
			Enabled:     true,
			Description: "CPU使用率持续超过80%达5分钟时触发告警",
		},
		{
			Name:        "CPU使用率严重过高",
			MetricType:  "cpu",
			Operator:    ">",
			Threshold:   95.0,
			Duration:    60, // 1分钟
			Severity:    "critical",
			Enabled:     true,
			Description: "CPU使用率持续超过95%达1分钟时触发严重告警",
		},
		{
			Name:        "内存使用率过高",
			MetricType:  "memory",
			Operator:    ">",
			Threshold:   85.0,
			Duration:    300, // 5分钟
			Severity:    "warning",
			Enabled:     true,
			Description: "内存使用率持续超过85%达5分钟时触发告警",
		},
		{
			Name:        "内存使用率严重过高",
			MetricType:  "memory",
			Operator:    ">",
			Threshold:   95.0,
			Duration:    60, // 1分钟
			Severity:    "critical",
			Enabled:     true,
			Description: "内存使用率持续超过95%达1分钟时触发严重告警",
		},
		{
			Name:        "磁盘使用率过高",
			MetricType:  "disk",
			Operator:    ">",
			Threshold:   90.0,
			Duration:    600, // 10分钟
			Severity:    "warning",
			Enabled:     true,
			Description: "磁盘使用率持续超过90%达10分钟时触发告警",
		},
		{
			Name:        "磁盘使用率严重过高",
			MetricType:  "disk",
			Operator:    ">",
			Threshold:   98.0,
			Duration:    60, // 1分钟
			Severity:    "critical",
			Enabled:     true,
			Description: "磁盘使用率持续超过98%达1分钟时触发严重告警",
		},
	}

	for _, rule := range defaultRules {
		// 使用 FirstOrCreate 避免重复插入
		var existingRule model.AlertRule
		if err := db.Where("name = ?", rule.Name).FirstOrCreate(&existingRule, rule).Error; err != nil {
			return fmt.Errorf("failed to create default alert rule %s: %w", rule.Name, err)
		}
	}

	return nil
}

// InitializeDefaultHosts 初始化默认主机数据
func (db *DB) InitializeDefaultHosts() error {
	defaultHosts := []model.Host{
		{
			Hostname:          "localhost",
			DisplayName:       "本地主机",
			IPAddress:         "127.0.0.1",
			Environment:       "dev",
			Location:          "本地",
			Tags:              `["local", "development"]`,
			Description:       "本地开发环境主机",
			Status:            "online",
			MonitoringEnabled: true,
			Agent:             false,
		},
	}

	for _, host := range defaultHosts {
		// 使用 FirstOrCreate 避免重复插入
		var existingHost model.Host
		if err := db.Where("hostname = ?", host.Hostname).FirstOrCreate(&existingHost, host).Error; err != nil {
			return fmt.Errorf("failed to create default host %s: %w", host.Hostname, err)
		}
	}

	return nil
}

// InitializeDefaultHostGroups 初始化默认主机组
func (db *DB) InitializeDefaultHostGroups() error {
	defaultGroups := []model.HostGroup{
		{
			Name:        "development",
			DisplayName: "开发环境",
			Description: "开发环境主机组",
			Environment: "dev",
			Tags:        `["dev", "testing"]`,
			Enabled:     true,
		},
		{
			Name:        "production",
			DisplayName: "生产环境",
			Description: "生产环境主机组",
			Environment: "prod",
			Tags:        `["prod", "critical"]`,
			Enabled:     true,
		},
		{
			Name:        "staging",
			DisplayName: "预发布环境",
			Description: "预发布环境主机组",
			Environment: "staging",
			Tags:        `["staging", "testing"]`,
			Enabled:     true,
		},
	}

	for _, group := range defaultGroups {
		// 使用 FirstOrCreate 避免重复插入
		var existingGroup model.HostGroup
		if err := db.Where("name = ?", group.Name).FirstOrCreate(&existingGroup, group).Error; err != nil {
			return fmt.Errorf("failed to create default host group %s: %w", group.Name, err)
		}
	}

	return nil
}

// InitializeDefaultHostConfigs 初始化默认主机配置
func (db *DB) InitializeDefaultHostConfigs() error {
	// 首先获取 localhost 主机
	var localhost model.Host
	if err := db.Where("hostname = ?", "localhost").First(&localhost).Error; err != nil {
		// 如果没有找到主机，跳过配置初始化
		return nil
	}

	defaultConfigs := []model.HostConfig{
		{
			HostID:      localhost.ID,
			Key:         "monitoring_interval",
			Value:       "60",
			Type:        "int",
			Category:    "monitoring",
			Description: "监控数据收集间隔（秒）",
			Editable:    true,
		},
		{
			HostID:      localhost.ID,
			Key:         "alert_enabled",
			Value:       "true",
			Type:        "bool",
			Category:    "alert",
			Description: "启用告警",
			Editable:    true,
		},
		{
			HostID:      localhost.ID,
			Key:         "agent_port",
			Value:       "9001",
			Type:        "int",
			Category:    "system",
			Description: "监控代理端口",
			Editable:    true,
		},
	}

	for _, config := range defaultConfigs {
		// 使用 FirstOrCreate 避免重复插入
		var existingConfig model.HostConfig
		if err := db.Where("host_id = ? AND key = ?", config.HostID, config.Key).FirstOrCreate(&existingConfig, config).Error; err != nil {
			return fmt.Errorf("failed to create default host config %s: %w", config.Key, err)
		}
	}

	return nil
}

// Setup 完整的数据库设置，包括迁移和初始化数据
func (db *DB) Setup() error {
	// 自动迁移
	if err := db.AutoMigrate(); err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}

	// 初始化默认配置
	if err := db.InitializeDefaultConfigs(); err != nil {
		return fmt.Errorf("initialize default configs failed: %w", err)
	}

	// 初始化默认告警规则
	if err := db.InitializeDefaultAlertRules(); err != nil {
		return fmt.Errorf("initialize default alert rules failed: %w", err)
	}

	// 初始化默认主机
	if err := db.InitializeDefaultHosts(); err != nil {
		return fmt.Errorf("initialize default hosts failed: %w", err)
	}

	// 初始化默认主机组
	if err := db.InitializeDefaultHostGroups(); err != nil {
		return fmt.Errorf("initialize default host groups failed: %w", err)
	}

	// 初始化默认主机配置
	if err := db.InitializeDefaultHostConfigs(); err != nil {
		return fmt.Errorf("initialize default host configs failed: %w", err)
	}

	return nil
}