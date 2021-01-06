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
	device.SetFrequency(50)

	//Creates channel to manipulate
	pwmChannel := device.GetChannels(14, 2)

	//Set ON period for 500 microseconds on channel 14
	pwmChannel.SetOnPeriodDuration(500)
	time.Sleep(time.Second)

	//Set on period for 2500 microseconds shifted on 500 microseconds from period's start
	pwmChannel.SetOnPeriodDuration(2500)
	time.Sleep(time.Second)
}
