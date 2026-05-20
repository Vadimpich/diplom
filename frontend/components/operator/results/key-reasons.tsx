import { AlertTriangle, BrainCircuit, Signal, Waves } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { ExaminationResult } from "@/lib/api/types";
import { getKeyReasons, getReasonAccentClasses } from "./helpers";

function iconForReason(key: string) {
  if (key.includes("baseline")) {
    return Signal;
  }
  if (key.includes("acoustic")) {
    return Waves;
  }
  if (key.includes("paralinguistic")) {
    return BrainCircuit;
  }
  return AlertTriangle;
}

const levelLabels = {
  low: "Низкая",
  medium: "Средняя",
  high: "Высокая",
};

export function KeyReasons({ result }: { result: ExaminationResult }) {
  const reasons = getKeyReasons(result);

  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle>Причины решения</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {reasons.map((reason) => {
          const Icon = iconForReason(reason.key);
          return (
            <div key={reason.key} className={`rounded-2xl border p-4 ${getReasonAccentClasses(reason.level)}`}>
              <div className="flex items-start gap-3">
                <div className="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-2xl bg-background/80">
                  <Icon className="h-5 w-5" />
                </div>
                <div className="min-w-0 space-y-2">
                  <p className="text-sm font-semibold text-foreground">{reason.title}</p>
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge variant="neutral" className="bg-background/80 text-[11px]">
                      {levelLabels[reason.level]} выраженность
                    </Badge>
                    <span className="text-xs text-muted-foreground">{reason.source}</span>
                  </div>
                  <p className="text-sm leading-5 text-muted-foreground">{reason.description}</p>
                </div>
              </div>
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}
