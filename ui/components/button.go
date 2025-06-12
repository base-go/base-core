package components

import "base/core/layout"

// ButtonComponent defines a reusable button component
func ButtonComponent() *layout.Component {
	return &layout.Component{
		Name: "Button",
		Template: `<button 
	type="{{.Props.type}}" 
	class="{{.Props.class}} {{.Component.StyleClasses}}"
	{{if .Props.onclick}}@click="{{.Props.onclick}}"{{end}}
	{{if .Props.disabled}}disabled{{end}}
	v-scope="{{.ComponentState}}">
	{{if .Props.icon}}<i class="{{.Props.icon}} mr-2"></i>{{end}}
	{{.Props.text}}
</button>`,
		Props: map[string]interface{}{
			"text":     "Button",
			"type":     "button",
			"class":    "",
			"onclick":  "",
			"icon":     "",
			"disabled": false,
		},
		StateKeys:    []string{"loading"},
		StyleClasses: "px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500",
	}
}

// PrimaryButtonComponent creates a primary styled button
func PrimaryButtonComponent() *layout.Component {
	btn := ButtonComponent()
	btn.Name = "PrimaryButton"
	btn.StyleClasses = "px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
	return btn
}

// SecondaryButtonComponent creates a secondary styled button
func SecondaryButtonComponent() *layout.Component {
	btn := ButtonComponent()
	btn.Name = "SecondaryButton"
	btn.StyleClasses = "px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500"
	return btn
}

// DangerButtonComponent creates a danger styled button
func DangerButtonComponent() *layout.Component {
	btn := ButtonComponent()
	btn.Name = "DangerButton"
	btn.StyleClasses = "px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500"
	return btn
}

func LinkComponent() *layout.Component {
	return &layout.Component{
		Name: "LinkButton",
		Template: `<a 
			class="{{.Props.class}} {{.Component.StyleClasses}}"
			v-scope="{{.ComponentState}}"
			>{{.Props.text}}</a>`,
		Props: map[string]interface{}{
			"text":  "Link",
			"class": "",
		},
		StyleClasses: "px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500",
	}
}
