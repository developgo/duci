package github_test

import (
	"github.com/duck8823/minimal-ci/infrastructure/logger"
	"github.com/duck8823/minimal-ci/service/github"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
)

var regex = regexp.MustCompile("\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}.\\d{3}")

func TestProgressLogger_Write(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("error occured. %+v", err)
	}

	logger.Init(writer, logging.DEBUG)

	progress := &github.ProgressLogger{}
	progress.Write([]byte("hoge\rfuga"))

	writer.Close()

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatalf("error occured. %+v", err)
	}

	log := string(bytes)
	if !regex.MatchString(log) {
		t.Fatalf("invalid format. %+v", log)
	}

	actual := strings.TrimRight(regex.ReplaceAllString(log, ""), "\n")
	expected := " [INFO    ]\033[0m hoge"

	if actual != expected {
		t.Errorf("must remove CR flag or later. wont: %+v, but got: %+v", expected, actual)
	}
}
