package main

import (
	"sync"
)

var manager sync.Map
var connNumber int
// store 存储
func store(uuid string, cl *Client) {
	manager.Store(uuid, cl)
	connNumber = connNumber+1
}

// load 获取
func load(uuid string) *Client {
	value, ok := manager.Load(uuid)
	if ok {
		return value.(*Client)
	}
	return nil
}

// delete 删除
func delete(uuid string) {
	connNumber = connNumber-1
	manager.Delete(uuid)
}
