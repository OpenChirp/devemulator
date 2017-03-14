package envbug

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/golang/protobuf/proto"
)

type Device struct {
	id string
}

func (d *Device) Init(devid string) {
	d.id = devid
}

func (d *Device) TX() []byte {
	m := &EnvBugMessage{
		Temperature: int32(rand.Intn(70)),
		Humidity:    uint32(rand.Intn(100)),
		Light:       uint32(rand.Intn(20)),
		Pir:         uint32(rand.Intn(10)),
		Mic:         uint32(rand.Intn(20)),
		AccX:        10 - uint32(rand.Intn(5)),   // some jitter
		AccY:        550 - uint32(rand.Intn(5)),  // some jitter
		AccZ:        1024 - uint32(rand.Intn(5)), // some jitter
	}
	data, err := proto.Marshal(m)
	if err != nil {
		log.Fatal(d.id+"(envbug): Error - Marshaling MIO Message: ", err)
	}
	return data
}

func (d *Device) RX(rxdata []byte) []byte {
	m := &EnvBugMessage{}

	err := proto.Unmarshal(rxdata, m)
	if err != nil {
		log.Println(d.id + "(envbug): Failed to unmarshal rxdata")
		return nil
	}

	fmt.Printf("%s (envbug): Instructed to change reporting duty cycle to %d\n", d.id, m.Duty)

	return nil
}
