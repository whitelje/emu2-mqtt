/*
 *  Serial to MQTT adapter for EMU2
 *
 *  Copyright Jake Whiteley
 *  SPDX-License-Identifier: Apache-2.0
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may
 *  not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 *  WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/srom/xmlstream"
	"go.bug.st/serial"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var example = `<InstantaneousDemand>
  <DeviceMacId>0xd8d5b9000001140f</DeviceMacId>
  <MeterMacId>0x001350050001f4b4</MeterMacId>
  <TimeStamp>0x29068265</TimeStamp>
  <Demand>0x000345</Demand>
  <Multiplier>0x00000001</Multiplier>
  <Divisor>0x000003e8</Divisor>
  <DigitsRight>0x03</DigitsRight>
  <DigitsLeft>0x0f</DigitsLeft>
  <SuppressLeadingZero>Y</SuppressLeadingZero>
</InstantaneousDemand>`

var mqttOut chan<- pubTuple
var topicHeader = "sensors/EMU2/"

func processXML(buff []byte) (err error) {

	scanner := xmlstream.NewScanner(bytes.NewReader(buff), new(XMLInstantaneousDemand),
		new(XMLConnectionStatus), new(DeviceInfo), new(ScheduleInfo), new(MeterList), new(MeterInfo),
		new(XMLNetworkInfo), new(XMLTimeCluster), new(XMLMessageCluster), new(XMLPriceCluster),
		new(XMLCurrentSummationDelivered), new(XMLCurrentPeriodUsage), new(XMLLastPeriodUsage), new(XMLProfileData))

	for scanner.Scan() {
		tag := scanner.Element()
		switch el := tag.(type) {
		case *XMLInstantaneousDemand:
			err = handleInstantaneousDemand(*el)

		case *XMLConnectionStatus:
			err = handleConnectionStatus(*el)

		case *DeviceInfo:
			err = handleDeviceInfo(*el)

		case *ScheduleInfo:
			err = handleScheduleInfo(*el)

		case *MeterList:
			err = handleMeterList(*el)

		case *MeterInfo:
			err = handleMeterInfo(*el)

		case *XMLNetworkInfo:
			err = handleNetworkInfo(*el)

		case *XMLTimeCluster:
			err = handleTimeCluster(*el)

		case *XMLMessageCluster:
			err = handleMessageCluster(*el)

		case *XMLPriceCluster:
			err = handlePriceCluster(*el)

		case *XMLCurrentSummationDelivered:
			err = handleCurrentSummationDelivered(*el)

		case *XMLCurrentPeriodUsage:
			err = handleCurrentPeriodUsage(*el)

		case *XMLLastPeriodUsage:
			err = handleLastPeriodUsage(*el)

		case *XMLProfileData:
			err = handleProfileData(*el)

		default:
			continue

		}
		if err != nil {
			break
		}
	}
	return err
}

var ctx context.Context

func processBytes(input chan byte, group *sync.WaitGroup) {
	defer group.Done()
	var err error = nil
	var startSeen bool = false
	var endSeen bool = false
	var possibleEnd bool = false
	var sol bool = true
	var invalid bool = false

	buff := make([]byte, 0)
	for {
		select {
		case i := <-input:
			if invalid {
				if i == '\n' {
					invalid = false
					sol = true
				}
			} else {
				if sol {
					if startSeen == false {
						if i == '<' {
							startSeen = true
							buff = append(buff, i)
						} else {
							invalid = true
						}
					} else {
						if i == '<' {
							possibleEnd = true
						}
						buff = append(buff, i)
					}
					sol = false
				} else {
					if i == '\n' {
						sol = true
					}
					if possibleEnd {
						if i == '/' {
							endSeen = true
						}
						possibleEnd = false
					}
					buff = append(buff, i)
					if endSeen {
						if i == '\n' {
							err = processXML(buff)
							if err != nil {
								log.Panic(err)
							}
							endSeen = false
							startSeen = false
							sol = true
							invalid = false
							possibleEnd = false
							buff = make([]byte, 0)
						}
					}
				}
			}
			if err != nil {
				log.Panic(err)
			}

		case <-ctx.Done():
			return

		}
	}
}

func readThread(port serial.Port, group *sync.WaitGroup) {
	defer group.Done()
	ch := make(chan byte, 250)
	group.Add(1)
	go processBytes(ch, group)

	for true {
		buf := make([]byte, 250)
		val, err := port.Read(buf)
		if err != nil {
			close(ch)
			log.Fatal(err)
		}
		for i := 0; i < val; i++ {
			ch <- buf[i]
		}

		if ctx.Err() != nil {
			close(ch)
			return
		}
	}
}

type pubTuple struct {
	topic string
	data  string
}

func mqttThread(data <-chan pubTuple, group *sync.WaitGroup, manager *autopaho.ConnectionManager) {
	defer group.Done()

	for {
		select {
		case out := <-data:
			go func(data pubTuple) {
				pr, err := manager.Publish(ctx, &paho.Publish{
					QoS:     2,
					Topic:   topicHeader + out.topic,
					Payload: []byte(out.data),
				})
				if err != nil {
					fmt.Printf("error publishing: %s\n", err)
				} else if pr.ReasonCode != 0 && pr.ReasonCode != 16 { // 16 = Server received message but there are no subscribers
					fmt.Printf("reason code %d received\n", pr.ReasonCode)
				}
			}(out)

		case <-ctx.Done():
			return
		}
	}
}

func main() {
	mode := &serial.Mode{
		BaudRate: 115200, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit,
	}

	port, err := serial.Open("/dev/emu2", mode)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer func(port serial.Port) {
		err := port.Close()
		if err != nil {
			panic(err)
		}
	}(port)

	brokerUrl, err := url.Parse("tcp://192.168.1.5:1883")
	connectRetryDelay := time.Duration(10000) * time.Millisecond

	config := autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{brokerUrl},
		KeepAlive:         30,
		ConnectRetryDelay: connectRetryDelay,
		OnConnectionUp:    func(*autopaho.ConnectionManager, *paho.Connack) { fmt.Println("mqtt connection up") },
		OnConnectError:    func(err error) { fmt.Printf("error whilst attempting connection: %s\n", err) },
		Debug:             log.Default(),
		ClientConfig: paho.ClientConfig{
			ClientID:      "emu2",
			OnClientError: func(err error) { fmt.Printf("server requested disconnect: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}
	config.SetUsernamePassword("emu2", []byte("veiv5a5c52FBxEE"))
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	cm, err := autopaho.NewConnection(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go readThread(port, wg)

	cmChan := make(chan pubTuple)
	mqttOut = cmChan

	wg.Add(1)
	go mqttThread(cmChan, wg, cm)

	// Wait for a signal before exiting
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	fmt.Println("signal caught - exiting")
	cancel()

	wg.Wait()
	fmt.Println("shutdown complete")
}

func handleProfileData(data XMLProfileData) error {
	return nil
}

func handleLastPeriodUsage(usage XMLLastPeriodUsage) error {
	lastUsage, err := convertXMLLastPeriodUsage(usage)
	if err != nil {
		return err
	}
	println(lastUsage.LastUsage.String())
	return nil
}

func handleCurrentPeriodUsage(usage XMLCurrentPeriodUsage) error {
	curUsage, err := convertXMLCurrentPeriodUsage(usage)
	if err != nil {
		return err
	}
	out := pubTuple{
		topic: "Period Usage",
		data:  curUsage.CurrentUsage.String(),
	}
	mqttOut <- out

	out.topic = "Period Timestamp"
	out.data = curUsage.TimeStamp.String()
	mqttOut <- out

	out.topic = "Period Started"
	out.data = curUsage.StartDate.String()
	mqttOut <- out

	return nil

}

func handleCurrentSummationDelivered(delivered XMLCurrentSummationDelivered) error {
	sumDelivered, err := convertXMLCurrentSummationDelivered(delivered)
	if err != nil {
		return err
	}
	println(sumDelivered.SummationDelivered.String())
	out := pubTuple{
		topic: "Total Delivered",
		data:  sumDelivered.SummationDelivered.String(),
	}
	mqttOut <- out

	out.topic = "Total Received"
	out.data = sumDelivered.SummationReceived.String()
	mqttOut <- out

	return nil

}

func handlePriceCluster(cluster XMLPriceCluster) error {
	priceCluster, err := convertXMLPriceCluster(cluster)
	if err != nil {
		return nil
	}
	out := pubTuple{"Price", priceCluster.Price.String()}
	mqttOut <- out
	return nil

}

func handleMessageCluster(cluster XMLMessageCluster) error {
	return nil

}

func handleTimeCluster(cluster XMLTimeCluster) error {
	timeCluster, err := convertXMLTimeCluster(cluster)
	if err != nil {
		return err
	}
	println(timeCluster.LocalTime.String())
	return nil

}

func handleNetworkInfo(info XMLNetworkInfo) error {
	return nil

}

func handleMeterInfo(info MeterInfo) error {
	return nil

}

func handleMeterList(list MeterList) error {
	return nil

}

func handleScheduleInfo(info ScheduleInfo) error {
	return nil

}

func handleDeviceInfo(info DeviceInfo) error {
	return nil

}

func handleConnectionStatus(status XMLConnectionStatus) error {
	connStatus, err := convertXMLConnectionStatus(status)
	if err != nil {
		return err
	}

	out := pubTuple{
		topic: "Status",
		data:  connStatus.Status,
	}

	mqttOut <- out
	return nil
}

func handleInstantaneousDemand(inst XMLInstantaneousDemand) error {
	instDemand, err := convertXMLInstantaneousDemand(inst)
	if err != nil {
		println(err)
		return err
	}

	out := pubTuple{
		topic: "Instantaneous Demand",
		data:  instDemand.Demand.String(),
	}
	mqttOut <- out

	out.topic = "Message Timestamp"
	out.data = instDemand.TimeStamp.String()
	mqttOut <- out

	return nil
}
