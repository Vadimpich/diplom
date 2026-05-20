import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { ExaminationResult } from "@/lib/api/types";
import { channelLabels, getScoreVisual } from "./helpers";

const positiveScoreKeys = new Set([
  "text_confidence_score",
  "text_coherence_score",
  "voice_stability_score",
]);

export function DetailedReportModal({
  open,
  onClose,
  result,
}: {
  open: boolean;
  onClose: () => void;
  result: ExaminationResult | undefined;
}) {
  if (!open || !result) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-primary/30 p-4 backdrop-blur-sm">
      <div className="absolute inset-0" onClick={onClose} aria-hidden="true" />
      <div className="relative z-10 max-h-[90vh] w-full max-w-6xl overflow-hidden rounded-[28px] border border-border bg-background shadow-soft">
        <div className="flex items-start justify-between gap-4 border-b border-border/70 px-6 py-5">
          <div>
            <h2 className="text-2xl font-semibold">Подробный отчёт</h2>
            <p className="mt-2 max-w-3xl text-sm text-muted-foreground">
              Здесь собраны сырые score-метрики аналитических каналов. Они нужны для детального разбора, а не для быстрого операторского решения.
            </p>
          </div>
          <Button variant="ghost" onClick={onClose}>
            Закрыть
          </Button>
        </div>

        <div className="max-h-[calc(90vh-96px)] overflow-y-auto px-6 py-6">
          <div className="space-y-6">
            {result.channel_reports.map((report) => {
              const qualityFlags = report.quality_flags ?? [];
              const evidenceItems = report.evidence ?? [];

              return (
                <Card key={report.channel}>
                  <CardHeader>
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <CardTitle>{channelLabels[report.channel]}</CardTitle>
                        <CardDescription>Модель: {report.model_version}</CardDescription>
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {qualityFlags.length ? (
                          qualityFlags.map((flag) => (
                            <Badge key={flag} variant="warning">
                              {flag}
                            </Badge>
                          ))
                        ) : (
                          <Badge variant="success">Качество без критичных флагов</Badge>
                        )}
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-5">
                    <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                      {report.scores.map((score) => {
                        const visual = getScoreVisual(score);
                        return (
                          <div key={score.key} className="rounded-2xl border border-border/70 p-4">
                            <div className="flex items-start justify-between gap-4">
                              <div>
                                <p className="font-medium">{score.label}</p>
                                <p className="mt-1 text-xs text-muted-foreground">{score.key}</p>
                              </div>
                              <Badge variant={visual.badge}>{score.value.toFixed(3)}</Badge>
                            </div>
                            <div className={`mt-4 h-2 rounded-full ${visual.trackClassName}`}>
                              <div
                                className={`h-2 rounded-full ${visual.barClassName}`}
                                style={{ width: `${Math.max(6, Math.min(100, score.value * 100))}%` }}
                              />
                            </div>
                            <p className={`mt-3 text-sm ${visual.valueClassName}`}>
                              {positiveScoreKeys.has(score.key)
                                ? "Для этого показателя более высокое значение обычно трактуется как более благоприятное."
                                : "Для этого показателя более высокое значение обычно трактуется как более тревожное."}
                            </p>
                          </div>
                        );
                      })}
                    </div>

                    {evidenceItems.length ? (
                      <div className="rounded-2xl border border-border/70 bg-secondary/30 p-4">
                        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Evidence</p>
                        <div className="mt-3 space-y-2 text-sm text-muted-foreground">
                          {evidenceItems.map((item, index) => (
                            <p key={`${report.channel}-evidence-${index}`}>{item}</p>
                          ))}
                        </div>
                      </div>
                    ) : null}
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
