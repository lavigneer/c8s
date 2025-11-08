# C8S Dashboard User Guide

**Version**: 1.0
**Audience**: C8S users, developers, pipeline operators
**Last Updated**: 2025-11-02

Complete guide to using the C8S web dashboard for managing pipelines, viewing logs, and managing artifacts.

## Table of Contents

- [Dashboard Overview](#dashboard-overview)
- [Getting Started](#getting-started)
- [Navigation & Filtering](#navigation--filtering)
- [Pipeline Management](#pipeline-management)
- [Log Viewer](#log-viewer)
- [Artifact Management](#artifact-management)
- [Keyboard Shortcuts](#keyboard-shortcuts)
- [Accessibility Features](#accessibility-features)
- [Troubleshooting](#troubleshooting)

---

## Dashboard Overview

The C8S Dashboard provides a web interface for:

- **Pipeline Monitoring**: View running and completed pipelines
- **Real-time Logs**: Stream logs as pipelines execute
- **Artifact Management**: Download and preview build artifacts
- **Project Management**: Create and manage CI/CD projects
- **Status Tracking**: Monitor pipeline success rates and performance

![Dashboard Overview](../screenshots/dashboard/pipeline-history.png)

### Key Components

```
Dashboard Layout
├── Header
│   ├── Logo & Title
│   ├── Search Bar
│   └── User Menu
│
├── Sidebar
│   ├── Projects
│   ├── Pipeline History
│   ├── Quick Stats
│   └── Navigation
│
└── Main Content Area
    ├── Pipeline List
    ├── Pipeline Details
    ├── Log Viewer
    └── Artifact Browser
```

---

## Getting Started

### Accessing the Dashboard

1. **Open Dashboard**
   - Navigate to: `https://c8s-dashboard.example.com`
   - Or use internal URL: `http://c8s-api:8080`

2. **Login**
   - Click "Login" button
   - Enter credentials (OAuth/OIDC or API token)
   - Accept if prompted for permissions

![C8S Dashboard Login](../screenshots/authentication/login-page.png)

3. **First Steps**
   - Browse "Projects" in sidebar
   - View "Pipeline History" to see recent runs
   - Click a pipeline run to view details

### Dashboard Layout

**Top Navigation**
- Logo/Title on left
- Quick search in center
- User profile menu on right

**Sidebar**
- Project list (filterable)
- Recent pipelines
- Quick stats panel
- Settings

**Main Area**
- Pipeline list by default
- Details pane on selection
- Full-screen modes available

---

## Navigation & Filtering

### Browsing Projects

1. **View All Projects**
   - Click "Projects" in sidebar
   - See all projects you have access to
   - Search by name or filter by status

2. **Project Details**
   - Click project name to expand
   - View recent runs for that project
   - See project configuration (if editor)

### Filtering Pipeline Runs

The dashboard supports filtering by multiple criteria:

**Filter Options**
- **Status**: Success, Failed, Running, Cancelled
- **Branch**: Filter by git branch
- **Date Range**: Custom date selection
- **Author**: Filter by commit author
- **Search**: Full-text search in pipeline names

**Example Filters**
```
status:success branch:main
Status = Failed since:2025-11-01
branch:feature/* running:true
```

### URL State Persistence

All filters are saved in the URL:
```
https://dashboard.example.com?status=success&branch=main&page=2
```

This allows:
- Bookmarking filtered views
- Sharing views with teammates
- Browser back/forward navigation

---

## Pipeline Management

### Viewing Pipeline History

1. **Open Pipeline List**
   - Default view on dashboard load
   - Shows recent runs for accessible projects

2. **View Run Details**
   - Click any pipeline run to expand
   - See:
     - Commit info (hash, author, message)
     - Start/end times and duration
     - Resource usage
     - Step-by-step status

3. **Quick Stats**
   - Success rate (last 7 days)
   - Average duration
   - Failure count
   - Running pipelines count

### Starting a Pipeline Manually

1. **Trigger Pipeline**
   - Click "Run Pipeline" button on project
   - Select branch to build
   - Set environment variables (if needed)
   - Click "Start Run"

2. **Monitor Execution**
   - Watch pipeline status update in real-time
   - See step progress indicator
   - Monitor resource usage

### Managing Pipelines

**Editor Actions** (requires editor role):
- Edit configuration
- Disable/enable pipeline
- Delete pipeline
- Manage webhooks
- Configure environment variables

**Admin Actions** (requires admin role):
- Transfer project ownership
- Manage team access
- View audit logs
- Configure RBAC

---

## Log Viewer

### Real-Time Log Streaming

1. **Open Log Viewer**
   - Click on a running pipeline run
   - Click "Logs" tab
   - Logs stream in real-time via Server-Sent Events (SSE)

2. **Log Features**
   - **Auto-scroll**: Automatically follows new logs
   - **Pause**: Pause scrolling to read content
   - **Search**: Find text in logs (Ctrl+F)
   - **Copy**: Copy log section to clipboard
   - **Download**: Download full logs as file

### Advanced Log Filtering

**Filter by Step**
```
Select step from dropdown to show only that step's logs
All steps (default) shows merged output
```

**Filter by Level**
```
Show all
Show info+
Show warn+
Show errors only
```

**Search in Logs**
```
Press Ctrl+F or click search icon
Highlights all matches
Navigate with enter key
```

### Log Formatting

Logs are color-coded:
- **Blue**: Information messages
- **Yellow**: Warnings
- **Red**: Errors
- **Gray**: Debug output

### Secret Masking

Sensitive values are automatically masked:
- Database passwords
- API tokens
- Private keys
- AWS credentials

Example output:
```
Connecting to database: postgres://user:****@db.example.com:5432
API token: sk_live_****...****
```

---

## Artifact Management

### Viewing Artifacts

1. **Open Artifacts Tab**
   - Go to completed pipeline run
   - Click "Artifacts" tab
   - Browse file browser view

2. **File Browser**
   - Folder hierarchy displayed
   - File icons show type
   - Size and created date shown
   - Click to expand folders

### Downloading Artifacts

1. **Single File Download**
   - Click file in artifact browser
   - Click "Download" button
   - File saved to downloads folder

2. **Batch Download**
   - Select multiple files (Ctrl+Click)
   - Click "Download Selected"
   - ZIP archive created

3. **Download All**
   - Click "Download All" button in artifacts header
   - All artifacts zipped and downloaded

### Previewing Artifacts

Supported preview types:
- **Images**: PNG, JPG, GIF, WebP
- **Text**: LOG, TXT, JSON, YAML, XML
- **Code**: Java, Python, Go, JS, etc.
- **Diffs**: Show unified diff view

Click any file to preview:
```
Preview opens in side panel
Syntax highlighting for code
Image zoom available
Text search supported
```

### Artifact Filtering

Filter artifacts by:
- **Type**: Images, logs, archives, etc.
- **Size**: Show only large files
- **Date**: Recently created artifacts
- **Step**: Artifacts from specific step

---

## Keyboard Shortcuts

### Navigation

| Shortcut | Action |
|----------|--------|
| `J` | Next pipeline run |
| `K` | Previous pipeline run |
| `G H` | Go to project home |
| `G P` | Go to projects |
| `G L` | Go to pipeline list |
| `/` | Focus search |
| `?` | Show help (shortcuts) |

### Pipeline View

| Shortcut | Action |
|----------|--------|
| `Space` | Toggle run details |
| `L` | Show logs |
| `A` | Show artifacts |
| `R` | Reload/refresh |
| `C` | Copy run ID |
| `T` | Show timestamps |

### Log Viewer

| Shortcut | Action |
|----------|--------|
| `F` | Find in logs |
| `Esc` | Clear search |
| `N` | Next match |
| `Shift+N` | Previous match |
| `C` | Copy visible logs |
| `P` | Pause/resume auto-scroll |

### General

| Shortcut | Action |
|----------|--------|
| `Ctrl+S` | Save/apply filters |
| `Ctrl+K` | Open command palette |
| `Esc` | Close modals/popovers |
| `Enter` | Confirm action |

---

## Accessibility Features

### Keyboard Navigation

- **Full keyboard navigation** throughout dashboard
- Tab through all interactive elements
- Enter/Space to activate buttons
- Arrow keys for lists and menus

### Screen Reader Support

- **Semantic HTML** for proper structure
- **ARIA labels** on all interactive elements
- **Role attributes** for complex components
- **Alt text** on images

### Visual Accessibility

- **High contrast mode**: Available in settings
- **Larger fonts**: Browser zoom support
- **Focus indicators**: Clear keyboard focus
- **Color-independent info**: Don't rely on color alone

### Responsive Design

Dashboard is fully responsive:
- **Desktop**: Full features
- **Tablet**: Optimized layout
- **Mobile**: Simplified mobile interface

### Settings

Navigate to **Settings** menu for:
- Theme: Light/Dark mode
- Contrast: Standard/High
- Font size: Small/Medium/Large
- Reduced motion: Minimize animations

---

## Troubleshooting

### Common Issues

#### Dashboard Won't Load

**Problem**: Blank page or connection error

**Solutions**:
1. Clear browser cache (Ctrl+Shift+Delete)
2. Check network connectivity
3. Verify dashboard URL is correct
4. Check browser console for errors (F12)
5. Try different browser

#### Logs Not Streaming

**Problem**: Logs are empty or frozen

**Solutions**:
1. Check pipeline is actually running
2. Click pause/resume in log viewer
3. Refresh page (F5)
4. Check WebSocket connection in DevTools
5. Verify firewall allows WebSocket traffic

#### Artifacts Not Showing

**Problem**: No artifacts in artifact tab

**Solutions**:
1. Check if pipeline completed successfully
2. Verify pipeline configured to save artifacts
3. Check storage backend is accessible
4. Review pipeline logs for errors

#### Performance Issues

**Problem**: Dashboard is slow or unresponsive

**Solutions**:
1. Disable auto-refresh in settings
2. Filter to show fewer runs
3. Close other browser tabs
4. Check browser memory usage (DevTools)
5. Try incognito/private mode

### Debug Information

Enable debug mode for troubleshooting:

**Browser Console**
```javascript
// Enable verbose logging
localStorage.setItem('DEBUG', '*');
location.reload();

// Disable when done
localStorage.removeItem('DEBUG');
```

**Check Network**
```
Open DevTools (F12) → Network tab
Monitor API requests
Check for failed requests (red)
Review response status codes
```

### Getting Help

If issues persist:
1. Check [Troubleshooting Guide](./TROUBLESHOOTING.md)
2. Review [Configuration Guide](./CONFIGURATION.md)
3. Check server logs: `kubectl logs -f deployment/c8s-api -n c8s-system`
4. Open GitHub issue with dashboard screenshots

---

## Related Documentation

- [Getting Started](./GETTING_STARTED.md) - Quick start guide
- [Configuration](./CONFIGURATION.md) - Configuration options
- [Operator Guide](./OPERATOR_GUIDE.md) - Deployment guide
- [Troubleshooting](./TROUBLESHOOTING.md) - Common issues and solutions
