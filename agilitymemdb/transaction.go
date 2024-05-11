package agilitymemdb

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
)

// 生成唯一的事务ID
func generateTransactionID() string {
	return uuid.New().String()
}
func (db *AgilityMemDB) BeginTransaction() error {
	db.lock.Lock()
	defer db.lock.Unlock()

	// 检查是否存在事务
	if db.transaction != nil {
		return errors.New("transaction already in progress")
	}

	// 创建一个新的事务结构体，保存事务相关的信息
	transaction := &Transaction{
		ID:         generateTransactionID(),
		Operations: make(map[string]*Item),
	}

	// 将当前事务保存到 AgilityMemDB 结构体中
	db.transaction = transaction

	return nil
}

func (db *AgilityMemDB) CommitTransaction() error {
	db.lock.Lock()
	defer db.lock.Unlock()

	// 检查是否存在事务
	if db.transaction == nil {
		return errors.New("no active transaction")
	}

	// 提交事务
	for key, item := range db.transaction.Operations {
		db.Data[key] = item
	}

	// 清除事务信息
	db.transaction = nil

	// 返回nil表示没有错误
	return nil
}

func (db *AgilityMemDB) RollbackTransaction() {
	db.lock.Lock()
	defer db.lock.Unlock()

	// 检查是否存在事务
	if db.transaction == nil {
		fmt.Println("No active transaction.")
		return
	}

	// 回滚事务
	for key := range db.transaction.Operations {
		delete(db.Data, key)
	}

	// 清除事务信息
	db.transaction = nil

	fmt.Println("Transaction rolled back successfully.")
}
