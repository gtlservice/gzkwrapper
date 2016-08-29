/*
* (C) 2001-2015 gtlService Inc.
*
* gzkwrapper source code
* version: 1.0.0
* author: bobliu0909@gmail.com
* datetime: 2016-2-25
 */

package gzkwrapper

//对zookeeper进行封装，提供服务器自动测试与心跳维护
//Server：服务器节点，维护Worker节点信息和心跳
//Worker: 工作节点，负责注册zookeeper和心跳发送
