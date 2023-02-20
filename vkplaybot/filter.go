package vkplaybot

import (
	"strings"
)

type FilterModule struct {
	BaseModule
	//name     string
	banwords []string
}

func NewFilterModule(censored []string) *FilterModule {
	return &FilterModule{
		BaseModule: BaseModule{
			Name:        "filter",
			IsActivated: true,
			Break:       true,
		},
		//name:     "filter",
		banwords: censored,
	}
}

// linear matches message with "bad" words and punish LUL
func (m *FilterModule) Match(msg *WSMessage) (found bool, fn ModuleFunc) {
	b := msg.ParseMessage()
	if b.Text == "" {
		return false, nil
	}
	for i := 0; i < len(m.banwords); i++ {
		if strings.Contains(b.Text, m.banwords[i]) {
			return true, Punish
		}
	}
	return false, nil
}

// after punishment no longer proceed
func (m *FilterModule) Stop() bool {
	return m.Break
}

func (m *FilterModule) GetName() string {
	return m.Name
}

func Punish(w ModuleHandler, m *WSMessage) {
	text := `осуждаю! нельзя так говорить в этом чате`
	w.SendChatMessage(text, w.GetChannel(), m.Result.Data.Data.Data.User)
}
