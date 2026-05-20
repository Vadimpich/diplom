import { AudioLines, FileText, Mic2, TrendingUp } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { ExaminationResult } from "@/lib/api/types";
import { getChannelImpactItems, getImpactLabel } from "./helpers";

function iconForChannel(key: string) {
  switch (key) {
    case "text":
      return FileText;
    case "acoustic":
      return AudioLines;
    case "paralinguistic":
      return Mic2;
    default:
      return TrendingUp;
  }
}

function progressClasses(emphasis: "low" | "medium" | "high") {
  switch (emphasis) {
    case "high":
      return {
        track: "bg-danger/10",
        bar: "bg-danger",
      };
    case "medium":
      return {
        track: "bg-warning/12",
        bar: "bg-warning",
      };
    default:
      return {
        track: "bg-success/12",
        bar: "bg-success",
      };
  }
}

export function ChannelContribution({ result }: { result: ExaminationResult }) {
  const items = getChannelImpactItems(result);

  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle>Сигналы по каналам анализа</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-4 lg:grid-cols-2">
        {items.map((item) => {
          const Icon = iconForChannel(item.key);
          const visual = progressClasses(item.emphasis);
          return (
            <div key={item.key} className="rounded-3xl border border-border/70 p-5">
              <div className="flex items-start justify-between gap-4">
                <div className="flex items-start gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-secondary">
                    <Icon className="h-5 w-5 text-foreground" />
                  </div>
                  <div>
                    <p className="text-base font-semibold text-foreground">{item.title}</p>
                    <p className="mt-1 text-sm font-medium text-foreground">{getImpactLabel(item.emphasis)}</p>
                    {item.summary ? <p className="mt-1 text-sm text-muted-foreground">{item.summary}</p> : null}
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-xs text-muted-foreground">{item.valueText}</p>
                </div>
              </div>

              <div className={`mt-4 h-2.5 rounded-full ${visual.track}`}>
                <div
                  className={`h-2.5 rounded-full ${visual.bar}`}
                  style={{ width: `${Math.max(10, Math.min(100, item.progress))}%` }}
                />
              </div>
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}
