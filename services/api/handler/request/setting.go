package request

type SettingReq struct {
	Group      string `json:"group"`
	Key        string `json:"key"`
	Label      string `json:"label"`
	SetType    string `json:"set_type"`
	SetContent string `json:"set_content"`
}
