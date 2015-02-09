package utils

import (
	"fmt"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"
)

var rg = rand.New(rand.NewSource(time.Now().Unix()))

func UuidV4String() string {
	d := uuid.NewV4()
	return d.String()
}

func randomByte(gradient, floor int) byte {
	if gradient == 0 {
		gradient = 1
	}
	max := int(255 / gradient)
	return byte((rg.Intn(max-floor) + floor) * gradient)
}

func RandomColor() [3]byte {
	return [3]byte{randomByte(5, 5), randomByte(5, 5), randomByte(5, 5)}
}

func RandomColorString() string {
	color := RandomColor()
	return fmt.Sprintf("#%x%x%x", color[0], color[1], color[2])
}
