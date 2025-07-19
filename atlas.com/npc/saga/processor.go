package saga

import (
	"atlas-npc-conversations/buddy"
	"atlas-npc-conversations/kafka/message/buddy/list"
	"atlas-npc-conversations/kafka/message/saga"
	localproducer "atlas-npc-conversations/kafka/producer"
	"context"
	"fmt"
	"github.com/Chronicle20/atlas-kafka/producer"
	"github.com/sirupsen/logrus"
)

type Processor interface {
	Create(s Saga) error
	ExecuteAction(action Action, payload any) error
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

func (p *ProcessorImpl) Create(s Saga) error {
	return localproducer.ProviderImpl(p.l)(p.ctx)(saga.EnvCommandTopic)(CreateCommandProvider(s))
}

func (p *ProcessorImpl) ExecuteAction(action Action, payload any) error {
	p.l.Debugf("Executing saga action [%s]", action)

	switch action {
	case IncreaseBuddyCapacity:
		return p.executeIncreaseBuddyCapacity(payload)
	default:
		return fmt.Errorf("unsupported saga action: %s", action)
	}
}

func (p *ProcessorImpl) executeIncreaseBuddyCapacity(payload any) error {
	buddyPayload, ok := payload.(IncreaseBuddyCapacityPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for IncreaseBuddyCapacity action: expected IncreaseBuddyCapacityPayload, got %T", payload)
	}

	p.l.Debugf("Increasing buddy capacity for character [%d] to [%d] in world [%d]",
		buddyPayload.CharacterId, buddyPayload.NewCapacity, buddyPayload.WorldId)

	// Initialize buddy producer with proper context and headers
	buddyProducer := buddy.ProviderImpl(p.l)(p.ctx)(list.EnvCommandTopic)

	// Create the buddy capacity increase message provider using the same pattern as in buddy/producer.go
	key := producer.CreateKey(int(buddyPayload.CharacterId))
	value := &list.Command[list.IncreaseCapacityCommandBody]{
		WorldId:     byte(buddyPayload.WorldId),
		CharacterId: buddyPayload.CharacterId,
		Type:        list.CommandTypeIncreaseCapacity,
		Body: list.IncreaseCapacityCommandBody{
			NewCapacity: buddyPayload.NewCapacity,
		},
	}
	messageProvider := producer.SingleMessageProvider(key, value)

	// Send the messages using the buddy producer
	err := buddyProducer(messageProvider)
	if err != nil {
		p.l.WithError(err).Errorf("Failed to send buddy capacity increase command for character [%d]", buddyPayload.CharacterId)
		return fmt.Errorf("failed to increase buddy capacity: %w", err)
	}

	p.l.Debugf("Successfully sent buddy capacity increase command for character [%d]", buddyPayload.CharacterId)
	return nil
}
