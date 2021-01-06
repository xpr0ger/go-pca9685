package pca9685

import "github.com/d2r2/go-i2c"

//I2CWrapper wrapper struct with ReadWriter implementation to separete I2C implementation
type I2CWrapper struct {
	Bus *i2c.I2C
}

//NewI2CWrapper wrapper constructor
func NewI2CWrapper(bus *i2c.I2C) *I2CWrapper {
	return &I2CWrapper{Bus: bus}
}

func (i *I2CWrapper) Read(p []byte) (n int, err error) {
	return i.Bus.ReadBytes(p)
}

func (i *I2CWrapper) Write(p []byte) (n int, err error) {
	return i.Bus.WriteBytes(p)
}
