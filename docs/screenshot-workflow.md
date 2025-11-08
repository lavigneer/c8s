# Screenshot Generation Workflow

This guide explains how to use the screenshot generation system to create and embed screenshots in C8S documentation.

## Overview

The screenshot workflow allows you to:
- Generate screenshots automatically using Playwright
- Capture responsive design variants (desktop, tablet, mobile)
- Organize screenshots by feature
- Generate markdown snippets for easy embedding in docs
- Maintain a central screenshot configuration

## Quick Start

### Generate All Screenshots

```bash
npm run screenshots
```

This command will:
1. Start the API server (if not already running)
2. Run the screenshot test suite
3. Capture all configured screenshots
4. Organize them into `docs/screenshots/` by feature
5. Print a summary of generated screenshots

### Generate Screenshots for Specific Feature

```bash
# Screenshot authentication pages only
npm run test:e2e -- --grep "@screenshots" --grep "authentication"

# Screenshot dashboard pages only
npm run test:e2e -- --grep "@screenshots" --grep "dashboard"
```

### Run Screenshots in UI Mode (Interactive)

```bash
npm run test:e2e:ui -- --grep "@screenshots"
```

This opens Playwright's interactive test UI, allowing you to:
- Watch screenshots being captured in real-time
- Inspect page state
- Re-run individual screenshot tests
- Debug issues

## Directory Structure

Screenshots are organized by feature in `docs/screenshots/`:

```
docs/
└── screenshots/
    ├── authentication/
    │   ├── login-page.png
    │   ├── login-page-desktop.png
    │   ├── login-page-tablet.png
    │   └── login-page-mobile.png
    ├── dashboard/
    │   ├── pipeline-list.png
    │   ├── quick-stats-panel.png
    │   └── filter-panel.png
    ├── pipeline/
    │   ├── pipeline-detail.png
    │   ├── pipeline-running.png
    │   └── pipeline-completed.png
    ├── logs/
    │   ├── log-viewer.png
    │   ├── log-search.png
    │   └── log-filters.png
    ├── artifacts/
    │   ├── artifact-list.png
    │   ├── artifact-preview.png
    │   └── artifact-download.png
    ├── projects/
    │   ├── project-list.png
    │   ├── project-detail.png
    │   └── webhook-config.png
    ├── keyboard-shortcuts/
    │   └── help-modal.png
    ├── responsive/
    │   ├── dashboard-desktop.png
    │   ├── dashboard-tablet.png
    │   ├── dashboard-mobile.png
    │   ├── pipeline-detail-desktop.png
    │   ├── pipeline-detail-tablet.png
    │   └── pipeline-detail-mobile.png
    └── components/
        ├── status-badges.png
        ├── error-messages.png
        └── loading-states.png
```

## Embedding Screenshots in Documentation

### Basic Image Embedding

Use this markdown syntax to embed a screenshot:

```markdown
![Login Page](./screenshots/authentication/login-page.png)
```

### With Title and Description

```markdown
#### Login Page

The login page provides a simple interface for user authentication.

![Login Page](./screenshots/authentication/login-page.png)
```

### Responsive Design Examples

For responsive screenshots, show all three variants:

```markdown
#### Dashboard on Different Devices

**Desktop View (1920x1080):**

![Dashboard - Desktop](./screenshots/responsive/dashboard-desktop.png)

**Tablet View (1024x1366):**

![Dashboard - Tablet](./screenshots/responsive/dashboard-tablet.png)

**Mobile View (390x844):**

![Dashboard - Mobile](./screenshots/responsive/dashboard-mobile.png)
```

## Adding New Screenshots

### 1. Define the Screenshot in Config

Add your test case to `tests/e2e/specs/screenshots.spec.ts`.

```typescript
{
  feature: 'my-feature',
  screenName: 'my-screen',
  filename: 'my-screen',
  description: 'Description of what this screen shows',
  setup: async () => {
    // Navigation logic here
  },
  responsive: true,  // For responsive variants
  readinessSelectors: ['[data-testid="content-loaded"]'],
  waitTime: 500,
}
```

### 2. Run the Screenshot Test

```typescript
test('capture my screen @screenshots', async ({ page }) => {
  // Navigate to your screen
  await page.goto('/my-feature');

  // Wait for content to load
  await waitForUIReady(page, ['[data-testid="content-loaded"]']);

  // Capture screenshot
  await captureScreenshot(page, 'my-feature', 'my-screen', {
    outputDir: 'docs/screenshots',
    waitTime: 500,
  });

  // For responsive variants, use:
  // await captureResponsiveScreenshots(page, 'my-feature', 'my-screen');
});
```

```bash
npm run test:e2e -- --grep "@screenshots" --grep "my screen"
```

### 3. Embed in Documentation

Find the appropriate documentation file and add:

```markdown
![My Screen](./screenshots/my-feature/my-screen.png)
```

## Advanced Usage

### Capturing Element-Specific Screenshots

To capture only a specific element instead of the full page:

```typescript
await captureScreenshot(page, 'dashboard', 'stats-panel', {
  outputDir: 'docs/screenshots',
  element: '[data-testid="quick-stats"]',  // CSS selector of element
  scrollIntoView: true,
  waitTime: 300,
});
```

### Full-Page Screenshots

For long pages that require scrolling:

```typescript
await captureScreenshot(page, 'documentation', 'full-guide', {
  outputDir: 'docs/screenshots',
  fullPage: true,  // Captures entire scrollable page
  waitTime: 1000,
});
```

### Waiting for Dynamic Content

When a page has dynamic content (real-time updates, animations):

```typescript
import { waitForUIReady } from '../utils/screenshot-utils';

await waitForUIReady(page, [
  '[data-testid="page-loaded"]',
  '.dashboard-content',
  'main',
]);
```

Or specify custom selectors:

```typescript
await waitForUIReady(page, [
  '[data-testid="custom-loader"]',
  '[data-testid="content"]',
]);
```

### Generating Markdown Snippets Programmatically

Use the utility functions to generate markdown:

```typescript
import {
  generateMarkdownSnippet,
  generateResponsiveMarkdownSnippet,
} from '../utils/screenshot-utils';

// Single screenshot
const markdown = generateMarkdownSnippet(
  'docs/screenshots/auth/login.png',
  'Login page',
  '### Login Page'
);

// Responsive screenshots
const responsiveMarkdown = generateResponsiveMarkdownSnippet(
  {
    desktop: 'docs/screenshots/dashboard/dashboard-desktop.png',
    tablet: 'docs/screenshots/dashboard/dashboard-tablet.png',
    mobile: 'docs/screenshots/dashboard/dashboard-mobile.png',
  },
  'Dashboard on different devices',
  '### Responsive Dashboard'
);
```

## Best Practices

### 1. Use Data-TestId Attributes

Mark important elements with `data-testid` for reliable selection:

```html
<div data-testid="quick-stats">
  <!-- Stats content -->
</div>
```

Then reference in screenshot tests:
```typescript
element: '[data-testid="quick-stats"]'
```

### 2. Wait for Content

Always ensure dynamic content is loaded before capturing:

```typescript
await waitForUIReady(page, ['[data-testid="page-loaded"]']);
await captureScreenshot(page, ...);
```

### 3. Consistent Viewport Sizes

Use standard viewport sizes for consistency:
- **Desktop**: 1920x1080 (default)
- **Tablet**: 1024x1366
- **Mobile**: 390x844

### 4. Meaningful Names

Use clear, descriptive screenshot names:
- ✅ `pipeline-list-with-filters`
- ❌ `screen1.png`

### 5. Organize by Feature

Group related screenshots in feature directories. This makes the screenshot folder self-documenting.

### 6. Keep Screenshots Up-to-Date

When you update UI:
1. Update the screenshot test
2. Re-run: `npm run screenshots`
3. Update the documentation with new screenshots

### 7. Use Meaningful Descriptions

In the screenshot config, add descriptions:
```typescript
description: 'Dashboard showing pipeline runs with status indicators and filtering options'
```

## Troubleshooting

### Screenshots Not Capturing?

1. Check if the test is actually navigating to the right page:
   ```bash
   npm run test:e2e:ui -- --grep "@screenshots"
   ```

2. Verify the page is ready before capturing:
   ```typescript
   await page.waitForURL(/dashboard/);
   await waitForUIReady(page);
   ```

3. Check browser console for errors:
   ```typescript
   const errors = await page.evaluate(() => {
     return console.errors;
   });
   ```

### Blurry Screenshots?

This is usually a device pixel ratio issue. The default settings should work, but if you need:

```typescript
await page.screenshot({
  path: 'screenshot.png',
  deviceScaleFactor: 2,  // For 2x resolution
});
```

### Inconsistent Screenshots?

Common causes:
- Animations still running: increase `waitTime`
- Dynamic content loading: use `waitForUIReady()`
- Font rendering differences: test on consistent OS

## Integration with CI/CD

The screenshot generation workflow runs during CI/CD:

1. **Before PR merge**: Screenshots are captured to ensure UI is documented
2. **Artifact storage**: Screenshots are uploaded as GitHub action artifacts
3. **Review**: Compare screenshots across branches to see UI changes

In GitHub Actions:
```yaml
- name: Generate Documentation Screenshots
  run: npm run screenshots

- name: Upload Screenshots
  uses: actions/upload-artifact@v3
  if: always()
  with:
    name: documentation-screenshots
    path: docs/screenshots/
```

## Reference

### Available Utilities

From `tests/e2e/utils/screenshot-utils.ts`:

- **`captureScreenshot()`** - Capture single viewport screenshot
- **`captureResponsiveScreenshots()`** - Capture desktop/tablet/mobile variants
- **`waitForUIReady()`** - Wait for page readiness
- **`generateMarkdownSnippet()`** - Generate markdown for embedding
- **`generateResponsiveMarkdownSnippet()`** - Generate markdown for responsive sets
- **`createScreenshotIndex()`** - Create index of all screenshots
- **`printScreenshotSummary()`** - Print summary to console

### Configuration Files

- **Screenshot specs**: `tests/e2e/screenshots-config.ts`
- **Screenshot tests**: `tests/e2e/specs/screenshots.spec.ts`
- **Utilities**: `tests/e2e/utils/screenshot-utils.ts`
- **Page Objects**: `tests/e2e/pages/`

## FAQ

**Q: Can I customize the output directory?**

A: Yes, pass `outputDir` option:
```typescript
await captureScreenshot(page, 'feature', 'screen', {
  outputDir: 'custom/path',
});
```

**Q: How do I capture the same screen at multiple zoom levels?**

A: Create separate tests or extend `captureResponsiveScreenshots()`:
```typescript
await captureResponsiveScreenshots(
  page,
  'feature',
  'screen',
  {
    desktop: { width: 1920, height: 1080 },
    'desktop-1440': { width: 1440, height: 900 },
    tablet: { width: 1024, height: 1366 },
  }
);
```

**Q: Can I run screenshots without starting the server?**

A: The `playwright.config.ts` has `webServer` configuration that auto-starts the server. To use an existing server:
```bash
BASE_URL=http://your-server:8080 npm run test:e2e -- --grep "@screenshots"
```

**Q: How do I exclude certain screenshots from CI/CD?**

A: Remove the `@screenshots` tag from the test or use:
```bash
npm run test:e2e -- --grep "@screenshots" --grep "desktop"  # Run desktop only
```

## Next Steps

1. Run `npm run screenshots` to generate initial screenshots
2. Add screenshots to your documentation
3. Commit `docs/screenshots/` to version control
4. Update screenshots when UI changes: `npm run screenshots`
5. Consider adding screenshot comparisons to CI/CD for detecting visual regressions
