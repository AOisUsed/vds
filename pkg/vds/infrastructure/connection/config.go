package connection

// Configurable 这是 connection 类型的一个可选接口，用于判断该连接是否可以获取配置
//
// 这个接口的作用是将一个连接 (如tcp, udp 等)的信息作为配置文件输出，这样可以持久化保存在数据库中，以供获取
//
// 在本项目中，一个使用场景是,如果一个vds的connection是以udp实现的，我们需要把这个udp的配置存在注册中心中，
// 以供其他vds获取，并以此为依据建立连接。
//
// mock connection 不需要实现这个接口，因为没有持久化需求
type Configurable interface {
	Config() *Config
}

// Config 连接信息， 用于序列化
type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`

	LocalHost string `json:"local_host,omitempty"` // 例如: "0.0.0.0" 或特定网卡 IP
	LocalPort int    `json:"local_port,omitempty"` // 例如: 6000

	// 连接类型 如 “udp”
	Type string `json:"type"`
}
