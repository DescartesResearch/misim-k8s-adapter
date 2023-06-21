package entity

// Information about a successfully binded pod
type BindingInformation struct {
	Pod  string
	Node string
}

// Information about a failed binding for a pod
type BindingFailureInformation struct {
	Pod     string
	Message string
}
