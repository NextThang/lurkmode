package message

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nextthang/lurkmode/internal/stylebuilder"
)

const (
	SubPlanPrime = iota
	SubPlanTier1
	SubPlanTier2
	SubPlanTier3
)

type SubPlan uint8

func parseSubPlan(plan string) SubPlan {
	switch plan {
	case "Prime":
		return SubPlanPrime
	case "1000":
		return SubPlanTier1
	case "2000":
		return SubPlanTier2
	case "3000":
		return SubPlanTier3
	default:
		return SubPlanTier1 // Default to Tier 1 if unknown
	}
}

func parseSubLength(num int) (string, int) {
	switch {
	case num > 3000:
		return "🥇", num - 3000
	case num > 2000:
		return "🥈", num - 2000
	default:
		return "🥉", num
	}
}

const timeFormat = "[" + time.Kitchen + "] "

var (
	timeStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("247"))
	broadcasterBadgeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e81815"))
	vipBadgeStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#e005b9"))
	modBadgeStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ad03"))
	subBadgeStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#6441a5"))
)

func renderColoredName(user twitch.User, style lipgloss.Style) string {
	if user.Color == "" {
		return style.Render(user.DisplayName)
	} else {
		return lipgloss.NewStyle().
			Inherit(style).
			Foreground(lipgloss.Color(user.Color)).
			Render(user.DisplayName)
	}
}

func renderUserTags(user twitch.User, style lipgloss.Style) string {
	var tags []string
	if user.IsBroadcaster {
		tags = append(tags, broadcasterBadgeStyle.Inherit(style).Render("[👑]"))
	}
	if user.IsVip {
		tags = append(tags, vipBadgeStyle.Inherit(style).Render("[💎]"))
	}
	if user.IsMod {
		tags = append(tags, modBadgeStyle.Inherit(style).Render("[⚔️]"))
	}
	if subLength, ok := user.Badges["subscriber"]; ok {
		// TODO: Add the length into the message. [%s%d] looks too cluttered though.
		emoji, _ := parseSubLength(subLength)
		tags = append(tags, subBadgeStyle.Inherit(style).Render(fmt.Sprintf("[%s]", emoji)))
	}

	if len(tags) > 0 {
		tags = append(tags, style.Render(" "))
	}

	return strings.Join(tags, style.Render(""))
}

func parseMsgParamsKeyUint(message *twitch.UserNoticeMessage, key string, defaultValue uint32) uint32 {
	val, ok := message.MsgParams[key]
	if !ok {
		return defaultValue
	}
	num, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return defaultValue
	}
	return uint32(num)
}

func parseMsgParamsKeyString(message *twitch.UserNoticeMessage, key string, defaultValue string) string {
	if val, ok := message.MsgParams[key]; ok {
		return val
	}
	return defaultValue
}

type Message interface {
	Render(renderTime bool, style lipgloss.Style) string
	ChannelName() string
}

type UserNotice interface {
	isUserNotice()
}

func NewMessage(message twitch.Message) Message {
	switch v := message.(type) {
	case *twitch.PrivateMessage:
		return &channelMessage{
			baseMessage: baseMessage{
				User:    v.User,
				Time:    v.Time,
				Channel: v.Channel,
			},
			Message: v.Message,
		}
	case *twitch.UserNoticeMessage:
		return parseUserNoticeMessage(v)
	default:
		return nil
	}
}

func newBaseMessageFromNotice(message *twitch.UserNoticeMessage) baseMessage {
	return baseMessage{
		User:    message.User,
		Time:    message.Time,
		Channel: message.Channel,
	}
}

func newChannelMessageFromNotice(message *twitch.UserNoticeMessage) channelMessage {
	return channelMessage{
		baseMessage: newBaseMessageFromNotice(message),
		Message:     message.Message,
	}
}

type userNoticeParser func(message *twitch.UserNoticeMessage) Message

var userNoticeParsers = map[string]userNoticeParser{
	"sub":            parseSubMessage,
	"resub":          parseResubMessage,
	"subgift":        parseSubGiftMessage,
	"submysterygift": parseSubMysteryGiftMessage,
	"raid":           parseRaidMessage,
}

func parseSubMessage(message *twitch.UserNoticeMessage) Message {
	return &subMessage{
		channelMessage: newChannelMessageFromNotice(message),
		Plan:           parseSubPlan(message.MsgParams["sub-plan"]),
	}
}

func parseResubMessage(message *twitch.UserNoticeMessage) Message {
	return &resubMessage{
		channelMessage:   newChannelMessageFromNotice(message),
		Plan:             parseSubPlan(message.MsgParams["sub-plan"]),
		CumulativeMonths: parseMsgParamsKeyUint(message, "msg-param-cumulative-months", 1),
		CurrentStreak:    parseMsgParamsKeyUint(message, "msg-param-streak-months", 0),
	}
}

func parseSubGiftMessage(message *twitch.UserNoticeMessage) Message {
	return &subGiftMessage{
		baseMessage: newBaseMessageFromNotice(message),
		Receiver: twitch.User{
			ID:          parseMsgParamsKeyString(message, "msg-param-recipient-id", ""),
			Name:        parseMsgParamsKeyString(message, "msg-param-recipient-user-name", ""),
			DisplayName: parseMsgParamsKeyString(message, "msg-param-recipient-display-name", ""),
		},
		Plan: parseSubPlan(message.MsgParams["sub-plan"]),
	}
}

func parseSubMysteryGiftMessage(message *twitch.UserNoticeMessage) Message {
	return &subMysteryGiftMessage{
		baseMessage:    newBaseMessageFromNotice(message),
		Plan:           parseSubPlan(message.MsgParams["sub-plan"]),
		GiftCount:      parseMsgParamsKeyUint(message, "msg-param-mass-gift-count", 1),
		TotalGiftCount: parseMsgParamsKeyUint(message, "msg-param-sender-count", 0),
	}
}

func parseRaidMessage(message *twitch.UserNoticeMessage) Message {
	return &raidMessage{
		baseMessage: newBaseMessageFromNotice(message),
		ViewerCount: parseMsgParamsKeyUint(message, "msg-param-viewerCount", 1),
	}
}

func parseUserNoticeMessage(message *twitch.UserNoticeMessage) Message {
	if parser, ok := userNoticeParsers[message.MsgID]; ok {
		return parser(message)
	}
	log.Printf("Unknown user notice message type: %s %+v", message.MsgID, message)
	return nil
}

type baseMessage struct {
	User    twitch.User
	Time    time.Time
	Channel string
}

func (m *baseMessage) renderHeader(renderTime bool, style lipgloss.Style, builder *stylebuilder.StyleBuilder) {
	if renderTime {
		builder.WriteStringWithStyle(m.Time.Format(timeFormat), timeStyle)
	}
	builder.WriteStyledString(renderUserTags(m.User, style))
	builder.WriteStyledString(renderColoredName(m.User, style))
}

func (m *baseMessage) ChannelName() string {
	return m.Channel
}

type channelMessage struct {
	baseMessage
	Message string
}

func (m *channelMessage) Render(renderTime bool, style lipgloss.Style) string {
	builder := stylebuilder.NewStyleBuilder(style)
	m.renderHeader(renderTime, style, builder)

	builder.WriteString(": ")
	builder.WriteString(m.Message)
	return builder.String()
}

type subMessage struct {
	channelMessage
	Plan SubPlan // msg-param-sub-plan
}

func (m *subMessage) Render(renderTime bool, style lipgloss.Style) string {
	builder := stylebuilder.NewStyleBuilder(style)
	m.renderHeader(renderTime, style, builder)

	builder.WriteString(" subscribed")
	if m.Plan == SubPlanPrime {
		builder.WriteString(" with Prime")
	} else {
		builder.WriteString(fmt.Sprintf("at Tier %d", m.Plan))
	}
	if m.Message != "" {
		builder.WriteString("\n")
		builder.WriteStyledString(m.channelMessage.Render(renderTime, style))
	}
	return builder.String()
}

func (m *subMessage) isUserNotice() {}

type resubMessage struct {
	channelMessage
	Plan             SubPlan // msg-param-sub-plan
	CumulativeMonths uint32  // msg-param-cumulative-months
	CurrentStreak    uint32  // msg-param-streak-months, if msg-param-should-share-streak is 1
}

func (m *resubMessage) Render(renderTime bool, style lipgloss.Style) string {
	builder := stylebuilder.NewStyleBuilder(style)
	m.renderHeader(renderTime, style, builder)

	builder.WriteString(" resubscribed")
	if m.Plan == SubPlanPrime {
		builder.WriteString(" with Prime!")
	} else {
		builder.WriteString(fmt.Sprintf(" at Tier %d!", m.Plan))
	}

	builder.WriteString(fmt.Sprintf(" They have been subscribed for %d months!", m.CumulativeMonths))
	if m.CurrentStreak > 0 {
		builder.WriteString(fmt.Sprintf(" Their current streak is %d months!", m.CurrentStreak))
	}

	if m.Message != "" {
		builder.WriteString("\n")
		builder.WriteStyledString(m.channelMessage.Render(renderTime, style))
	}
	return builder.String()
}

func (m *resubMessage) isUserNotice() {}

type subGiftMessage struct {
	baseMessage
	Receiver twitch.User
	Plan     SubPlan // msg-param-sub-plan
}

func (m *subGiftMessage) Render(renderTime bool, style lipgloss.Style) string {
	builder := stylebuilder.NewStyleBuilder(style)
	m.renderHeader(renderTime, style, builder)
	builder.WriteString(fmt.Sprintf(" gifted a Tier %d subscription to %s!", m.Plan, m.Receiver.DisplayName))
	return builder.String()
}

func (m *subGiftMessage) isUserNotice() {}

type subMysteryGiftMessage struct {
	baseMessage
	Plan           SubPlan // msg-param-sub-plan
	GiftCount      uint32  // msg-param-mass-gift-count
	TotalGiftCount uint32  // msg-param-sender-count
}

func (m *subMysteryGiftMessage) Render(renderTime bool, style lipgloss.Style) string {
	builder := stylebuilder.NewStyleBuilder(style)
	m.renderHeader(renderTime, style, builder)
	builder.WriteString(fmt.Sprintf(" gifted %d Tier %d subscriptions!", m.GiftCount, m.Plan))
	if m.TotalGiftCount > 0 {
		builder.WriteString(fmt.Sprintf(" Total gifted subscriptions: %d", m.TotalGiftCount))
	}

	return builder.String()
}

func (m *subMysteryGiftMessage) isUserNotice() {}

type raidMessage struct {
	baseMessage
	ViewerCount uint32 // msg-param-viewerCount
}

func (m *raidMessage) Render(renderTime bool, style lipgloss.Style) string {
	builder := stylebuilder.NewStyleBuilder(style)
	m.renderHeader(renderTime, style, builder)

	builder.WriteString(fmt.Sprintf(" raided with %d viewers!", m.ViewerCount))

	return builder.String()
}

func (m *raidMessage) isUserNotice() {}
