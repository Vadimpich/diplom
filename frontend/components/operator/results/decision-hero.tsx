import Link from "next/link";
import { ShieldAlert, ShieldCheck, ShieldQuestion, ShieldX } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import type { ExaminationResult } from "@/lib/api/types";
import {
  formatDecisionCode,
  getDecisionHeroCopy,
  getPrimaryAction,
  getToneClasses,
} from "./helpers";

function resolveIcon(result: ExaminationResult) {
  const decisionCode = result.decision.decision_code;
  const riskClass = result.decision.risk_class;

  if (decisionCode === "no_access" || riskClass === "critical") {
    return ShieldX;
  }
  if (decisionCode === "extended_check" || riskClass === "high") {
    return ShieldAlert;
  }
  if (decisionCode === "allow" || riskClass === "low") {
    return ShieldCheck;
  }
  return ShieldQuestion;
}

export function DecisionHero({
  result,
  specialistHref,
}: {
  result: ExaminationResult;
  specialistHref: string;
}) {
  const hero = getDecisionHeroCopy(result);
  const tone = getToneClasses(hero.tone);
  const action = getPrimaryAction(result);
  const Icon = resolveIcon(result);

  return (
    <Card className={`${tone.border} ${tone.background} overflow-hidden`}>
      <CardContent className="p-0">
        <div className="grid gap-5 p-5 lg:grid-cols-[1.35fr_0.65fr] lg:p-6">
          <div className="space-y-4">
            <div className="flex items-start gap-4">
              <div className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl ${tone.panel}`}>
                <Icon className="h-6 w-6" />
              </div>
              <div className="space-y-2">
                <Badge variant={tone.badge}>{hero.badge}</Badge>
                <h2 className="text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">{hero.title}</h2>
                <p className="max-w-3xl text-sm leading-6 text-muted-foreground">{hero.summary}</p>
              </div>
            </div>

            <div className="flex flex-wrap gap-2">
              {hero.symptoms.length ? (
                hero.symptoms.map((symptom) => (
                  <Badge key={symptom} variant="neutral" className="bg-background/85">
                    {symptom}
                  </Badge>
                ))
              ) : (
                <Badge variant="neutral" className="bg-background/85">
                  Явные симптомы не выделены
                </Badge>
              )}
            </div>
          </div>

          <div className="rounded-3xl border border-border/70 bg-background/82 p-5">
            <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Принимаемое решение</p>
            <p className="mt-2 text-lg font-semibold text-foreground">{formatDecisionCode(result)}</p>
            <Button asChild variant={tone.button} className="mt-4 w-full" size="lg">
              <Link href={specialistHref}>{action.label}</Link>
            </Button>
            <p className="mt-3 text-xs leading-5 text-muted-foreground">{action.description}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
