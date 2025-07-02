package npc

import (
	npc2 "atlas-npc-conversations/kafka/message/npc"
	"atlas-npc-conversations/kafka/producer"
	"context"
	"github.com/sirupsen/logrus"
)

const (
	MessageTypeSimple        = "SIMPLE"
	MessageTypeNext          = "NEXT"
	MessageTypeNextPrevious  = "NEXT_PREVIOUS"
	MessageTypePrevious      = "PREVIOUS"
	MessageTypeYesNo         = "YES_NO"
	MessageTypeOk            = "OK"
	MessageTypeNum           = "NUM"
	MessageTypeText          = "TEXT"
	MessageTypeStyle         = "STYLE"
	MessageTypeAcceptDecline = "ACCEPT_DECLINE"

	SpeakerNPCLeft        = "NPC_LEFT"
	SpeakerNPCRight       = "NPC_RIGHT"
	SpeakerCharacterLeft  = "CHARACTER_LEFT"
	SpeakerCharacterRight = "CHARACTER_RIGHT"
	SpeakerUnknown        = "UNKNOWN"
	SpeakerUnknown2       = "UNKNOWN2"
)

type Processor interface {
	Dispose(worldId byte, channelId byte, characterId uint32)
	SendSimple(worldId byte, channelId byte, characterId uint32, npcId uint32) TalkFunc
	SendNext(worldId byte, channelId byte, characterId uint32, npcId uint32) TalkFunc
	SendNextPrevious(worldId byte, channelId byte, characterId uint32, npcId uint32) TalkFunc
	SendOk(worldId byte, channelId byte, characterId uint32, npcId uint32) TalkFunc
	SendYesNo(worldId byte, channelId byte, characterId uint32, npcId uint32) TalkFunc
	SendNPCTalk(worldId byte, channelId byte, characterId uint32, npcId uint32, config *TalkConfig) func(message string, configurations ...TalkConfigurator)
}

type ProcessorImpl struct {
	l   logrus.FieldLogger
	ctx context.Context
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context) Processor {
	return &ProcessorImpl{
		l:   l,
		ctx: ctx,
	}
}

func (p *ProcessorImpl) Dispose(worldId byte, channelId byte, characterId uint32) {
	_ = producer.ProviderImpl(p.l)(p.ctx)(npc2.EnvEventTopicCharacterStatus)(enableActionsProvider(worldId, channelId, characterId))
}

type TalkConfig struct {
	messageType string
	speaker     string
}

func (c TalkConfig) MessageType() string {
	return c.messageType
}

func (c TalkConfig) Speaker() string {
	return c.speaker
}

type TalkConfigurator func(config *TalkConfig)

type TalkFunc func(message string, configurations ...TalkConfigurator)

func (p *ProcessorImpl) SendSimple(worldId byte, channelId byte, characterId uint32, npcId uint32) TalkFunc {
	return p.SendNPCTalk(worldId, channelId, characterId, npcId, &TalkConfig{messageType: MessageTypeSimple, speaker: SpeakerNPCLeft})
}

func (p *ProcessorImpl) SendNext(worldId byte, channelId byte, characterId uint32, npcId uint32) TalkFunc {
	return p.SendNPCTalk(worldId, channelId, characterId, npcId, &TalkConfig{messageType: MessageTypeNext, speaker: SpeakerNPCLeft})
}

func (p *ProcessorImpl) SendNextPrevious(worldId byte, channelId byte, characterId uint32, npcId uint32) TalkFunc {
	return p.SendNPCTalk(worldId, channelId, characterId, npcId, &TalkConfig{messageType: MessageTypeNextPrevious, speaker: SpeakerNPCLeft})
}

func (p *ProcessorImpl) SendOk(worldId byte, channelId byte, characterId uint32, npcId uint32) TalkFunc {
	return p.SendNPCTalk(worldId, channelId, characterId, npcId, &TalkConfig{messageType: MessageTypeOk, speaker: SpeakerNPCLeft})
}

func (p *ProcessorImpl) SendYesNo(worldId byte, channelId byte, characterId uint32, npcId uint32) TalkFunc {
	return p.SendNPCTalk(worldId, channelId, characterId, npcId, &TalkConfig{messageType: MessageTypeYesNo, speaker: SpeakerNPCLeft})
}

func (p *ProcessorImpl) SendNPCTalk(worldId byte, channelId byte, characterId uint32, npcId uint32, config *TalkConfig) func(message string, configurations ...TalkConfigurator) {
	return func(message string, configurations ...TalkConfigurator) {
		for _, configuration := range configurations {
			configuration(config)
		}
		_ = producer.ProviderImpl(p.l)(p.ctx)(npc2.EnvConversationCommandTopic)(simpleConversationProvider(worldId, channelId, characterId, npcId, message, config.MessageType(), config.Speaker()))
	}
}
