package main

import (
	"AgilityMemDB/agilitymemdb"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"log"
	"net/http"
	"os"
	"runtime"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var db agilitymemdb.AgilityMemDB

func main() {
	// 初始化 AgilityMemDB 实例
	db = agilitymemdb.AgilityMemDB{
		FilePath: "data.json", // 数据库文件路径
		Data:     make(map[string]*agilitymemdb.Item),
	}
	// 加载数据
	err := db.LoadData()
	if err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	// 设置路由和处理程序
	http.HandleFunc("/get", handleGet)
	http.HandleFunc("/put", handlePut)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/begin", handleBeginTransaction)
	http.HandleFunc("/commit", handleCommitTransaction)
	http.HandleFunc("/rollback", handleRollbackTransaction)
	http.HandleFunc("/persist", handlePersist) // 添加持久化处理程序
	// 启动HTTP服务器
	printLogoAndInfo()
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	// 在程序退出前将数据持久化到磁盘
	err = db.Persist()
	if err != nil {
		log.Fatalf("Failed to persist data: %v", err)
	}
}

func printLogoAndInfo() {
	logo := `
                        __  __  
 /\  _ .|.|_  |\/| _ _ |  \|__) 
/--\(_)||||_\/|  |(-||||__/|__) 
    _/      /        
`
	fmt.Println(logo)
	port := "8080"
	apiPath := "/api/v1"
	jsonPath := "/data/db.json"
	infoTable := [][]string{
		{"Total RAM", fmt.Sprintf("%d MB", getTotalRAM())},
		{"Free RAM", fmt.Sprintf("%d MB", getFreeRAM())},
		{"CPUs", fmt.Sprintf("%d", getCPUCount())},
		{"Port", fmt.Sprintf(port)},
		{"API Path", fmt.Sprintf(apiPath)},
		{"JSON Path", fmt.Sprintf(jsonPath)},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Category", "Value"})
	table.SetBorder(true)

	table.AppendBulk(infoTable)
	table.Render()
}

func getTotalRAM() int {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return int(memStats.TotalAlloc / 1024 / 1024)
}

func getFreeRAM() int {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return int(memStats.HeapReleased / 1024 / 1024)
}

func getCPUCount() int {
	return runtime.NumCPU()
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	key := r.URL.Query().Get("key")

	// 获取数据
	value, exists := db.Get(key)
	if exists {
		// 返回响应
		response := Response{
			Success: true,
			Message: "Data retrieved successfully",
			Data: struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}{key, value},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		// 返回404 Not Found错误
		response := Response{
			Success: false,
			Message: "Data not found",
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
	}
}

func handlePut(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var data struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		response := Response{
			Success: false,
			Message: fmt.Sprintf("Failed to parse request body: %v", err),
		}

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 开启事务
	err = db.BeginTransaction()
	if err != nil {
		response := Response{
			Success: false,
			Message: fmt.Sprintf("Failed to begin transaction: %v", err),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 添加事务操作
	err = db.Put(data.Key, data.Value)
	if err != nil {
		response := Response{
			Success: false,
			Message: fmt.Sprintf("Failed to put data: %v", err),
		}

		// 回滚事务
		db.RollbackTransaction()

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 提交事务
	err = db.CommitTransaction()
	if err != nil {
		response := Response{
			Success: false,
			Message: fmt.Sprintf("Failed to commit transaction: %v", err),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 返回响应
	response := Response{
		Success: true,
		Message: "Data added successfully",
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	key := r.URL.Query().Get("key")

	// 开启事务
	err := db.BeginTransaction()
	if err != nil {
		response := Response{
			Success: false,
			Message: fmt.Sprintf("Failed to begin transaction: %v", err),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 添加事务操作
	db.Delete(key)

	// 提交事务
	err = db.CommitTransaction()
	if err != nil {
		response := Response{
			Success: false,
			Message: fmt.Sprintf("Failed to commit transaction: %v", err),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 返回响应
	response := Response{
		Success: true,
		Message: "Data deleted successfully",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleBeginTransaction(w http.ResponseWriter, r *http.Request) {
	// 开启事务
	err := db.BeginTransaction()
	if err != nil {
		response := Response{
			Success: false,
			Message: fmt.Sprintf("Failed to begin transaction: %v", err),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 返回响应
	response := Response{
		Success: true,
		Message: "Transaction began successfully",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleCommitTransaction(w http.ResponseWriter, r *http.Request) {
	// 提交事务
	err := db.CommitTransaction()
	if err != nil {
		response := Response{
			Success: false,
			Message: fmt.Sprintf("Failed to commit transaction: %v", err),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 返回响应
	response := Response{
		Success: true,
		Message: "Transaction committed successfully",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleRollbackTransaction(w http.ResponseWriter, r *http.Request) {
	// 回滚事务
	db.RollbackTransaction()

	// 返回响应
	response := Response{
		Success: true,
		Message: "Transaction rolled back successfully",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handlePersist 处理持久化请求
func handlePersist(w http.ResponseWriter, r *http.Request) {
	// 调用 Persist() 方法将数据持久化到磁盘
	err := db.Persist()
	if err != nil {
		response := Response{
			Success: false,
			Message: fmt.Sprintf("Failed to persist data: %v", err),
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 返回响应
	response := Response{
		Success: true,
		Message: "Data persisted successfully",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
