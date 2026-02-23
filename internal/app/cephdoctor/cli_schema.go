package cephdoctor

type cli struct {
	Cluster clusterCmd `kong:"cmd,help='Cluster operations.'"`
}

type clusterCmd struct {
	Register   clusterRegisterCmd   `kong:"cmd,help='Register a cluster.'"`
	Unregister clusterUnregisterCmd `kong:"cmd,help='Unregister a cluster.'"`
	List       clusterListCmd       `kong:"cmd,help='List clusters.'"`
}

type clusterRegisterCmd struct {
	Name string `kong:"arg,help='Cluster name.'"`
	Host string `kong:"arg,help='Host in host[:port] format.'"`
	Key  string `kong:"arg,help='Access key.'"`
}

type clusterUnregisterCmd struct {
	Name string `kong:"arg,help='Cluster name.'"`
}

type clusterListCmd struct{}
