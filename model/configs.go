package model

type Configs struct {
	// Mq的所有配置信息
	MqConfig       MqConfig `json:"MqConfig"`

	// Secret 用来标识Provider，防止恶意调用支付服务，代表只有provider能够调用支付程序
	ProviderSecret string   `json:"ProviderSecret"`
}

// MqConfig Mq配置信息
type MqConfig struct {
	MqAk string `json:"MqAk"`
	MqSk string `json:"MqSk"`

	// MqHost 杭州
	MqHost     string `json:"MqHost"`
	InstanceId string `json:"InstanceId"`
}
