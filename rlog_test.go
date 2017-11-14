package rlog

import (
	"fmt"
	"os"
	"testing"
	"time"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateFileName(t *testing.T) {
	w := Writer{filePattern: "Year$YMonth$MDay$D.txt"}
	period := time.Date(2017, 2, 3, 1, 0, 0, 0, time.UTC)
	w.updateFilename(period)

	assert.EqualValues(t, "Year2017Month02Day03.txt", w.thisFile)
}

func TestRotate(t *testing.T) {
	testTime := time.Date(2017, 2, 3, 1, 0, 0, 0, time.UTC)
	rotated := ""
	now = func() time.Time {
		return testTime
	}
	tempDir := os.TempDir()
	pid := os.Getpid()

	var (
		testPattern   = fmt.Sprintf("%s/TestRotate-%d/$Y-$M-$D.txt", tempDir, pid)
		expectedDir   = fmt.Sprintf("%s/TestRotate-%d", tempDir, pid)
		expectedFile1 = fmt.Sprintf("%s/2017-02-03.txt", expectedDir)
		expectedFile2 = fmt.Sprintf("%s/2017-02-04.txt", expectedDir)
	)
	//defer os.RemoveAll(expectedDir)

	w, err := NewWriter(testPattern, 24*time.Hour, func(old string) { rotated = old })
	require.NoError(t, err)
	defer w.Close()

	require.NotNil(t, w)

	fmt.Fprintln(w, "First Day, First Line")
	fmt.Fprintln(w, "First Day, Second Line")
	info, err := os.Stat(expectedFile1)
	require.NoError(t, err)
	assert.False(t, info.IsDir())

	// Now move onto the next day
	testTime = testTime.Add(24 * time.Hour)
	fmt.Fprintln(w, "Second Day, First Line")
	fmt.Fprintln(w, "Second Day, Second Line")
	assert.EqualValues(t, expectedFile1, rotated)
	info, err = os.Stat(expectedFile2)
	require.NoError(t, err)

	w.Close()

	contents, err := ioutil.ReadFile(expectedFile1)
	require.NoError(t, err)
	assert.EqualValues(t, "First Day, First Line\nFirst Day, Second Line\n", string(contents))
	contents, err = ioutil.ReadFile(expectedFile2)
	require.NoError(t, err)
	assert.EqualValues(t, "Second Day, First Line\nSecond Day, Second Line\n", string(contents))

	// Now reopen and verify that we append.
	w, err = NewWriter(testPattern, 24*time.Hour, func(old string) { rotated = old })
	require.NoError(t, err)
	fmt.Fprintln(w, "Appended")
	w.Close()
	contents, err = ioutil.ReadFile(expectedFile2)
	require.NoError(t, err)
	assert.EqualValues(t, "Second Day, First Line\nSecond Day, Second Line\nAppended\n", string(contents))
}

func TestReopen(t *testing.T) {
}
