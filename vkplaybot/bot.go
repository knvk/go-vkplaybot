package vkplaybot

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// Some constant URLs needed to auth bot
const (
	SignInURL   string = "https://auth-ac.vkplay.ru/sign_in"
	GetTokenURL string = "https://account.vkplay.ru/oauth2/?redirect_uri=" +
		"https%3A%2F%2Fvkplay.live%2Fapp%2Foauth_redirect_vkplay&client_id=vkplay.live&response_type=code&skip_grants=1"
	GetWSTokenURL string = "https://api.vkplay.live/v1/ws/connect"
	WSURL         string = "wss://pubsub.vkplay.live/connection/websocket"
)

type VKPlayBot struct {
	Client    *http.Client
	Config    BotConfig
	ClientID  string
	wsToken   string
	token     *AuthToken
	WSConn    *websocket.Conn
	WSCounter int
	Cancel    chan struct{}
	Channel   *Channel
	Modules   []Module
}

type AuthToken struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int    `json:"expiresAt"`
}

// Keep-alive ticker to persist WS connection
var (
	ticker = time.NewTicker(60 * time.Second)
)

// DoReq is a wrapper around bot's *http.Client Do method
func (bot *VKPlayBot) DoReq(r *http.Request) (*http.Response, error) {
	return bot.Client.Do(r)
}

// auth is private method using for authentication
// I havent found yet any valid method to use VK's oauth2
// so simple http post/get methods are used here
func (bot *VKPlayBot) auth() error {

	r, err := http.NewRequest("POST", SignInURL, strings.NewReader(fmt.Sprintf("login=%s&"+
		"password=%s", bot.Config.User, bot.Config.Pass)))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("origin", "https://account.vkplay.ru")
	r.Header.Add("referer", "https://account.vkplay.ru/")
	_, err = bot.DoReq(r)

	if err != nil {
		return err
	}

	r, err = http.NewRequest("GET", GetTokenURL, nil)
	r.Header.Add("origin", "https://account.vkplay.ru")
	r.Header.Add("referer", "https://account.vkplay.ru/")
	_, err = bot.DoReq(r)
	if err != nil {
		return err
	}

	u, _ := url.Parse("https://vkplay.live")
	for _, cookie := range bot.Client.Jar.Cookies(u) {
		if cookie.Name == "_clientId" {
			bot.ClientID = cookie.Value
		}
		if cookie.Name == "auth" {
			token := &AuthToken{}
			t, _ := url.QueryUnescape(cookie.Value)
			err := json.Unmarshal([]byte(t), &token)
			if err != nil {
				return err
			}
			bot.token = token

			return nil
		}
	}
	return errors.New("Can't Auth")
}

// ReadWSMessage  is a wrapper that returns byte slice read from bot Web Socket Conn
func (bot *VKPlayBot) ReadWSMessage() (p []byte, err error) {
	_, p, err = bot.WSConn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return p, nil

}

// SendWSMessage send slice of byte to bot WebSocket connection
// Mostly its just text message so no others are used
// Also we need to use unique ID  every message, to do that we increase counter every time
func (bot *VKPlayBot) SendWSMessage(p []byte) error {
	bot.WSCounter++
	err := bot.WSConn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// connectWS is a method that used to connect to websocket
// before calling it we need to be sure auth token is present
func (bot *VKPlayBot) connectWS() error {
	r, _ := http.NewRequest("GET", GetWSTokenURL, nil)
	// set headers
	r.Header.Add("Authorization", "Bearer "+bot.token.AccessToken)
	r.Header.Add("X-From-Id", bot.ClientID)
	r.Header.Add("Origin", "https://vkplay.live")
	r.Header.Add("Referer", "https://vkplay.live/")

	resp, err := bot.DoReq(r)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return (err)
	}

	var token map[string]interface{}
	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal(b, &token)
	bot.wsToken = token["token"].(string)

	h := http.Header{
		"Origin": {"https://vkplay.live"},
	}
	c, resp, err := websocket.DefaultDialer.Dial(WSURL, h)
	if err != nil {
		log.Printf("handshake failed with status %d", resp.StatusCode, resp.Body)
		return err
	}
	bot.WSConn = c

	// register
	t := fmt.Sprintf(`{"params": {"token": "%s", "name": "js"}, "id": %d}`, bot.wsToken, bot.WSCounter)
	bot.SendWSMessage([]byte(t))

	_, _, err = bot.WSConn.ReadMessage()
	if err != nil {
		return err
	}
	return nil
}

// GetChannel returns bot's current channel struct
func (bot *VKPlayBot) GetChannel() *Channel {
	return bot.Channel
}

// This complex method is needed to bootstrap bot before serving
// first of all we check if auth token present, if so we skip auth method
// Then get and init modules/channels from config file
func (bot *VKPlayBot) Start() error {

	log.Println("Starting bot")
	// preps
	bot.WSCounter = 1
	bot.Cancel = make(chan struct{})
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	bot.Client = &http.Client{Jar: jar}

	// auth
	if bot.Config.Token != "" {
		log.Println("Auth using token")
		token := &AuthToken{}
		err := json.Unmarshal([]byte(bot.Config.Token), &token)
		if err != nil {
			return err
		}
		bot.token = token
	} else {
		err := bot.auth()
		log.Println("Auth using login/pass")
		if err != nil {
			log.Println("Auth error: %s\n", err)
			return err
		}
	}
	// ws
	log.Println("Connecting to WS")
	err = bot.connectWS()
	if err != nil {
		log.Println("WS error: %s\n", err)
		return err
	}
	log.Println("Activating modules")
	bot.Modules, err = bot.Config.activateModules()
	// channels
	bot.Channel, err = bot.Config.getChannelsFromConfig()
	if err != nil {
		log.Println("Chat channels error: %s\n", err)
		return err
	}
	return nil
}

// connectToChat used to connect to the chat by writing spec message to WS
// also we start coroutine which periodicly sends "keep alives" to persist connection
func (bot *VKPlayBot) connectToChat(ch *Channel) error {

	p := fmt.Sprintf(`{"method": 1,"params":{"channel": "public-chat:%d"},"id":%d}`, (*ch).Owner.ID, bot.WSCounter)
	err := bot.SendWSMessage([]byte(p))
	if err != nil {
		return err
	}

	go func() error {
		for {
			select {
			// sometimes
			//case <- bot.Cancel:
			//		fmt.Println("stopping connect")
			//bot.WSConn.Close()
			//break
			case <-ticker.C:
				p := fmt.Sprintf(`{"method":7,"id":%d}`, bot.WSCounter)
				err := bot.SendWSMessage([]byte(p))
				if err != nil {
					return err
				}
			default:
			}
		}
	}()

	return nil
}

// handle is the core function of the bot
// it starts in separate routine after connecting to the chat
// and executing modules sequentially (modules order matter)
// and then matches and executes sth
func (bot *VKPlayBot) Handle() {

	log.Printf("Conntecting to chat %s", bot.Channel.URL)
	err := bot.connectToChat(bot.Channel)
	if err != nil {
		panic(err)
	}
	log.Println("Success! Start serving...")

	go func() {
		for {
			select {
			case <-bot.Cancel:
				//break
				log.Println("Stop handling")
				bot.WSConn.Close()
				return
			default:
				m, err := bot.ReadChatMessage()

				if err != nil {
					panic(err)
				}

				if m != nil {
					for i := len(bot.Modules) - 1; i >= 0; i-- {
						match, fn := bot.Modules[i].Match(m)
						//log.Println("executing module", bot.Modules[i].GetName())
						if match {
							if fn != nil {
								bot.onMessage(m, fn)
							}
							if bot.Modules[i].Stop() {
								break
							}
						}
					}
				}
			}
		}

	}()
}

// wrapper to call matching function on success
func (bot *VKPlayBot) onMessage(m *WSMessage, fn ModuleFunc) {
	fn(bot, m)
}
