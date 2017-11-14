// Package rlog provides a rotated log file. Log files will be rotated on a given time boundary (e.g.
// daily), and an action (such as compression) may be specified to be performed when rotation happens.
package rlog

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ErrClosed is the error used for write operations on a closed pipe.
var ErrClosedPipe = errors.New("write on closed pipe")

type RotationFunction func(filename string)

type Writer struct {
	OnRotate    RotationFunction
	w           io.WriteCloser
	closed      bool
	filePattern string
	thisFile    string
	nextRotate  time.Time
	period      time.Duration
}

// NewWriter creates a rotating log based on the given filename pattern. The log will be
// rotated at the specified period. If supplied, the RotationFunction will be called once
// the old file has been closed.
//
// The filename pattern may contain special characters that are replaced:
//		Pattern		Replacement
//		$Y			4-digit year
//		$M			2-digit month
//		$D			2-digit day of month
//		$h			2-digit hour
//		$m			2-digit minute
//		$s			2-digit second
func NewWriter(filePattern string, period time.Duration, onRotate RotationFunction) (io.WriteCloser, error) {
	w := &Writer{filePattern: filePattern, period: period, OnRotate: onRotate}
	return w, w.rotate()
}

// NewDailyWriter is a convenience method for the common case where logs are rotated daily and old
// logs are gzipped.
func NewDailyWriter(filePattern string) (io.WriteCloser, error) {
	return NewWriter(filePattern, 24*time.Hour, GZIPOnRotate)
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, os.ErrClosed
	}

	if now().After(w.nextRotate) {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	return w.w.Write(p)
}

func (w *Writer) Close() error {
	w.closed = true
	return w.w.Close()
}

func (w *Writer) rotate() error {
	var err error
	prevFile := w.thisFile
	if w.w != nil {
		//w.w will be nil only when we are called to create the first file.
		if err := w.w.Close(); err != nil {
			return err
		}
		if w.OnRotate != nil {
			go w.OnRotate(prevFile)
		}
	}

	thisPeriod := now().Truncate(w.period)
	w.updateFilename(thisPeriod)
	os.MkdirAll(filepath.Dir(w.thisFile), 0755)
	w.w, err = os.OpenFile(w.thisFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	w.nextRotate = thisPeriod.Add(w.period)
	return err
}

func (w *Writer) updateFilename(period time.Time) {
	thisFile := w.filePattern
	thisFile = strings.Replace(thisFile, "$Y", fmt.Sprintf("%04d", period.Year()), -1)
	thisFile = strings.Replace(thisFile, "$M", fmt.Sprintf("%02d", period.Month()), -1)
	thisFile = strings.Replace(thisFile, "$D", fmt.Sprintf("%02d", period.Day()), -1)
	thisFile = strings.Replace(thisFile, "$h", fmt.Sprintf("%02d", period.Hour()), -1)
	thisFile = strings.Replace(thisFile, "$m", fmt.Sprintf("%02d", period.Minute()), -1)
	thisFile = strings.Replace(thisFile, "$s", fmt.Sprintf("%02d", period.Second()), -1)
	w.thisFile = thisFile
}

func GZIPOnRotate(filename string) {
	runAndLog(filename, "gzip", "-n")
}

func XZIPOnRotate(filename string) {
	runAndLog(filename, "xz")
}

func runAndLog(filename string, cmd string, args ...string) {
	out, err := exec.Command(cmd, append(args, filename)...).Output()
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Stdout.Write(out)

	switch err := err.(type) {
	case *exec.ExitError:
		os.Stderr.Write(err.Stderr)
	}
}

// For testing
var now = func() time.Time {
	return time.Now()
}
