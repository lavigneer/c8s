"use client"

import Link from "next/link"

export default function ProjectsList() {
  const projects = [
    {
      id: "web-app",
      name: "web-application",
      description: "Main web application for customer portal",
      repository: "github.com/acme/web-application",
      lastBuild: "2 minutes ago",
      status: "success",
      activeBuilds: 0,
    },
    {
      id: "api-service",
      name: "api-service",
      description: "Backend API service",
      repository: "github.com/acme/api-service",
      lastBuild: "5 minutes ago",
      status: "running",
      activeBuilds: 1,
    },
    {
      id: "mobile-app",
      name: "mobile-app",
      description: "iOS and Android mobile applications",
      repository: "github.com/acme/mobile-app",
      lastBuild: "23 minutes ago",
      status: "success",
      activeBuilds: 0,
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

  return (
    <div className="rounded-lg border border-border bg-card">
      <div className="border-b border-border p-4">
        <h2 className="text-lg font-semibold">Projects</h2>
      </div>
      <div className="grid gap-4 p-4 md:grid-cols-2 lg:grid-cols-3">
        {projects.map((project) => (
          <Link
            key={project.id}
            href={`/project/${project.id}`}
            className="group rounded-lg border border-border bg-secondary p-4 hover:border-primary/50 hover:bg-accent"
          >
            <div className="space-y-3">
              <div className="flex items-start justify-between gap-2">
                <div className="flex size-10 items-center justify-center rounded-lg bg-primary/10 group-hover:bg-primary/20">
                  <svg
                    className="size-5 text-primary"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                  >
                    <path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4" />
                    <path d="M9 18c-4.51 2-5-2-7-2" />
                  </svg>
                </div>
                <div className={`size-2 rounded-full ${getStatusColor(project.status)}`} />
              </div>
              <div>
                <h3 className="font-semibold group-hover:text-primary">{project.name}</h3>
                <p className="text-xs text-muted-foreground">{project.description}</p>
              </div>
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>Last build: {project.lastBuild}</span>
                {project.activeBuilds > 0 && (
                  <span className="rounded bg-blue-500/20 px-1.5 py-0.5 text-blue-500">
                    {project.activeBuilds} active
                  </span>
                )}
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}
