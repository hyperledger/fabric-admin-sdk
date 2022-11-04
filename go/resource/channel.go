package resource

type ChannelList struct {
	SystemChannel interface{} `json:"systemChannel"`
	Channels      []struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"channels"`
}
