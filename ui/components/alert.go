package components

import "base/core/layout"

// AlertComponent defines a reusable alert component
func AlertComponent() *layout.Component {
	return &layout.Component{
		Name: "Alert",
		Template: `<div class="{{.Props.class}} {{.Component.StyleClasses}}" v-scope="{{.ComponentState}}" v-show="!dismissed">
	<div class="flex">
		<div class="flex-shrink-0">
			{{if eq .Props.type "success"}}
			<svg class="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path>
			</svg>
			{{else if eq .Props.type "error"}}
			<svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"></path>
			</svg>
			{{else if eq .Props.type "warning"}}
			<svg class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path>
			</svg>
			{{else}}
			<svg class="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"></path>
			</svg>
			{{end}}
		</div>
		<div class="ml-3 flex-1">
			{{if .Props.title}}<h3 class="text-sm font-medium">{{.Props.title}}</h3>{{end}}
			<div class="text-sm">{{.Props.message}}</div>
		</div>
		{{if .Props.dismissible}}
		<div class="ml-auto pl-3">
			<div class="-mx-1.5 -my-1.5">
				<button @click="dismissed = true" class="inline-flex rounded-md p-1.5 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-offset-2">
					<span class="sr-only">Dismiss</span>
					<svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
					</svg>
				</button>
			</div>
		</div>
		{{end}}
	</div>
</div>`,
		Props: map[string]interface{}{
			"type":        "info", // success, error, warning, info
			"title":       "",
			"message":     "",
			"class":       "",
			"dismissible": true,
		},
		StateKeys:    []string{"dismissed"},
		StyleClasses: "rounded-md p-4",
	}
}

// SuccessAlertComponent creates a success alert variant
func SuccessAlertComponent() *layout.Component {
	alert := AlertComponent()
	alert.Name = "SuccessAlert"
	alert.Props["type"] = "success"
	alert.StyleClasses = "rounded-md p-4 bg-green-50 border border-green-200"
	return alert
}

// ErrorAlertComponent creates an error alert variant
func ErrorAlertComponent() *layout.Component {
	alert := AlertComponent()
	alert.Name = "ErrorAlert"
	alert.Props["type"] = "error"
	alert.StyleClasses = "rounded-md p-4 bg-red-50 border border-red-200"
	return alert
}

// WarningAlertComponent creates a warning alert variant
func WarningAlertComponent() *layout.Component {
	alert := AlertComponent()
	alert.Name = "WarningAlert"
	alert.Props["type"] = "warning"
	alert.StyleClasses = "rounded-md p-4 bg-yellow-50 border border-yellow-200"
	return alert
}
