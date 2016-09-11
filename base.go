/*
* (C) 2001-2015 gtlService Inc.
*
* gzkwrapper source code
* version: 1.0.0
* author: bobliu0909@gmail.com
* datetime: 2016-1-20
*
 */

package gzkwrapper

import (
	"bytes"
	"encoding/json"
	"errors"
	"runtime"
	"sync"
)

var (
	ErrKeyInvalid      = errors.New("key invalid.")
	ErrArgsInvalid     = errors.New("args invalid.")
	ErrNodeIsNull      = errors.New("node is nil.")
	ErrNodeConnInvalid = errors.New("node conn invalid.")
)

type NodeType int

const (
	NODE_SERVER NodeType = iota + 1 //服务节点
	NODE_WORKER                     //工作节点
)

func (t NodeType) String() string {
	switch t {
	case NODE_SERVER:
		return "NODE_SERVER"
	case NODE_WORKER:
		return "NODE_WORKER"
	}
	return ""
}

var buffer_pool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1<<10))
	},
}

func encode(nodedata *NodeData) ([]byte, error) {

	buffer := buffer_pool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer buffer_pool.Put(buffer)
	if err := json.NewEncoder(buffer).Encode(nodedata); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decode(b []byte) (*NodeData, error) {

	if len(b) <= 0 {
		return nil, errors.New("nodedata invalid.")
	}

	nodedata := &NodeData{}
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(nodedata); err != nil {
		return nil, err
	}
	return nodedata, nil
}

type BaseNode struct {
	NodeType `json:"type"` //节点类型
	HostName string        `json:"hostname"` //主机名称
}

type NodeData struct {
	BaseNode
	Location  string      `json:"location"`  //节点位置
	OS        string      `json:"os"`        //节点系统
	Platform  string      `json:"platform"`  //节点平台
	IpAddr    string      `json:"ipaddr"`    //网络地址
	ProcessId int         `json:"pid"`       //节点进程号
	Singin    bool        `json:"singin"`    //签到状态
	Timestamp int64       `json:"timestamp"` //心跳时间戳
	Attach    interface{} `json:"attach"`    //附加数据
}

func NewNodeData(nodetype NodeType, hostname string, location string, os string, platform string,
	ipaddr string, processid int) *NodeData {

	if os == "" {
		os = runtime.GOOS
	}

	if platform == "" {
		platform = runtime.GOARCH
	}

	return &NodeData{
		BaseNode: BaseNode{
			NodeType: nodetype,
			HostName: hostname,
		},
		Location:  location,
		OS:        os,
		Platform:  platform,
		IpAddr:    ipaddr,
		ProcessId: processid,
		Singin:    false,
		Timestamp: 0,
		Attach:    nil,
	}
}

type NodeMapper struct {
	mutex *sync.RWMutex
	keys  []string
	items map[string]*NodeData
}

func NewNodeMapper() *NodeMapper {

	return &NodeMapper{
		mutex: new(sync.RWMutex),
		keys:  make([]string, 0),
		items: make(map[string]*NodeData),
	}
}

func (mapper *NodeMapper) Count() int {

	mapper.mutex.RLock()
	defer mapper.mutex.RUnlock()
	return len(mapper.items)
}

func (mapper *NodeMapper) GetKeys() []string {

	mapper.mutex.RLock()
	defer mapper.mutex.RUnlock()
	return mapper.keys
}

func (mapper *NodeMapper) Contains(key string) bool {

	mapper.mutex.RLock()
	defer mapper.mutex.RUnlock()
	for _, k := range mapper.keys {
		if k == key {
			return true
		}
	}
	return false
}

func (mapper *NodeMapper) Get(key string) *NodeData {

	mapper.mutex.RLock()
	defer mapper.mutex.RUnlock()
	if _, ret := mapper.items[key]; !ret {
		return nil
	}
	return mapper.items[key]
}

func (mapper *NodeMapper) Copy(m map[string]*NodeData) {

	if len(m) <= 0 {
		return
	}

	mapper.mutex.Lock()
	defer mapper.mutex.Unlock()
	mapper.items = m
	mapper.keys = mapper.keys[0:0]
	for key := range mapper.items {
		mapper.keys = append(mapper.keys, key)
	}
}

func (mapper *NodeMapper) Append(key string, value *NodeData) int {

	if value == nil {
		return -1
	}

	mapper.mutex.Lock()
	defer mapper.mutex.Unlock()
	if _, ret := mapper.items[key]; !ret {
		mapper.items[key] = value
		mapper.keys = append(mapper.keys, key)
		return 0
	}
	return -1
}

func (mapper *NodeMapper) Remove(key string) int {

	mapper.mutex.Lock()
	defer mapper.mutex.Unlock()
	if _, ret := mapper.items[key]; ret {
		delete(mapper.items, key)
		for i, k := range mapper.keys {
			if k == key {
				mapper.keys = append(mapper.keys[:i], mapper.keys[i+1:]...)
				break
			}
		}
		return 0
	}
	return -1
}

func (mapper *NodeMapper) Set(key string, value *NodeData) int {

	if value == nil {
		return -1
	}

	mapper.mutex.Lock()
	defer mapper.mutex.Unlock()
	if _, ret := mapper.items[key]; ret {
		mapper.items[key] = value
		return 0
	}
	return -1
}

func (mapper *NodeMapper) Clear() {

	mapper.mutex.Lock()
	defer mapper.mutex.Unlock()
	for key := range mapper.items {
		delete(mapper.items, key)
	}
	mapper.keys = mapper.keys[0:0]
}
