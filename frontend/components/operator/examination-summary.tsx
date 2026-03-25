import { ExaminationStatusBadge } from "@/components/operator/status-badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { Answer, Examination, Questionnaire, Specialist } from "@/lib/api/types";
import { formatDateTime } from "@/lib/utils";

export function ExaminationSummary({
  examination,
  specialist,
  questionnaire,
  answers,
}: {
  examination: Examination;
  specialist?: Specialist;
  questionnaire?: Questionnaire | null;
  answers: Answer[];
}) {
  const totalQuestions = questionnaire?.questions.length ?? 0;
  const savedAnswers = answers.length;
  const remainingAnswers = totalQuestions > 0 ? Math.max(totalQuestions - savedAnswers, 0) : null;
  const nextQuestion = questionnaire?.questions[savedAnswers] ?? null;
  const completionPercent =
    totalQuestions > 0 ? Math.min(Math.round((savedAnswers / totalQuestions) * 100), 100) : 0;

  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle>Ход обследования</CardTitle>
        <CardDescription>Текущий статус, прогресс по вопросам и уже сохранённые ответы сеанса.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Статус</p>
            <div className="mt-3">
              <ExaminationStatusBadge status={examination.status} />
            </div>
            <p className="mt-3 text-sm text-muted-foreground">
              {specialist?.full_name ?? `Специалист #${examination.specialist_id}`}
            </p>
          </div>
          <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Прогресс</p>
            <p className="mt-3 text-2xl font-semibold">
              {savedAnswers}
              {totalQuestions > 0 ? ` / ${totalQuestions}` : ""}
            </p>
            <p className="mt-2 text-sm text-muted-foreground">
              {totalQuestions > 0
                ? remainingAnswers === 0
                  ? "Все вопросы уже сохранены."
                  : `Осталось ответов: ${remainingAnswers}.`
                : "Работа идёт без привязанного опросника."}
            </p>
          </div>
        </div>

        {totalQuestions > 0 ? (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-xs uppercase tracking-[0.18em] text-muted-foreground">
              <span>Готовность сессии</span>
              <span>{completionPercent}%</span>
            </div>
            <div className="h-2.5 overflow-hidden rounded-full bg-secondary">
              <div className="h-full rounded-full bg-primary transition-all" style={{ width: `${completionPercent}%` }} />
            </div>
            <p className="text-sm text-muted-foreground">
              {nextQuestion
                ? `Следом идёт вопрос ${nextQuestion.position}.`
                : "После проверки ответов можно завершать сбор и переходить к обработке."}
            </p>
          </div>
        ) : null}

        <div className="grid gap-3 sm:grid-cols-2">
          <div className="rounded-2xl border border-border/70 p-4">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Начато</p>
            <p className="mt-2 text-sm font-medium">
              {examination.started_at ? formatDateTime(examination.started_at) : "Ещё не начато"}
            </p>
          </div>
          <div className="rounded-2xl border border-border/70 p-4">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Завершено</p>
            <p className="mt-2 text-sm font-medium">
              {examination.finished_at ? formatDateTime(examination.finished_at) : "Сбор ещё идёт"}
            </p>
          </div>
        </div>

        <div className="space-y-3">
          <div className="flex items-center justify-between gap-3">
            <p className="text-sm font-medium">Сохранённые ответы</p>
            <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">{savedAnswers} записей</p>
          </div>
          {answers.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-border/70 bg-secondary/20 p-4 text-sm text-muted-foreground">
              Пока нет сохранённых ответов. После записи каждый ответ сразу появится в этом списке.
            </div>
          ) : (
            <div className="space-y-3">
              {answers.map((answer, index) => {
                const question = questionnaire?.questions[index];

                return (
                  <div key={answer.id} className="rounded-2xl border border-border/70 p-4">
                    <div className="flex items-start justify-between gap-4">
                      <div>
                        <p className="text-sm font-medium">
                          {question ? `Вопрос ${question.position}` : `Ответ ${index + 1}`}
                        </p>
                        {question ? (
                          <p className="mt-1 text-xs text-muted-foreground line-clamp-2">{question.text}</p>
                        ) : null}
                      </div>
                      <p className="text-xs text-muted-foreground">{formatDateTime(answer.created_at)}</p>
                    </div>
                    <p className="mt-3 text-sm text-foreground/80">{answer.text || "Текст ответа не указан."}</p>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
