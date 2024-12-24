package entity

type Cluster struct {
	Name         string
	URLKubeApi   string
	KubePort     int32
	TlsCheck     bool
	ExternalPort int32
	Nodes        []Node
}
