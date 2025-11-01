/**
 * Clipboard Utility Functions
 * Provides copy-to-clipboard functionality with visual feedback
 */

/**
 * Copy text to clipboard with visual feedback
 * @param {string} text - Text to copy
 * @param {HTMLElement} element - Element to show feedback near
 */
function copyToClipboard(text, element) {
    // Use modern Clipboard API if available
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text)
            .then(() => {
                showCopyFeedback(element);
            })
            .catch(err => {
                console.error('Failed to copy:', err);
                fallbackCopyToClipboard(text, element);
            });
    } else {
        // Fallback for older browsers
        fallbackCopyToClipboard(text, element);
    }
}

/**
 * Fallback method for copying to clipboard (older browsers)
 * @param {string} text - Text to copy
 * @param {HTMLElement} element - Element to show feedback near
 */
function fallbackCopyToClipboard(text, element) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    document.body.appendChild(textArea);

    try {
        textArea.select();
        document.execCommand('copy');
        showCopyFeedback(element);
    } catch (err) {
        console.error('Fallback copy failed:', err);
        alert('Failed to copy. Please try again.');
    } finally {
        document.body.removeChild(textArea);
    }
}

/**
 * Show visual feedback when copy is successful
 * @param {HTMLElement} element - Element to show feedback near
 */
function showCopyFeedback(element) {
    // Create tooltip
    const tooltip = document.createElement('div');
    tooltip.className = 'fixed bg-gray-900 text-white px-3 py-2 rounded text-sm z-50';
    tooltip.textContent = '✓ Copied!';
    tooltip.style.pointerEvents = 'none';
    tooltip.style.animation = 'fadeInOut 2s ease-in-out';

    // Position tooltip near element
    if (element && element.getBoundingClientRect) {
        const rect = element.getBoundingClientRect();
        tooltip.style.left = (rect.left + rect.width / 2 - 30) + 'px';
        tooltip.style.top = (rect.top - 40) + 'px';
    }

    document.body.appendChild(tooltip);

    // Remove after animation
    setTimeout(() => {
        tooltip.remove();
    }, 2000);
}

/**
 * Add copy button to an element
 * @param {HTMLElement} element - Element to add copy button to
 * @param {string} textToCopy - Text to copy when button is clicked
 * @param {string} buttonText - Button text (default: "Copy")
 */
function addCopyButton(element, textToCopy, buttonText = 'Copy') {
    const button = document.createElement('button');
    button.className = 'ml-2 px-2 py-1 text-xs bg-gray-200 hover:bg-gray-300 rounded transition';
    button.textContent = buttonText;
    button.onclick = (e) => {
        e.preventDefault();
        copyToClipboard(textToCopy, button);
    };

    element.appendChild(button);
    return button;
}

/**
 * Add copy functionality to all elements with data-copy attribute
 * Usage: <span data-copy="text to copy">Visible Text</span>
 */
function initializeCopyElements() {
    document.querySelectorAll('[data-copy]').forEach(element => {
        element.style.cursor = 'pointer';
        element.title = 'Click to copy';

        element.addEventListener('click', () => {
            const textToCopy = element.getAttribute('data-copy');
            copyToClipboard(textToCopy, element);
        });
    });
}

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', initializeCopyElements);

// Add CSS animation for tooltip
const style = document.createElement('style');
style.textContent = `
    @keyframes fadeInOut {
        0% {
            opacity: 0;
            transform: translateY(-5px);
        }
        10% {
            opacity: 1;
            transform: translateY(0);
        }
        90% {
            opacity: 1;
            transform: translateY(0);
        }
        100% {
            opacity: 0;
            transform: translateY(-5px);
        }
    }

    [data-copy] {
        position: relative;
    }

    [data-copy]:hover {
        text-decoration: underline;
    }
`;
document.head.appendChild(style);
