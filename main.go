package main

import (
	"flag"
	"fmt"
	"log-forward/conf"
	"log-forward/kafka"
	"log-forward/us3"
	"time"

	"gopkg.in/ini.v1"
)

var config = new(conf.Config)
var kafkaMsg = make(chan string)

func main() {
	var (
		confFile    string
		us3ConfFile string
	)
	flag.StringVar(&confFile, "confFile", "config.ini", "set config file")
	flag.StringVar(&us3ConfFile, "us3ConfFile", "config.json", "set the us3 config json file")
	flag.Parse()

	fmt.Println(us3ConfFile)
	fmt.Println(confFile)

	err := ini.MapTo(config, confFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		return
	}

	if _, err = kafka.Init(config.Kafka.Dsn); err != nil {
		fmt.Println("init kafka failed, err:%v\n", err)
		return
	}
	fmt.Println("init kafka success.")

	us3.Run(us3ConfFile, kafkaMsg)
	run()
}

func run() {

	for {
		select {
		case line := <-kafkaMsg:
			kafka.SendToKafka(config.Kafka.Topic, line)
		default:
			time.Sleep(time.Second)
		}
	}
}
