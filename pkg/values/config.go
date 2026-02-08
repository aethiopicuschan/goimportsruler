package values

type configFileName string
type ext string

const (
	jsonExt ext = ".json"
	yamlExt ext = ".yaml"
	ymlExt  ext = ".yml"
)

func (e ext) string() string {
	return string(e)
}

func (c configFileName) Hidden() configFileName {
	return configFileName("." + c.String())
}

func (c configFileName) Ext(ext ext) configFileName {
	return configFileName(c.String() + ext.string())
}

func (c configFileName) String() string {
	return string(c)
}

const (
	configFileNamePrefix = "goimportsruler"
)

func GetAllConfigFileNames() []string {
	return []string{
		configFileName(configFileNamePrefix).Ext(jsonExt).String(),
		configFileName(configFileNamePrefix).Ext(yamlExt).String(),
		configFileName(configFileNamePrefix).Ext(ymlExt).String(),
		configFileName(configFileNamePrefix).Ext(jsonExt).Hidden().String(),
		configFileName(configFileNamePrefix).Ext(yamlExt).Hidden().String(),
		configFileName(configFileNamePrefix).Ext(ymlExt).Hidden().String(),
	}
}
