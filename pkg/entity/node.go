package entity

type Node struct {
	Name       string
	InternalIP string
	ExternalIP string
	Capacity   Capacity
	Allocated  Allocated
}
