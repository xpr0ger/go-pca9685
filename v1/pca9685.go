package pca9685

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/pkg/errors"
)

const (
	//ClockDeviceHZ is device Clock speed, 25 MHz according to datasheet
	ClockDeviceHZ = 25_000_000

	//ClockPeriodsFractionsCount is clock period resolution according to datasheet is 12-bit value 0 - 4095
	ClockPeriodsFractionsCount = 4096

	//ClockPeriodsFractionsMaxValue is clock period resolution according to datasheet is 12-bit value 0 - 4095
	ClockPeriodsFractionsMaxValue = 0xfff

	//DefaultAddress is default device address to open i2c communication
	DefaultAddress = byte(0x40)

	//PrescaleMinValue is minimum prescale value allowed by device
	PrescaleMinValue = 3

	//PrescaleMaxValue is maximum prescale value allowed by device
	PrescaleMaxValue = 255

	//AddressMode1 is address of MODE1 register
	AddressMode1 = byte(0x00)

	//AddressMode2 is address of MODE2 register
	AddressMode2 = byte(0x01)

	//AddressSubAddr1 is address of SUBADDR1  see datasheet of PCA9685
	AddressSubAddr1 = byte(0x02)

	//AddressSubAddr2 is address of SUBADDR1  see datasheet of PCA9685
	AddressSubAddr2 = byte(0x03)

	//AddressSubAddr3 is address of SUBADDR1  see datasheet of PCA9685
	AddressSubAddr3 = byte(0x04)

	//AddressAllCallAdr is address of ALLCALLADR  see datasheet of PCA9685
	AddressAllCallAdr = byte(0x05)

	//AddressLed1OnLow is first chanel addres, each channel has 4 addresses:
	//0x06 - low on addres for chanel 1
	//0x07 - hi on addres for chanel 1
	//0x08 - low off address for chanel 1
	//0x09 - hi off address for chanel 1
	//
	//0x0A - low on address for channel 2
	//and so no
	AddressLed1OnLow = 0x06

	//AddressLedAllOnLow is the same as AddressLed1OnLow but for all channel
	AddressLedAllOnLow = 0xFA

	//AddressPrescale is PRE_SCALE addres it allow to define period frequency
	AddressPrescale = 0xFE

	bitMode1Reset             = 0b10000000
	bitMode1ExternalClock     = 0b01000000
	bitMode1AutoIncrement     = 0b00100000
	bitMode1Sleep             = 0b00010000
	bitMode1RespondToSubAddr1 = 0b00001000
	bitMode1RespondToSubAddr2 = 0b00000100
	bitMode1RespondToSubAddr3 = 0b00000010
	bitMode1RespondToAllCall  = 0b00000001
	fullOnOffValue            = 0x1000

	channelByteResolution = 0x04
)

//PCA9685 Structure that represent device
type PCA9685 struct {
	bus       io.ReadWriter
	frequency uint16
}

//NewPCA9685 Construct
//bus - i2c implementation to communicate with PCA9685
func NewPCA9685(bus io.ReadWriter) (*PCA9685, error) {
	b := &PCA9685{bus: bus}
	return b, b.Reset()
}

//Reset MODE1 register
func (p *PCA9685) Reset() error {
	_, err := p.bus.Write([]byte{AddressMode1, 0x0})
	return errors.Wrapf(err, "failed to reset device")
}

//SetFrequency how many impulses will be generated per second
//Min value is 25 Hz max value 1526 Hz
func (p *PCA9685) SetFrequency(clockSpeed uint) error {
	//Calculating prescale value for given frequency according to datasheet equation
	prescaler := (float64(ClockDeviceHZ) / float64(ClockPeriodsFractionsCount*clockSpeed)) - 1.0
	prescaler = math.RoundToEven(prescaler)

	//Check prescaler compliance
	if prescaler < float64(PrescaleMinValue) || prescaler > float64(PrescaleMaxValue) {
		return fmt.Errorf("device cannot oscillate on %d Hz", int(clockSpeed))
	}

	value := byte(prescaler)

	_, err := p.bus.Write([]byte{AddressMode1})
	if err != nil {
		return errors.Wrap(err, "failed to set address for read")
	}

	buf := make([]byte, 1)
	_, err = p.bus.Read(buf)
	mode1 := buf[0]
	if err != nil {
		return errors.Wrap(err, "failed to read MODE1 register")
	}

	//Turn Off reset bit and put device into sleep mode to change prescaler
	sleepMode1 := (mode1 & 0b01111111) | bitMode1Sleep

	_, err = p.bus.Write([]byte{AddressMode1, sleepMode1})
	if err != nil {
		return errors.Wrap(err, "failed to put device in sleep mode")
	}

	//Write prescaler value
	_, err = p.bus.Write([]byte{AddressPrescale, value})
	if err != nil {
		return errors.Wrap(err, "failed to change prescaler value")
	}

	//Restore initial state
	_, err = p.bus.Write([]byte{AddressMode1, mode1})
	if err != nil {
		return errors.Wrap(err, "failed to restore MODE1 register")
	}

	//Wait for wake up
	time.Sleep(2000)

	//Set up mode bit for restert
	resetMode1 := mode1 | bitMode1Reset | bitMode1AutoIncrement | bitMode1RespondToAllCall
	_, err = p.bus.Write([]byte{AddressMode1, resetMode1})
	if err != nil {
		return errors.Wrap(err, "failed to put device in restart mode")
	}

	//TODO: Read frequency in channel
	p.frequency = uint16(clockSpeed)
	return nil
}

//TurnOff shortcat to turn off all channels
func (p *PCA9685) TurnOff() error {
	return p.GetAllChannels().FullOff()
}

//GetChannel get single channel
func (p *PCA9685) GetChannel(channelNumber byte) *Channels {
	return NewChannels(p.bus, channelNumber*4+AddressLed1OnLow, p.frequency, 1)
}

//GetChannels get N chanels in row
func (p *PCA9685) GetChannels(channelNumber byte, channelsCount uint16) *Channels {
	return NewChannels(p.bus, channelNumber*4+AddressLed1OnLow, p.frequency, channelsCount)
}

//GetAllChannels allows to write to all channels at the same time
func (p *PCA9685) GetAllChannels() *Channels {
	return NewChannels(p.bus, AddressLedAllOnLow, p.frequency, 1)
}

//Channels hold all data of channel or chanels` series
type Channels struct {
	bus           io.ReadWriter
	addresLowOn   byte
	frequency     uint16
	channelsCount uint16
}

//NewChannels - channels` constructor
func NewChannels(bus io.ReadWriter, addresLowOn byte, frequency uint16, channelsCount uint16) *Channels {
	return &Channels{
		bus:           bus,
		addresLowOn:   addresLowOn,
		frequency:     frequency,
		channelsCount: channelsCount,
	}
}

//SetOnPeriodDuration set on period in microseconds
func (c *Channels) SetOnPeriodDuration(microseconds uint16) error {
	//Get total period duration in microseconds
	periodDuration := float64(time.Millisecond) / float64(c.frequency)
	if periodDuration < float64(microseconds) {
		return fmt.Errorf("on period duration can not be longer then %d ms", uint16(periodDuration))
	}

	return c.SetOnPeriodDurationWithShift(0, microseconds)
}

//SetOnPeriodDurationWithShift set on period with shift from start in milliseconds
func (c *Channels) SetOnPeriodDurationWithShift(onShiftMicroseconds, microseconds uint16) error {
	//Get total period duration in microseconds
	periodDuration := float64(time.Millisecond) / float64(c.frequency)
	if periodDuration < float64(onShiftMicroseconds+microseconds) {
		return fmt.Errorf("on period duration can not be longer then %d ms", uint16(periodDuration))
	}

	//Get how load one fraction is lasting
	periodFractionDuration := periodDuration / float64(ClockPeriodsFractionsMaxValue)

	onFractionNumber := math.RoundToEven(float64(onShiftMicroseconds) / periodFractionDuration)
	offFractionNumber := math.RoundToEven(float64(onShiftMicroseconds+microseconds) / periodFractionDuration)

	return c.SetPeriod(uint16(onFractionNumber), uint16(offFractionNumber))
}

//SetPeriod set On/Off period in fractions
func (c *Channels) SetPeriod(on, off uint16) error {
	if on > ClockPeriodsFractionsMaxValue {
		return fmt.Errorf("on value cannot be greater then %d", ClockPeriodsFractionsMaxValue)
	}

	if off > ClockPeriodsFractionsMaxValue {
		return fmt.Errorf("off value cannot be greater then %d", ClockPeriodsFractionsMaxValue)
	}

	return c.writeOnOffValue(on, off)

}

//FullOn make PWM output constant ON
func (c *Channels) FullOn() error {
	return c.writeOnOffValue(fullOnOffValue, 0)
}

//FullOff make PWM output constant OFF
func (c *Channels) FullOff() error {
	return c.writeOnOffValue(0, fullOnOffValue)
}

func (c *Channels) writeOnOffValue(on, off uint16) error {
	//1 byte reserved for address
	buf := make([]byte, c.channelsCount*channelByteResolution+1)
	buf[0] = c.addresLowOn

	for i := uint16(0); i < c.channelsCount; i++ {
		//First two bytes for on state
		binary.LittleEndian.PutUint16(buf[1+i*channelByteResolution:3+i*channelByteResolution], on)
		//Second two bytes for off state
		binary.LittleEndian.PutUint16(buf[3+i*channelByteResolution:5+i*channelByteResolution], off)
	}

	_, err := c.bus.Write(buf)
	return errors.Wrapf(err, "failed to write new state")
}
