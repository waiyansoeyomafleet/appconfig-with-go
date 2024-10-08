package utils

import (
	"os"
	"strconv"
)

func GetX() int {
	env := os.Getenv("X")
	i, err := strconv.Atoi(env)
	Check(err)
	return i
}
