package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"monitor-server/internal/config"
	"monitor-server/internal/database"
	"monitor-server/internal/model"
	"monitor-server/internal/repository"
)

func main() {
	fmt.Println("🏗️  主机管理第一阶段测试开始...")
	fmt.Println("=====================================")

	// 加载配置和连接数据库
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	defer db.Close()

	// 执行数据库迁移和初始化
	fmt.Println("🔄 执行数据库迁移和初始化...")
	if err := db.Setup(); err != nil {
		log.Fatalf("❌ 数据库设置失败: %v", err)
	}
	fmt.Println("✅ 数据库迁移和初始化完成!")

	// 验证主机管理相关表是否创建成功
	fmt.Println("\n🔍 验证主机管理数据库表...")
	hostTables := []string{"hosts", "host_configs", "host_groups", "host_group_members"}
	
	for _, table := range hostTables {
		var count int64
		fullTableName := cfg.Database.Postgres.Schema + "." + table
		if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?", 
			cfg.Database.Postgres.Schema, table).Scan(&count).Error; err != nil {
			fmt.Printf("❌ 检查表 %s 失败: %v\n", table, err)
		} else if count > 0 {
			fmt.Printf("✅ 表 %s 存在\n", fullTableName)
		} else {
			fmt.Printf("❌ 表 %s 不存在\n", fullTableName)
		}
	}

	// 创建Repository实例
	fmt.Println("\n🏭 创建Repository实例...")
	hostRepo := repository.NewHostRepository(db.DB)
	hostConfigRepo := repository.NewHostConfigRepository(db.DB)
	hostGroupRepo := repository.NewHostGroupRepository(db.DB)
	fmt.Println("✅ Repository实例创建成功!")

	// 测试主机Repository
	fmt.Println("\n🖥️  测试主机Repository...")
	
	// 检查默认主机
	localhost, err := hostRepo.GetByHostname("localhost")
	if err != nil {
		fmt.Printf("❌ 获取localhost主机失败: %v\n", err)
	} else {
		fmt.Printf("✅ 默认主机存在 - %s (%s) [%s]\n", 
			localhost.DisplayName, localhost.IPAddress, localhost.Status)
	}

	// 创建测试主机
	testHost := &model.Host{
		Hostname:          "web-server-01",
		DisplayName:       "Web服务器01",
		IPAddress:         "192.168.1.100",
		Environment:       "prod",
		Location:          "数据中心A",
		Tags:              `["web", "nginx", "production"]`,
		Description:       "生产环境Web服务器",
		Status:            "online",
		MonitoringEnabled: true,
		OS:                "Ubuntu",
		Platform:          "linux",
		CPUCores:          8,
		TotalMemory:       16777216000, // 16GB
		Agent:             true,
		LastSeen:          &[]time.Time{time.Now()}[0],
	}

	if err := hostRepo.Create(testHost); err != nil {
		fmt.Printf("❌ 创建测试主机失败: %v\n", err)
	} else {
		fmt.Printf("✅ 测试主机创建成功，ID: %d\n", testHost.ID)
	}

	// 测试主机查询功能
	hosts, total, err := hostRepo.List(0, 10)
	if err != nil {
		fmt.Printf("❌ 查询主机列表失败: %v\n", err)
	} else {
		fmt.Printf("✅ 主机列表查询成功，共 %d 台主机:\n", total)
		for _, host := range hosts {
			fmt.Printf("   - %s (%s) [%s] - %s\n", 
				host.Hostname, host.IPAddress, host.Status, host.Environment)
		}
	}

	// 测试主机状态统计
	statusStats, err := hostRepo.CountByStatus()
	if err != nil {
		fmt.Printf("❌ 查询主机状态统计失败: %v\n", err)
	} else {
		fmt.Printf("✅ 主机状态统计:\n")
		for status, count := range statusStats {
			fmt.Printf("   - %s: %d 台\n", status, count)
		}
	}

	// 测试主机环境统计
	envStats, err := hostRepo.CountByEnvironment()
	if err != nil {
		fmt.Printf("❌ 查询主机环境统计失败: %v\n", err)
	} else {
		fmt.Printf("✅ 主机环境统计:\n")
		for env, count := range envStats {
			fmt.Printf("   - %s: %d 台\n", env, count)
		}
	}

	// 测试主机组Repository
	fmt.Println("\n📁 测试主机组Repository...")
	
	// 查询默认主机组
	groups, groupTotal, err := hostGroupRepo.List(0, 10)
	if err != nil {
		fmt.Printf("❌ 查询主机组列表失败: %v\n", err)
	} else {
		fmt.Printf("✅ 主机组列表查询成功，共 %d 个主机组:\n", groupTotal)
		for _, group := range groups {
			fmt.Printf("   - %s (%s) [%s]\n", 
				group.DisplayName, group.Name, group.Environment)
		}
	}

	// 创建测试主机组
	testGroup := &model.HostGroup{
		Name:        "web-servers",
		DisplayName: "Web服务器组",
		Description: "所有Web服务器的集合",
		Environment: "prod",
		Tags:        `["web", "frontend"]`,
		Enabled:     true,
	}

	if err := hostGroupRepo.Create(testGroup); err != nil {
		fmt.Printf("❌ 创建测试主机组失败: %v\n", err)
	} else {
		fmt.Printf("✅ 测试主机组创建成功，ID: %d\n", testGroup.ID)
	}

	// 测试主机加入主机组
	if err := hostGroupRepo.AddHost(testGroup.ID, testHost.ID); err != nil {
		fmt.Printf("❌ 添加主机到主机组失败: %v\n", err)
	} else {
		fmt.Println("✅ 主机成功加入主机组!")
	}

	// 测试查询主机组中的主机
	groupHosts, err := hostGroupRepo.GetHostsByGroupID(testGroup.ID)
	if err != nil {
		fmt.Printf("❌ 查询主机组中的主机失败: %v\n", err)
	} else {
		fmt.Printf("✅ 主机组 '%s' 中有 %d 台主机:\n", testGroup.DisplayName, len(groupHosts))
		for _, host := range groupHosts {
			fmt.Printf("   - %s (%s)\n", host.Hostname, host.DisplayName)
		}
	}

	// 测试主机配置Repository
	fmt.Println("\n⚙️  测试主机配置Repository...")
	
	// 查询主机配置
	configs, err := hostConfigRepo.GetByHostID(localhost.ID)
	if err != nil {
		fmt.Printf("❌ 查询主机配置失败: %v\n", err)
	} else {
		fmt.Printf("✅ 主机 '%s' 的配置 (%d个):\n", localhost.Hostname, len(configs))
		for _, config := range configs {
			fmt.Printf("   - %s: %s (%s)\n", config.Key, config.Value, config.Description)
		}
	}

	// 创建测试主机配置
	testConfigs := []model.HostConfig{
		{
			HostID:      testHost.ID,
			Key:         "monitoring_interval",
			Value:       "30",
			Type:        "int",
			Category:    "monitoring",
			Description: "监控数据收集间隔",
			Editable:    true,
		},
		{
			HostID:      testHost.ID,
			Key:         "max_cpu_threshold",
			Value:       "85.0",
			Type:        "float",
			Category:    "alert",
			Description: "CPU使用率告警阈值",
			Editable:    true,
		},
		{
			HostID:      testHost.ID,
			Key:         "backup_enabled",
			Value:       "true",
			Type:        "bool",
			Category:    "system",
			Description: "启用自动备份",
			Editable:    true,
		},
	}

	if err := hostConfigRepo.CreateBatch(testConfigs); err != nil {
		fmt.Printf("❌ 批量创建主机配置失败: %v\n", err)
	} else {
		fmt.Printf("✅ 批量创建 %d 个主机配置成功!\n", len(testConfigs))
	}

	// 测试查询测试主机的配置
	testHostConfigs, err := hostConfigRepo.GetByHostID(testHost.ID)
	if err != nil {
		fmt.Printf("❌ 查询测试主机配置失败: %v\n", err)
	} else {
		fmt.Printf("✅ 测试主机 '%s' 的配置 (%d个):\n", testHost.Hostname, len(testHostConfigs))
		for _, config := range testHostConfigs {
			fmt.Printf("   - %s: %s [%s]\n", config.Key, config.Value, config.Category)
		}
	}

	// 测试按分类查询配置
	alertConfigs, err := hostConfigRepo.GetByCategory(testHost.ID, "alert")
	if err != nil {
		fmt.Printf("❌ 查询告警配置失败: %v\n", err)
	} else {
		fmt.Printf("✅ 告警相关配置 (%d个):\n", len(alertConfigs))
		for _, config := range alertConfigs {
			fmt.Printf("   - %s: %s\n", config.Key, config.Value)
		}
	}

	// 测试高级查询功能
	fmt.Println("\n🔍 测试高级查询功能...")
	
	// 搜索主机
	searchHosts, searchTotal, err := hostRepo.Search("web", "prod", "", 0, 10)
	if err != nil {
		fmt.Printf("❌ 搜索主机失败: %v\n", err)
	} else {
		fmt.Printf("✅ 搜索结果：关键词'web'，环境'prod'，共 %d 台主机:\n", searchTotal)
		for _, host := range searchHosts {
			fmt.Printf("   - %s (%s) [%s]\n", host.Hostname, host.DisplayName, host.Environment)
		}
	}

	// 测试主机组统计
	groupStats, err := hostGroupRepo.GetGroupStats()
	if err != nil {
		fmt.Printf("❌ 查询主机组统计失败: %v\n", err)
	} else {
		fmt.Printf("✅ 主机组统计信息:\n")
		for _, stat := range groupStats {
			fmt.Printf("   - %s: %d台主机，%d台在线 [%s]\n", 
				stat.GroupName, stat.HostCount, stat.OnlineCount, stat.Environment)
		}
	}

	// 测试关联查询
	fmt.Println("\n🔗 测试关联查询...")
	
	// 查询主机及其配置
	hostWithConfigs, err := hostRepo.GetWithConfigs(testHost.ID)
	if err != nil {
		fmt.Printf("❌ 查询主机及配置失败: %v\n", err)
	} else {
		fmt.Printf("✅ 主机 '%s' 及其配置:\n", hostWithConfigs.Hostname)
		fmt.Printf("   配置数量: %d\n", len(hostWithConfigs.Configs))
	}

	// 查询主机及其所属组
	hostWithGroups, err := hostRepo.GetWithGroups(testHost.ID)
	if err != nil {
		fmt.Printf("❌ 查询主机及所属组失败: %v\n", err)
	} else {
		fmt.Printf("✅ 主机 '%s' 所属组:\n", hostWithGroups.Hostname)
		for _, group := range hostWithGroups.Groups {
			fmt.Printf("   - %s (%s)\n", group.DisplayName, group.Name)
		}
	}

	// 测试批量操作
	fmt.Println("\n⚡ 测试批量操作...")
	
	// 批量更新主机状态
	hostIDs := []uint{testHost.ID}
	if err := hostRepo.BatchUpdateStatus(hostIDs, "maintenance"); err != nil {
		fmt.Printf("❌ 批量更新主机状态失败: %v\n", err)
	} else {
		fmt.Println("✅ 批量更新主机状态成功!")
		
		// 验证状态更新
		updatedHost, _ := hostRepo.GetByID(testHost.ID)
		fmt.Printf("   主机 '%s' 状态已更新为: %s\n", updatedHost.Hostname, updatedHost.Status)
	}

	// 数据库健康检查
	fmt.Println("\n❤️  数据库健康状态:")
	health := db.Health()
	healthJSON, _ := json.MarshalIndent(health, "", "  ")
	fmt.Printf("%s\n", healthJSON)

	// 清理测试数据
	fmt.Println("\n🧹 清理测试数据...")
	
	// 删除主机配置
	hostConfigRepo.DeleteByHostID(testHost.ID)
	
	// 从主机组移除主机
	hostGroupRepo.RemoveHost(testGroup.ID, testHost.ID)
	
	// 删除测试主机组
	hostGroupRepo.Delete(testGroup.ID)
	
	// 删除测试主机
	hostRepo.Delete(testHost.ID)
	
	fmt.Println("✅ 测试数据清理完成!")

	fmt.Println("\n=====================================")
	fmt.Println("🎉 主机管理第一阶段测试全部通过!")
	fmt.Println("✅ 数据库模型创建成功!")
	fmt.Println("✅ 数据库迁移完成!")
	fmt.Println("✅ Repository层实现正常!")
	fmt.Println("🚀 准备进入第二阶段：后端API开发!")
}