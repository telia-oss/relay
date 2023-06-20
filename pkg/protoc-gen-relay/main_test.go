package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProtocGenRelay(t *testing.T) {
	protoFile, err := ioutil.TempFile(".", "test_*.proto")
	require.NoError(t, err)
	defer os.Remove(protoFile.Name())

	_, err = protoFile.WriteString(`
	syntax = "proto3";
	package test;
	option go_package = "/pb";
	service TestService {}
	`)
	require.NoError(t, err)

	outputDir, err := ioutil.TempDir(".", "output")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	cmd := exec.Command("protoc", "--plugin", "./protoc-gen-relay", "--relay_out="+outputDir, protoFile.Name())
	err = cmd.Run()
	require.NoError(t, err)

	filename := strings.TrimPrefix(protoFile.Name(), ".")
	filename = strings.Replace(filename, ".proto", "_relay.pb.go", 1)
	output, err := ioutil.ReadFile(outputDir + "/pb" + filename)
	require.NoError(t, err)
	require.Contains(t, string(output), "func RegisterServiceRelay(config relay.Config) *relay.Relay {")
	require.Contains(t, string(output), "return relay.Must(relay.New(*config.Name, config))")
}
