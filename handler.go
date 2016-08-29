/*
* (C) 2001-2015 gtlService Inc.
*
* gzkwrapper source code
* version: 1.0.0
* author: bobliu0909@gmail.com
* datetime: 2016-2-22
*
 */

package gzkwrapper

type NodeInfo struct {
	Key  string
	Data *NodeData
}

type INodeNotifyHandler interface {
	OnZkWrapperPulseHandlerFunc(key string, nodedata *NodeData, err error)
	OnZkWrapperNodeHandlerFunc(online []*NodeInfo, offline []*NodeInfo)
	OnZkWrapperWatchHandlerFunc(path string, data []byte, err error)
}

type PulseHandlerFunc func(key string, nodedata *NodeData, err error)
type NodeHandlerFunc func(online []*NodeInfo, offline []*NodeInfo)
type WatchHandlerFunc func(path string, data []byte, err error)

func (fn PulseHandlerFunc) OnZkWrapperPulseHandlerFunc(key string, nodedata *NodeData, err error) {
	fn(key, nodedata, err)
}

func (fn NodeHandlerFunc) OnZkWrapperNodeHandlerFunc(online []*NodeInfo, offline []*NodeInfo) {
	fn(online, offline)
}

func (fn WatchHandlerFunc) OnZkWrapperWatchHandlerFunc(path string, data []byte, err error) {
	fn(path, data, err)
}
