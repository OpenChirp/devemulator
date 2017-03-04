package labdoor

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
	m := &LabDoorMessage{}
	if rand.Intn(1) == 0 {
		m.Doorstatus = LabDoorMessage_OPEN
	} else {
		m.Doorstatus = LabDoorMessage_CLOSED
	}
	data, err := proto.Marshal(m)
	if err != nil {
		log.Fatal(d.id+"(labdoor): Error - Marshaling MIO Message: ", err)
	}
	return data
}

func (d *Device) RX(rxdata []byte) []byte {
	m := &LabDoorMessage{}

	err := proto.Unmarshal(rxdata, m)
	if err != nil {
		log.Println(d.id + "(labdoor): Failed to unmarshal rxdata")
		return nil
	}
	fmt.Println(d.id + "(labdoor): ")

	return nil
}
