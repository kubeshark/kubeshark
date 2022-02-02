package holder

import "github.com/up9inc/mizu/agent/pkg/resolver"

var k8sResolver *resolver.Resolver

func SetResolver(param *resolver.Resolver) {
	k8sResolver = param
}

func GetResolver() *resolver.Resolver {
	return k8sResolver
}
