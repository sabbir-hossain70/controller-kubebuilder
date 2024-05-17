package v1alpha1

import "strings"

func (b *Bookserver) DeploymentName() string {
	return strings.Join([]string{b.Name, "dep"}, "-")
}

func (b *Bookserver) ServiceName() string {
	return strings.Join([]string{b.Name, "svc"}, "-")
}
