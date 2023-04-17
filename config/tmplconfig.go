package config

type TemplateConfigFile struct {
	Target struct {
		RelativeDir struct {
			Proto  []string `yaml:"proto"`
			Client []string `yaml:"client"`
			Server []string `yaml:"server"`
		} `yaml:"relativeDir"`
	} `yaml:"target"`
	Template struct {
		FilesFormatSuffix string `yaml:"filesFormatSuffix"`
		ProtoFilename     string `yaml:"protoFilename"`
		RelativeDir       struct {
			Proto  []string `yaml:"proto"`
			Client []string `yaml:"client"`
			Server []string `yaml:"server"`
		} `yaml:"relativeDir"`
	} `yaml:"template"`
}
