package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

//	returns a random string
func GetRand() string {
	randInt := randInt(0, 9999)
	hash := md5.New()
	io.WriteString(hash, string(randInt))
	return fmt.Sprintf("%x_%v", hash.Sum(nil), time.Now().UnixNano())
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
