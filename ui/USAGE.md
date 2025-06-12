# UI Component Library - Usage Guide

This is a comprehensive guide for using the standalone UI component library that bridges Go templates with Petite-Vue.

## Quick Start

### 1. Integration Setup

In your Go application, import and register the UI components:

```go
import "base/ui"

// In your template registration function
func RegisterTemplates(engine *layout.Engine) error {
    // Register UI components
    if engine.GetComponentRegistry() != nil {
        ui.RegisterUIComponents(engine.GetComponentRegistry())
    }
    
    // ... rest of your template setup
    return nil
}
```

### 2. Template Helper Functions

The library provides helper functions for creating component props:

```go
// dict - creates a map for component properties
{{component "Card" (dict "title" "My Card" "content" "Hello World")}}

// list - creates an array for lists
{{component "List" (dict "items" (list "Item 1" "Item 2" "Item 3"))}}

// c - a more intuitive syntax for components (recommended)
{{c "Dropdown" buttonContent="Click Me" content="<a href='#'>Option 1</a>" list=.listGoObject}}
```

### 3. Improved Component Syntax

The `c` helper provides a more intuitive, attribute-style syntax for components that's easier to read and write:

```html
<!-- Basic usage with simple attributes -->
{{c "Button" text="Click me" color="blue"}}

<!-- Using with complex content -->
{{c "Dropdown" buttonContent="Click Me" content="<a href='#'>Option 1</a>" position="bottom-right"}}
```

### 4. XML-like Component Syntax

For an even cleaner, more HTML-like syntax, you can use the XML-like component syntax which is automatically preprocessed into standard Go template syntax:

```html
<!-- Using equals sign syntax -->
<{ Button 
    variant="primary" 
    text="Click me" 
    disabled=false 
}>

<!-- Using arrow operator syntax (alternative) -->
<{ Dropdown 
    buttonContent->"Click Me" 
    content->"<a href='#'>Option 1</a><a href='#'>Option 2</a>" 
    position->"bottom-right" 
}>
```

The arrow operator (`->`) can be used as an alternative to the equals sign (`=`) if you encounter any issues with the equals sign in certain contexts.

```html
<!-- Using with complex data types -->
<{ Dropdown 
    buttonContent="Select a user" 
    items=[
        {name: "John Doe", email: "john@example.com"},
        {name: "Jane Smith", email: "jane@example.com"}
    ]
    renderItem="{{.name}} ({{.email}})" 
}>
```

<!-- Passing Go template variables -->
{{c "Avatar" 
    src=.user.ProfileImage 
    initials=.user.Initials 
    size="10"
}}

<!-- Using with lists -->
{{c "List" items=.notifications}}
```

This syntax is recommended for all new component usage as it's more readable and maintainable than the dictionary-based approach.

## Component Reference

### Basic Components

#### Card Component
Flexible container with optional header, content, and footer.

```html
<!-- Simple card -->
{{component "Card" (dict "title" "Simple Card" "content" "This is a basic card.")}}

<!-- Card with subtitle and footer -->
{{component "Card" (dict 
    "title" "Advanced Card" 
    "subtitle" "With more features"
    "content" "Card content goes here."
    "footer" "<button class=\"bg-blue-600 text-white px-4 py-2 rounded\">Action</button>"
)}}

<!-- Card with list items -->
{{component "Card" (dict 
    "title" "Feature List"
    "items" (list "Feature 1" "Feature 2" "Feature 3")
)}}
```

**Props:**
- `title` (string): Card title
- `subtitle` (string): Card subtitle
- `content` (string): Main content (HTML safe)
- `footer` (string): Footer content (HTML safe)
- `items` ([]string): List items to display

#### Dropdown Component
Dropdown menu with click-outside behavior.

```html
<!-- Basic dropdown -->
{{component "Dropdown" (dict 
    "buttonContent" "Click me"
    "content" "<a href=\"#\" class=\"block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100\">Option 1</a><a href=\"#\" class=\"block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100\">Option 2</a>"
)}}

<!-- Using defineComponent for cleaner HTML -->
{{$buttonContent := defineComponent `
    <span class="flex items-center">
        <svg class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
        Select an option
    </span>
` "{}"}}

{{$menuContent := defineComponent `
    <a href="#" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Profile</a>
    <a href="#" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Settings</a>
    <div class="border-t border-gray-100"></div>
    <a href="#" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Sign out</a>
` "{}"}}

{{component "Dropdown" (dict 
    "buttonContent" $buttonContent
    "content" $menuContent
    "buttonClass" "inline-flex justify-center w-full rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white"
)}}
```

**Props:**
- `buttonContent` (string): Content for the dropdown button (HTML safe)
- `content` (string): Content for the dropdown menu (HTML safe)
- `buttonClass` (string): Custom CSS classes for the button
- `menuClass` (string): Custom CSS classes for the dropdown menu
- `contentClass` (string): Custom CSS classes for the content container
- `showChevron` (bool): Whether to show the dropdown chevron icon
- `defaultOpen` (bool): Whether the dropdown should be open by default

#### Button Components
Clickable buttons with various styles and behaviors.

```html
<!-- Basic button -->
{{component "Button" (dict "text" "Click Me")}}

<!-- Button with click handler -->
{{component "Button" (dict 
    "text" "Submit" 
    "onclick" "handleSubmit()"
    "type" "submit"
)}}

<!-- Styled button variants -->
{{component "PrimaryButton" (dict "text" "Primary Action")}}
{{component "SecondaryButton" (dict "text" "Secondary")}}
{{component "DangerButton" (dict "text" "Delete")}}

<!-- Disabled button -->
{{component "Button" (dict 
    "text" "Disabled" 
    "disabled" true
    "class" "opacity-50 cursor-not-allowed"
)}}
```

**Props:**
- `text` (string): Button text
- `type` (string): Button type (button, submit, reset)
- `onclick` (string): JavaScript click handler
- `class` (string): Additional CSS classes
- `icon` (string): Icon class (optional)
- `disabled` (bool): Whether button is disabled

#### List Component
Dynamic lists with optional actions and empty states.

```html
<!-- Simple list -->
{{component "List" (dict 
    "title" "Todo List"
    "items" (list 
        (dict "title" "Task 1" "description" "First task to complete")
        (dict "title" "Task 2" "description" "Second task")
    )
)}}

<!-- List with actions -->
{{component "List" (dict 
    "title" "Action Items"
    "items" (list 
        (dict 
            "title" "Review Code" 
            "description" "Check pull request #123"
            "action" "window.open('/pr/123')"
            "actionText" "View"
        )
    )
)}}

<!-- Empty list with custom message -->
{{component "List" (dict 
    "title" "Empty List"
    "items" (list)
    "emptyText" "No items available"
)}}
```

**Props:**
- `title` (string): List title
- `items` ([]object): List items with title, description, action, actionText
- `emptyText` (string): Text to show when list is empty

#### Modal Component
Overlay modal dialogs with reactive open/close state.

```html
<!-- Basic modal -->
{{component "Modal" (dict 
    "title" "Confirmation"
    "content" "<p>Are you sure you want to continue?</p>"
    "footer" "<button onclick=\"closeModal()\" class=\"bg-gray-600 text-white px-4 py-2 rounded mr-2\">Cancel</button><button class=\"bg-blue-600 text-white px-4 py-2 rounded\">Confirm</button>"
)}}

<!-- Modal without header -->
{{component "Modal" (dict 
    "content" "<div class=\"text-center\"><h3 class=\"text-lg font-bold mb-4\">Custom Modal</h3><p>Content without predefined header.</p></div>"
)}}
```

**Props:**
- `title` (string): Modal title (optional)
- `content` (string): Modal content (HTML safe)
- `footer` (string): Footer content (HTML safe, optional)

**JavaScript Integration:**
```javascript
// Open modal
document.querySelector('[data-modal] [v-scope]').__vue__.open = true;

// Close modal
document.querySelector('[data-modal] [v-scope]').__vue__.open = false;
```

### Alert Components
Dismissible alert messages with different severity levels.

```html
<!-- Basic alert -->
{{component "Alert" (dict 
    "type" "info"
    "title" "Information"
    "message" "This is an informational message."
)}}

<!-- Pre-styled alert variants -->
{{component "SuccessAlert" (dict 
    "title" "Success!"
    "message" "Operation completed successfully."
)}}

{{component "ErrorAlert" (dict 
    "title" "Error"
    "message" "Something went wrong."
)}}

{{component "WarningAlert" (dict 
    "message" "This is a warning message."
    "dismissible" false
)}}
```

**Props:**
- `type` (string): Alert type (success, error, warning, info)
- `title` (string): Alert title (optional)
- `message` (string): Alert message
- `dismissible` (bool): Whether alert can be dismissed
- `class` (string): Additional CSS classes

### Interactive Components

#### Toggle Component
Toggle switch for boolean options.

```html
<!-- Basic toggle -->
{{component "Toggle" (dict 
    "label" "Enable notifications"
)}}

<!-- Toggle with default state and change handler -->
{{component "Toggle" (dict 
    "label" "Dark mode"
    "defaultEnabled" true
    "onChange" "toggleDarkMode()"
)}}

<!-- Custom styled toggle -->
{{component "Toggle" (dict 
    "label" "Custom toggle"
    "enabledClass" "bg-green-600"
    "disabledClass" "bg-red-200"
)}}
```

**Props:**
- `label` (string): Accessible label for the toggle
- `defaultEnabled` (bool): Initial state of the toggle
- `onChange` (string): JavaScript to execute when toggled
- `class` (string): Custom CSS classes for the toggle container
- `enabledClass` (string): Custom CSS classes when enabled
- `disabledClass` (string): Custom CSS classes when disabled
- `knobClass` (string): Custom CSS classes for the toggle knob

#### Avatar Component
User profile pictures or initials.

```html
<!-- Avatar with image -->
{{component "Avatar" (dict 
    "src" "/images/user.jpg"
    "alt" "User profile"
    "size" "10"
)}}

<!-- Avatar with initials -->
{{component "Avatar" (dict 
    "initials" "JD"
    "color" "blue"
)}}

<!-- Custom styled avatar -->
{{component "Avatar" (dict 
    "initials" "AB"
    "color" "purple"
    "rounded" "md"
    "size" "12"
)}}
```

**Props:**
- `src` (string): Image URL (optional, falls back to initials)
- `alt` (string): Image alt text
- `initials` (string): Text to display when no image is available
- `size` (string): Size in Tailwind units (8, 10, 12, etc.)
- `rounded` (string): Border radius (full, md, lg, etc.)
- `color` (string): Color scheme for the placeholder
- `class` (string): Custom CSS classes for the container
- `imgClass` (string): Custom CSS classes for the image
- `placeholderClass` (string): Custom CSS classes for the placeholder
- `textClass` (string): Custom CSS classes for the initials text

#### Badge Component
Labels and notification indicators.

```html
<!-- Simple badge -->
{{component "Badge" (dict 
    "text" "New"
    "color" "blue"
)}}

<!-- Badge with dot indicator -->
{{component "Badge" (dict 
    "text" "3 notifications"
    "color" "red"
    "dot" true
)}}

<!-- Custom styled badge -->
{{component "Badge" (dict 
    "text" "Premium"
    "color" "yellow"
    "rounded" "md"
    "class" "px-3 py-1 text-sm"
)}}
```

**Props:**
- `text` (string): Badge text
- `color` (string): Color scheme (gray, red, blue, green, etc.)
- `rounded` (string): Border radius (full, md, etc.)
- `dot` (bool): Whether to show a dot indicator
- `class` (string): Custom CSS classes
- `dotClass` (string): Custom CSS classes for the dot

#### Tooltip Component
Informational tooltips on hover.

```html
<!-- Basic tooltip -->
{{component "Tooltip" (dict 
    "content" "This is a helpful tooltip"
    "children" "<button class=\"px-4 py-2 bg-blue-600 text-white rounded\">Hover me</button>"
)}}

<!-- Positioned tooltip -->
{{component "Tooltip" (dict 
    "content" "Appears on the right"
    "children" "<span class=\"underline cursor-help\">Help text</span>"
    "position" "right"
)}}
```

**Props:**
- `content` (string): Tooltip content
- `children` (string): Element that triggers the tooltip
- `position` (string): Tooltip position (top, bottom, left, right)
- `class` (string): Custom CSS classes for the tooltip
- `arrowClass` (string): Custom CSS classes for the tooltip arrow

### Layout Components

#### Container Component
Responsive container wrapper with consistent padding.

```html
{{component "Container" (dict 
    "content" "<h1>Page Content</h1><p>This content is wrapped in a responsive container.</p>"
)}}

<!-- With custom classes -->
{{component "Container" (dict 
    "content" "<div>Custom container</div>"
    "class" "bg-gray-100 py-8"
)}}
```

#### Grid and Flex Components
Layout helpers for modern CSS layouts.

```html
<!-- Grid layout -->
{{component "Grid" (dict 
    "content" "<div>Item 1</div><div>Item 2</div><div>Item 3</div>"
    "class" "grid-cols-3 gap-4"
)}}

<!-- Flex layout -->
{{component "Flex" (dict 
    "content" "<div>Left</div><div>Center</div><div>Right</div>"
    "class" "justify-between items-center"
)}}
```

## Advanced Usage

### Custom Component State

Components support reactive state using Petite-Vue. State is automatically managed:

```html
<!-- Component with loading state -->
{{component "Button" (dict 
    "text" "Save" 
    "onclick" "this.loading = true; saveData().finally(() => this.loading = false)"
)}}

<!-- Component with conditional content -->
{{component "Card" (dict 
    "title" "Dynamic Card"
    "content" "Content that can change based on state"
)}}
```

### Inline Components

For quick custom components without creating new definitions:

```html
{{defineComponent `
    <div class="bg-blue-100 border-l-4 border-blue-500 p-4">
        <div class="flex items-center">
            <div class="flex-shrink-0">
                <svg class="h-5 w-5 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"></path>
                </svg>
            </div>
            <div class="ml-3">
                <p class="text-sm text-blue-700">This is a custom inline component!</p>
            </div>
        </div>
    </div>
` "{}"}}
```

### Component Composition

Components can be nested and combined:

```html
{{component "Card" (dict 
    "title" "User Profile"
    "content" (printf "%s%s" 
        (component "Alert" (dict "type" "success" "message" "Profile updated"))
        "<div class=\"mt-4\"><p>User information goes here</p></div>"
    )
    "footer" (component "Button" (dict "text" "Edit Profile" "onclick" "editProfile()"))
)}}
```

## Best Practices

### 1. Component Props
- Always provide default values for optional props
- Use semantic prop names that describe the purpose
- Keep prop interfaces consistent across similar components

### 2. Styling
- Use Tailwind CSS classes for consistency
- Leverage the `StyleClasses` property for default styling
- Allow customization through the `class` prop

### 3. JavaScript Integration
- Use Petite-Vue reactive features for dynamic behavior
- Keep JavaScript simple and focused on the component's purpose
- Prefer declarative over imperative approaches

### 4. Performance
- Use `v-show` instead of `v-if` for frequently toggled elements
- Minimize the number of reactive state variables
- Consider component reusability when designing APIs

## Examples in Practice

### Form with Validation
```html
<form v-scope="{ errors: {}, loading: false }">
    {{component "Input" (dict 
        "label" "Email Address"
        "type" "email"
        "name" "email"
        "required" true
        "error" "{{ if .errors.email }}{{.errors.email}}{{ end }}"
    )}}
    
    {{component "Textarea" (dict 
        "label" "Message"
        "name" "message"
        "rows" 4
        "required" true
    )}}
    
    {{component "Button" (dict 
        "text" "Send Message"
        "type" "submit"
        "onclick" "submitForm()"
    )}}
</form>
```

### Dashboard Card Grid
```html
{{component "Container" (dict 
    "content" (printf `
        <h1 class="text-2xl font-bold mb-6">Dashboard</h1>
        %s
    ` (component "Grid" (dict 
        "class" "grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
        "content" (printf "%s%s%s"
            (component "Card" (dict "title" "Users" "content" "1,234 total users"))
            (component "Card" (dict "title" "Orders" "content" "567 orders today"))
            (component "Card" (dict "title" "Revenue" "content" "$12,345 this month"))
        )
    )))
)}}
```

This component library provides a powerful foundation for building consistent, interactive web interfaces while maintaining the simplicity of Go templates and the reactivity of modern JavaScript frameworks.