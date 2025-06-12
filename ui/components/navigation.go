package components

import "base/core/layout"

// NavbarComponent defines a reusable navbar component
func NavbarComponent() *layout.Component {
	return &layout.Component{
		Name: "Navbar",
		Template: `<nav class="{{.Props.class}} {{.Component.StyleClasses}}" v-scope="{{.ComponentState}}">
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="flex justify-between h-16">
			<div class="flex items-center">
				{{if .Props.logo}}
				<a href="{{.Props.logoHref}}" class="flex-shrink-0">
					{{.Props.logo | safe}}
				</a>
				{{end}}
				{{if .Props.brand}}
				<a href="{{.Props.brandHref}}" class="ml-4 text-xl font-bold text-gray-900">{{.Props.brand}}</a>
				{{end}}
			</div>
			<div class="hidden md:flex items-center space-x-4">
				{{range .Props.links}}
				<a href="{{.href}}" class="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium">{{.text}}</a>
				{{end}}
			</div>
			<div class="md:hidden">
				<button @click="mobileOpen = !mobileOpen" class="text-gray-700 hover:text-gray-900 p-2">
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
					</svg>
				</button>
			</div>
		</div>
	</div>
	<!-- Mobile menu -->
	<div v-show="mobileOpen" class="md:hidden">
		<div class="px-2 pt-2 pb-3 space-y-1 sm:px-3">
			{{range .Props.links}}
			<a href="{{.href}}" class="text-gray-700 hover:text-gray-900 block px-3 py-2 rounded-md text-base font-medium">{{.text}}</a>
			{{end}}
		</div>
	</div>
</nav>`,
		Props: map[string]interface{}{
			"class":     "",
			"logo":      "",
			"logoHref":  "/",
			"brand":     "",
			"brandHref": "/",
			"links":     []map[string]interface{}{},
		},
		StateKeys:    []string{"mobileOpen"},
		StyleClasses: "bg-white shadow",
	}
}

// BreadcrumbComponent defines a reusable breadcrumb component
func BreadcrumbComponent() *layout.Component {
	return &layout.Component{
		Name: "Breadcrumb",
		Template: `<nav class="{{.Props.class}} {{.Component.StyleClasses}}" aria-label="Breadcrumb">
	<ol class="flex items-center space-x-4">
		{{range $index, $item := .Props.items}}
		<li>
			<div class="flex items-center">
				{{if ne $index 0}}
				<svg class="flex-shrink-0 h-5 w-5 text-gray-300" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
					<path d="M5.555 17.776l8-16 .894.448-8 16-.894-.448z" />
				</svg>
				{{end}}
				{{if $item.href}}
				<a href="{{$item.href}}" class="ml-4 text-sm font-medium text-gray-500 hover:text-gray-700">{{$item.text}}</a>
				{{else}}
				<span class="ml-4 text-sm font-medium text-gray-900">{{$item.text}}</span>
				{{end}}
			</div>
		</li>
		{{end}}
	</ol>
</nav>`,
		Props: map[string]interface{}{
			"class": "",
			"items": []map[string]interface{}{},
		},
		StateKeys:    []string{},
		StyleClasses: "",
	}
}

// TabsComponent defines a reusable tabs component
func TabsComponent() *layout.Component {
	return &layout.Component{
		Name: "Tabs",
		Template: `<div class="{{.Component.StyleClasses}}" v-scope="{{.ComponentState}}">
	<div class="border-b border-gray-200">
		<nav class="-mb-px flex space-x-8" aria-label="Tabs">
			{{range $index, $tab := .Props.tabs}}
			<button 
				@click="activeTab = {{$index}}"
				class="whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm"
				:class="activeTab === {{$index}} ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'">
				{{$tab.label}}
			</button>
			{{end}}
		</nav>
	</div>
	<div class="mt-4">
		{{range $index, $tab := .Props.tabs}}
		<div v-show="activeTab === {{$index}}">
			{{$tab.content | safe}}
		</div>
		{{end}}
	</div>
</div>`,
		Props: map[string]interface{}{
			"tabs": []map[string]interface{}{},
		},
		StateKeys:    []string{"activeTab"},
		StyleClasses: "",
	}
}
