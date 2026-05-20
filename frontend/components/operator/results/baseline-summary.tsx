import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { ExaminationResult } from "@/lib/api/types";
import { Activity } from "lucide-react";
import { formatDelta, getBaselineDescription, getBaselineHeadline } from "./helpers";

export function BaselineSummary({ result }: { result: ExaminationResult }) {
  const personalAvailable = !!result.baseline_snapshot.personal.baseline_available;
  const baselineCount = result.baseline_snapshot.personal.baseline_exam_count ?? 0;
  const hasLimitedHistory = !personalAvailable;
  const description = getBaselineDescription(result);

  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle>Baseline</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-start gap-3 rounded-2xl border border-border/70 bg-secondary/35 p-4">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-background">
            <Activity className="h-5 w-5" />
          </div>
          <div className="space-y-1">
            <p className="text-base font-semibold text-foreground">{getBaselineHeadline(result)}</p>
          </div>
        </div>

        <div className="grid gap-3 sm:grid-cols-3">
          <div className="rounded-2xl border border-border/70 p-4">
            <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Общее отклонение</p>
            <p className="mt-2 text-2xl font-semibold text-foreground">
              {formatDelta(result.baseline_snapshot.general.delta)}
            </p>
          </div>
          <div className="rounded-2xl border border-border/70 p-4">
            <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
              {personalAvailable ? "Личное отклонение" : "Личная норма"}
            </p>
            <p className="mt-2 text-2xl font-semibold text-foreground">
              {personalAvailable ? formatDelta(result.baseline_snapshot.personal.delta) : "Нет"}
            </p>
          </div>
          <div className="rounded-2xl border border-border/70 p-4">
            <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">История</p>
            <p className="mt-2 text-2xl font-semibold text-foreground">
              {baselineCount}
            </p>
            <p className="mt-1 text-sm text-muted-foreground">обследования</p>
          </div>
        </div>

        {hasLimitedHistory && description ? (
          <p className="text-sm leading-5 text-muted-foreground">{description}</p>
        ) : null}
      </CardContent>
    </Card>
  );
}
