package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/charmbracelet/lipgloss"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/voormedia/voormedia-toolkit/pkg/util"
)

var (
	readyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	hintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// Run Google Cloud SQL proxy container
func Run(log *util.Logger, port string) error {
	if project, err := util.GetCurrentGCPProject(); err == nil && project != "" {
		fmt.Println(util.GCPBanner(project))
	}

	sqlInstances, err := util.FindSQLInstances()
	if err != nil {
		return err
	}

	q := []*survey.Question{
		{
			Name: "instance",
			Prompt: &survey.Select{
				Message: "Choose a SQL instance:",
				Options: sqlInstances,
			},
		},
	}

	selection := struct{ Instance string }{}

	err = survey.Ask(q, &selection)
	if err != nil {
		return err
	}

	connectionName, err := util.FindConnectionName(selection.Instance)
	if err != nil {
		return err
	}

	proxyFile, err := findProxyFile()
	if err != nil {
		return err
	}

	args := []string{
		"-instances", connectionName + "=tcp:" + port,
	}

	cmd := exec.Command(proxyFile, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		scanProxyStderr(stderrPipe, port)
	}()

	<-done
	cmd.Wait()
	return nil
}

func scanProxyStderr(r io.Reader, port string) {
	scanner := bufio.NewScanner(r)
	ready := false
	for scanner.Scan() {
		line := scanner.Text()
		if !ready && strings.Contains(line, "Ready for new connections") {
			ready = true
			fmt.Fprintf(os.Stderr, "%s Cloud SQL proxy listening on localhost:%s %s\n",
				readyStyle.Render("✓"),
				port,
				hintStyle.Render("(Ctrl+C to stop)"),
			)
			continue
		}
		fmt.Fprintln(os.Stderr, line)
	}
}

func findProxyFile() (string, error) {
	proxyFile, err := homedir.Expand("~/cloud_sql_proxy")
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(proxyFile); os.IsNotExist(err) {
		if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			return "", errors.Errorf("Unsupported OS: %s", runtime.GOOS)
		}
		url := fmt.Sprintf("https://dl.google.com/cloudsql/cloud_sql_proxy.%s.%s", runtime.GOOS, runtime.GOARCH)
		if err := downloadProxy(url, proxyFile); err != nil {
			return "", errors.Errorf("Failed to download proxy file: %s", err.Error())
		}
		if err := os.Chmod(proxyFile, 0755); err != nil {
			return "", errors.Errorf("Failed to make proxy file executable: %s", err.Error())
		}
	}

	return proxyFile, nil
}

func downloadProxy(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("Unexpected response: %s", resp.Status)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	pw := util.NewProgressWriter(f, "", resp.ContentLength).SetHeader("Downloading Cloud SQL proxy")
	if _, err := io.Copy(pw, resp.Body); err != nil {
		pw.FinishFail()
		return err
	}
	pw.Finish()
	return nil
}
