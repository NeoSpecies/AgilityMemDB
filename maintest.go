package main

import (
	"AgilityMemDB/agilitymemdb"
	"fmt"
	"log"
	//"github.com/your/package/path/agilitymemdb"
)

func main() {
	// 初始化 AgilityMemDB 实例
	db := agilitymemdb.AgilityMemDB{
		FilePath: "data.json", // 数据库文件路径
		Data:     make(map[string]*agilitymemdb.Item),
	}

	// 加载数据
	err := db.LoadData()
	if err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	// 添加测试数据
	db.Put("key1", "value11231321321")
	db.Put("key2", "value2")

	// 获取数据
	value, exists := db.Get("key1")
	if exists {
		fmt.Printf("Value for key11: %s\n", value)
	} else {
		fmt.Println("Key1 does not exist.")
	}

	// 删除数据
	db.Delete("key2")

	// 开启事务
	db.BeginTransaction()

	// 添加事务操作
	db.Put("key3", "value3")
	db.Put("key4", "value4")

	// 提交事务
	db.CommitTransaction()

	// 获取数据
	value, exists = db.Get("key3")
	if exists {
		fmt.Printf("Value for key3: %s\n", value)
	} else {
		fmt.Println("Key3 does not exist.")
	}

	// 回滚事务
	db.BeginTransaction()
	db.Put("key5", "value5")
	db.Put("key6", "value6")
	db.RollbackTransaction()

	// 获取数据
	value, exists = db.Get("key5")
	if exists {
		fmt.Printf("Value for key5: %s\n", value)
	} else {
		fmt.Println("Key5 does not exist.")
	}

	// 保存数据到文件
	err = db.Persist()
	if err != nil {
		log.Fatalf("Failed to persist data: %v", err)
	}

	fmt.Println("Test completed.")
}
