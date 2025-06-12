package components

import "base/core/layout"

// ModalComponent defines a reusable modal component
func ModalComponent() *layout.Component {
	return &layout.Component{
		Name: "Modal",
		Template: `<div v-scope="{{.ComponentState}}" v-show="open" class="fixed inset-0 z-50 overflow-y-auto">
	<div class="flex items-center justify-center min-h-screen px-4">
		<div class="fixed inset-0 bg-black opacity-50" @click="open = false"></div>
		<div class="relative bg-white rounded-lg max-w-md w-full mx-auto shadow-xl">
			{{if .Props.title}}
			<div class="px-6 py-4 border-b">
				<h3 class="text-lg font-medium">{{.Props.title}}</h3>
				<button @click="open = false" class="absolute top-4 right-4 text-gray-400 hover:text-gray-600">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>
			{{end}}
			<div class="px-6 py-4">
				{{.Props.content | safe}}
			</div>
			{{if .Props.footer}}
			<div class="px-6 py-4 border-t bg-gray-50 rounded-b-lg">
				{{.Props.footer | safe}}
			</div>
			{{end}}
		</div>
	</div>
</div>`,
		Props: map[string]interface{}{
			"title":   "",
			"content": "",
			"footer":  "",
		},
		StateKeys:    []string{"open"},
		StyleClasses: "",
	}
}
