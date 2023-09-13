/**
 * @Author: koulei
 * @Description:
 * @File: kafka
 * @Version: 1.0.0
 * @Date: 2023/9/7 16:27
 */

package connector

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"

	"github.com/flash520/pusher/pkg/pusher"
)

func NewKafkaConfig(GroupID, topic string, brokers ...string) kafka.ReaderConfig {
	return kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  GroupID,
		Topic:    topic,
		MaxBytes: 10e6,
	}
}

type Kafka struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	reader     *kafka.Reader
	eventChan  chan<- pusher.Data
}

func NewKafkaReader(config kafka.ReaderConfig) pusher.Reader {
	reader := kafka.NewReader(config)
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &Kafka{
		ctx:        ctx,
		cancelFunc: cancelFunc,
		reader:     reader,
	}
}

func (k *Kafka) Name() string {
	return "kafka"
}

func (k *Kafka) SetChannel(channel chan<- pusher.Data) {
	k.eventChan = channel
}

func (k *Kafka) Start() {
	defer func() {
		logrus.Warnf("Stoped Connector: %s", k.Name())
		_ = k.reader.Close()
	}()
	stat := k.reader.Stats()
	logrus.Infof("Started Connector: %s -> Broker: %v Topic: %s Partition: %s",
		k.Name(),
		k.reader.Config().Brokers,
		stat.Topic,
		stat.Partition,
	)
	for {
		select {
		case <-k.ctx.Done():
			return
		default:
		}
		message, err := k.reader.ReadMessage(k.ctx)
		if err != nil {
			logrus.Errorf("kafka reader error: %s", err.Error())
			time.Sleep(time.Second * 5)
			continue
		}
		k.eventChan <- pusher.NewData("kafka", message)
	}
}

func (k *Kafka) Stop() {
	k.cancelFunc()
}
