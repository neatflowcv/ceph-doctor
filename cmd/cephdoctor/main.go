package main

import (
	"os"

	"github.com/neatflowcv/ceph-doctor/internal/app/cephdoctor"
)

func main() {
	err := cephdoctor.Execute()
	if err == nil {
		return
	}

	if exit, ok := cephdoctor.ExitCode(err); ok {
		os.Exit(exit)
	}

	os.Exit(1)
}
