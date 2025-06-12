// BaseUI - Main Entry Point
import './styles/main.css';
import { App } from './components/App.js';
import { DropdownManager } from './components/Dropdown.js';
import { ToggleManager } from './components/Toggle.js';
import { TooltipManager } from './components/Tooltip.js';
import './utils/auth.js';

// Initialize BaseUI when DOM is ready
function initBaseUI() {
  // Initialize main app
  window.baseApp = new App();
  
  // Initialize component managers
  DropdownManager.init();
  ToggleManager.init();
  TooltipManager.init();
  
  console.log('BaseUI initialized');
}

// Auto-initialize when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initBaseUI);
} else {
  initBaseUI();
}

// Export for manual initialization if needed
export { initBaseUI, App, DropdownManager, ToggleManager, TooltipManager };