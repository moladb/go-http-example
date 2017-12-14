package rest

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Resource
	HandlerFunc gin.HandlerFunc
}

type Service interface {
	GetAPIGroup() string
	ListHandlers() []Handler
}

type Resource struct {
	Name   string
	Path   string
	Method string
}

type APIGroup struct {
	Name      string
	Resources []Resource
}

type serviceRegistry struct {
	apiGroups map[string]APIGroup
}

func newServiceRegistry() *serviceRegistry {
	return &serviceRegistry{
		apiGroups: make(map[string]APIGroup),
	}
}

func (r *serviceRegistry) AddResource(group string, res Resource) {
	group = strings.Trim(group, "/")
	apiGroup, ok := r.apiGroups[group]
	if !ok {
		apiGroup = APIGroup{Name: group}
	}
	apiGroup.Resources = append(apiGroup.Resources, res)
	r.apiGroups[group] = apiGroup
}

func (r *serviceRegistry) ListAPIGroups() []string {
	names := make([]string, 0, len(r.apiGroups))
	for k := range r.apiGroups {
		names = append(names, k)
	}
	return names
}

func (r *serviceRegistry) ListResources(apiGroup string) (APIGroup, bool) {
	apiGroup = strings.Trim(apiGroup, "/")
	g, ok := r.apiGroups[apiGroup]
	return g, ok
}
