package asyncapi

type Document struct {
	AsyncAPI   string                 `yaml:"asyncapi"`
	Info       Info                   `yaml:"info"`
	Channels   map[string]Channel     `yaml:"channels"`
	Operations map[string]Operation   `yaml:"operations"`
	Components Components             `yaml:"components"`
}

type Info struct {
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

type Channel struct {
	Address     string                 `yaml:"address"`
	Messages    map[string]interface{} `yaml:"messages"`
}

type Operation struct {
	Action      string        `yaml:"action"`
	ChannelRef  Ref           `yaml:"channel"`
	XSteps      []interface{} `yaml:"x-integron-steps"`
}

type Ref struct {
	Ref string `yaml:"$ref"`
}

type Components struct {
	Messages map[string]interface{} `yaml:"messages"`
	Schemas  map[string]interface{} `yaml:"schemas"`
}

type ResolvedOperation struct {
	OperationKey string
	Action       string
	Topic        string
	Steps        []interface{}
}
