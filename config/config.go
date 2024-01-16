package config

//main face
type Config struct {
	ini *IniConfig
	json *JsonConfig
}

//construct
func NewConfig(rootPaths ...string) *Config {
	//get key param
	cfgRootPath := "."
	if rootPaths != nil && len(rootPaths) > 0 {
		cfgRootPath = rootPaths[0]
	}
	//self init
	this := &Config{
		ini: NewIniConfigWithPara(cfgRootPath),
		json: NewJsonConfigWithPara(cfgRootPath),
	}
	return this
}

//get sub face
func (c *Config) GetIniConf() *IniConfig {
	return c.ini
}

func (c *Config) GetJsonConf() *JsonConfig {
	return c.json
}