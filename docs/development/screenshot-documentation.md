# Screenshot Documentation Workflow

Guide for capturing and managing screenshots of the C8S dashboard for documentation.

## Quick Start

Generate desktop screenshots of the dashboard:

```bash
npm run screenshots
```

This will:
- Start the API server automatically
- Capture screenshots (desktop size: 1920x1080)
- Save to `docs/screenshots/` organized by feature
- Print a summary when complete

Screenshots are tagged with `@screenshots` for easy test filtering:

```bash
# Run only screenshot tests
npm run test:e2e -- --grep @screenshots

# Run specific feature screenshots
npm run test:e2e -- --grep "@screenshots" --grep "dashboard"
```

## NPM Scripts

| Command | Purpose |
|---------|---------|
| `npm run screenshots` | Generate all screenshots |
| `npm run screenshots:ui` | Generate with interactive Playwright UI |
| `npm run screenshots:debug` | Generate with Playwright debugger |
| `npm run screenshots:report` | View HTML test report |

## How It Works

### Screenshot Generation Architecture

**Utilities** (`tests/e2e/utils/screenshot-utils.ts`):
- `captureScreenshot()` - Capture single viewport screenshot
- `captureResponsiveScreenshots()` - Generate multiple viewport sizes (currently unused)
- `waitForUIReady()` - Wait for dynamic content before capturing
- `generateMarkdownSnippet()` - Generate markdown for embedding
- `printScreenshotSummary()` - Print capture summary

**Tests** (`tests/e2e/specs/screenshots.spec.ts`):
- Login page (always accessible)
- Dashboard/pipeline history (main view)
- Pipeline detail, logs, artifacts (if pipelines exist)
- Projects page (if accessible)

**Configuration** (`tests/e2e/screenshots-config.ts`):
- Pre-defined screenshot specifications
- Easy to extend with new pages

### Screenshot Organization

Generated screenshots go to `docs/screenshots/` organized by feature:

```
docs/screenshots/
├── authentication/
│   └── login-page.png
├── dashboard/
│   └── pipeline-history.png
├── pipeline/
│   └── pipeline-detail.png
├── logs/
│   └── log-viewer.png
├── artifacts/
│   └── artifact-list.png
└── projects/
    └── projects-list.png
```

## Embedding Screenshots in Documentation

### Basic Usage

In any documentation file:

```markdown
![Dashboard Overview](../screenshots/dashboard/pipeline-history.png)
```

### With Context

```markdown
## Pipeline History

The dashboard shows your recent pipeline runs with status indicators.

![Dashboard Overview](../screenshots/dashboard/pipeline-history.png)

Use the filters to find specific pipelines by status, branch, or date.
```

### Common Locations

- **User guides** (`docs/guides/`): Embed screenshots showing features
- **Getting started** (`docs/guides/getting-started.md`): Show login and dashboard
- **Feature documentation**: Screenshot of each feature with explanation

## Adding New Screenshots

### 1. Add Test Case

Edit `tests/e2e/specs/screenshots.spec.ts`:

```typescript
test('capture my feature @screenshots', async ({ page }) => {
  await page.goto('/my-feature');
  await waitForUIReady(page);

  await captureScreenshot(page, 'my-feature', 'my-screen', {
    outputDir: 'docs/screenshots',
    waitTime: 500,
  });
});
```

### 2. Run Test

```bash
npm run test:e2e -- --grep "@screenshots" --grep "my feature"
```

### 3. Embed Screenshot

Screenshot will be saved to: `docs/screenshots/my-feature/my-screen.png`

Add to documentation:

```markdown
![My Feature](../screenshots/my-feature/my-screen.png)
```

## Best Practices

### Before Capturing Screenshots

1. **Ensure page elements have `data-testid` attributes** for reliable targeting:
   ```html
   <div data-testid="quick-stats">...</div>
   ```

2. **Set up test authentication** - Tests use `setupTestAuth()` to avoid login flow
   ```typescript
   await setupTestAuth(page);
   ```

3. **Wait for dynamic content** - Use `waitForUIReady()` before capturing:
   ```typescript
   await waitForUIReady(page, ['[data-testid="content-loaded"]']);
   ```

4. **Gracefully handle optional content** - Skip if elements don't exist:
   ```typescript
   const element = page.locator('[data-testid="optional"]');
   if (await element.isVisible({ timeout: 2000 }).catch(() => false)) {
     // Capture this element
   }
   ```

### File Naming

- Use descriptive names: `pipeline-detail.png` not `screen1.png`
- Group by feature in directories
- Keep names lowercase and hyphenated
- Match documentation context

### Documentation Integration

- **Always add context** - Explain what the screenshot shows
- **Keep screenshots current** - Regenerate when UI changes
- **Commit screenshots** - Include in version control with code changes
- **Link related content** - Point to detailed guides from screenshots

## Maintenance

### When UI Changes

1. Update the UI code
2. Run `npm run screenshots` to regenerate affected screenshots
3. Update documentation if needed
4. Commit screenshots with code changes

### Troubleshooting

**Screenshots not generating?**
- Check if the test is running: `npm run screenshots:ui`
- Verify the page loads: Check browser console for errors
- Ensure auth is set up: `setupTestAuth(page)` in beforeEach

**Screenshots are blank?**
- Increase `waitTime` in capture options
- Use `waitForUIReady()` with custom selectors
- Check if page elements have `data-testid` attributes

**Blurry screenshots?**
- Ensure consistent viewport size (1920x1080)
- Run on same hardware for consistency
- Playwright handles DPI scaling automatically

## Viewing Results

### HTML Report

```bash
npm run screenshots:report
```

Opens an interactive Playwright report showing:
- Test results
- Screenshots captured
- Timing information
- Error details

### Inspect Generated Files

```bash
# List all screenshots
find docs/screenshots -type f -name "*.png" | sort

# Check file sizes
du -sh docs/screenshots/
```

## Documentation Standards

From `.claude/skills/documentation.md`:

- **All markdown docs**: lowercase, hyphenated naming
- **All docs**: organized in `docs/` directory with subdirectories
- **Screenshots**: stored in `docs/screenshots/` by feature
- **Related docs**: kept in `docs/development/` for developer workflows

## Testing Integration

Screenshots are run as part of the E2E test suite:

```bash
# Include in full test run
npm run test:e2e

# Run only screenshots
npm run test:e2e -- --grep @screenshots
```

In CI/CD, screenshots are:
- Captured on each run
- Stored as artifacts for review
- Used to detect visual regressions

## Future Enhancements

Potential improvements:

- [ ] Visual regression testing (compare against baseline)
- [ ] Automated image optimization/compression
- [ ] Screenshot comparison in PRs
- [ ] Multi-viewport support (currently desktop-only)
- [ ] Accessibility report generation from screenshots
- [ ] Animated GIF recording for interactive features

## References

- **Tests**: `tests/e2e/specs/screenshots.spec.ts`
- **Utilities**: `tests/e2e/utils/screenshot-utils.ts`
- **Configuration**: `tests/e2e/screenshots-config.ts`
- **Skills**: `.claude/skills/documentation.md`

## Related Documentation

- [Getting Started Guide](../guides/getting-started.md) - Includes login and dashboard screenshots
- [Dashboard Features Guide](../guides/dashboard-features.md) - Includes UI screenshots
- [Development Guide](./development.md) - General development setup
