package types

// Types used internally to define the objects needed to print data, etc

type Allocated struct {
	Cpu    string
	Memory string
}

type Capacity struct {
	Memory float64
	Pods   int64
	Cpu    int64
}

type Cluster struct {
	Name         string
	URLKubeApi   string
	KubePort     int32
	TlsCheck     bool
	ExternalPort int32
	Nodes        []Node
}

type Node struct {
	Name       string
	InternalIP string
	ExternalIP string
	Capacity   Capacity
	Allocated  Allocated
}

type Package struct {
	Name             string
	Namespace        string
	Type             string
	GitRepository    string
	ArgocdRepository string
	Status           string
}

type Secret struct {
	IsCore    bool
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Username  string            `json:"username,omitempty"`
	Password  string            `json:"password,omitempty"`
	Token     string            `json:"token,omitempty"`
	Data      map[string]string `json:"data,omitempty"`
}
