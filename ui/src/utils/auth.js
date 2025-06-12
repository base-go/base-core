// Authentication utility functions

export function logout() {
    if (confirm('Are you sure you want to logout?')) {
        // Create a form and submit it as POST request
        const form = document.createElement('form');
        form.method = 'POST';
        form.action = '/auth/logout';
        
        // Add CSRF token if available
        const csrfToken = document.querySelector('meta[name="csrf-token"]');
        if (csrfToken) {
            const csrfInput = document.createElement('input');
            csrfInput.type = 'hidden';
            csrfInput.name = '_token';
            csrfInput.value = csrfToken.getAttribute('content');
            form.appendChild(csrfInput);
        }
        
        document.body.appendChild(form);
        form.submit();
    }
}

// Export to global scope for template access
window.logout = logout;