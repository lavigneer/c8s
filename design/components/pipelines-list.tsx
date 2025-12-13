"use client"

export default function PipelinesList() {
  const pipelines = [
    {
      id: 1,
      name: "main",
      commit: "feat: add user authentication",
      author: "john.doe",
      status: "success",
      duration: "4m 12s",
      time: "2 minutes ago",
    },
    {
      id: 2,
      name: "develop",
      commit: "fix: resolve database connection issue",
      author: "jane.smith",
      status: "running",
      duration: "2m 45s",
      time: "5 minutes ago",
    },
    {
      id: 3,
      name: "feature/new-ui",
      commit: "chore: update dependencies",
      author: "bob.wilson",
      status: "failed",
      duration: "3m 28s",
      time: "10 minutes ago",
    },
    {
      id: 4,
      name: "main",
      commit: "docs: update README with setup instructions",
      author: "alice.chen",
      status: "success",
      duration: "3m 55s",
      time: "23 minutes ago",
    },
  ]

  const getStatusColor = (status: string) => {
    switch (status) {
      case "success":
        return "bg-green-500"
      case "failed":
        return "bg-red-500"
      case "running":
        return "bg-blue-500"
      default:
        return "bg-muted"
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "success":
        return (
          <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="20 6 9 17 4 12" />
          </svg>
        )
      case "failed":
        return (
          <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        )
      case "running":
        return (
          <svg className="size-4 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 12a9 9 0 1 1-6.219-8.56" />
          </svg>
        )
      default:
        return null
    }
  }

  return (
    <div className="rounded-lg border border-border bg-card">
      <div className="border-b border-border p-4">
        <h2 className="text-lg font-semibold">Recent Builds</h2>
      </div>
      <div className="divide-y divide-border">
        {pipelines.map((pipeline) => (
          <div key={pipeline.id} className="p-4 hover:bg-accent/50">
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1 space-y-1">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-sm font-semibold">{pipeline.name}</span>
                  <span
                    className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs text-white ${getStatusColor(pipeline.status)}`}
                  >
                    {getStatusIcon(pipeline.status)}
                    {pipeline.status}
                  </span>
                </div>
                <div className="text-sm text-foreground">{pipeline.commit}</div>
                <div className="flex items-center gap-4 text-xs text-muted-foreground">
                  <span>{pipeline.author}</span>
                  <span>•</span>
                  <span>{pipeline.duration}</span>
                  <span>•</span>
                  <span>{pipeline.time}</span>
                </div>
              </div>
              <button className="rounded-lg border border-border bg-secondary px-3 py-1 text-sm hover:bg-accent">
                View
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
