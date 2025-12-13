"use client"

export default function BuildsOverview() {
  const stats = [
    { label: "Total Builds", value: "1,234", change: "+12%", positive: true },
    { label: "Success Rate", value: "94.2%", change: "+2.1%", positive: true },
    { label: "Avg Duration", value: "4m 32s", change: "-18s", positive: true },
    { label: "Failed", value: "12", change: "-3", positive: true },
  ]

  return (
    <div className="grid gap-4 md:grid-cols-4">
      {stats.map((stat) => (
        <div key={stat.label} className="rounded-lg border border-border bg-card p-4">
          <div className="text-sm text-muted-foreground">{stat.label}</div>
          <div className="mt-2 flex items-end justify-between">
            <div className="text-2xl font-semibold">{stat.value}</div>
            <div className={`text-sm ${stat.positive ? "text-green-500" : "text-red-500"}`}>{stat.change}</div>
          </div>
        </div>
      ))}
    </div>
  )
}
