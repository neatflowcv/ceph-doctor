package cephdoctor

import (
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/neatflowcv/ceph-doctor/internal/domain"
)

func renderClusterTable(w io.Writer, clusters []*domain.Cluster) {
	tableWriter := table.NewWriter()
	tableWriter.SetOutputMirror(w)
	tableWriter.AppendHeader(table.Row{"Name", "Hosts"})

	for _, cluster := range clusters {
		tableWriter.AppendRow(table.Row{
			cluster.Name(),
			strings.Join(cluster.Hosts(), ","),
		})
	}

	tableWriter.Render()
}
