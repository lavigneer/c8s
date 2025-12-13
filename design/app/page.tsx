import DashboardHeader from "@/components/dashboard-header"
import BuildsOverview from "@/components/builds-overview"
import PipelinesList from "@/components/pipelines-list"
import RecentActivity from "@/components/recent-activity"

export default function DashboardPage() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <DashboardHeader />

      <main className="container mx-auto space-y-6 p-6">
        <BuildsOverview />

        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <PipelinesList />
          </div>

          <div>
            <RecentActivity />
          </div>
        </div>
      </main>
    </div>
  )
}
