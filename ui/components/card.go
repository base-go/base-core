package components

import "base/core/layout"

// CardComponent defines a reusable card component
func CardComponent() *layout.Component {
	return &layout.Component{
		Name: "Card",
		Template: `<div class="bg-white rounded-lg shadow {{.Component.StyleClasses}}" v-scope="{{.ComponentState}}">
	{{if .Props.title}}
	<div class="px-6 py-4 border-b">
		<h3 class="text-lg font-medium text-gray-900">{{.Props.title}}</h3>
		{{if .Props.subtitle}}<p class="text-sm text-gray-600">{{.Props.subtitle}}</p>{{end}}
	</div>
	{{end}}
	<div class="px-6 py-4">
		{{if .Props.content}}{{.Props.content | safe}}{{end}}
		{{if .Props.items}}
		<ul class="space-y-2">
			{{range .Props.items}}
			<li class="text-sm text-gray-700">{{.}}</li>
			{{end}}
		</ul>
		{{end}}
	</div>
	{{if .Props.footer}}
	<div class="px-6 py-3 bg-gray-50 rounded-b-lg">
		{{.Props.footer | safe}}
	</div>
	{{end}}
</div>`,
		Props: map[string]interface{}{
			"title":    "",
			"subtitle": "",
			"content":  "",
			"footer":   "",
			"items":    []string{},
		},
		StateKeys:    []string{"loading"},
		StyleClasses: "max-w-sm",
	}
}
