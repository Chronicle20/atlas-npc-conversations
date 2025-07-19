package buddy

import (
	"atlas-npc-conversations/kafka/message/buddy/list"
	"github.com/Chronicle20/atlas-constants/world"
	"github.com/Chronicle20/atlas-kafka/producer"
	"github.com/Chronicle20/atlas-model/model"
	"github.com/segmentio/kafka-go"
)

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