"use client"

import Link from "next/link"

interface PipelineDetailProps {
  projectId: string
  id: string
}

export default function PipelineDetail({ projectId, id }: PipelineDetailProps) {
  // Mock data - in a real app, this would fetch based on the id
  const pipeline = {
    id,
    projectId,
    projectName: "web-application",
    name: "main",
    commit: "feat: add user authentication",
    commitHash: "a1b2c3d",
    author: "john.doe",
    status: "success",
    duration: "4m 12s",
    time: "2 minutes ago",
    stages: [
      {
        name: "Build",
        status: "success",
        duration: "1m 23s",
        steps: [
          { name: "Checkout code", status: "success", duration: "12s" },
          { name: "Install dependencies", status: "success", duration: "45s" },
          { name: "Build application", status: "success", duration: "26s" },
        ],
      },
      {
        name: "Test",
        status: "success",
        duration: "2m 15s",
        steps: [
          { name: "Run unit tests", status: "success", duration: "1m 34s" },
          { name: "Run integration tests", status: "success", duration: "41s" },
        ],
      },
      {
        name: "Deploy",
        status: "success",
        duration: "34s",
        steps: [
          { name: "Build Docker image", status: "success", duration: "18s" },
          { name: "Push to registry", status: "success", duration: "16s" },
        ],
      },
    ],
  }

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
    <div className="space-y-6">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Link href="/" className="hover:text-foreground">
          Dashboard
        </Link>
        <span>/</span>
        <Link href={`/project/${projectId}`} className="hover:text-foreground">
          {pipeline.projectName}
        </Link>
        <span>/</span>
        <span className="text-foreground">Pipeline #{id}</span>
      </div>

      {/* Pipeline Header */}
      <div className="rounded-lg border border-border bg-card p-6">
        <div className="flex items-start justify-between">
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold">{pipeline.name}</h1>
              <span
                className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-sm font-medium text-white ${getStatusColor(pipeline.status)}`}
              >
                {getStatusIcon(pipeline.status)}
                {pipeline.status}
              </span>
            </div>
            <div className="text-base text-foreground">{pipeline.commit}</div>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <span className="flex items-center gap-1.5">
                <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <circle cx="12" cy="12" r="10" />
                  <polyline points="12 6 12 12 16 14" />
                </svg>
                {pipeline.duration}
              </span>
              <span>•</span>
              <span className="flex items-center gap-1.5">
                <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                  <circle cx="12" cy="7" r="4" />
                </svg>
                {pipeline.author}
              </span>
              <span>•</span>
              <span className="flex items-center gap-1.5">
                <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <circle cx="12" cy="12" r="10" />
                  <polyline points="12 6 12 12 16 14" />
                </svg>
                {pipeline.time}
              </span>
              <span>•</span>
              <span className="font-mono text-xs">{pipeline.commitHash}</span>
            </div>
          </div>
          <div className="flex gap-2">
            <button className="rounded-lg border border-border bg-secondary px-4 py-2 text-sm hover:bg-accent">
              Rebuild
            </button>
            <button className="rounded-lg border border-border bg-secondary px-4 py-2 text-sm hover:bg-accent">
              View Logs
            </button>
          </div>
        </div>
      </div>

      {/* Pipeline Stages */}
      <div className="space-y-4">
        {pipeline.stages.map((stage, stageIndex) => (
          <div key={stageIndex} className="rounded-lg border border-border bg-card">
            <div className="flex items-center justify-between border-b border-border p-4">
              <div className="flex items-center gap-3">
                <span
                  className={`flex size-8 items-center justify-center rounded-full text-white ${getStatusColor(stage.status)}`}
                >
                  {getStatusIcon(stage.status)}
                </span>
                <div>
                  <h3 className="font-semibold">{stage.name}</h3>
                  <p className="text-sm text-muted-foreground">{stage.duration}</p>
                </div>
              </div>
            </div>

            <div className="divide-y divide-border">
              {stage.steps.map((step, stepIndex) => (
                <div key={stepIndex} className="flex items-center justify-between p-4 hover:bg-accent/50">
                  <div className="flex items-center gap-3">
                    <span
                      className={`flex size-6 items-center justify-center rounded-full text-white ${getStatusColor(step.status)}`}
                    >
                      {getStatusIcon(step.status)}
                    </span>
                    <span className="text-sm">{step.name}</span>
                  </div>
                  <span className="text-sm text-muted-foreground">{step.duration}</span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Build Artifacts */}
      <div className="rounded-lg border border-border bg-card">
        <div className="border-b border-border p-4">
          <h3 className="font-semibold">Build Artifacts</h3>
        </div>
        <div className="divide-y divide-border">
          <div className="flex items-center justify-between p-4 hover:bg-accent/50">
            <div className="flex items-center gap-3">
              <svg
                className="size-5 text-muted-foreground"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z" />
                <polyline points="13 2 13 9 20 9" />
              </svg>
              <div>
                <div className="text-sm font-medium">build-output.zip</div>
                <div className="text-xs text-muted-foreground">2.4 MB</div>
              </div>
            </div>
            <button className="text-sm text-blue-500 hover:underline">Download</button>
          </div>
          <div className="flex items-center justify-between p-4 hover:bg-accent/50">
            <div className="flex items-center gap-3">
              <svg
                className="size-5 text-muted-foreground"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z" />
                <polyline points="13 2 13 9 20 9" />
              </svg>
              <div>
                <div className="text-sm font-medium">test-coverage.html</div>
                <div className="text-xs text-muted-foreground">128 KB</div>
              </div>
            </div>
            <button className="text-sm text-blue-500 hover:underline">Download</button>
          </div>
        </div>
      </div>
    </div>
  )
}
