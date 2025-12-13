"use client"

export default function RecentActivity() {
  const activities = [
    {
      id: 1,
      type: "deployment",
      message: "Deployed to production",
      user: "john.doe",
      time: "3 minutes ago",
    },
    {
      id: 2,
      type: "build",
      message: "Build completed successfully",
      user: "jane.smith",
      time: "8 minutes ago",
    },
    {
      id: 3,
      type: "error",
      message: "Build failed on test stage",
      user: "bob.wilson",
      time: "12 minutes ago",
    },
    {
      id: 4,
      type: "commit",
      message: "Pushed 3 commits to main",
      user: "alice.chen",
      time: "25 minutes ago",
    },
    {
      id: 5,
      type: "deployment",
      message: "Deployed to staging",
      user: "john.doe",
      time: "1 hour ago",
    },
  ]

  const getActivityIcon = (type: string) => {
    switch (type) {
      case "deployment":
        return (
          <div className="flex size-8 items-center justify-center rounded-full bg-green-500/10 text-green-500">
            <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
              <polyline points="22 4 12 14.01 9 11.01" />
            </svg>
          </div>
        )
      case "build":
        return (
          <div className="flex size-8 items-center justify-center rounded-full bg-blue-500/10 text-blue-500">
            <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />
            </svg>
          </div>
        )
      case "error":
        return (
          <div className="flex size-8 items-center justify-center rounded-full bg-red-500/10 text-red-500">
            <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
          </div>
        )
      case "commit":
        return (
          <div className="flex size-8 items-center justify-center rounded-full bg-purple-500/10 text-purple-500">
            <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="4" />
              <line x1="1.05" y1="12" x2="7" y2="12" />
              <line x1="17.01" y1="12" x2="22.96" y2="12" />
            </svg>
          </div>
        )
      default:
        return null
    }
  }

  return (
    <div className="rounded-lg border border-border bg-card">
      <div className="border-b border-border p-4">
        <h2 className="text-lg font-semibold">Activity Feed</h2>
      </div>
      <div className="divide-y divide-border">
        {activities.map((activity) => (
          <div key={activity.id} className="flex gap-3 p-4">
            {getActivityIcon(activity.type)}
            <div className="flex-1 space-y-1">
              <div className="text-sm">{activity.message}</div>
              <div className="text-xs text-muted-foreground">
                {activity.user} • {activity.time}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
