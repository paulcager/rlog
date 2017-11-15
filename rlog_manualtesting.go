// +build manual_test

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/paulcager/rlog"
)

func main() {
	tmp := os.TempDir()
	w, err := rlog.NewWriter(tmp+"/rlog-test-$h$m$s.txt", time.Minute, rlog.GZIPOnRotate)
	if err != nil {
		panic(err)
	}

	defer w.Close()
	for i := 0; i < 6*5; i++ {
		time.Sleep(10 * time.Second)
		fmt.Fprintln(w, time.Now())
	}
}
