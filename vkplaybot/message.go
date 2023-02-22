package vkplaybot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// returns viewers list according to "native" data scheme
func (bot *VKPlayBot) GetViewers() (*Viewers, error) {
	r, err := http.NewRequest("GET", fmt.Sprintf(`https://api.vkplay.live/v1/blog/%s/`+
		`public_video_stream/chat/user/?with_bans=true`, bot.Channel.URL), nil)
	if err != nil {
		log.Printf("[Error] %s\n", err)
		return nil, err
	}
	r.Header.Add("Origin", "https://vkplay.live")
	r.Header.Add("Referer", fmt.Sprintf("https://vkplay.live/%s", bot.Channel.URL))
	r.Header.Add("Authorization", "Bearer "+bot.token.AccessToken)
	r.Header.Add("X-From-Id", bot.ClientID)
	resp, err := bot.DoReq(r)
	if err != nil {
		log.Printf("[Error] %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	viewers := &Viewers{}
	json.Unmarshal(rb, &viewers)
	return viewers, nil
}

// ReadChatMessage returns WSMessage struct which represent chat message
// its blocking so after SIGTERM(^C) it stuck for some time before actually closing
// only chat messages are supported rn
func (bot *VKPlayBot) ReadChatMessage() (m *WSMessage, err error) {
	msgRaw, err := bot.ReadWSMessage()
	if err != nil {
		return nil, err
	}
	m = &WSMessage{}
	json.Unmarshal(msgRaw, &m)
	if len(m.Result.Channel) > 0 {
		if m.Result.Data.Data.Type == "message" {
			return m, nil
		}
	}
	return nil, nil
}

// SendChatMessage sends message to the channel-chat
// currently only text messages are supported
// also unescaping needed
func (bot *VKPlayBot) SendChatMessage(p string, c *Channel, mention *User) {
	msg := []interface{}{}
	if mention != nil {
		m := &MentionContent{
			Type:        "mention",
			ID:          mention.ID,
			Nick:        mention.Nick,
			DisplayName: mention.DisplayName,
			Name:        mention.Name,
		}
		msg = append(msg, m)
		txt := &TextContent{
			Modificator: "",
			Type:        "text",
			Content:     "[\", \",\"unstyled\",[]]",
		}

		msg = append(msg, txt)
	}
	txt := &TextContent{
		Modificator: "",
		Type:        "text",
		Content:     fmt.Sprintf("[\"%s \",\"unstyled\",[]]", p),
	}

	msg = append(msg, txt)
	txt = &TextContent{
		Modificator: "BLOCK_END",
		Type:        "text",
		Content:     "",
	}
	msg = append(msg, txt)

	b, _ := json.Marshal(msg)
	body := strings.NewReader("data=" + string(b))
	r, err := http.NewRequest("POST", fmt.Sprintf("https://api.vkplay.live/v1/blog/%s/public_video_stream/chat", c.URL), body)
	if err != nil {
		log.Printf("[Error] %s\n", err)
	}

	r.Header.Add("Origin", "https://vkplay.live")
	r.Header.Add("Referer", fmt.Sprintf("https://vkplay.live/%s", c.URL))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Authorization", "Bearer "+bot.token.AccessToken)
	r.Header.Add("X-From-Id", bot.ClientID)

	resp, err := bot.DoReq(r)
	if err != nil {
		log.Printf("[Error] %s\n", err)
	}
	if resp.StatusCode != 200 {
		rd, _ := io.ReadAll(resp.Body)
		log.Println(string(rd))
	}
	defer resp.Body.Close()
}

// ParseMessage returns Parsed message
// moslty it called on raw WSMessage which represents chat message
// stream messages are not supported yet
func (m *WSMessage) ParseMessage() *Message {
	text := []interface{}{}
	msg := &Message{Smile: []string{}, User: m.Result.Data.Data.Data.User}
	for _, t := range m.Result.Data.Data.Data.MsgContent {
		switch t.Type {
		case "mention":
			msg.To = t.Name
		case "text":
			if t.Modificator == "" {
				json.Unmarshal([]byte(t.Content), &text)
				msg.Text = text[0].(string)
			}
		case "smile":
			msg.Smile = append(msg.Smile, t.Name)
		case "link":
			msg.Link = t.URL
		default:
			log.Printf("unknown type: %v", t)
		}
	}
	return msg
}
