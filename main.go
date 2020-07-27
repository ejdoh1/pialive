package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	SendIntervalSec      int    `split_words:"true" default:"60"`
	ReconnectIntervalSec int    `split_words:"true" default:"30"`
	TopicPrefix          string `split_words:"true" default:"pialive/"`
	TopicSuffix          string `split_words:"true"`
	Brokers              string `default:"tcp://broker.hivemq.com:1883,tcp://test.mosquitto.org:1883"`
	Qos                  int    `split_words:"true" default:"1"`
	Retained             bool   `split_words:"true" default:"true"`
	Command              string `split_words:"true" default:"ifconfig"`
	StartupPause         int    `split_words:"true" default:"5"`
	Base64Encode         bool   `split_words:"true" default:"true"`
}

type piAliveMsg struct {
	CommandOutput string `json:"cmdOutput"`
	Command       string `json:"command"`
	Timestamp     string `json:"ts"`
}

func main() {
	var cfg config
	err := envconfig.Process("pialive", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	topicSuffix := cfg.TopicSuffix
	if topicSuffix == "" {
		topicSuffix = getMacAddr()
	}
	topicPub := cfg.TopicPrefix + topicSuffix
	log.Println("publish topic:", topicPub)

	clientID := uuid.New().String()
	log.Println("client id:", clientID)

	brokers := strings.Split(cfg.Brokers, ",")
	for _, broker := range brokers {
		log.Println("Starting connection for broker:", broker)
		go func() {
			opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID).SetAutoReconnect(true)
			c := mqtt.NewClient(opts)
			for {
				if token := c.Connect(); token.Wait() && token.Error() != nil {
					log.Println("Failed to connect to broker:", broker)
					time.Sleep(time.Duration(cfg.ReconnectIntervalSec) * time.Second)
				} else {
					log.Println("Connected to broker:", broker)
					break
				}
			}
			for {
				m, e := runCmd(cfg.Command)
				cmdOutput := string(m)
				if e != nil {
					log.Println("error running command:", e.Error())
					cmdOutput = cmdOutput + e.Error()
				}

				if cfg.Base64Encode {
					cmdOutput = base64.StdEncoding.EncodeToString([]byte(cmdOutput))
				}

				pubMessage := piAliveMsg{
					Timestamp:     time.Now().String(),
					Command:       cfg.Command,
					CommandOutput: cmdOutput,
				}

				x, _ := json.Marshal(pubMessage)
				if token := c.Publish(topicPub, byte(cfg.Qos), cfg.Retained, x); token.Wait() && token.Error() != nil {
					log.Println(token.Error())
				} else {
					log.Println("Published message")
				}
				time.Sleep(time.Duration(cfg.SendIntervalSec) * time.Second)
			}
		}()
		time.Sleep(time.Duration(cfg.StartupPause) * time.Second)
	}
	s := make(chan string)
	<-s
}

func runCmd(bashCmd string) ([]byte, error) {
	cmd := exec.Command("bash", "-c", bashCmd)
	return cmd.Output()
}

func getMacAddr() (addr string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return
}
