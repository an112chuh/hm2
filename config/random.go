package config

import (
	"math/rand"
	"time"
)

func InitRandom() {
	rand.Seed(time.Now().UnixNano())
}
