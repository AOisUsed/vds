package virtual_device

// 虚拟设备属性
type Flag struct {
	Mode             ModeType        // 模式
	WorkMode         WorkModeType    // 工作模式
	FixedFrequency   uint16          // 定频
	HoppingFrequency uint8           // 跳频
	ChannelNumber    uint8           //信道号
	CryptoMode       CryptoModeType  //保密方式
	NoiseMode        NoiseModeType   //静噪方式
	NoiseLevel       uint8           //静噪等级
	CallMode         CallModeType    //呼叫方式
	CallNumber       uint8           //呼叫号码
	NoiseSwitch      NoiseSwitchType //静噪开关
	NetworkMode      NetworkModeType //网络模式
	VoiceEncoding    string          //话音编码
	OptionText       string          //选项按钮名称

	KeySwitch      KeySwitchType   //明密切换
	KeyAreaCode    KeyAreaCodeType //密钥区号
	KeyGroupNumber uint8           //密钥组号
	KeyType        KeyTypeType     //注钥类型
	// 添加密钥组号位置跟踪字段: 0表示设置十位，1表示设置个位
	KeyGroupNumberPosition uint8
	// 添加临时模式变量，用于模式选择菜单
	TempMode        ModeType
	TempKeySwitch   KeySwitchType   // 临时明密切换值
	TempKeyAreaCode KeyAreaCodeType // 临时密钥区号值
	PowerLevel      uint8           // 功率级别：0-低功率，1-高功率
	VolumeLevel     uint8           // 音量级别：0-10
}

type ModeType uint8

const (
	Mode504 ModeType = iota
	Mode503
	Mode121
	CenterControl // 中控
)

type WorkModeType uint8

const (
	FixedFrequency   WorkModeType = iota // 定频
	HoppingFrequency                     // 跳频
)

// KeySwitch      uint8 //明密切换
type KeySwitchType uint8

const (
	KeySwitchOn  KeySwitchType = iota // 明
	KeySwitchOff                      // 密
)

// KeyAreaCode    uint8 //密钥区号
type KeyAreaCodeType uint8

const (
	KeyAreaCodeA KeyAreaCodeType = iota // A区
	KeyAreaCodeB                        // B区
)

// KeyType        uint8 //注钥类型
type KeyTypeType uint8

const (
	KeyType513A           KeyTypeType = iota // 513A区
	KeyType513B                              // 513B区
	KeyTypeCollaborativeA                    // 协同A区
	KeyTypeCollaborativeB                    // 协同B区
)

// NetworkMode      uint8        //网络模式
type NetworkModeType uint8

const (
	NetworkModePointToPoint NetworkModeType = iota // 点对点
	NetworkModeBattleNet                           //战斗网
)

// NoiseMode        uint8        //静噪方式
type NoiseModeType uint8

const (
	NoiseModeNoise NoiseModeType = iota // 噪声
	//NoiseModeSilence                    // 静音
)

// CallMode         uint8        //呼叫方式
type CallModeType uint8

const (
	CallModeGroup CallModeType = iota // 群呼
	// CallModeSingle                    // 单呼
)

// CryptoMode       uint8        //保密方式
type CryptoModeType uint8

const (
	CryptoModeNavy   CryptoModeType = iota // 海军
	CryptoModeCommon                       // 通装
)

type NoiseSwitchType uint8

const (
	NoiseSwitchOn  NoiseSwitchType = iota // 开
	NoiseSwitchOff                        // 关
)
