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
	addr     string
	url      string
	process  *os.Process
	token    string
	agentUrl string
}

type Options struct {
	Addr      string
	AuthToken string
}

func Open(ctx context.Context, opts Options) (*Tunnel, error) {
	t := &Tunnel{addr: opts.Addr, token: opts.AuthToken}
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
	if t.token != "" {
		cmd.Args = append(cmd.Args, "--authtoken", t.token)
	}

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
		Msg  string
		Url  string
		Addr string
	}

	for sc.Scan() {
		err := json.Unmarshal(sc.Bytes(), &log)
		if err != nil {
			continue
		}

		if log.Msg == "starting web service" {
			t.agentUrl = log.Addr
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

func (t *Tunnel) AgentUrl() string {
	return t.agentUrl
}
