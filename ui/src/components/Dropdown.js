// Object-oriented Dropdown Component
export class Dropdown {
    constructor(element) {
        this.element = element;
        this.trigger = element.querySelector('[data-dropdown-trigger]');
        this.menu = element.querySelector('.dropdown-menu');
        this.chevron = element.querySelector('.dropdown-chevron');
        this.isOpen = false;
        
        if (this.trigger && this.menu) {
            this.init();
        }
    }
    
    init() {
        this.bindEvents();
    }
    
    bindEvents() {
        this.trigger.addEventListener('click', (e) => {
            e.stopPropagation();
            this.toggle();
        });
        
        // Prevent dropdown from closing when clicking inside
        this.menu.addEventListener('click', (e) => {
            e.stopPropagation();
        });
    }
    
    toggle() {
        if (this.isOpen) {
            this.close();
        } else {
            // Close all other dropdowns first
            DropdownManager.closeAll(this);
            this.open();
        }
    }
    
    open() {
        this.isOpen = true;
        this.menu.classList.remove('hidden');
        this.trigger.setAttribute('aria-expanded', 'true');
        if (this.chevron) {
            this.chevron.classList.add('rotate-180');
        }
    }
    
    close() {
        this.isOpen = false;
        this.menu.classList.add('hidden');
        this.trigger.setAttribute('aria-expanded', 'false');
        if (this.chevron) {
            this.chevron.classList.remove('rotate-180');
        }
    }
}

// Dropdown Manager to handle multiple dropdowns
export class DropdownManager {
    static dropdowns = [];
    
    static init() {
        // Clear existing dropdowns
        this.dropdowns = [];
        
        // Initialize all dropdowns
        document.querySelectorAll('.dropdown-wrapper').forEach(element => {
            const dropdown = new Dropdown(element);
            this.dropdowns.push(dropdown);
        });
        
        // Remove existing listeners to prevent duplicates
        document.removeEventListener('click', this.globalClickHandler);
        document.removeEventListener('keydown', this.globalKeyHandler);
        
        // Add global event listeners
        document.addEventListener('click', this.globalClickHandler);
        document.addEventListener('keydown', this.globalKeyHandler);
    }
    
    static globalClickHandler = () => {
        DropdownManager.closeAll();
    }
    
    static globalKeyHandler = (e) => {
        if (e.key === 'Escape') {
            DropdownManager.closeAll();
        }
    }
    
    static closeAll(except = null) {
        this.dropdowns.forEach(dropdown => {
            if (dropdown !== except && dropdown.isOpen) {
                dropdown.close();
            }
        });
    }
    
    static addDropdown(dropdown) {
        this.dropdowns.push(dropdown);
    }
}

// Global function for backwards compatibility
export function closeDropdown(element) {
    const dropdown = element.closest('.dropdown-wrapper');
    if (dropdown) {
        const dropdownInstance = DropdownManager.dropdowns.find(d => d.element === dropdown);
        if (dropdownInstance) {
            dropdownInstance.close();
        }
    }
}

// Export to global scope for template access
window.DropdownManager = DropdownManager;
window.Dropdown = Dropdown;
window.closeDropdown = closeDropdown;