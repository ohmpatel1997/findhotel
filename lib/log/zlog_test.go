package zlog_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/ohmpatel1997/findhotel/lib/log"
)

func captureOutput(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()
	os.Stdout = writer
	os.Stderr = writer
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	writer.Close()
	return <-out
}

func TestLog_Info(t *testing.T) {
	output := captureOutput(func() {
		zlogger := zlog.New()
		zlogger.Info(
			"msg",
			map[string]interface{}{
				"req": "req",
				"res": "res",
			},
		)
	})

	fieldsMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(output), &fieldsMap)
	if err != nil {
		t.Error(err)
	}

	expectedFields := []string{"status", "time", "message", "req", "res"}
	for _, field := range expectedFields {
		if _, ok := fieldsMap[field]; !ok {
			t.Errorf(field + " doesn't exist")
		}
	}
}

func TestLog_Error(t *testing.T) {
	output := captureOutput(func() {
		zlogger := zlog.New()
		zlogger.Error(
			"msg",
			errors.New("generic"),
			map[string]interface{}{
				"key1": "val1",
				"key2": "val2",
			},
		)
	})

	fieldsMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(output), &fieldsMap)
	if err != nil {
		t.Error(err)
	}

	expectedFields := []string{"status", "time", "message", "key1", "key2"}
	for _, field := range expectedFields {
		if _, ok := fieldsMap[field]; !ok {
			t.Errorf(field + " doesn't exist")
		}
	}
}

func TestLog_Warning(t *testing.T) {
	output := captureOutput(func() {
		zlogger := zlog.New()
		zlogger.Warn(
			"msg",
			map[string]interface{}{
				"key1": "val1",
				"key2": "val2",
			},
		)
	})

	fieldsMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(output), &fieldsMap)
	if err != nil {
		t.Error(err)
	}

	expectedFields := []string{"status", "time", "message", "key1", "key2"}
	for _, field := range expectedFields {
		if _, ok := fieldsMap[field]; !ok {
			t.Errorf(field + " doesn't exist")
		}
	}
}

func TestLog_Debug(t *testing.T) {
	output := captureOutput(func() {
		zlogger := zlog.New()
		zlogger.Warn(
			"msg",
			map[string]interface{}{
				"key1": "val1",
				"key2": "val2",
			},
		)
	})

	fieldsMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(output), &fieldsMap)
	if err != nil {
		t.Error(err)
	}

	expectedFields := []string{"status", "time", "message", "key1", "key2"}
	for _, field := range expectedFields {
		if _, ok := fieldsMap[field]; !ok {
			t.Errorf(field + " doesn't exist")
		}
	}
}
