package vkplaybot

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type CommandsModule struct {
	BaseModule
	//name     string
	commands map[string]*Command
}

type Command struct {
	name       string
	handler    ModuleFunc
	Cooldown   int
	InCooldown bool
	mutex      *sync.RWMutex
}

var (
	commandList = map[string]ModuleFunc{
		"!лет":    FollowAge,
		"!член":   Dick,
		"!привет": Greet,
		"!анек":   Joke,
		"!зрилы":  Viewer,
	}
)

func NewCommandsModule() *CommandsModule {
	cm := &CommandsModule{
		BaseModule: BaseModule{
			Name:        "commands",
			IsActivated: true,
			Break:       false,
		},
		commands: make(map[string]*Command),
	}
	for k, v := range commandList {
		cm.HandleCommand(k, v)
	}
	cm.HandleCommand("!хелп", Help)
	return cm
}

func (m *CommandsModule) GetName() string {
	return m.Name
}

// matches command and sets its cooldown
func (m *CommandsModule) Match(msg *WSMessage) (found bool, fn ModuleFunc) {
	b := msg.ParseMessage()
	if b.Text == "" {
		return false, nil
	}
	cmdTxt := strings.Fields(b.Text)
	if len(cmdTxt) < 1 {
		return false, nil
	}
	v, found := m.commands[cmdTxt[0]]

	if found {
		if v.InCooldown {
			fmt.Println("in cooldown")
			return false, nil
		}
		v.AddCooldown()
		return found, v.handler
	}
	return
}

func (m *CommandsModule) Stop() bool {
	return m.Break
}

func (m *CommandsModule) HandleCommand(cmd string, fn ModuleFunc) *Command {
	return m.newCommand(cmd, fn)
}

func (m *CommandsModule) newCommand(n string, fn ModuleFunc) *Command {
	cmd := &Command{name: n, handler: fn, Cooldown: 5, mutex: &sync.RWMutex{}}
	m.commands[n] = cmd
	return cmd
}

func (c *Command) AddCooldown() {
	c.addCooldown()

	time.AfterFunc(time.Duration(c.Cooldown)*time.Second, func() {
		c.removeCooldown()
	})
}

func (c *Command) addCooldown() {
	c.mutex.Lock()
	c.InCooldown = true
	c.mutex.Unlock()
}

func (c *Command) removeCooldown() {
	c.mutex.Lock()
	c.InCooldown = false
	c.mutex.Unlock()
}

func Help(w ModuleHandler, m *WSMessage) {
	ch := w.GetChannel()
	text := "Список команд: "
	for k := range commandList {
		text = text + k + " "
	}
	w.SendChatMessage(text, ch, nil)
}

//
// Commands list
// To add new one write your own function and place its also to the commandList variable
// !хелп command is permanent and generates automatically
//

func Greet(w ModuleHandler, m *WSMessage) {
	b := m.ParseMessage()
	ch := w.GetChannel()
	text := fmt.Sprintf(`привет %s, пошел в жопу`, b.User.Nick)
	//🪓
	w.SendChatMessage(text, ch, nil)
}

func FollowAge(w ModuleHandler, m *WSMessage) {
	b := m.ParseMessage()
	epoch := int64(b.User.Subscription.OnTime)
	ch := w.GetChannel()
	if epoch > 0 {
		t := time.Unix(epoch, 0).Format("2006 Jan 02 15:04:05")
		w.SendChatMessage(fmt.Sprintf("%s following channel since %s", b.User.Nick, t), ch, nil)
	} else {
		w.SendChatMessage(fmt.Sprintf("%s not following this channel yet", b.User.Nick), ch, nil)
	}
}

func Dick(w ModuleHandler, m *WSMessage) {
	b := m.ParseMessage()
	a := rand.Intn(11-4) + 4
	ch := w.GetChannel()
	w.SendChatMessage(fmt.Sprintf("пиписька %s %d см", b.User.Nick, a), ch, b.User)
}

func Joke(w ModuleHandler, m *WSMessage) {
	res, _ := http.Get(`https://www.anekdot.ru/rss/randomu.html`)
	resp, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	text := string(resp)
	text = strings.Split(text, "JSON.parse('[\\\"")[1]
	text = strings.Split(text, "\\\",\\\"")[0]
	text = strings.ReplaceAll(text, "\\\\\\\"", "\"")
	br := regexp.MustCompile(`([а-я])<br>([а-я])`)
	text = br.ReplaceAllString(text, `$1 $2`)
	text = strings.ReplaceAll(text, "<br>", "")
	text = strings.ReplaceAll(text, `"`, "")
	ch := w.GetChannel()
	w.SendChatMessage(text, ch, nil)
}

func Viewer(w ModuleHandler, m *WSMessage) {
	v, _ := w.GetViewers()
	ml := []string{}
	ul := []string{}
	for _, v := range v.Data.Moderators {
		ml = append(ml, v.Nick)
	}
	for _, v := range v.Data.Users {
		ul = append(ul, v.Nick)
	}
	t := fmt.Sprintf(`Заведующий:%s Mодеры:%s Смотрящие:%s`, v.Data.Owner.Nick, strings.Join(ml, ","), strings.Join(ul, ","))
	gg := Chunks(t, 200)
	ch := w.GetChannel()
	for _, v := range gg {
		w.SendChatMessage(v, ch, nil)
	}
}

func Chunks(s string, chunkSize int) []string {
	if len(s) == 0 {
		return nil
	}
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks []string = make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}
