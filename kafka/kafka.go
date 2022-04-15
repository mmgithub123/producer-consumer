package kafka

import (
    "crypto/sha256"
    "fmt"
    "strings"

    "github.com/Shopify/sarama"
    "github.com/wjiec/gdsn"
    "github.com/xdg-go/scram"
)

var (
    client sarama.SyncProducer
)

var (
    SHA256 scram.HashGeneratorFcn = sha256.New
)

type XDGSCRAMClient struct {
    *scram.Client
    *scram.ClientConversation
    scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
    x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
    if err != nil {
        return err
    }
    x.ClientConversation = x.Client.NewConversation()
    return nil
}

func (x *XDGSCRAMClient) Step(challenge string) (response string, err error) {
    response, err = x.ClientConversation.Step(challenge)
    return
}

func (x *XDGSCRAMClient) Done() bool {
    return x.ClientConversation.Done()
}

func Init(kafkaDSN string) (sarama.SyncProducer, error) {
    d, err := gdsn.Parse(kafkaDSN)
    if err != nil {
        return nil, err
    }

    config := sarama.NewConfig()

    config.Producer.RequiredAcks = sarama.WaitForAll

    config.Producer.Return.Successes = true

    if d.User.Username() != "" {
        config.Metadata.Full = true
        config.Net.SASL.Enable = true
        config.Net.SASL.User = d.User.Username()
        config.Net.SASL.Password, _ = d.User.Password()

        config.Net.SASL.Handshake = true
        config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
            return &XDGSCRAMClient{
                HashGeneratorFcn: SHA256,
            }
        }
        config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
    }

    client, err = sarama.NewSyncProducer(strings.Split(d.Address(), ","), config)
    if err != nil {
        fmt.Println("producer closed, err:", err)
    }

    return client, err

}

func SendToKafka(topic, data string) {
    msg := &sarama.ProducerMessage{}
    msg.Topic = topic
    msg.Value = sarama.StringEncoder(data)

    pid, offset, err := client.SendMessage(msg)
    if err != nil {
        fmt.Println("sned mage failed, err:", err)
    }
    fmt.Printf("pid:%v offset:%v\n", pid, offset)
    fmt.Println("send ok")
}
