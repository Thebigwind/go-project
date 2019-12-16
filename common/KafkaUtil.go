package common

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Shopify/sarama"
	etcdclient "github.com/coreos/etcd/client"
	"strings"
)

const KAFKA_SERVER_KEY = "/datamgmt/bus/kafka-cluster0"

type KafkaClient struct {
	kclient sarama.SyncProducer
}

func GetKafkaClient(servers []string, topic string) (*KafkaClient, error) {
	Logger.Info("servers", servers)

	//connect kafa
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	client, err := sarama.NewSyncProducer(servers, config)
	if err != nil {
		Logger.Errorf("err:", err)
		return nil, err
	}

	Kclient := &KafkaClient{
		kclient: client,
	}

	return Kclient, err
}

func SendToKafkaString(servers []string, topic string, tagMsg string) error {

	//struct msg
	msg := &sarama.ProducerMessage{}
	msg.Topic = topic
	msg.Value = sarama.StringEncoder(tagMsg)

	client, err := GetKafkaClient(servers, topic)
	if err != nil {
		Logger.Errorf("err:", err)
		return err
	}
	defer client.kclient.Close()

	//send msg
	pid, offset, err := client.kclient.SendMessage(msg)
	if err != nil {
		Logger.Errorf("err:", err)
		return err
	}

	Logger.Debugf("pid:%+v,offset:%+v", pid, offset)
	return err
}

func SendToKafkaBytes(servers []string, topic string, tagMsg []byte) error {

	//struct msg
	msg := &sarama.ProducerMessage{}
	msg.Topic = topic
	msg.Value = sarama.ByteEncoder(tagMsg)

	client, err := GetKafkaClient(servers, topic)
	if err != nil {
		Logger.Errorf("err:", err)
		return err
	}
	defer client.kclient.Close()

	//send msg
	pid, offset, err := client.kclient.SendMessage(msg)
	if err != nil {
		Logger.Errorf("err:", err)
		return err
	}

	Logger.Debugf("pid:%+v,offset:%+v", pid, offset)
	return err
}

type KafkaEntry struct {
	BootstrapServers string `json:"bootstrap_servers"` //Servers string `json:"servers"`
	Type             string `json:"type"`
}

////////////get kafkaserver address from etcd//////////////////
//get servers from etcd.
func GetkafkaEtcdConfig() (error, []string, string) {
	sopt := etcdclient.GetOptions{
		Recursive: false,
		Sort:      false,
		Quorum:    false,
	}
	ctx := context.Background()
	resp, err := GlobalContext.Kapi.Get(ctx, KAFKA_SERVER_KEY, &sopt)

	if err != nil {
		Logger.Errorf("get etcd config error:%v", err)
		return err, nil, ""
	}
	//log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
	value := strings.Replace(resp.Node.Value, "bootstrap.servers", "bootstrap_servers", -1)

	var kafka KafkaEntry
	err = json.Unmarshal([]byte(value), &kafka)

	if err != nil {
		Logger.Printf("Can't decode DB etcd config: %s\n", err.Error())
		return err, nil, ""
	}
	//return kafka.BootstrapServers
	return err, strings.Split(kafka.BootstrapServers, ","), kafka.Type
}

func GetkafkaJsonConfig() (error, []string, string) {
	servers := GlobalMetaviewConfig.KafkaConfig.Servers
	types := GlobalMetaviewConfig.KafkaConfig.Type

	var err error
	if servers == "" {
		err = errors.New("the server address is empty.")
	}

	return err, strings.Split(servers, ","), types
}
