/**
 * Usage: devemulater --runtime     <runtime_in_seconds>
 *                    --mqtt_broker <broker_uri>
 *                    --mqtt_user   <username>
 *                    --mqtt_pass   <password>
 *                    --mqtt_qos    <qos_value>
 *
 * The runtime argument is the number of seconds to run the simulation for.
 * @author Craig Hesling <craig@hesling.com>
 */
package main

import (
	CRAND "crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"reflect"

	"github.com/openchirp/devemulator/devices/envbug"
	"github.com/openchirp/devemulator/devices/labdoor"
)

const (
	defaultframeworkserver = "http://localhost"
	mqttdefaultbroker      = "tcp://localhost:1883"
	mqttclientidprefix     = "device_simulator"
	mqttdefaultqos         = 0
)

/* Options to be filled in by arguments */
var mqttBroker string
var mqttUser string
var mqttPass string
var mqttQos uint
var runtime uint64

/* Generate a random client id for mqtt */
func genclientid() string {
	r, err := CRAND.Int(CRAND.Reader, new(big.Int).SetInt64(100000))
	if err != nil {
		log.Fatal("Couldn't generate a random number for MQTT client ID")
	}
	return mqttclientidprefix + r.String()
}

/* Setup argument flags and help prompt */
func init() {
	flag.StringVar(&mqttBroker, "mqtt_broker", mqttdefaultbroker, "Sets the MQTT broker")
	flag.StringVar(&mqttUser, "mqtt_user", "", "Sets the MQTT username")
	flag.StringVar(&mqttPass, "mqtt_pass", "", "Sets the MQTT password")
	flag.UintVar(&mqttQos, "mqtt_qos", mqttdefaultqos, "Sets the MQTT QOS to use when publishing and subscribing [0, 1, or 2]")
	flag.Uint64Var(&runtime, "runtime", 0, "Sets the duration in seconds to run the simulator for")
}

// OpenChirp MQTT Info
const (
	mqtt_topic_root   = "devices"
	mqtt_raw_rx_label = "rawrx" // rx from end user perspective
	mqtt_raw_tx_label = "rawtx" // tx from end user perspective
)

const (
	// Max random sleep time in Milliseconds
	device_max_sleep = 1000
	// device_max_sleep = 100
	device_max_rand_field = 3
)

type deviceConfig struct {
	mqttNode    string
	txPeriodMs  uint32
	devInstance interface{} // Should implement SimpleDevice or AdvancedDevice
}

func (dev deviceConfig) txTopic() string {
	return mqtt_topic_root + "/" + dev.mqttNode + "/" + mqtt_raw_rx_label
}

func (dev deviceConfig) rxTopic() string {
	return mqtt_topic_root + "/" + dev.mqttNode + "/" + mqtt_raw_tx_label
}

// List of virtual devices - mqtt topic, period in ms, device type runtime
var devices = []deviceConfig{
	{"13541c75-2066-482a-970e-0a89d712ad91", 1 * 60 * 1000, &labdoor.Device{}},
	{"e394dbc9-faca-4715-b5f2-3fd0cbc7467a", 1000, &envbug.Device{}},
	{"abfc2077-d78c-4081-8a0f-c3c5e0b557c0", 1000, &envbug.Device{}},
	{"3acbbd36-59cc-4e91-8cdc-49e21ab22a31", 5000, &envbug.Device{}},
	{"ab93c086-d494-4b4b-8cbe-182a8e482f3e", 500, &envbug.Device{}},
	{"8808d035-77a3-4576-ab0e-03384126d3cb", 500, &envbug.Device{}},
	{"a2fe5f15-36fd-4e09-9717-634e23d5bc45", 250, &envbug.Device{}},
}

func deviceTopic(dev string) string {
	return mqtt_topic_root + "/" + dev + "/" + mqtt_raw_rx_label
}

// Message type between simulation devices and mqtt publisher
type message struct {
	device string
	base64 string
}

// SimpleDevice specifies the interface for vitual device that periodically
// transmits and never shutdown. The simulation will call TX at some defined period.
type SimpleDevice interface {
	// funtion that will be called once to initialize each new device
	Init(devid string)
	// should return the bytes that the device would tx
	TX() []byte
	// input data the device should receive
	// should return any data the device would like to transmit in response
	RX(rxdata []byte) []byte
}

// AdvancedDevice specifies the interface for vitual device that randomly
// transmits. The simulation will call TX randomly
type AdvancedDevice interface {
	// funtion that will be called once to initialize each new device
	Init(devid string)
	Run(devrx <-chan []byte, devtx chan<- []byte)
}

func runSimpleDevice(client MQTT.Client, devCfg deviceConfig) {
	devCfg.devInstance.(SimpleDevice).Init(devCfg.mqttNode)

	// Sleep some random time to emulate random turn on time
	time.Sleep(time.Duration(rand.Intn(device_max_sleep)) * time.Millisecond)

	client.Subscribe(devCfg.rxTopic(), byte(mqttQos), func(c MQTT.Client, m MQTT.Message) {
		bytes, err := base64.StdEncoding.DecodeString(string(m.Payload()))
		if err != nil {
			log.Println("Error base64 decoding message for " + m.Topic())
			return
		}
		resp := devCfg.devInstance.(SimpleDevice).RX(bytes)
		if resp != nil {
			resp64 := base64.StdEncoding.EncodeToString(resp)
			client.Publish(devCfg.txTopic(), byte(mqttQos), false, resp64)
		}
	})

	for {
		tx := devCfg.devInstance.(SimpleDevice).TX()
		if tx != nil {
			tx64 := base64.StdEncoding.EncodeToString(tx)
			client.Publish(devCfg.txTopic(), byte(mqttQos), false, tx64)
		}
		time.Sleep(time.Duration(devCfg.txPeriodMs) * time.Millisecond)
	}

}

func runAdvancedDevice(client MQTT.Client, devCfg deviceConfig) {
	log.Fatalln("AdvancedDevice functionality is not implemented")
	// devrxdata := make(chan []byte)
	// devtxdata := make(chan []byte)
	// quit := make(chan bool)

	// go func() {
	// 	for {
	// 		select {
	// 		// case tx <- devtxdata:

	// 		case <-quit:
	// 			return
	// 		}
	// 	}

	// }()

	// devCfg.devInstance.(AdvancedDevice).Init(devCfg.mqttNode)
	// devCfg.devInstance.(AdvancedDevice).Run(devrxdata, devtxdata)
}

func main() {
	/* Parse Arguments */
	flag.Parse()

	fmt.Println("# Starting up")

	/* Setup basic MQTT connection */
	opts := MQTT.NewClientOptions().AddBroker(mqttBroker)
	opts.SetClientID(genclientid())
	opts.SetUsername(mqttUser)
	opts.SetPassword(mqttPass)

	/* Create and start a client using the above ClientOptions */
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer c.Disconnect(250)

	/* Start a go routine for each virtual device */
	for _, dev := range devices {
		fmt.Print("# Running device " + dev.mqttNode + " type ")
		fmt.Println(reflect.TypeOf(dev.devInstance))
		if _, ok := dev.devInstance.(SimpleDevice); ok {
			go runSimpleDevice(c, dev)
		} else if _, ok := dev.devInstance.(AdvancedDevice); ok {
			go runAdvancedDevice(c, dev)
		} else {
			log.Fatal("Device " + dev.mqttNode + " is of invalid type")
		}
	}

	/* Setup Stuff to Manage a Safe Exit */

	timeout := make(chan bool)
	// signal for SIGINT
	signals := make(chan os.Signal)

	if runtime > 0 {
		fmt.Println("# Running for ", runtime, " second(s)")
		// run for argument seconds if specified
		go func() {
			<-time.After(time.Duration(runtime) * time.Second)
			timeout <- true
		}()
	}

	signal.Notify(signals, os.Interrupt)

	select {
	case <-timeout:
		fmt.Println("# Reached end of timed life")
	case sig := <-signals:
		fmt.Println("# Caught signal ", sig)
	}

	fmt.Println("# Ending")
}
