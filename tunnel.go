package ngrok

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type Tunnel struct {
	addr    string
	url     string
	process *os.Process
}

func Open(ctx context.Context, addr string) (*Tunnel, error) {
	t := &Tunnel{addr: addr}
	done := make(chan bool, 1)
	var err error
	go func() {
		err = t.Start()
		done <- true
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return t, err
	}
}

func (t *Tunnel) Start() error {

	err := check()
	if err != nil {
		return fmt.Errorf("failed checking ngrok executable: %w", err)
	}

	cmd := exec.Command(ngrok, "http", t.addr, "--log", "stdout", "--log-format", "json")

	out, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed getting stdout pipe: %w", err)
	}

	sc := bufio.NewScanner(out)

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed starting ngrok: %w", err)
	}

	var log struct {
		Msg string
		Url string
	}

	for sc.Scan() {
		err := json.Unmarshal(sc.Bytes(), &log)
		if err != nil {
			continue
		}

		if log.Msg == "started tunnel" {
			t.url = log.Url
			break
		}
	}

	t.process = cmd.Process

	return nil

}

func (t *Tunnel) Close() error {
	if t.process != nil {
		return t.process.Kill()
	}
	return nil
}

func (t *Tunnel) Url() string {
	return t.url
}
