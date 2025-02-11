// Load shared header into any page that includes this script
async function loadSharedHeader() {
    try {
        const response = await fetch('/static/shared-header.html');
        const html = await response.text();
        
        // Create a container for the header
        const headerContainer = document.createElement('div');
        headerContainer.innerHTML = html;
        
        // Insert at the start of the body
        document.body.insertBefore(headerContainer, document.body.firstChild);
    } catch (error) {
        console.error('Error loading shared header:', error);
    }
}

// Load header when the DOM is ready
document.addEventListener('DOMContentLoaded', loadSharedHeader);
