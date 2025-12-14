import DashboardHeader from "@/components/dashboard-header"
import PipelineDetail from "@/components/pipeline-detail"

export default function PipelinePage({ params }: { params: { projectId: string; id: string } }) {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <DashboardHeader />
      <main className="container mx-auto p-6">
        <PipelineDetail projectId={params.projectId} id={params.id} />
      </main>
    </div>
  )
}
