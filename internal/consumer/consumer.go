package consumer

import (
	"WBTechL0/internal/config"
	"WBTechL0/internal/models"
	"WBTechL0/internal/service"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"log"
)

type Consumer struct {
	srv      *service.OrderService
	cfgKafka config.Kafka
}

type consumerGroupHandler struct {
	srv *service.OrderService
}

// New создает новый экземпляр Kafka-консюмера
func New(orderService *service.OrderService, cfgKafka config.Kafka) *Consumer {
	return &Consumer{
		srv:      orderService,
		cfgKafka: cfgKafka,
	}
}

// Start запускает консюмера
func (c *Consumer) Start(ctx context.Context) error {
	cfg := sarama.NewConfig()
	//config.Version = sarama.V2_0_0_0
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(c.cfgKafka.Brokers, c.cfgKafka.GroupId, cfg)
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	handler := consumerGroupHandler{srv: c.srv}

	for {
		if err = consumerGroup.Consume(ctx, []string{c.cfgKafka.Topic}, &handler); err != nil {
			log.Printf("Error from consumer: %v", err)
			return err
		}

		// Если контекст завершен, выходим из цикла
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var order models.Order

		// Десериализация сообщения из Kafka
		err := json.Unmarshal(msg.Value, &order)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		// Обработка и сохранение ордера через сервис
		h.srv.SaveOrder(order)

		// Подтверждаем обработку сообщения
		sess.MarkMessage(msg, "")
	}
	return nil
}
