package main

import (
	"time"

	"github.com/d2r2/go-i2c"
	pca9685 "github.com/xpr0ger/go-pca9685/v1"
)

//This is example for Sg90 servo
func main() {
	// Create I2C bus
	bus, err := i2c.NewI2C(pca9685.DefaultAddress, 1)
	if err != nil {
		println(err.Error())
		return
	}
	defer bus.Close()

	//Create ReadWriter wrapper
	busWrapper := pca9685.NewI2CWrapper(bus)

	//Create PCA9685 device
	device, err := pca9685.NewPCA9685(busWrapper)
	if err != nil {
		println(err.Error())
		return
	}

	//Turn off all channels at the end
	defer device.TurnOff()

	//Set PWM frequency
	err = device.SetFrequency(1526)
	time.Sleep(time.Second * 5)
	if err != nil {
		println(err.Error())
		return
	}

	//Create two channels 0 and 1 to manipulate
	chs := device.GetChannels(0, 2)

	for i := uint16(0); i <= 4095; i++ {
		time.Sleep(2000)
		chs.SetPeriod(0, i)
	}

	time.Sleep(time.Second * 5)

	for i := uint16(0); i <= 4095; i++ {
		time.Sleep(2000)
		chs.SetPeriod(i, 4095)
	}
}
