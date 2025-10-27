/**
 * Keyboard Shortcuts Manager
 * Implements keyboard shortcuts for common dashboard actions per FR-013
 */
class KeyboardShortcutManager {
  constructor() {
    // Platform-aware key combinations (Cmd on Mac, Ctrl on Windows/Linux)
    const modKey = /Mac|iPhone|iPad|iPod/.test(navigator.platform) ? 'cmd' : 'ctrl';

    this.shortcuts = {
      '?': { handler: () => this.showHelp(), description: 'Show keyboard shortcuts help' },
      [modKey + '+k']: { handler: () => this.focusSearch(), description: 'Focus search input' },
      [modKey + '+r']: { handler: () => this.refreshPage(), description: 'Refresh data' },
      [modKey + '+l']: { handler: () => this.jumpToLatestLog(), description: 'Jump to latest log' },
      'escape': { handler: () => this.closeModal(), description: 'Close modal' },
      [modKey + '+enter']: { handler: () => this.submitForm(), description: 'Submit active form' },
      'j': { handler: () => this.navigateDown(), description: 'Next pipeline run' },
      'k': { handler: () => this.navigateUp(), description: 'Previous pipeline run' },
      'x': { handler: () => this.cancelPipeline(), description: 'Cancel selected pipeline' },
      'd': { handler: () => this.downloadArtifact(), description: 'Download artifact' },
      'v': { handler: () => this.viewArtifact(), description: 'View artifact' },
      [modKey + '+']: { handler: () => this.toggleFilterPanel(), description: 'Toggle filter panel' },
      [modKey + '+s']: { handler: () => this.saveFilterPreset(), description: 'Save filter preset' },
    };

    this.isInputFocused = false;
  }

  /**
   * Register keyboard event listener
   */
  register() {
    document.addEventListener('keydown', (e) => this.handleKeyPress(e));
    document.addEventListener('focus', (e) => this.updateInputFocus(), true);
    document.addEventListener('blur', (e) => this.updateInputFocus(), true);
  }

  /**
   * Update whether input is focused
   */
  updateInputFocus() {
    const activeEl = document.activeElement;
    this.isInputFocused = activeEl && (
      activeEl.tagName === 'INPUT' ||
      activeEl.tagName === 'TEXTAREA' ||
      activeEl.contentEditable === 'true'
    );
  }

  /**
   * Build key combination string from keyboard event
   */
  getKeyCombo(event) {
    const parts = [];

    // Add modifier keys
    if (event.ctrlKey) parts.push('ctrl');
    if (event.metaKey) parts.push('cmd');
    if (event.altKey) parts.push('alt');
    if (event.shiftKey) parts.push('shift');

    // Add the key itself
    const key = event.key.toLowerCase();
    if (key !== 'control' && key !== 'meta' && key !== 'alt' && key !== 'shift') {
      parts.push(key === ' ' ? 'space' : key);
    }

    return parts.join('+');
  }

  /**
   * Handle keyboard press
   */
  handleKeyPress(event) {
    // Don't process shortcuts when typing in input fields (except special cases)
    const keyCombo = this.getKeyCombo(event);
    const shortcut = this.shortcuts[keyCombo];

    if (!shortcut) {
      return;
    }

    // Allow some shortcuts in input fields
    const allowedInInput = ['escape', 'ctrl+enter', 'cmd+enter'];
    if (this.isInputFocused && !allowedInInput.includes(keyCombo)) {
      return;
    }

    event.preventDefault();
    shortcut.handler();
  }

  /**
   * Show help modal with available shortcuts
   */
  showHelp() {
    const modal = document.getElementById('shortcuts-help-modal');
    if (modal) {
      modal.classList.remove('hidden');
    } else {
      console.warn('Shortcuts help modal not found');
    }
  }

  /**
   * Focus on search input
   */
  focusSearch() {
    const searchInput = document.querySelector('input[type="search"], input[placeholder*="Search"], input[placeholder*="search"]');
    if (searchInput) {
      searchInput.focus();
    }
  }

  /**
   * Refresh page data
   */
  refreshPage() {
    // Trigger HTMX refresh on main content area
    const content = document.getElementById('content');
    if (content) {
      htmx.ajax('GET', window.location.pathname, { target: '#content', swap: 'innerHTML' });
    } else {
      location.reload();
    }
  }

  /**
   * Jump to latest log line
   */
  jumpToLatestLog() {
    const logSection = document.getElementById('log-section');
    if (logSection) {
      logSection.scrollIntoView({ behavior: 'smooth' });
    }
  }

  /**
   * Close currently open modal
   */
  closeModal() {
    const modal = document.querySelector('.modal:not(.hidden)');
    if (modal) {
      modal.classList.add('hidden');
    }
  }

  /**
   * Submit active form
   */
  submitForm() {
    const form = document.querySelector('form:focus-within') || document.querySelector('form');
    if (form) {
      form.submit();
    }
  }

  /**
   * Navigate to next pipeline run
   */
  navigateDown() {
    // Find next pipeline row and scroll to it
    const current = document.querySelector('.pipeline-row.focused');
    if (current && current.nextElementSibling) {
      current.classList.remove('focused');
      current.nextElementSibling.classList.add('focused');
      current.nextElementSibling.scrollIntoView({ behavior: 'smooth' });
    }
  }

  /**
   * Navigate to previous pipeline run
   */
  navigateUp() {
    const current = document.querySelector('.pipeline-row.focused');
    if (current && current.previousElementSibling) {
      current.classList.remove('focused');
      current.previousElementSibling.classList.add('focused');
      current.previousElementSibling.scrollIntoView({ behavior: 'smooth' });
    }
  }

  /**
   * Cancel selected pipeline
   */
  cancelPipeline() {
    const current = document.querySelector('.pipeline-row.focused');
    if (current) {
      const cancelBtn = current.querySelector('[data-action="cancel"]');
      if (cancelBtn) {
        cancelBtn.click();
      }
    }
  }

  /**
   * Download artifact
   */
  downloadArtifact() {
    const downloadBtn = document.querySelector('[data-action="download"]');
    if (downloadBtn) {
      downloadBtn.click();
    }
  }

  /**
   * View artifact
   */
  viewArtifact() {
    const previewBtn = document.querySelector('[data-action="preview"]');
    if (previewBtn) {
      previewBtn.click();
    }
  }

  /**
   * Toggle filter panel
   */
  toggleFilterPanel() {
    const filterPanel = document.getElementById('filter-panel');
    if (filterPanel) {
      filterPanel.classList.toggle('hidden');
    }
  }

  /**
   * Save filter preset
   */
  saveFilterPreset() {
    const saveBtn = document.querySelector('[data-action="save-filter"]');
    if (saveBtn) {
      saveBtn.click();
    }
  }

  /**
   * Get available shortcuts for display in help
   */
  getShortcuts() {
    const shortcuts = [];
    for (const [keyCombo, { description }] of Object.entries(this.shortcuts)) {
      shortcuts.push({ keys: keyCombo, description });
    }
    return shortcuts;
  }
}

// Auto-initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  const manager = new KeyboardShortcutManager();
  manager.register();
  window.keyboardShortcutManager = manager;
});
