import Link from "next/link";
import { ClipboardCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { ExaminationResult } from "@/lib/api/types";
import { getOperatorSteps, getPrimaryAction } from "./helpers";

export function OperatorActionCard({
  result,
  specialistHref,
}: {
  result: ExaminationResult;
  specialistHref: string;
}) {
  const action = getPrimaryAction(result);
  const steps = getOperatorSteps(result);

  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle>Что сделать дальше</CardTitle>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="flex items-start gap-3 rounded-3xl border border-border/70 bg-secondary/35 p-5">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-background">
            <ClipboardCheck className="h-5 w-5" />
          </div>
          <p className="text-sm leading-6 text-muted-foreground">
            Используйте итог как опору для решения, но принимайте его вместе с контекстом специалиста и историей обследований.
          </p>
        </div>

        <ol className="space-y-3">
          {steps.map((step, index) => (
            <li key={step} className="flex items-start gap-3 rounded-3xl border border-border/70 p-4">
              <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-secondary text-sm font-semibold text-foreground">
                {index + 1}
              </div>
              <p className="text-sm leading-6 text-foreground">{step}</p>
            </li>
          ))}
        </ol>

        <Button asChild className="w-full sm:w-auto">
          <Link href={specialistHref}>{action.label}</Link>
        </Button>
      </CardContent>
    </Card>
  );
}
