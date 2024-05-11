package agilitymemdb

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
)

type AgilityMemDB struct {
	Data        map[string]*Item
	lock        sync.RWMutex
	transaction *Transaction
	FilePath    string // 新增字段：文件路径
}

// LoadData 从文件中加载数据
func (db *AgilityMemDB) LoadData() error {
	// 打开文件进行读取
	file, err := os.Open(db.FilePath)
	if os.IsNotExist(err) {
		// 如果文件不存在，则忽略错误
		return nil
	} else if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			err = cerr
		}
	}()

	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// 解析JSON数据
	err = json.Unmarshal(data, &db.Data)
	if err != nil {
		return err
	}

	return nil
}

func (db *AgilityMemDB) Get(key string) (string, bool) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	item, ok := db.Data[key]
	if !ok {
		return "", false
	}
	return item.Value, true
}

func (db *AgilityMemDB) Put(key string, value string) error {
	// 在适当的位置返回错误
	if db.transaction == nil {
		return errors.New("no active transaction")
	}

	// 执行事务操作
	db.transaction.Operations[key] = &Item{Value: value}

	// 返回nil表示没有错误
	return nil
}

func (db *AgilityMemDB) Delete(key string) {
	db.lock.Lock()
	defer db.lock.Unlock()

	delete(db.Data, key)
}

// Persist 将数据持久化到文件
func (db *AgilityMemDB) Persist() error {
	// 将内存数据库中的数据转换为JSON格式
	jsonData, err := json.Marshal(db.Data)
	if err != nil {
		return err
	}

	// 打开文件进行写入
	file, err := os.Create(db.FilePath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// 写入JSON数据到文件
	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}
