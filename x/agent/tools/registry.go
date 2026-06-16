package tools

import (
	"fmt"
	"os"
	"sort"

	"github.com/MD-Mushfiqur123/lychee/api"
)

type Tool interface {
	Name() string
	Description() string
	Schema() api.ToolFunction
	Execute(args map[string]any) (string, error)
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

func (r *Registry) Unregister(name string) {
	delete(r.tools, name)
}

func (r *Registry) Has(name string) bool {
	_, ok := r.tools[name]
	return ok
}

func (r *Registry) RegisterBash() {
	r.Register(&BashTool{})
}

func (r *Registry) RegisterWebSearch() {
	r.Register(&WebSearchTool{})
}

func (r *Registry) RegisterWebFetch() {
	r.Register(&WebFetchTool{})
}

func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *Registry) Tools() api.Tools {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)

	var tools api.Tools
	for _, name := range names {
		tool := r.tools[name]
		tools = append(tools, api.Tool{
			Type:     "function",
			Function: tool.Schema(),
		})
	}
	return tools
}

func (r *Registry) Execute(call api.ToolCall) (string, error) {
	tool, ok := r.tools[call.Function.Name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", call.Function.Name)
	}
	return tool.Execute(call.Function.Arguments.ToMap())
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func DefaultRegistry() *Registry {
	r := NewRegistry()
	if os.Getenv("LYCHEE_AGENT_DISABLE_BASH") == "" {
		r.Register(&BashTool{})
	}
	return r
}
