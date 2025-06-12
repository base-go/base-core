package components

import "base/core/layout"

// InputComponent defines a reusable input component
func InputComponent() *layout.Component {
	return &layout.Component{
		Name: "Input",
		Template: `<div class="{{.Component.StyleClasses}}">
	{{if .Props.label}}<label for="{{.Props.id}}" class="block text-sm font-medium text-gray-700 mb-1">{{.Props.label}}</label>{{end}}
	<input 
		type="{{.Props.type}}" 
		id="{{.Props.id}}"
		name="{{.Props.name}}"
		placeholder="{{.Props.placeholder}}"
		value="{{.Props.value}}"
		class="{{.Props.class}} block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
		{{if .Props.required}}required{{end}}
		{{if .Props.disabled}}disabled{{end}}
		v-scope="{{.ComponentState}}">
	{{if .Props.error}}<p class="mt-1 text-sm text-red-600">{{.Props.error}}</p>{{end}}
	{{if .Props.help}}<p class="mt-1 text-sm text-gray-500">{{.Props.help}}</p>{{end}}
</div>`,
		Props: map[string]interface{}{
			"type":        "text",
			"id":          "",
			"name":        "",
			"label":       "",
			"placeholder": "",
			"value":       "",
			"class":       "",
			"required":    false,
			"disabled":    false,
			"error":       "",
			"help":        "",
		},
		StateKeys:    []string{"focused", "value"},
		StyleClasses: "",
	}
}

// TextareaComponent defines a reusable textarea component
func TextareaComponent() *layout.Component {
	return &layout.Component{
		Name: "Textarea",
		Template: `<div class="{{.Component.StyleClasses}}">
	{{if .Props.label}}<label for="{{.Props.id}}" class="block text-sm font-medium text-gray-700 mb-1">{{.Props.label}}</label>{{end}}
	<textarea 
		id="{{.Props.id}}"
		name="{{.Props.name}}"
		placeholder="{{.Props.placeholder}}"
		rows="{{.Props.rows}}"
		class="{{.Props.class}} block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
		{{if .Props.required}}required{{end}}
		{{if .Props.disabled}}disabled{{end}}
		v-scope="{{.ComponentState}}">{{.Props.value}}</textarea>
	{{if .Props.error}}<p class="mt-1 text-sm text-red-600">{{.Props.error}}</p>{{end}}
	{{if .Props.help}}<p class="mt-1 text-sm text-gray-500">{{.Props.help}}</p>{{end}}
</div>`,
		Props: map[string]interface{}{
			"id":          "",
			"name":        "",
			"label":       "",
			"placeholder": "",
			"value":       "",
			"class":       "",
			"rows":        3,
			"required":    false,
			"disabled":    false,
			"error":       "",
			"help":        "",
		},
		StateKeys:    []string{"focused", "value"},
		StyleClasses: "",
	}
}

// SelectComponent defines a reusable select component
func SelectComponent() *layout.Component {
	return &layout.Component{
		Name: "Select",
		Template: `<div class="{{.Component.StyleClasses}}">
	{{if .Props.label}}<label for="{{.Props.id}}" class="block text-sm font-medium text-gray-700 mb-1">{{.Props.label}}</label>{{end}}
	<select 
		id="{{.Props.id}}"
		name="{{.Props.name}}"
		class="{{.Props.class}} block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
		{{if .Props.required}}required{{end}}
		{{if .Props.disabled}}disabled{{end}}
		v-scope="{{.ComponentState}}">
		{{if .Props.placeholder}}<option value="">{{.Props.placeholder}}</option>{{end}}
		{{range .Props.options}}
		<option value="{{.value}}" {{if eq .value $.Props.value}}selected{{end}}>{{.label}}</option>
		{{end}}
	</select>
	{{if .Props.error}}<p class="mt-1 text-sm text-red-600">{{.Props.error}}</p>{{end}}
	{{if .Props.help}}<p class="mt-1 text-sm text-gray-500">{{.Props.help}}</p>{{end}}
</div>`,
		Props: map[string]interface{}{
			"id":          "",
			"name":        "",
			"label":       "",
			"placeholder": "",
			"value":       "",
			"class":       "",
			"options":     []map[string]interface{}{},
			"required":    false,
			"disabled":    false,
			"error":       "",
			"help":        "",
		},
		StateKeys:    []string{"focused", "value"},
		StyleClasses: "",
	}
}

// CheckboxComponent defines a reusable checkbox component
func CheckboxComponent() *layout.Component {
	return &layout.Component{
		Name: "Checkbox",
		Template: `<div class="{{.Component.StyleClasses}}">
	<div class="flex items-center">
		<input 
			type="checkbox" 
			id="{{.Props.id}}"
			name="{{.Props.name}}"
			value="{{.Props.value}}"
			class="{{.Props.class}} h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
			{{if .Props.checked}}checked{{end}}
			{{if .Props.required}}required{{end}}
			{{if .Props.disabled}}disabled{{end}}
			v-scope="{{.ComponentState}}">
		{{if .Props.label}}
		<label for="{{.Props.id}}" class="ml-2 block text-sm text-gray-900">{{.Props.label}}</label>
		{{end}}
	</div>
	{{if .Props.error}}<p class="mt-1 text-sm text-red-600">{{.Props.error}}</p>{{end}}
	{{if .Props.help}}<p class="mt-1 text-sm text-gray-500">{{.Props.help}}</p>{{end}}
</div>`,
		Props: map[string]interface{}{
			"id":       "",
			"name":     "",
			"label":    "",
			"value":    "",
			"class":    "",
			"checked":  false,
			"required": false,
			"disabled": false,
			"error":    "",
			"help":     "",
		},
		StateKeys:    []string{"checked"},
		StyleClasses: "",
	}
}
