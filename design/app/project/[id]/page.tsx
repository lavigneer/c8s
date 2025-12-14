import DashboardHeader from "@/components/dashboard-header"
import ProjectDetail from "@/components/project-detail"

export default function ProjectPage({ params }: { params: { id: string } }) {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <DashboardHeader />
      <main className="container mx-auto p-6">
        <ProjectDetail id={params.id} />
      </main>
    </div>
  )
}
