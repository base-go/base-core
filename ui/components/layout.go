package components

import "base/core/layout"

// ContainerComponent defines a reusable container component
func ContainerComponent() *layout.Component {
	return &layout.Component{
		Name: "Container",
		Template: `<div class="{{.Props.class}} {{.Component.StyleClasses}}">
	{{.Props.content | safe}}
</div>`,
		Props: map[string]interface{}{
			"content": "",
			"class":   "",
		},
		StateKeys:    []string{},
		StyleClasses: "max-w-7xl mx-auto px-4 sm:px-6 lg:px-8",
	}
}

// GridComponent defines a reusable grid component
func GridComponent() *layout.Component {
	return &layout.Component{
		Name: "Grid",
		Template: `<div class="{{.Props.class}} {{.Component.StyleClasses}}">
	{{.Props.content | safe}}
</div>`,
		Props: map[string]interface{}{
			"content": "",
			"class":   "",
			"cols":    "1",
		},
		StateKeys:    []string{},
		StyleClasses: "grid grid-cols-1 gap-4",
	}
}

// FlexComponent defines a reusable flex component
func FlexComponent() *layout.Component {
	return &layout.Component{
		Name: "Flex",
		Template: `<div class="{{.Props.class}} {{.Component.StyleClasses}}" v-scope="{{.ComponentState}}">
	{{.Props.content | safe}}
</div>`,
		Props: map[string]interface{}{
			"content":   "",
			"class":     "",
			"direction": "row",
			"justify":   "start",
			"align":     "start",
		},
		StateKeys:    []string{},
		StyleClasses: "flex",
	}
}
