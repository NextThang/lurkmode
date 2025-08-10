package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/nextthang/lurkmode/internal/message"
	"github.com/nextthang/lurkmode/internal/twitch"
	"github.com/nextthang/lurkmode/pkg/ringbuffer"
)

type model struct {
	ready        bool
	viewport     viewport.Model
	channelName  string
	messageChan  <-chan message.Message
	messages     *ringbuffer.RingBuffer[message.Message]
	twitchClient *twitch.Client
	footer       footer
	header       header
	renderTime   bool
	shuttingDown bool
}

// TODO: We should probably make this configurable.
const historySize = 200

func (m model) receiveMessage() tea.Cmd {
	return func() tea.Msg {
		for {
			msg := <-m.messageChan
			if msg != nil {
				return msg
			}
		}
	}
}

func closeTwitchClient(twitchClient *twitch.Client) tea.Cmd {
	return func() tea.Msg {
		if err := twitchClient.Disconnect(); err != nil {
			time.Sleep(100 * time.Millisecond)
			return closeTwitchClient(twitchClient)()
		}

		return tea.QuitMsg{}
	}
}

func (m model) Init() tea.Cmd {
	return m.receiveMessage()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		switch k {
		case "ctrl+c", "q":
			if err := m.twitchClient.Disconnect(); err != nil {
				m.shuttingDown = true
			}
			return m, closeTwitchClient(m.twitchClient)
		case "t":
			m.renderTime = !m.renderTime
			m.viewport.SetContent(m.renderChatHistory())
		}
	case tea.QuitMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport.SetContent(m.renderChatHistory())
			m.ready = true
		}
		m.viewport.SetWidth(msg.Width)
		// TODO: We need extra space for the header and footer, but we should probably do this dynamically instead of hardcoding it.
		m.viewport.SetHeight(msg.Height - 2)
		m.viewport.GotoBottom()
	case message.Message:
		m.messages.Add(msg)
		m.viewport.SetContent(m.renderChatHistory())
		m.viewport.GotoBottom()
		return m, m.receiveMessage()
	}

	var viewportCmd tea.Cmd
	m.viewport, viewportCmd = m.viewport.Update(msg)
	var headerCmd tea.Cmd
	m.header, headerCmd = m.header.Update(msg)

	return m, tea.Batch(viewportCmd, headerCmd)
}

func (m model) View() string {
	if !m.ready {
		return "Preparing to lurk..."
	}
	if m.shuttingDown {
		return "Shutting down..."
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.header.View(),
		m.viewport.View(),
		m.footer.View(),
	)
}

func (m model) renderChatHistory() string {
	history := m.messages.Get()
	if len(history) == 0 {
		return "*Crickets*"
	}

	regularMesasgeStyle := lipgloss.NewStyle()
	userNoticeStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1f1f23"))

	var builder strings.Builder
	for i, msg := range history {
		if i > 0 {
			builder.WriteString("\n")
		}

		style := regularMesasgeStyle
		if _, ok := msg.(message.UserNotice); ok {
			style = userNoticeStyle
		}

		builder.WriteString(msg.Render(m.renderTime, style))
	}
	return builder.String()
}

func newModel(channelName string, messageChan <-chan message.Message, twitchIrcClient *twitch.Client) model {
	m := model{
		channelName:  channelName,
		messageChan:  messageChan,
		viewport:     viewport.New(),
		twitchClient: twitchIrcClient,
		footer:       newFooter(),
		header:       newHeader(fmt.Sprintf("LurkMode - #%s", channelName)),
		messages:     ringbuffer.NewBuffer[message.Message](historySize),
	}

	m.viewport.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6441a5")).
		Padding(0, 1)
	m.viewport.SoftWrap = true

	return m
}

func Run(channelName string) error {
	tea.LogToFile("debug.log", "")

	messageChan := make(chan message.Message, 100)
	twitchIrcClient := twitch.NewClient(messageChan, channelName)

	program := tea.NewProgram(newModel(channelName, messageChan, twitchIrcClient), tea.WithAltScreen(), tea.WithMouseCellMotion())

	ircClientReturnChan := make(chan error)
	go func() { ircClientReturnChan <- twitchIrcClient.Connect() }()

	if _, err := program.Run(); err != nil {
		return err
	}

	if err := <-ircClientReturnChan; err != nil {
		return err
	}

	return nil
}
