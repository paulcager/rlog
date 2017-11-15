# github.com/paulcager/rlog

## Lightweight rolling log file package.

Package rlog provides a rotated log file. Log files will be rotated on a given time boundary (e.g.
daily), and an action (such as compression) may be specified to be performed when rotation happens.

### Usage

	w, err := rlog.NewWriter(tmp+"/rlog-test-$h$m$s.txt", time.Minute, rlog.GZIPOnRotate)
	if err != nil {
		panic(err)
	}
  defer w.Close()
  fmt.FPrintln(w, "Hello, world")
