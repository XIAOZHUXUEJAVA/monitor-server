package main

import (
	"encoding/json"
	"fmt"
	"log"
	"monitor-server/internal/config"
	"monitor-server/internal/database"
	"monitor-server/internal/model"
	"monitor-server/internal/repository"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化数据库
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 创建 repository
	hostRepo := repository.NewHostRepository(db.DB)

	fmt.Println("=== Testing Host API ===")

	// 1. 测试创建主机
	fmt.Println("\n1. Creating test host...")
	testHost := &model.Host{
		Hostname:          "test-server-01",
		DisplayName:       "Test Server 01",
		IPAddress:         "192.168.1.100",
		Environment:       "testing",
		Location:          "Test Lab",
		Description:       "Test server for debugging",
		Status:            "online",
		MonitoringEnabled: true,
		OS:                "Ubuntu",
		Platform:          "22.04",
		CPUCores:          2,
		TotalMemory:       4096,
		Agent:             true,
	}

	err = hostRepo.Create(testHost)
	if err != nil {
		fmt.Printf("Error creating host: %v\n", err)
	} else {
		fmt.Printf("Created host with ID: %d\n", testHost.ID)
	}

	// 2. 测试查询主机列表
	fmt.Println("\n2. Querying host list...")
	hosts, total, err := hostRepo.List(0, 10)
	if err != nil {
		fmt.Printf("Error querying hosts: %v\n", err)
	} else {
		fmt.Printf("Total hosts: %d\n", total)
		for _, host := range hosts {
			hostJSON, _ := json.MarshalIndent(host, "", "  ")
			fmt.Printf("Host: %s\n", hostJSON)
		}
	}

	// 3. 测试查询统计信息
	fmt.Println("\n3. Testing basic host count...")
	fmt.Printf("Found %d hosts in database\n", total)

	fmt.Println("\n=== Test completed ===")
}