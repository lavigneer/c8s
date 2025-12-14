"use client"

import Link from "next/link"
import { useState } from "react"

interface ProjectDetailProps {
  id: string
}

export default function ProjectDetail({ id }: ProjectDetailProps) {
  const [filter, setFilter] = useState<string>("all")

  // Mock project data
  const project = {
    id,
    name: "web-application",
    repository: "github.com/acme/web-application",
    description: "Main web application for customer portal",
    defaultBranch: "main",
    lastUpdated: "2 minutes ago",
  }

  const pipelines = [
    {
      id: "1",
      projectId: id,
      branch: "main",
      commit: "a3f82e1",
      message: "feat: add user authentication",
      author: "john.doe",
      status: "success",
      duration: "4m 12s",
      time: "2 minutes ago",
      triggeredBy: "Push",
    },
    {
      id: "2",
      projectId: id,
      branch: "develop",
      commit: "b9c14f3",
      message: "fix: resolve database connection issue",
      author: "jane.smith",
      status: "running",
      duration: "2m 45s",
      time: "5 minutes ago",
      triggeredBy: "Push",
    },
    {
      id: "3",
      projectId: id,
      branch: "feature/new-ui",
      commit: "c7d2a89",
      message: "chore: update dependencies",
      author: "bob.wilson",
      status: "failed",
      duration: "3m 28s",
      time: "10 minutes ago",
      triggeredBy: "Pull Request",
    },
    {
      id: "4",
      projectId: id,
      branch: "main",
      commit: "d1e5b42",
      message: "docs: update README with setup instructions",
      author: "alice.chen",
      status: "success",
      duration: "3m 55s",
      time: "23 minutes ago",
      triggeredBy: "Push",
    },
    {
      id: "5",
      projectId: id,
      branch: "hotfix/security",
      commit: "e8f3c91",
      message: "security: patch vulnerability in auth module",
      author: "mike.jones",
      status: "success",
      duration: "5m 02s",
      time: "1 hour ago",
      triggeredBy: "Manual",
    },
  ]

  const filteredPipelines = pipelines.filter((pipeline) => {
    if (filter === "all") return true
    return pipeline.status === filter
  })

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
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Link href="/" className="hover:text-foreground">
          Dashboard
        </Link>
        <span>/</span>
        <span className="text-foreground">Projects</span>
        <span>/</span>
        <span className="text-foreground">{project.name}</span>
      </div>

      {/* Project Header */}
      <div className="rounded-lg border border-border bg-card p-6">
        <div className="flex items-start justify-between">
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <div className="flex size-12 items-center justify-center rounded-lg bg-primary/10">
                <svg
                  className="size-6 text-primary"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4" />
                  <path d="M9 18c-4.51 2-5-2-7-2" />
                </svg>
              </div>
              <div>
                <h1 className="text-2xl font-bold">{project.name}</h1>
                <p className="text-sm text-muted-foreground">{project.repository}</p>
              </div>
            </div>
            <p className="text-sm text-foreground">{project.description}</p>
            <div className="flex items-center gap-4 text-xs text-muted-foreground">
              <span className="flex items-center gap-1">
                <svg className="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
                </svg>
                Default: {project.defaultBranch}
              </span>
              <span>•</span>
              <span>Last updated: {project.lastUpdated}</span>
            </div>
          </div>
          <button className="rounded-lg border border-border bg-secondary px-4 py-2 text-sm hover:bg-accent">
            Settings
          </button>
        </div>
      </div>

      {/* Pipeline Runs */}
      <div className="rounded-lg border border-border bg-card">
        <div className="flex items-center justify-between border-b border-border p-4">
          <h2 className="text-lg font-semibold">Pipeline Runs</h2>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setFilter("all")}
              className={`rounded-lg px-3 py-1 text-sm ${
                filter === "all" ? "bg-primary text-primary-foreground" : "bg-secondary hover:bg-accent"
              }`}
            >
              All
            </button>
            <button
              onClick={() => setFilter("success")}
              className={`rounded-lg px-3 py-1 text-sm ${
                filter === "success" ? "bg-primary text-primary-foreground" : "bg-secondary hover:bg-accent"
              }`}
            >
              Success
            </button>
            <button
              onClick={() => setFilter("failed")}
              className={`rounded-lg px-3 py-1 text-sm ${
                filter === "failed" ? "bg-primary text-primary-foreground" : "bg-secondary hover:bg-accent"
              }`}
            >
              Failed
            </button>
            <button
              onClick={() => setFilter("running")}
              className={`rounded-lg px-3 py-1 text-sm ${
                filter === "running" ? "bg-primary text-primary-foreground" : "bg-secondary hover:bg-accent"
              }`}
            >
              Running
            </button>
          </div>
        </div>
        <div className="divide-y divide-border">
          {filteredPipelines.map((pipeline) => (
            <div key={pipeline.id} className="p-4 hover:bg-accent/50">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 space-y-2">
                  <div className="flex items-center gap-3">
                    <span
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs text-white ${getStatusColor(pipeline.status)}`}
                    >
                      {getStatusIcon(pipeline.status)}
                      {pipeline.status}
                    </span>
                    <span className="font-mono text-sm font-semibold">{pipeline.branch}</span>
                    <span className="rounded bg-secondary px-2 py-0.5 font-mono text-xs text-muted-foreground">
                      {pipeline.commit}
                    </span>
                  </div>
                  <div className="text-sm text-foreground">{pipeline.message}</div>
                  <div className="flex items-center gap-4 text-xs text-muted-foreground">
                    <span className="flex items-center gap-1">
                      <svg className="size-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
                        <circle cx="12" cy="7" r="4" />
                      </svg>
                      {pipeline.author}
                    </span>
                    <span>•</span>
                    <span className="flex items-center gap-1">
                      <svg className="size-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <circle cx="12" cy="12" r="10" />
                        <polyline points="12 6 12 12 16 14" />
                      </svg>
                      {pipeline.duration}
                    </span>
                    <span>•</span>
                    <span>{pipeline.time}</span>
                    <span>•</span>
                    <span className="flex items-center gap-1">
                      <svg className="size-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <circle cx="12" cy="12" r="10" />
                        <polyline points="8 12 12 16 16 12" />
                        <line x1="12" y1="8" x2="12" y2="16" />
                      </svg>
                      {pipeline.triggeredBy}
                    </span>
                  </div>
                </div>
                <Link
                  href={`/project/${pipeline.projectId}/pipeline/${pipeline.id}`}
                  className="rounded-lg border border-border bg-secondary px-3 py-1 text-sm hover:bg-accent"
                >
                  View Details
                </Link>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
