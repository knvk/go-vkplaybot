package vkplaybot

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PollModule struct {
	BaseModule
	Poll *Poll
}

type Poll struct {
	Title    string
	Running  bool
	Voted    map[string]struct{}
	Options  []option
	VoteTime int
}

type option struct {
	Name  string
	ID    string
	Count int
}

var (
	CurPoll = make(chan Poll, 1)
)

func NewPollModule() *PollModule {
	return &PollModule{
		BaseModule: BaseModule{
			Name:        "poll",
			IsActivated: false,
			Break:       false,
		},
		Poll: &Poll{
			Running: false,
		},
	}
}

// linear matches message with "bad" words and punish LUL
// func (m *PollModule) Match(msg *WSMessage) (found bool, fn ModuleFunc) {
func (m *PollModule) Match(msg *WSMessage) (found bool, fn ModuleFunc) {
	b := msg.ParseMessage()
	if b.Text == "" {
		return false, nil
	}
	cmdTxt := strings.Fields(b.Text)
	if len(cmdTxt) < 1 {
		return false, nil
	}

	if m.Poll.Running {
		//log.Println("poll already running")
		for i := 0; i < len(m.Poll.Options); i++ {
			if cmdTxt[0] == m.Poll.Options[i].ID {
				if _, found := m.Poll.Voted[b.User.Nick]; !found {
					m.Poll.Voted[b.User.Nick] = struct{}{}
					m.Poll.Options[i].Count++
					log.Printf("%s voted for %s", b.User.Nick, m.Poll.Options[i].Name)
				}
			}
		}
	}
	re := regexp.MustCompile(`!poll|([^|]+)`)
	pp := re.FindAllStringSubmatch(b.Text, -1)

	if len(pp) < 3 {
		return false, nil
	} else {
		cmdID := 1
		m.Poll = &Poll{
			Title:    strings.Trim(pp[1][0], " "),
			Running:  true,
			Options:  []option{},
			VoteTime: 60,
			Voted:    make(map[string]struct{}),
		}
		for i := 2; i < len(pp); i++ {
			opt := &option{
				Name:  strings.Trim(pp[i][0], " "),
				ID:    "!" + strconv.Itoa(cmdID),
				Count: 0,
			}
			cmdID++
			m.Poll.Options = append(m.Poll.Options, *opt)
		}
		CurPoll <- *m.Poll
		return true, VoteStart
	}
}

func VoteStart(w ModuleHandler, m *WSMessage) {
	p := <-CurPoll
	optText := ""
	for i := 0; i < len(p.Options); i++ {
		optText += p.Options[i].ID + " for " + p.Options[i].Name + "\\n"
	}
	text := fmt.Sprintf("Poll for %s started. Vote:\\n%s", p.Title, optText)
	time.AfterFunc(time.Duration(p.VoteTime)*time.Second, func() {
		p.VoteEnd(w)
	})
	w.SendChatMessage(text, w.GetChannel(), nil)
}

func (p *Poll) VoteEnd(w ModuleHandler) {
	log.Println("ending vote")
	text := ""
	for _, o := range p.Options {
		text += o.Name + ": " + strconv.Itoa(o.Count) + "\\n"
	}
	text = fmt.Sprintf("Poll for %s ended. Results:\\n%s", p.Title, text)
	w.SendChatMessage(text, w.GetChannel(), nil)
	p = &Poll{
		Running: false,
	}
}

// after punishment no longer proceed
func (m *PollModule) Stop() bool {
	return m.Break
}

func (m *PollModule) GetName() string {
	return m.Name
}
