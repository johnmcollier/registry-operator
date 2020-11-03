package config

// ControllerCfg logic borrowed from https://github.com/devfile/devworkspace-operator/blob/master/pkg/config/config.go
var ControllerCfg ControllerConfig

type ControllerConfig struct {
	isOpenShift bool
}

func (c *ControllerConfig) IsOpenShift() bool {
	return c.isOpenShift
}

func (c *ControllerConfig) SetIsOpenShift(isOpenShift bool) {
	c.isOpenShift = isOpenShift
}
