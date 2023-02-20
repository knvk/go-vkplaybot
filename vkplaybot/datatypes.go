package vkplaybot

//
// Chat related
//

type Channel struct {
	URL       string `json:"blogUrl"`
	WSChannel string `json:"publicWebSocketChannel"`
	Owner     Owner  `json:"owner"`
}

type Owner struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type User struct {
	DisplayName     string        `json:"displayName"`
	HasAvatar       bool          `json:"hasAvatar"`
	AvatarURL       string        `json:"avatarUrl"`
	ID              int           `json:"id"`
	Nick            string        `json:"nick"`
	Name            string        `json:"name"`
	CreatedAt       int           `json:"createdAt"`
	Badges          []interface{} `json:"badges"`
	IsChatModerator bool          `json:"isChatModerator"`
	Subscription    struct {
		OnTime int `json:"onTime"`
	} `json:"subscription"`
}

type Viewers struct {
	Data struct {
		Count struct {
			Users         int `json:"users"`
			TemporaryBans int `json:"temporaryBans"`
			Moderators    int `json:"moderators"`
			PermanentBans int `json:"permanentBans"`
		} `json:"count"`
		Moderators []*User `json:"moderators"`
		Owner      *User   `json:"owner"`
		Users      []*User `json:"users"`
	} `json:"data"`
}

type WSMessage struct {
	Result struct {
		Channel string `json:"channel"`
		Data    struct {
			Data struct {
				Data struct {
					User       *User             `json:"user"`
					ID         int               `json:"id"`
					Author     *User             `json:"author"`
					CreatedAt  int               `json:"createdAt"`
					IsPrivate  bool              `json:"isPrivate"`
					Styles     []interface{}     `json:"styles"`
					MsgContent []*MessageContent `json:"data"`
				} `json:"data"`
				Type string `json:"type"`
			} `json:"data"`
			Offset int `json:"offset"`
		} `json:"data"`
	} `json:"result"`
}

type MessageContent struct {
	Modificator string `json:"modificator"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Name        string `json:"name"`
	URL         string `json:"url"`
}

type TextContent struct {
	Modificator string `json:"modificator"`
	Type        string `json:"type"`
	Content     string `json:"content"`
}

type MentionContent struct {
	Type        string `json:"type"`
	ID          int    `json:"id"`
	Nick        string `json:"nick"`
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
}

type Message struct {
	User  *User
	Chat  string
	Text  string
	To    string
	Smile []string
	Link  string
}

//
// Module related
//

type BaseModule struct {
	Name        string
	IsActivated bool
	Break       bool
}

type Module interface {
	Match(*WSMessage) (bool, ModuleFunc)
	GetName() string
	Stop() bool
}

type ModuleHandler interface {
	SendChatMessage(string, *Channel, *User)
	GetChannel() *Channel
	GetViewers() (*Viewers, error)
}

type ModuleFunc func(ModuleHandler, *WSMessage)
