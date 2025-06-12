// Object-oriented Tooltip Component
export class Tooltip {
    constructor(element) {
        this.wrapper = element;
        this.trigger = element.querySelector('.tooltip-trigger');
        this.content = element.querySelector('.tooltip-content');
        this.isVisible = false;
        
        if (this.trigger && this.content) {
            this.init();
        }
    }
    
    init() {
        this.bindEvents();
    }
    
    bindEvents() {
        this.trigger.addEventListener('mouseenter', () => this.show());
        this.trigger.addEventListener('mouseleave', () => this.hide());
        this.trigger.addEventListener('focus', () => this.show());
        this.trigger.addEventListener('blur', () => this.hide());
    }
    
    show() {
        this.isVisible = true;
        this.content.classList.remove('hidden');
    }
    
    hide() {
        this.isVisible = false;
        this.content.classList.add('hidden');
    }
}

// Tooltip Manager
export class TooltipManager {
    static tooltips = [];
    
    static init() {
        // Clear existing tooltips
        this.tooltips = [];
        
        document.querySelectorAll('.tooltip-wrapper').forEach(element => {
            const tooltip = new Tooltip(element);
            this.tooltips.push(tooltip);
        });
    }
    
    static addTooltip(tooltip) {
        this.tooltips.push(tooltip);
    }
}

// Export to global scope for template access
window.TooltipManager = TooltipManager;
window.Tooltip = Tooltip;