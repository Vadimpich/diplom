import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { Answer, Examination, Specialist } from "@/lib/api/types";
import { formatDateTime } from "@/lib/utils";
import { ExaminationStatusBadge } from "@/components/operator/status-badge";

export function ExaminationSummary({
  examination,
  specialist,
  answers,
}: {
  examination: Examination;
  specialist?: Specialist;
  answers: Answer[];
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Сводка обследования</CardTitle>
        <CardDescription>Текущий backend-статус обследования и уже сохранённые ответы.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="grid gap-4 md:grid-cols-3">
          <div>
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Статус</p>
            <div className="mt-2">
              <ExaminationStatusBadge status={examination.status} />
            </div>
          </div>
          <div>
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Специалист</p>
            <p className="mt-2 text-sm font-medium">{specialist?.full_name ?? `ID ${examination.specialist_id}`}</p>
          </div>
          <div>
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Создано</p>
            <p className="mt-2 text-sm font-medium">{formatDateTime(examination.created_at)}</p>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Начато</p>
            <p className="mt-2 text-sm font-medium">
              {examination.started_at ? formatDateTime(examination.started_at) : "Ещё не начато"}
            </p>
          </div>
          <div>
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Завершено</p>
            <p className="mt-2 text-sm font-medium">
              {examination.finished_at ? formatDateTime(examination.finished_at) : "Ещё не завершено"}
            </p>
          </div>
        </div>

        <div className="space-y-3">
          <p className="text-sm font-medium">Ответы сеанса</p>
          {answers.length === 0 ? (
            <p className="text-sm text-muted-foreground">Ответы ещё не сохранены.</p>
          ) : (
            <div className="space-y-3">
              {answers.map((answer, index) => (
                <div key={answer.id} className="rounded-2xl border border-border/70 bg-secondary/40 p-4">
                  <div className="flex items-center justify-between gap-4">
                    <p className="text-sm font-medium">Ответ {index + 1}</p>
                    <p className="text-xs text-muted-foreground">{formatDateTime(answer.created_at)}</p>
                  </div>
                  <p className="mt-2 text-sm text-foreground/80">{answer.text}</p>
                  <p className="mt-2 font-mono text-xs text-muted-foreground">{answer.audio_s3_key}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
