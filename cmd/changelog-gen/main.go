package main

import (
	"yunion.io/x/log"

	"github.com/yunionio/git-tools/pkg/changelog-gen/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("execute error: %v", err)
	}
}
