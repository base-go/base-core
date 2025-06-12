package components

import "base/core/layout"

// DropdownComponent defines a reusable dropdown component with array-based items (NuxtUI style)
func DropdownComponent() *layout.Component {
	return &layout.Component{
		Name: "Dropdown",
		Template: `<div class="relative inline-block text-left dropdown-wrapper" data-dropdown-id="{{.Props.dropdownId}}">
	<!-- Trigger button -->
	<button 
		type="button" 
		class="dropdown-trigger {{.Props.buttonClass}} {{if not .Props.buttonClass}}inline-flex justify-center w-full rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500{{end}}"
		aria-expanded="false"
		aria-haspopup="true"
		data-dropdown-trigger>
		{{if .Props.buttonIcon}}<i class="{{.Props.buttonIcon}} mr-2"></i>{{end}}
		{{.Props.buttonText}}
		{{if .Props.showChevron}}
		<!-- Chevron icon -->
		<svg class="dropdown-chevron -mr-1 ml-2 h-5 w-5 transition-transform duration-200" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
			<path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
		</svg>
		{{end}}
	</button>

	<!-- Dropdown menu -->
	<div 
		class="dropdown-menu hidden {{.Props.menuClass}} {{if not .Props.menuClass}}origin-top-right absolute right-0 mt-2 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none z-50{{end}}"
		role="menu" 
		aria-orientation="vertical" 
		tabindex="-1">
		<div class="{{.Props.contentClass}} {{if not .Props.contentClass}}py-1{{end}}" role="none">
			{{range $index, $item := .Props.items}}
				{{if eq $item.type "divider"}}
					<div class="border-t border-gray-100 my-1"></div>
				{{else if eq $item.type "header"}}
					<div class="px-4 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wide">{{$item.label}}</div>
				{{else}}
					{{if $item.href}}
						<a href="{{$item.href}}" 
						   class="{{$item.class}} {{if not $item.class}}group flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 hover:text-gray-900{{end}}"
						   {{if $item.target}}target="{{$item.target}}"{{end}}
						   role="menuitem">
							{{if $item.icon}}<i class="{{$item.icon}} mr-3 h-4 w-4"></i>{{end}}
							{{$item.label}}
							{{if $item.badge}}<span class="ml-auto inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">{{$item.badge}}</span>{{end}}
						</a>
					{{else}}
						<button type="button"
						        class="{{$item.class}} {{if not $item.class}}group flex w-full items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 hover:text-gray-900{{end}}"
						        {{if $item.onclick}}onclick="{{$item.onclick}}{{if not $item.keepOpen}}; closeDropdown(this);{{end}}"{{end}}
						        role="menuitem">
							{{if $item.icon}}<i class="{{$item.icon}} mr-3 h-4 w-4"></i>{{end}}
							{{$item.label}}
							{{if $item.badge}}<span class="ml-auto inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">{{$item.badge}}</span>{{end}}
						</button>
					{{end}}
				{{end}}
			{{end}}
		</div>
	</div>
</div>

`,
		Props: map[string]interface{}{
			"buttonText":    "Dropdown",
			"buttonIcon":    "",
			"buttonClass":   "",
			"menuClass":     "",
			"contentClass":  "",
			"showChevron":   true,
			"defaultOpen":   false,
			"items":         []map[string]interface{}{},
			"dropdownId":    "",
		},
		StateKeys: []string{"open"},
	}
}

// LanguageDropdownComponent creates a language selector dropdown
func LanguageDropdownComponent() *layout.Component {
	dropdown := DropdownComponent()
	dropdown.Name = "LanguageDropdown"
	dropdown.Props["buttonText"] = "Language"
	dropdown.Props["buttonIcon"] = "fas fa-globe"
	dropdown.Props["items"] = []map[string]interface{}{
		{"label": "English", "href": "/switch-language/en", "icon": "fi fi-us"},
		{"label": "Español", "href": "/switch-language/es", "icon": "fi fi-es"},
		{"label": "Shqip", "href": "/switch-language/sq", "icon": "fi fi-al"},
	}
	return dropdown
}

// UserMenuDropdownComponent creates a user menu dropdown
func UserMenuDropdownComponent() *layout.Component {
	dropdown := DropdownComponent()
	dropdown.Name = "UserMenuDropdown"
	dropdown.Props["buttonText"] = "Account"
	dropdown.Props["buttonIcon"] = "fas fa-user-circle"
	dropdown.Props["items"] = []map[string]interface{}{
		{"type": "header", "label": "Account"},
		{"label": "Profile", "href": "/profile", "icon": "fas fa-user"},
		{"label": "Settings", "href": "/settings", "icon": "fas fa-cog"},
		{"type": "divider"},
		{"label": "Logout", "onclick": "logout()", "icon": "fas fa-sign-out-alt", "class": "text-red-600 hover:bg-red-50"},
	}
	return dropdown
}

// ActionDropdownComponent creates an action menu dropdown
func ActionDropdownComponent() *layout.Component {
	dropdown := DropdownComponent()
	dropdown.Name = "ActionDropdown"
	dropdown.Props["buttonText"] = "Actions"
	dropdown.Props["buttonIcon"] = "fas fa-ellipsis-v"
	dropdown.Props["items"] = []map[string]interface{}{
		{"label": "Edit", "onclick": "edit()", "icon": "fas fa-edit"},
		{"label": "Duplicate", "onclick": "duplicate()", "icon": "fas fa-copy"},
		{"type": "divider"},
		{"label": "Delete", "onclick": "confirmDelete()", "icon": "fas fa-trash", "class": "text-red-600 hover:bg-red-50"},
	}
	return dropdown
}

// ToggleComponent defines a simple toggle switch component
func ToggleComponent() *layout.Component {
	return &layout.Component{
		Name: "Toggle",
		Template: `<div class="toggle-wrapper" data-enabled="{{.Props.defaultEnabled}}">
	<button 
		type="button" 
		class="toggle-button {{.Props.class}} {{if not .Props.class}}relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500{{end}} {{if .Props.defaultEnabled}}{{.Props.enabledClass}} {{if not .Props.enabledClass}}bg-indigo-600{{end}}{{else}}{{.Props.disabledClass}} {{if not .Props.disabledClass}}bg-gray-200{{end}}{{end}}"
		role="switch"
		aria-checked="{{.Props.defaultEnabled}}"
		data-toggle-target
		data-on-change="{{.Props.onChange}}">
		<span class="sr-only">{{.Props.label}}</span>
		<span 
			aria-hidden="true" 
			class="toggle-knob {{.Props.knobClass}} {{if not .Props.knobClass}}pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200{{end}} {{if .Props.defaultEnabled}}translate-x-5{{else}}translate-x-0{{end}}">
		</span>
	</button>
</div>

`,
		Props: map[string]interface{}{
			"label":          "Toggle",
			"defaultEnabled": false,
			"onChange":       "",
			"class":          "",
			"enabledClass":   "",
			"disabledClass":  "",
			"knobClass":      "",
		},
		StateKeys: []string{"enabled"},
	}
}

// AvatarComponent defines a user avatar component
func AvatarComponent() *layout.Component {
	return &layout.Component{
		Name: "Avatar",
		Template: `<div class="{{.Props.class}} {{if not .Props.class}}inline-flex items-center justify-center{{end}}">
	{{if .Props.src}}
		<img 
			src="{{.Props.src}}" 
			alt="{{.Props.alt}}"
			class="{{.Props.imgClass}} {{if not .Props.imgClass}}h-{{.Props.size}} w-{{.Props.size}} rounded-{{.Props.rounded}}{{end}}"
		/>
	{{else}}
		<div class="{{.Props.placeholderClass}} {{if not .Props.placeholderClass}}h-{{.Props.size}} w-{{.Props.size}} rounded-{{.Props.rounded}} bg-{{.Props.color}}-100 flex items-center justify-center{{end}}">
			<span class="{{.Props.textClass}} {{if not .Props.textClass}}text-{{.Props.color}}-700 font-medium text-sm{{end}}">
				{{.Props.initials}}
			</span>
		</div>
	{{end}}
</div>`,
		Props: map[string]interface{}{
			"src":             "",
			"alt":             "User avatar",
			"initials":        "U",
			"size":            "8",
			"rounded":         "full",
			"color":           "indigo",
			"class":           "",
			"imgClass":        "",
			"placeholderClass": "",
			"textClass":       "",
		},
	}
}

// BadgeComponent defines a badge component for labels and counts
func BadgeComponent() *layout.Component {
	return &layout.Component{
		Name: "Badge",
		Template: `<span class="{{.Props.class}} {{if not .Props.class}}inline-flex items-center px-2.5 py-0.5 rounded-{{.Props.rounded}} text-xs font-medium bg-{{.Props.color}}-100 text-{{.Props.color}}-800{{end}}">
	{{if .Props.dot}}
		<span class="{{.Props.dotClass}} {{if not .Props.dotClass}}flex-shrink-0 h-2 w-2 mr-1.5 rounded-full bg-{{.Props.color}}-400{{end}}"></span>
	{{end}}
	{{.Props.text}}
</span>`,
		Props: map[string]interface{}{
			"text":     "Badge",
			"color":    "gray",
			"rounded":  "full",
			"dot":      false,
			"class":    "",
			"dotClass": "",
		},
	}
}

// TooltipComponent defines a tooltip component
func TooltipComponent() *layout.Component {
	return &layout.Component{
		Name: "Tooltip",
		Template: `<div class="relative inline-block tooltip-wrapper">
	<div class="tooltip-trigger">
		{{.Props.children}}
	</div>
	<div 
		class="tooltip-content hidden {{.Props.class}} {{if not .Props.class}}absolute z-10 px-3 py-2 text-sm font-medium text-white bg-gray-900 rounded-lg shadow-sm {{.Props.position}}-tooltip{{end}}"
		role="tooltip">
		{{.Props.content}}
		<div class="{{.Props.arrowClass}} {{if not .Props.arrowClass}}tooltip-arrow{{end}}"></div>
	</div>
</div>


<style>
.top-tooltip { bottom: calc(100% + 5px); left: 50%; transform: translateX(-50%); }
.bottom-tooltip { top: calc(100% + 5px); left: 50%; transform: translateX(-50%); }
.left-tooltip { right: calc(100% + 5px); top: 50%; transform: translateY(-50%); }
.right-tooltip { left: calc(100% + 5px); top: 50%; transform: translateY(-50%); }
.tooltip-arrow {
	position: absolute;
	width: 8px;
	height: 8px;
	background: inherit;
}
.top-tooltip .tooltip-arrow { bottom: -4px; left: 50%; transform: translateX(-50%) rotate(45deg); }
.bottom-tooltip .tooltip-arrow { top: -4px; left: 50%; transform: translateX(-50%) rotate(45deg); }
.left-tooltip .tooltip-arrow { right: -4px; top: 50%; transform: translateY(-50%) rotate(45deg); }
.right-tooltip .tooltip-arrow { left: -4px; top: 50%; transform: translateY(-50%) rotate(45deg); }
</style>`,
		Props: map[string]interface{}{
			"content":    "Tooltip content",
			"children":   "Hover me",
			"position":   "top", // top, bottom, left, right
			"class":      "",
			"arrowClass": "",
		},
		StateKeys: []string{"show"},
	}
}
