package types

type Package struct {
	Name             string
	Namespace        string
	Type             string
	GitRepository    string
	ArgocdRepository string
	Status           string
}
