/**
 * Theme Manager
 * Handles light/dark mode switching with persistence
 */

class ThemeManager {
    constructor() {
        this.STORAGE_KEY = 'c8s-theme-preference';
        this.LIGHT_CLASS = 'light';
        this.DARK_CLASS = 'dark';
        this.THEME_ATTR = 'data-theme';

        this.init();
    }

    /**
     * Initialize theme manager
     */
    init() {
        // Check localStorage for saved preference
        const savedTheme = localStorage.getItem(this.STORAGE_KEY);

        if (savedTheme) {
            this.setTheme(savedTheme);
        } else {
            // Check system preference
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            this.setTheme(prefersDark ? this.DARK_CLASS : this.LIGHT_CLASS);
        }

        // Listen for system theme changes
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            if (!localStorage.getItem(this.STORAGE_KEY)) {
                this.setTheme(e.matches ? this.DARK_CLASS : this.LIGHT_CLASS);
            }
        });

        // Set up theme toggle buttons
        this.setupToggleButtons();
    }

    /**
     * Set the theme
     * @param {string} theme - 'light' or 'dark'
     */
    setTheme(theme) {
        const root = document.documentElement;

        if (theme === this.DARK_CLASS) {
            root.classList.remove(this.LIGHT_CLASS);
            root.classList.add(this.DARK_CLASS);
            document.body.classList.add('dark');
            localStorage.setItem(this.STORAGE_KEY, this.DARK_CLASS);
        } else {
            root.classList.remove(this.DARK_CLASS);
            root.classList.add(this.LIGHT_CLASS);
            document.body.classList.remove('dark');
            localStorage.setItem(this.STORAGE_KEY, this.LIGHT_CLASS);
        }

        // Update theme attribute for CSS selectors
        root.setAttribute(this.THEME_ATTR, theme);

        // Dispatch custom event
        window.dispatchEvent(new CustomEvent('themechange', { detail: { theme } }));

        // Update toggle buttons
        this.updateToggleButtons(theme);
    }

    /**
     * Get current theme
     * @returns {string} 'light' or 'dark'
     */
    getTheme() {
        return localStorage.getItem(this.STORAGE_KEY) ||
               (window.matchMedia('(prefers-color-scheme: dark)').matches ? this.DARK_CLASS : this.LIGHT_CLASS);
    }

    /**
     * Toggle between light and dark theme
     */
    toggle() {
        const currentTheme = this.getTheme();
        const newTheme = currentTheme === this.LIGHT_CLASS ? this.DARK_CLASS : this.LIGHT_CLASS;
        this.setTheme(newTheme);
    }

    /**
     * Set up theme toggle buttons
     */
    setupToggleButtons() {
        document.querySelectorAll('[data-theme-toggle]').forEach(button => {
            button.addEventListener('click', () => this.toggle());
        });
    }

    /**
     * Update toggle button appearance
     * @param {string} theme - Current theme
     */
    updateToggleButtons(theme) {
        document.querySelectorAll('[data-theme-toggle]').forEach(button => {
            if (theme === this.DARK_CLASS) {
                button.setAttribute('title', 'Switch to light mode');
                button.textContent = '☀️ Light';
            } else {
                button.setAttribute('title', 'Switch to dark mode');
                button.textContent = '🌙 Dark';
            }
        });
    }

    /**
     * Force a specific theme
     * @param {string} theme - 'light' or 'dark'
     */
    force(theme) {
        localStorage.setItem(this.STORAGE_KEY, theme);
        this.setTheme(theme);
    }

    /**
     * Remove theme preference (use system default)
     */
    useSystemDefault() {
        localStorage.removeItem(this.STORAGE_KEY);
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        this.setTheme(prefersDark ? this.DARK_CLASS : this.LIGHT_CLASS);
    }
}

// Initialize theme manager on DOM content loaded
document.addEventListener('DOMContentLoaded', () => {
    window.themeManager = new ThemeManager();
});
