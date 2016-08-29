/*
* (C) 2001-2015 gtlService Inc.
*
* gzkwrapper source code
* version: 1.0.0
* author: bobliu0909@gmail.com
* datetime: 2016-1-24
*
 */

package gzkwrapper

import "github.com/gtlservice/gutils/network"

import (
	"os"
	"strings"
	"sync"
	"time"
)

type ServerArgs struct {
	Hosts     string
	Root      string
	Device    string
	Location  string
	OS        string
	Platform  string
	Pulse     string
	Timeout   string
	Threshold int
}

var server *Server
var refcache *NodeMapper
var suspicioncache *SuspicionMapper
var waitgroup sync.WaitGroup

var online = make([]*NodeInfo, 0)
var offline = make([]*NodeInfo, 0)

type Server struct {
	Key        string
	Root       string
	Pulse      time.Duration
	TimeoutSec float64
	Node       *Node
	Data       *NodeData
	Cache      *NodeMapper
	Handler    INodeNotifyHandler
	Quit       chan bool
}

func NewServer(key string, args *ServerArgs, handler INodeNotifyHandler) (*Server, error) {

	if len(strings.TrimSpace(key)) == 0 {
		return nil, ErrKeyInvalid
	}

	if args == nil {
		return nil, ErrArgsInvalid
	}

	addr, err := network.GetLocalNetDriveInfo(args.Device)
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	pulse, err := time.ParseDuration(args.Pulse)
	if err != nil {
		return nil, err
	}

	timeout, err := time.ParseDuration(args.Timeout)
	if err != nil {
		return nil, err
	}
	timeoutsec := timeout.Seconds() * float64(args.Threshold) //超时时长*阀值

	refcache = NewNodeMapper()
	suspicioncache = NewSuspicionMapper()
	server = &Server{
		Key:        key,
		Root:       args.Root,
		Pulse:      pulse,
		TimeoutSec: timeoutsec,
		Node:       NewNode(args.Hosts, handler.OnZkWrapperWatchHandlerFunc),
		Data:       NewNodeData(NODE_SERVER, hostname, args.Location, args.OS, args.Platform, addr.IP, os.Getpid()),
		Cache:      NewNodeMapper(),
		Handler:    handler,
		Quit:       make(chan bool),
	}
	return server, nil
}

func (s *Server) Open() error {

	if s.Node != nil {
		err := s.Node.Open()
		if err == nil {
			ret, err := s.Node.Exists(s.Root)
			if err != nil {
				return err
			}
			if !ret {
				s.Node.Create(s.Root, nil)
			}
			go s.pulseKeepAlive() //开启保活检测
		}
		return err
	}
	return ErrNodeIsNull
}

func (s *Server) Close() error {

	refcache.Clear()
	suspicioncache.Clear()
	if s.Node != nil {
		s.Quit <- true
		close(s.Quit)
		s.Node.Close()
		s.Cache.Clear()
		return nil
	}
	return ErrNodeIsNull
}

func (s *Server) GetLocation() string {

	return s.Data.Location
}

func (s *Server) GetOS() string {

	return s.Data.OS
}

func (s *Server) GetPlatform() string {

	return s.Data.Platform
}

func (w *Server) Watch(path string) error {

	return w.Node.Watch(path)
}

func (s *Server) Exists(path string) (bool, error) {

	return s.Node.Exists(path)
}

func (s *Server) Children(path string) ([]string, error) {

	return s.Node.Children(path)
}

func (s *Server) Get(path string) ([]byte, error) {

	return s.Node.Get(path)
}

func (s *Server) Create(path string, buffer []byte) error {

	return s.Node.Create(path, buffer)
}

func (s *Server) Remove(path string) error {

	return s.Node.Remove(path)
}

func (s *Server) Set(path string, buffer []byte) error {

	return s.Node.Set(path, buffer)
}

func (s *Server) RefreshCache() error {

	//更新本地refcache
	if err := pullRefCache(); err != nil {
		return err
	}

	lockeys := s.Cache.GetKeys()
	for i := len(lockeys) - 1; i >= 0; i-- {
		key := lockeys[i]
		if ret := refcache.Contains(key); !ret {
			waitgroup.Add(1)
			go func(k string) {
				s.Node.Remove(s.Root + "/WORKER-" + k)
				waitgroup.Done()
			}(key)
			offline = append(offline, &NodeInfo{Key: key, Data: s.Cache.Get(key)})
			s.Cache.Remove(key)
			suspicioncache.Del(key) //删除怀疑列表
			lockeys = s.Cache.GetKeys()
		}
	}
	waitgroup.Wait()

	temp_keys := make([]string, 0)
	lockeys = s.Cache.GetKeys()
	for _, key := range lockeys {
		temp_keys = append(temp_keys, key)
	}

	refkeys := refcache.GetKeys()
	for _, key := range refkeys { //合并到本地Cache
		refvalue := refcache.Get(key)
		if locvalue := s.Cache.Get(key); locvalue == nil {
			s.Cache.Append(key, refvalue)
		} else {
			if refvalue.Timestamp == locvalue.Timestamp {
				suspicioncache.Add(key) //加入怀疑列表
			} else {
				suspicioncache.Del(key) //删除怀疑列表
			}
			s.Cache.Set(key, refcache.Get(key))
		}
	}

	timestamp := time.Now().Unix()
	lockeys = s.Cache.GetKeys()
	for i := len(lockeys) - 1; i >= 0; i-- {
		key := lockeys[i]
		value := s.Cache.Get(key)
		if !value.Singin || checkTimeout(key, timestamp) { //删除本地退出或异常节点
			waitgroup.Add(1)
			go func(k string) {
				s.Node.Remove(s.Root + "/WORKER-" + k)
				waitgroup.Done()
			}(key)
			offline = append(offline, &NodeInfo{Key: key, Data: s.Cache.Get(key)})
			s.Cache.Remove(key)
			suspicioncache.Del(key) //删除怀疑列表
			lockeys = s.Cache.GetKeys()
		}
	}
	waitgroup.Wait()

	ret := false
	lockeys = s.Cache.GetKeys()
	for _, key := range lockeys { //找出新加入节点
		ret = false
		for _, k := range temp_keys {
			if key == k {
				ret = true
				break
			}
		}
		if !ret {
			online = append(online, &NodeInfo{Key: key, Data: s.Cache.Get(key)})
		}
	}

	if len(online) > 0 || len(offline) > 0 {
		s.Handler.OnZkWrapperNodeHandlerFunc(online, offline)
		online = online[0:0]
		offline = offline[0:0]
	}
	return nil
}

func pullRefCache() error {

	refcache.Clear()
	keys, err := server.Node.Children(server.Root)
	if err != nil {
		return err
	}

	for i := 0; i < len(keys); i++ {
		if !strings.HasPrefix(keys[i], "WORKER-") {
			continue
		}
		waitgroup.Add(1)
		go func(key string) { //根据节点名称获取节点数据并筛选WORKER类型节点
			if buffer, err := server.Node.Get(server.Root + "/" + key); err == nil {
				if value, err := decode(buffer); err == nil && value.NodeType == NODE_WORKER {
					refcache.Append(strings.TrimPrefix(key, "WORKER-"), value)
				}
			}
			waitgroup.Done()
		}(keys[i])
	}
	waitgroup.Wait()
	return nil
}

func checkTimeout(key string, timestamp int64) bool {

	jointimestamp := suspicioncache.Get(key)
	if jointimestamp == 0 {
		return false
	}

	seedt := time.Unix(timestamp, 0)
	nodet := time.Unix(jointimestamp, 0)
	diffsec := seedt.Sub(nodet).Seconds()
	if diffsec < server.TimeoutSec {
		return false
	}
	return true
}

func (s *Server) pulseKeepAlive() {

	var quit bool = false
NEW_TICK_DURATION:
	ticker := time.NewTicker(s.Pulse)
	for !quit {
		select {
		case <-s.Quit: //退出
			{
				ticker.Stop()
				quit = true
			}
		case <-ticker.C: //检测node心跳
			{
				ticker.Stop()
				s.RefreshCache()
				s.Handler.OnZkWrapperPulseHandlerFunc(s.Key, s.Data, nil)
				goto NEW_TICK_DURATION
			}
		}
	}
}
