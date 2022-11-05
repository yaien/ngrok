package ngrok

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
)

func TestCheck(t *testing.T) {

	server := func(target string) (*httptest.Server, *bool) {
		t.Helper()
		called := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res, err := http.Get(target)
			if err != nil {
				t.Fatalf("failed downloading file: %s", err)
			}
			res.Write(w)
			called = true
		}))

		return server, &called
	}

	t.Run("NotDownloaded", func(t *testing.T) {
		os.RemoveAll(config)
		target := url
		srv, called := server(target)
		defer srv.Close()

		t.Cleanup(func() {
			url = target
		})

		url = srv.URL
		check()
		if !*called {
			t.Error("expected server to be called")
		}
	})

	t.Run("Downloaded", func(t *testing.T) {
		target := url
		srv, called := server(target)
		defer srv.Close()

		t.Cleanup(func() {
			url = target
		})

		url = srv.URL
		check()
		if *called {
			t.Error("expected server to not be called")
		}
	})
}

func TestDownload(t *testing.T) {

	err := download()
	if err != nil {
		t.Fatalf("download failed: %s", err)
	}

	cmd := exec.Command(ngrok, "-v")

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("command execution failed: %s", err)
	}

	t.Log(string(output))

}
