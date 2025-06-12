// Global App State and Functionality
export class App {
    constructor() {
        this.state = {
            sidebarOpen: false
        };
        this.init();
    }

    init() {
        this.bindEvents();
    }

    bindEvents() {
        // Mobile sidebar toggle
        const sidebarToggle = document.querySelector('[data-sidebar-toggle]');
        if (sidebarToggle) {
            sidebarToggle.addEventListener('click', () => {
                this.toggleSidebar();
            });
        }

        // Close sidebar when clicking overlay
        const sidebarOverlay = document.querySelector('[data-sidebar-overlay]');
        if (sidebarOverlay) {
            sidebarOverlay.addEventListener('click', () => {
                this.closeSidebar();
            });
        }

        // Close sidebar on escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.state.sidebarOpen) {
                this.closeSidebar();
            }
        });
    }

    toggleSidebar() {
        this.state.sidebarOpen = !this.state.sidebarOpen;
        this.updateSidebarUI();
    }

    closeSidebar() {
        this.state.sidebarOpen = false;
        this.updateSidebarUI();
    }

    updateSidebarUI() {
        const sidebar = document.querySelector('[data-sidebar]');
        const overlay = document.querySelector('[data-sidebar-overlay]');

        if (!sidebar || !overlay) return;

        if (this.state.sidebarOpen) {
            sidebar.classList.remove('hidden');
            overlay.classList.remove('hidden');
        } else {
            sidebar.classList.add('hidden');
            overlay.classList.add('hidden');
        }
    }
}

// Export to global scope for template access
window.App = App;