package components

import "base/core/layout"

// ListComponent defines a reusable list component
func ListComponent() *layout.Component {
	return &layout.Component{
		Name: "List",
		Template: `<div class="{{.Component.StyleClasses}}" v-scope="{{.ComponentState}}">
	{{if .Props.title}}<h4 class="text-lg font-medium mb-4">{{.Props.title}}</h4>{{end}}
	{{if .Props.items}}
	<ul class="space-y-2">
		{{range $index, $item := .Props.items}}
		<li class="flex items-center justify-between p-3 bg-white rounded border hover:bg-gray-50">
			<div>
				{{if $item.title}}<div class="font-medium">{{$item.title}}</div>{{end}}
				{{if $item.description}}<div class="text-sm text-gray-600">{{$item.description}}</div>{{end}}
			</div>
			{{if $item.action}}
			<button @click="{{$item.action}}" class="text-blue-600 hover:text-blue-800">
				{{if $item.actionText}}{{$item.actionText}}{{else}}Action{{end}}
			</button>
			{{end}}
		</li>
		{{end}}
	</ul>
	{{else}}
	<div class="text-center py-8 text-gray-500">
		{{if .Props.emptyText}}{{.Props.emptyText}}{{else}}No items found{{end}}
	</div>
	{{end}}
</div>`,
		Props: map[string]interface{}{
			"title":     "",
			"items":     []map[string]interface{}{},
			"emptyText": "No items found",
		},
		StateKeys:    []string{"loading", "filter"},
		StyleClasses: "",
	}
}
