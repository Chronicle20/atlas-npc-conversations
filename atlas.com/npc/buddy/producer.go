package buddy

import (
	"atlas-npc-conversations/kafka/message/buddy/list"
	"context"
	"github.com/Chronicle20/atlas-constants/world"
	"github.com/Chronicle20/atlas-kafka/producer"
	"github.com/Chronicle20/atlas-kafka/topic"
	"github.com/Chronicle20/atlas-model/model"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// ProviderImpl creates a producer with proper context and header decorators for buddy list commands
func ProviderImpl(l logrus.FieldLogger) func(ctx context.Context) func(token string) producer.MessageProducer {
	return func(ctx context.Context) func(token string) producer.MessageProducer {
		sd := producer.SpanHeaderDecorator(ctx)
		td := producer.TenantHeaderDecorator(ctx)
		return func(token string) producer.MessageProducer {
			return producer.Produce(l)(producer.WriterProvider(topic.EnvProvider(l)(token)))(sd, td)
		}
	}
}

// increaseCapacityProvider creates a Kafka message provider for increasing buddy list capacity
func increaseCapacityProvider(worldId world.Id, characterId uint32, newCapacity byte) model.Provider[[]kafka.Message] {
	key := producer.CreateKey(int(characterId))
	value := &list.Command[list.IncreaseCapacityCommandBody]{
		WorldId:     byte(worldId),
		CharacterId: characterId,
		Type:        list.CommandTypeIncreaseCapacity,
		Body: list.IncreaseCapacityCommandBody{
			NewCapacity: newCapacity,
		},
	}
	return producer.SingleMessageProvider(key, value)
}