// Package router 包括了进站路由，出站路由和消息分发器三个部分,
//
// 消息进入vds后，由进站路由ingress router转发到对应的虚拟设备中,
//
// vds中虚拟设备发送的消息由出站路由egress router发送到统一出口，
//
// 消息分发器接收vds统一出口消息，并根据注册中心信息转发到对应的vds中
package router
