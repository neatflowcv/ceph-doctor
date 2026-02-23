package main

import (
	"fmt"
	"os"

	"github.com/neatflowcv/ceph-doctor/internal/app/cephdoctor"
)

func main() {
	err := cephdoctor.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)

		os.Exit(1)
	}
}
