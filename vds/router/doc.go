// Package router 包括了消息入口路由，消息集合器,
//
// 消息进入vds后，由进站路由ingress router转发到对应的虚拟设备中,
//
// vds中虚拟设备发送的消息由消息集合器aggregator发送到dispatcher处理，
package router
