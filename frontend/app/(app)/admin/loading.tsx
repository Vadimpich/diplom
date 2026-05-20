import { Skeleton } from "@/components/ui/skeleton";

export default function AdminLoading() {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between border-b border-border/70 pb-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-10 w-40" />
      </div>

      <div className="overflow-hidden rounded-2xl border border-border/70">
        <div className="grid grid-cols-[minmax(220px,1.2fr)_160px_160px_180px] gap-3 border-b border-border/70 bg-secondary/20 px-4 py-3">
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} className="h-4 w-full" />
          ))}
        </div>
        {Array.from({ length: 6 }).map((_, index) => (
          <div
            key={index}
            className="grid grid-cols-[minmax(220px,1.2fr)_160px_160px_180px] gap-3 border-b border-border/70 px-4 py-3 last:border-b-0"
          >
            {Array.from({ length: 4 }).map((_, cellIndex) => (
              <Skeleton key={cellIndex} className="h-5 w-full" />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}
