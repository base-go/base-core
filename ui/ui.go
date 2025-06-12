package ui

import (
	"base/core/layout"
	"base/ui/components"
)

// RegisterUIComponents registers all UI components with the layout engine
func RegisterUIComponents(registry *layout.ComponentRegistry) {
	// Basic components
	registry.RegisterComponent(components.CardComponent())
	registry.RegisterComponent(components.ButtonComponent())
	registry.RegisterComponent(components.LinkComponent())
	registry.RegisterComponent(components.PrimaryButtonComponent())
	registry.RegisterComponent(components.SecondaryButtonComponent())
	registry.RegisterComponent(components.DangerButtonComponent())
	registry.RegisterComponent(components.ListComponent())
	registry.RegisterComponent(components.ModalComponent())

	// Alert components
	registry.RegisterComponent(components.AlertComponent())
	registry.RegisterComponent(components.SuccessAlertComponent())
	registry.RegisterComponent(components.ErrorAlertComponent())
	registry.RegisterComponent(components.WarningAlertComponent())

	// Form components
	registry.RegisterComponent(components.InputComponent())
	registry.RegisterComponent(components.TextareaComponent())
	registry.RegisterComponent(components.SelectComponent())
	registry.RegisterComponent(components.CheckboxComponent())

	// Layout components
	registry.RegisterComponent(components.ContainerComponent())
	registry.RegisterComponent(components.GridComponent())
	registry.RegisterComponent(components.FlexComponent())

	// Navigation components
	registry.RegisterComponent(components.NavbarComponent())
	registry.RegisterComponent(components.BreadcrumbComponent())
	registry.RegisterComponent(components.TabsComponent())

	// New components
	registry.RegisterComponent(components.DropdownComponent())
	registry.RegisterComponent(components.ToggleComponent())
	registry.RegisterComponent(components.AvatarComponent())
	registry.RegisterComponent(components.BadgeComponent())
	registry.RegisterComponent(components.TooltipComponent())
	registry.RegisterComponent(components.UserMenuDropdownComponent())
	registry.RegisterComponent(components.LanguageDropdownComponent())
}
