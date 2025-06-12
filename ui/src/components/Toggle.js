// Object-oriented Toggle Component
export class Toggle {
    constructor(element) {
        this.wrapper = element;
        this.button = element.querySelector('[data-toggle-target]');
        this.knob = element.querySelector('.toggle-knob');
        this.enabled = element.dataset.enabled === 'true';
        this.onChange = this.button?.dataset.onChange;
        
        if (this.button) {
            this.init();
        }
    }
    
    init() {
        this.bindEvents();
    }
    
    bindEvents() {
        this.button.addEventListener('click', () => {
            this.toggle();
        });
    }
    
    toggle() {
        this.enabled = !this.enabled;
        this.updateUI();
        this.executeOnChange();
    }
    
    updateUI() {
        this.button.setAttribute('aria-checked', this.enabled);
        this.wrapper.dataset.enabled = this.enabled;
        
        if (this.enabled) {
            this.button.classList.remove('bg-gray-200');
            this.button.classList.add('bg-indigo-600');
            this.knob.classList.remove('translate-x-0');
            this.knob.classList.add('translate-x-5');
        } else {
            this.button.classList.remove('bg-indigo-600');
            this.button.classList.add('bg-gray-200');
            this.knob.classList.remove('translate-x-5');
            this.knob.classList.add('translate-x-0');
        }
    }
    
    executeOnChange() {
        if (this.onChange) {
            try {
                eval(this.onChange);
            } catch (e) {
                console.error('Toggle onChange error:', e);
            }
        }
    }
}

// Toggle Manager
export class ToggleManager {
    static toggles = [];
    
    static init() {
        // Clear existing toggles
        this.toggles = [];
        
        document.querySelectorAll('.toggle-wrapper').forEach(element => {
            const toggle = new Toggle(element);
            this.toggles.push(toggle);
        });
    }
    
    static addToggle(toggle) {
        this.toggles.push(toggle);
    }
}

// Export to global scope for template access
window.ToggleManager = ToggleManager;
window.Toggle = Toggle;