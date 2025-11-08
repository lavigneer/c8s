# Screenshot Documentation Guide

This document provides a quick reference for working with the automated screenshot generation system for C8S documentation.

## For Documentation Writers

### Embedding Screenshots

Once screenshots are generated, embed them in your markdown documentation:

```markdown
![Description of image](./screenshots/feature-name/screenshot-name.png)
```

Example:
```markdown
## Dashboard Overview

The main dashboard displays your pipeline runs with real-time status updates.

![Pipeline Dashboard](./screenshots/dashboard/pipeline-list.png)
```

### Screenshot Locations

All documentation screenshots are stored in:
```
docs/screenshots/
├── authentication/      # Login, logout, sessions
├── dashboard/          # Main dashboard views
├── pipeline/           # Pipeline detail pages
├── logs/              # Log viewer interface
├── artifacts/         # Artifact browser
├── projects/          # Project management
├── keyboard-shortcuts/ # Keyboard help modal
├── responsive/        # Mobile/tablet views
└── components/        # UI component examples
```

### Best Practices for Documentation

1. **Use descriptive alt text**: This helps accessibility and SEO
   ```markdown
   ![Dashboard showing pipeline runs with green success badges and status indicators](./screenshots/dashboard/pipeline-list.png)
   ```

2. **Provide context**: Explain what the screenshot shows
   ```markdown
   #### Pipeline Dashboard

   The main dashboard displays your recent pipeline runs. You can filter by status, branch, and date range using the controls above.

   ![Pipeline Dashboard](./screenshots/dashboard/pipeline-list.png)
   ```

3. **Link to detailed guides**: Screenshots work best with explanatory text
   ```markdown
   See the [Pipeline Management Guide](./guides/dashboard-features.md) for detailed instructions on filtering and running pipelines.
   ```

4. **Show responsive variants**: For responsive screenshots, display all three:
   ```markdown
   #### Mobile-Friendly Design

   C8S works seamlessly on desktop, tablet, and mobile devices.

   **Desktop View:**
   ![Dashboard on desktop](./screenshots/responsive/dashboard-desktop.png)

   **Tablet View:**
   ![Dashboard on tablet](./screenshots/responsive/dashboard-tablet.png)

   **Mobile View:**
   ![Dashboard on mobile](./screenshots/responsive/dashboard-mobile.png)
   ```

## For Developers

### Regenerating Screenshots

When you update the UI, regenerate screenshots:

```bash
npm run screenshots
```

This will:
- Start the development server
- Run all screenshot capture tests
- Generate new screenshots in `docs/screenshots/`
- Print a summary of what was captured

### Running Specific Screenshots

To capture only certain features:

```bash
# Authentication screenshots only
npm run test:e2e -- --grep "@screenshots" --grep "authentication"

# Dashboard and pipeline screenshots
npm run test:e2e -- --grep "@screenshots" --grep "dashboard|pipeline"
```

### Interactive Mode

Use the Playwright UI to visually monitor screenshot capture:

```bash
npm run screenshots:ui
```

This opens a browser-based test interface where you can:
- Watch tests run in real-time
- Step through individual tests
- Inspect page state
- Re-run failed screenshot captures

### Adding New Screenshots

1. **Add test case** in `tests/e2e/specs/screenshots.spec.ts`:
   ```typescript
   test('capture my new feature @screenshots', async ({ page }) => {
     await page.goto('/my-feature');
     await waitForUIReady(page);
     await captureScreenshot(page, 'my-feature', 'my-screen', {
       outputDir: 'docs/screenshots',
     });
   });
   ```

2. **Run the screenshot test**:
   ```bash
   npm run test:e2e -- --grep "@screenshots" --grep "my new feature"
   ```

3. **Find the generated image** in `docs/screenshots/my-feature/my-screen.png`

4. **Commit** the screenshot to version control:
   ```bash
   git add docs/screenshots/
   git commit -m "Add screenshots for my new feature"
   ```

### Advanced Techniques

#### Element-Specific Screenshots

Capture just one element on a page:

```typescript
await captureScreenshot(page, 'dashboard', 'stats-panel', {
  element: '[data-testid="quick-stats"]',
});
```

#### Full-Page Screenshots

For pages longer than viewport:

```typescript
await captureScreenshot(page, 'guide', 'full-page', {
  fullPage: true,  // Captures entire scrollable page
});
```

#### Responsive Variants

Generate desktop/tablet/mobile versions automatically:

```typescript
await captureResponsiveScreenshots(page, 'my-feature', 'my-screen');
```

This creates:
- `my-screen-desktop.png`
- `my-screen-tablet.png`
- `my-screen-mobile.png`

#### Wait for Dynamic Content

Wait for specific selectors before capturing:

```typescript
import { waitForUIReady } from '../utils/screenshot-utils';

await waitForUIReady(page, [
  '[data-testid="page-loaded"]',
  '[data-testid="content"]',
]);
```

## Screenshot Guidelines

### What to Screenshot

✅ **Good candidates**:
- Main feature pages (dashboard, pipelines, logs)
- Key UI workflows (login, create pipeline)
- Important UI components (filters, status badges)
- Responsive design variants
- Error states and validation messages
- Help modals and keyboard shortcuts

❌ **Avoid**:
- Sensitive data (API keys, real credentials)
- Third-party services
- Transient UI states (loading spinners, temporary notifications)
- Very large pages (use element-specific screenshots instead)

### Naming Conventions

Use clear, descriptive names that indicate the feature and state:

- ✅ `pipeline-list.png` - Good: describes what's shown
- ✅ `pipeline-running.png` - Good: includes state
- ✅ `login-error.png` - Good: shows error scenario
- ❌ `screen1.png` - Poor: not descriptive
- ❌ `test.png` - Poor: vague

### Size and Format

- **Format**: PNG (lossless, good for UI)
- **Resolution**: 1920x1080 (desktop), 1024x1366 (tablet), 390x844 (mobile)
- **Size**: Optimize with tools like TinyPNG before committing
- **Compression**: GitHub will compress for efficient storage

## Maintenance

### Keeping Screenshots Updated

1. **After UI changes**: Run `npm run screenshots` to regenerate
2. **Review diff**: Check git diff to see what changed
3. **Update docs**: Add/modify documentation to match new screenshots
4. **Commit together**: Commit code changes and updated screenshots

### Version Control

Screenshots should be committed to git:

```bash
git add docs/screenshots/
git commit -m "Update screenshots for new dashboard layout"
```

Alternatively, configure `.gitattributes` for image compression:

```
docs/screenshots/*.png filter=lfs diff=lfs merge=lfs -text
```

## File Structure Reference

```
c8s/
├── docs/
│   ├── screenshots/                    # All generated screenshots
│   │   ├── authentication/
│   │   ├── dashboard/
│   │   ├── pipeline/
│   │   ├── logs/
│   │   ├── artifacts/
│   │   ├── projects/
│   │   ├── keyboard-shortcuts/
│   │   ├── responsive/
│   │   └── components/
│   ├── screenshot-workflow.md         # Detailed workflow guide
│   ├── screenshots-guide.md           # This file
│   ├── guides/
│   │   ├── dashboard-features.md      # Main dashboard documentation
│   │   ├── pipeline-syntax.md
│   │   └── ...
│   └── ...
├── tests/
│   └── e2e/
│       ├── specs/
│       │   └── screenshots.spec.ts    # Screenshot capture tests
│       ├── utils/
│       │   └── screenshot-utils.ts    # Screenshot utilities
│       ├── screenshots-config.ts      # Screenshot specifications
│       └── pages/                     # Page objects for tests
└── package.json                        # npm scripts including screenshot commands
```

## Troubleshooting

### Screenshots look different on CI vs locally

**Cause**: Different operating system or browser version

**Solution**:
- Use Chromium (configured by default)
- The playwright.config.ts sets consistent viewport sizes
- CI runs with consistent environment

### Screenshot capture is timing out

**Cause**: Page taking too long to load

**Solution**:
```typescript
await waitForUIReady(page, ['[data-testid="page-loaded"]']);
await page.waitForTimeout(1000);  // Additional wait if needed
```

### Some screenshots are blurry

**Cause**: Device pixel ratio rendering

**Solution**: Usually resolves by running on consistent hardware. CI uses standardized environments.

## Quick Command Reference

| Command | Purpose |
|---------|---------|
| `npm run screenshots` | Generate all screenshots |
| `npm run screenshots:ui` | Generate with interactive UI |
| `npm run screenshots:debug` | Generate with debugger |
| `npm run screenshots:report` | View screenshot test report |
| `npm run test:e2e` | Run all E2E tests |
| `npm run test:e2e:ui` | Run tests interactively |

## Support

For issues or questions:

1. **Screenshot utility issues**: Check `tests/e2e/utils/screenshot-utils.ts`
2. **Test failures**: Review `tests/e2e/specs/screenshots.spec.ts`
3. **Workflow questions**: See `docs/screenshot-workflow.md`
4. **Documentation**: See `docs/guides/dashboard-features.md`

## Next Steps

1. Run `npm run screenshots` to generate initial screenshots
2. Add screenshots to your documentation files
3. Commit `docs/screenshots/` to version control
4. Update screenshots whenever UI changes
5. Use in PR reviews to visualize UI changes
