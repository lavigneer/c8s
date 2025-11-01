/**
 * Search Highlight Utility
 * Highlights search terms in pipeline list when filters are applied
 */

class SearchHighlighter {
    constructor() {
        this.highlightClass = 'search-highlight';
        this.init();
    }

    /**
     * Initialize search highlighter
     */
    init() {
        // Listen for filter changes
        document.addEventListener('htmx:afterSwap', (event) => {
            if (event.detail.target.id === 'pipeline-rows') {
                this.highlightSearchTerms();
            }
        });

        // Also highlight on initial page load
        document.addEventListener('DOMContentLoaded', () => {
            this.highlightSearchTerms();
        });
    }

    /**
     * Highlight search terms in the pipeline list
     */
    highlightSearchTerms() {
        // Get search term from filter input
        const searchInput = document.querySelector('input[name="search"]');
        if (!searchInput || !searchInput.value) {
            this.removeHighlights();
            return;
        }

        const searchTerm = searchInput.value.toLowerCase();
        if (searchTerm.length === 0) {
            this.removeHighlights();
            return;
        }

        // Find all searchable elements in pipeline rows
        const pipelineRows = document.querySelectorAll('[id^="pipeline-"]');

        pipelineRows.forEach(row => {
            // Highlight in pipeline name
            this.highlightInElement(row.querySelector('.text-lg'), searchTerm);

            // Highlight in commit SHA
            const shaElement = row.querySelector('span[data-copy]');
            if (shaElement) {
                this.highlightInElement(shaElement.parentElement, searchTerm);
            }

            // Highlight in branch name
            const branchElement = row.querySelector('span.font-medium');
            if (branchElement) {
                this.highlightInElement(branchElement, searchTerm);
            }

            // Highlight in author
            const authorElement = row.querySelector('.text-xs:last-of-type');
            if (authorElement) {
                this.highlightInElement(authorElement, searchTerm);
            }
        });
    }

    /**
     * Highlight search term in a specific element
     * @param {HTMLElement} element - Element to search in
     * @param {string} searchTerm - Term to highlight
     */
    highlightInElement(element, searchTerm) {
        if (!element) return;

        // Get text content
        const text = element.textContent;

        // Create regex for case-insensitive search
        const regex = new RegExp(`(${this.escapeRegex(searchTerm)})`, 'gi');

        // Only highlight if term is found
        if (regex.test(text)) {
            // Reset the regex test
            regex.lastIndex = 0;

            // Create highlighted HTML
            const highlightedHTML = text.replace(regex, '<mark class="' + this.highlightClass + '">$1</mark>');

            // Only update if content changed
            if (highlightedHTML !== text) {
                element.innerHTML = highlightedHTML;
            }
        }
    }

    /**
     * Remove all highlights from pipeline list
     */
    removeHighlights() {
        document.querySelectorAll('.' + this.highlightClass).forEach(element => {
            const parent = element.parentNode;
            while (element.firstChild) {
                parent.insertBefore(element.firstChild, element);
            }
            parent.removeChild(element);
            parent.normalize();
        });
    }

    /**
     * Escape special regex characters
     * @param {string} str - String to escape
     * @returns {string} Escaped string
     */
    escapeRegex(str) {
        return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    }

    /**
     * Get search stats (how many matches found)
     * @returns {object} Stats object with counts
     */
    getStats() {
        const highlights = document.querySelectorAll('.' + this.highlightClass);
        return {
            highlightCount: highlights.length,
            pipelineCount: document.querySelectorAll('[id^="pipeline-"]').length
        };
    }
}

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    window.searchHighlighter = new SearchHighlighter();
});

// Add CSS for highlight styling
const style = document.createElement('style');
style.textContent = `
    .search-highlight {
        background-color: #fcd34d;
        color: #1f2937;
        font-weight: 600;
        padding: 0 2px;
        border-radius: 2px;
        transition: background-color 0.2s ease;
    }

    .dark .search-highlight {
        background-color: #f59e0b;
        color: #1f2937;
    }

    .search-highlight:hover {
        background-color: #fbbf24;
    }

    .dark .search-highlight:hover {
        background-color: #d97706;
    }
`;
document.head.appendChild(style);
