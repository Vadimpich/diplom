import type {
  ExaminationProcessingStatus,
  ExaminationProcessingStatusChannel,
  ExaminationResult,
  ProcessingChannelName,
  ProcessingChannelStatus,
  ResultChannelReportScore,
} from "@/lib/api/types";

export const channelLabels: Record<ProcessingChannelName, string> = {
  text: "Текстовый анализ",
  acoustic: "Акустический анализ",
  paralinguistic: "Паралингвистический анализ",
};

export const decisionCodeLabels: Record<string, string> = {
  allow: "Допуск",
  monitoring: "Допуск с наблюдением",
  extended_check: "Повторная углублённая проверка",
  no_access: "Недопуск",
};

export const riskClassLabels: Record<string, string> = {
  low: "Низкий риск",
  attention: "Зона внимания",
  medium: "Средний риск",
  high: "Высокий риск",
  critical: "Критический риск",
};

export const patternLabels: Record<string, string> = {
  hidden_stress: "Скрытый стресс",
  open_stress: "Открытый стресс",
  cognitive_overload: "Когнитивная перегрузка",
  fatigue_state: "Состояние утомления",
  depressive_pattern: "Депрессивный паттерн",
  emotional_cross: "Эмоциональное противоречие",
  contradictory_profile: "Противоречивый профиль",
  compensation_breakdown: "Срыв компенсации",
  baseline_sensitive_case: "Чувствительный baseline-кейс",
  masked_distress: "Маскируемый дистресс",
  atypical_response: "Атипичный ответ",
};

const positiveScoreKeys = new Set([
  "text_confidence_score",
  "text_coherence_score",
  "voice_stability_score",
]);

type RiskTone = "low" | "medium" | "high" | "neutral";
type Emphasis = "low" | "medium" | "high";

type PatternReason = {
  title: string;
  description: string;
  level: Emphasis;
  source: string;
};

export type KeyReason = PatternReason & {
  key: string;
};

export type ChannelImpactItem = {
  key: string;
  title: string;
  summary?: string;
  valueText: string;
  progress: number;
  emphasis: Emphasis;
};

export const channelStatusConfig: Record<
  ProcessingChannelStatus,
  { label: string; variant: "neutral" | "warning" | "info" | "success" | "danger" }
> = {
  queued: { label: "В очереди", variant: "neutral" },
  processing: { label: "Обрабатывается", variant: "info" },
  succeeded: { label: "Готово", variant: "success" },
  retry_scheduled: { label: "Назначен повтор", variant: "warning" },
  failed_temporary: { label: "Временная ошибка", variant: "warning" },
  temporary_error: { label: "Временная ошибка", variant: "warning" },
  failed_fatal: { label: "Критическая ошибка", variant: "danger" },
  fatal_error: { label: "Критическая ошибка", variant: "danger" },
  exhausted: { label: "Попытки исчерпаны", variant: "danger" },
};

const patternReasonCatalog: Record<string, PatternReason> = {
  cognitive_overload: {
    title: "Когнитивная перегрузка",
    description: "Паузы и речевая дезорганизация.",
    level: "high",
    source: "Паралингвистика",
  },
  fatigue_state: {
    title: "Состояние утомления",
    description: "Есть заметное отклонение от нормы.",
    level: "high",
    source: "Baseline",
  },
  contradictory_profile: {
    title: "Противоречивый профиль",
    description: "Уверенный текст расходится с голосом и поведением.",
    level: "medium",
    source: "Сводный профиль",
  },
  hidden_stress: {
    title: "Скрытый стресс",
    description: "Спокойный ответ сочетается с напряжённым голосом.",
    level: "medium",
    source: "Голос",
  },
  open_stress: {
    title: "Открытый стресс",
    description: "Несколько каналов одновременно показывают напряжение.",
    level: "high",
    source: "Сводный профиль",
  },
  depressive_pattern: {
    title: "Сниженный эмоциональный тон",
    description: "Снижение энергичности и связности ответа.",
    level: "medium",
    source: "Содержание ответа",
  },
  emotional_cross: {
    title: "Эмоциональное противоречие",
    description: "Сигналы разных каналов расходятся между собой.",
    level: "medium",
    source: "Сводный профиль",
  },
  compensation_breakdown: {
    title: "Срыв компенсации",
    description: "Привычный самоконтроль выражен слабее обычного.",
    level: "high",
    source: "Сводный профиль",
  },
  baseline_sensitive_case: {
    title: "Отклонение от baseline",
    description: "Текущее состояние отличается от обычного профиля.",
    level: "high",
    source: "Baseline",
  },
  masked_distress: {
    title: "Маскируемый дистресс",
    description: "Прямых жалоб мало, но косвенные сигналы напряжения есть.",
    level: "medium",
    source: "Сводный профиль",
  },
  atypical_response: {
    title: "Атипичный ответ",
    description: "Профиль ответа выбивается из ожидаемой нормы.",
    level: "medium",
    source: "Сводный профиль",
  },
};

function getRiskTone(result: ExaminationResult): RiskTone {
  const riskClass = result.decision.risk_class;
  const decisionCode = result.decision.decision_code;

  if (riskClass === "critical" || riskClass === "high" || decisionCode === "no_access" || decisionCode === "extended_check") {
    return "high";
  }

  if (riskClass === "medium" || riskClass === "attention" || decisionCode === "monitoring") {
    return "medium";
  }

  if (riskClass === "low" || decisionCode === "allow" || result.decision.recommendation === "allowed") {
    return "low";
  }

  return "neutral";
}

function getBandLabel(band?: string | null) {
  switch (band) {
    case "low":
      return "низкое";
    case "mild":
      return "умеренное";
    case "medium":
      return "среднее";
    case "high":
      return "высокое";
    case "critical":
      return "критическое";
    default:
      return "не определено";
  }
}

function toEmphasisByContribution(value: number): Emphasis {
  if (value >= 0.18) {
    return "high";
  }
  if (value >= 0.07) {
    return "medium";
  }
  return "low";
}

function toEmphasisByBaselineBand(band?: string | null): Emphasis {
  if (band === "high" || band === "critical") {
    return "high";
  }
  if (band === "medium" || band === "mild") {
    return "medium";
  }
  return "low";
}

function getBaselineProgress(result: ExaminationResult) {
  const delta = Math.max(
    Math.abs(result.baseline_snapshot.personal.delta),
    Math.abs(result.baseline_snapshot.general.delta),
  );
  return Math.max(12, Math.min(100, Math.round((delta / 4) * 100)));
}

export function formatDecisionCode(result: ExaminationResult) {
  return decisionCodeLabels[result.decision.decision_code ?? ""] ?? "Требуется решение оператора";
}

export function formatRiskClass(result: ExaminationResult) {
  return riskClassLabels[result.decision.risk_class ?? ""] ?? "Требует оценки";
}

export function formatPatterns(result: ExaminationResult) {
  return (result.decision.patterns ?? []).map((pattern) => ({
    key: pattern,
    label: patternLabels[pattern] ?? pattern.replaceAll("_", " "),
  }));
}

export function formatDelta(delta: number) {
  const prefix = delta > 0 ? "+" : "";
  return `${prefix}${delta.toFixed(3)}`;
}

export function getResultStatusText(data?: ExaminationProcessingStatus) {
  if (!data) {
    return "статус обработки уточняется";
  }
  if (data.terminal && data.status === "completed") {
    return "обработка завершена";
  }
  if (data.status === "failed") {
    return "обработка завершилась с ошибкой";
  }
  return "идёт обработка";
}

export function getPrimaryAction(result: ExaminationResult) {
  switch (result.decision.decision_code) {
    case "allow":
      return {
        label: "Завершить",
        description: "Решение можно зафиксировать в карточке специалиста.",
      };
    case "monitoring":
      return {
        label: "Перейти к наблюдению",
        description: "Нужно сохранить результат и наблюдать состояние в динамике.",
      };
    case "extended_check":
      return {
        label: "Перейти к повторной проверке",
        description: "Нужна дополнительная оценка оператора.",
      };
    case "no_access":
      return {
        label: "Оформить ограничение допуска",
        description: "Требуется зафиксировать ограничение допуска.",
      };
    default:
      return {
        label: "Вернуться к специалисту",
        description: "Продолжите работу со специалистом в его карточке.",
      };
  }
}

export function getDecisionHeroCopy(result: ExaminationResult) {
  const tone = getRiskTone(result);
  const patterns = formatPatterns(result);
  const symptomLabels = patterns.map((item) => item.label);

  if (tone === "high") {
    return {
      tone,
      badge: formatRiskClass(result),
      title: formatDecisionCode(result),
      summary: "Обнаружены признаки повышенного риска. Требуется повторная углублённая проверка.",
      symptoms: symptomLabels,
    };
  }

  if (tone === "medium") {
    return {
      tone,
      badge: formatRiskClass(result),
      title: formatDecisionCode(result),
      summary: "Есть сигналы, требующие внимания. Решение стоит принимать с учётом контекста и истории обследований.",
      symptoms: symptomLabels,
    };
  }

  if (tone === "low") {
    return {
      tone,
      badge: formatRiskClass(result),
      title: formatDecisionCode(result),
      summary: "Критичных признаков не выявлено. Результат не указывает на выраженное отклонение по основным каналам.",
      symptoms: symptomLabels,
    };
  }

  return {
    tone,
    badge: "Нужно решение оператора",
    title: "Итог требует интерпретации",
    summary: "Автоматическая рекомендация недоступна. Ориентируйтесь на baseline и вклад каналов ниже.",
    symptoms: symptomLabels,
  };
}

export function getReasonAccentClasses(level: Emphasis) {
  switch (level) {
    case "high":
      return "border-danger/20 bg-danger/6 text-danger-foreground";
    case "medium":
      return "border-warning/25 bg-warning/10 text-warning-foreground";
    default:
      return "border-border/70 bg-secondary/35 text-foreground";
  }
}

export function getToneClasses(tone: RiskTone) {
  switch (tone) {
    case "high":
      return {
        border: "border-danger/20",
        background: "bg-[linear-gradient(180deg,rgba(239,68,68,0.08),rgba(255,255,255,0))]",
        badge: "danger" as const,
        button: "danger" as const,
        accent: "bg-danger",
        panel: "bg-danger/6",
      };
    case "medium":
      return {
        border: "border-warning/25",
        background: "bg-[linear-gradient(180deg,rgba(245,158,11,0.08),rgba(255,255,255,0))]",
        badge: "warning" as const,
        button: "secondary" as const,
        accent: "bg-warning",
        panel: "bg-warning/8",
      };
    case "low":
      return {
        border: "border-success/20",
        background: "bg-[linear-gradient(180deg,rgba(34,197,94,0.08),rgba(255,255,255,0))]",
        badge: "success" as const,
        button: "default" as const,
        accent: "bg-success",
        panel: "bg-success/8",
      };
    default:
      return {
        border: "border-accent/20",
        background: "bg-[linear-gradient(180deg,rgba(8,145,178,0.08),rgba(255,255,255,0))]",
        badge: "info" as const,
        button: "default" as const,
        accent: "bg-accent",
        panel: "bg-accent/8",
      };
  }
}

export function getKeyReasons(result: ExaminationResult): KeyReason[] {
  const reasons: KeyReason[] = [];
  const seen = new Set<string>();
  const personalBaselineAvailable = !!result.baseline_snapshot.personal.baseline_available;

  for (const pattern of result.decision.patterns ?? []) {
    const reason = patternReasonCatalog[pattern];
    if (!reason || seen.has(reason.title)) {
      continue;
    }
    seen.add(reason.title);
    reasons.push({ key: pattern, ...reason });
  }

  const baselineBand = result.baseline_snapshot.personal.band ?? result.baseline_snapshot.general.band;
  if (
    personalBaselineAvailable &&
    (baselineBand === "high" || baselineBand === "critical" || baselineBand === "medium") &&
    !seen.has("Отклонение от baseline")
  ) {
    seen.add("Отклонение от baseline");
    reasons.push({
      key: "baseline",
      title: "Отклонение от baseline",
      description: "Состояние заметно отличается от обычного профиля.",
      level: toEmphasisByBaselineBand(baselineBand),
      source: "Baseline",
    });
  }

  const strongestChannels = [...result.channel_contributions]
    .sort((left, right) => right.contribution - left.contribution)
    .filter((item) => item.contribution >= 0.09)
    .slice(0, 2);

  for (const item of strongestChannels) {
    const title =
      item.channel === "acoustic"
        ? "Акустический профиль"
        : item.channel === "paralinguistic"
          ? "Речевое поведение"
          : "Текстовый профиль";

    if (seen.has(title)) {
      continue;
    }

    seen.add(title);
    reasons.push({
      key: `${item.channel}-${item.metric_key}`,
      title,
      description:
        item.channel === "acoustic"
          ? "Напряжение и нестабильность голоса."
        : item.channel === "paralinguistic"
          ? "Паузы и организация речи дали выраженный сигнал."
          : "В содержании ответа есть значимый сигнал.",
      level: toEmphasisByContribution(item.contribution),
      source:
        item.channel === "acoustic"
          ? "Голос"
          : item.channel === "paralinguistic"
            ? "Темп и паузы"
            : "Содержание ответа",
    });
  }

  if (!reasons.length) {
    reasons.push({
      key: "stable-profile",
      title: "Критичных признаков не выявлено",
      description: "Выраженных отклонений по основным каналам не обнаружено.",
      level: "low",
      source: "Сводный профиль",
    });
  }

  return reasons.slice(0, 5);
}

export function getChannelImpactItems(result: ExaminationResult): ChannelImpactItem[] {
  const items: ChannelImpactItem[] = result.channel_contributions.map((item) => ({
    key: item.channel,
    title:
      item.channel === "acoustic"
        ? "Голос"
        : item.channel === "paralinguistic"
          ? "Темп и паузы"
          : "Содержание ответа",
    summary:
      item.channel === "text"
        ? "Текстовый сигнал"
        : item.channel === "acoustic"
          ? "Напряжение и нестабильность голоса"
          : "Паузы и речевая организация",
    valueText: item.contribution.toFixed(3),
    progress: Math.max(10, Math.round(item.contribution * 100)),
    emphasis: toEmphasisByContribution(item.contribution),
  }));

  const personalBaselineAvailable = !!result.baseline_snapshot.personal.baseline_available;
  items.push(
    personalBaselineAvailable
      ? {
          key: "baseline",
          title: "Отклонение от нормы",
          summary: "Сравнение с личной нормой",
          valueText: formatDelta(result.baseline_snapshot.personal.delta),
          progress: getBaselineProgress(result),
          emphasis: toEmphasisByBaselineBand(result.baseline_snapshot.personal.band),
        }
      : {
          key: "baseline",
          title: "Отклонение от нормы",
          summary: "Личная норма ещё формируется",
          valueText: `${result.baseline_snapshot.personal.baseline_exam_count ?? 0}/5`,
          progress: Math.max(
            10,
            Math.min(100, Math.round(((result.baseline_snapshot.personal.baseline_exam_count ?? 0) / 5) * 100)),
          ),
          emphasis: "low",
        },
  );

  return items;
}

export function getOperatorSteps(result: ExaminationResult) {
  const tone = getRiskTone(result);

  if (tone === "high") {
    return [
      "Провести повторную углублённую проверку.",
      "Сравнить результат с предыдущими обследованиями специалиста.",
      "При повторном высоком риске передать результат на дополнительную оценку.",
    ];
  }

  if (tone === "medium") {
    return [
      "Сверить результат с текущим рабочим контекстом специалиста.",
      "Принять решение о допуске с наблюдением или о повторной оценке.",
      "Зафиксировать результат и наблюдать динамику в следующих обследованиях.",
    ];
  }

  return [
    "Подтвердить результат обследования.",
    "Сохранить его в истории специалиста для будущего сравнения.",
    "При изменении состояния запустить повторную оценку.",
  ];
}

export function getScoreVisual(score: ResultChannelReportScore) {
  const normalized = Math.max(0, Math.min(1, score.value));
  const isPositive = positiveScoreKeys.has(score.key);
  const severity = isPositive ? 1 - normalized : normalized;

  if (severity >= 0.75) {
    return {
      badge: "danger" as const,
      barClassName: "bg-danger",
      trackClassName: "bg-danger/12",
      valueClassName: "text-danger-foreground",
    };
  }
  if (severity >= 0.5) {
    return {
      badge: "warning" as const,
      barClassName: "bg-warning",
      trackClassName: "bg-warning/15",
      valueClassName: "text-warning-foreground",
    };
  }
  if (severity >= 0.25) {
    return {
      badge: "info" as const,
      barClassName: "bg-accent",
      trackClassName: "bg-accent/15",
      valueClassName: "text-accent-foreground",
    };
  }
  return {
    badge: "success" as const,
    barClassName: "bg-success",
    trackClassName: "bg-success/15",
    valueClassName: "text-success-foreground",
  };
}

export function buildChannelStates(data?: ExaminationProcessingStatus) {
  return (["text", "acoustic", "paralinguistic"] as ProcessingChannelName[]).map((channel) => ({
    channel,
    state: data?.channels.find((item) => item.channel === channel),
  }));
}

export function getChannelStatusBadge(state?: ExaminationProcessingStatusChannel) {
  if (!state) {
    return { label: "Нет данных", variant: "neutral" as const };
  }

  return channelStatusConfig[state.status] ?? { label: state.status, variant: "neutral" as const };
}

export function getBaselineHeadline(result: ExaminationResult) {
  if (!result.baseline_snapshot.personal.baseline_available) {
    return "Личная норма ещё формируется";
  }

  switch (result.baseline_snapshot.personal.band) {
    case "low":
      return "Отклонение от личной нормы: в пределах нормы";
    case "mild":
    case "medium":
      return "Отклонение от личной нормы: умеренное";
    case "high":
    case "critical":
      return "Отклонение от личной нормы: выраженное";
    default:
      return `Отклонение от личной нормы: ${getBandLabel(result.baseline_snapshot.personal.band)}`;
  }
}

export function getBaselineDescription(result: ExaminationResult) {
  const personalCount = result.baseline_snapshot.personal.baseline_exam_count ?? 0;
  if (!result.baseline_snapshot.personal.baseline_available) {
    if (personalCount <= 0) {
      return "Сравнение пока выполняется только с общей нормой. Baseline ещё не влияет на итоговое решение.";
    }
    return `Собрано ${personalCount} из 5 качественных обследований. До формирования личной нормы сравнение с общей нормой остаётся только справочным сигналом.`;
  }
  return "";
}

export function getImpactLabel(level: Emphasis) {
  switch (level) {
    case "high":
      return "Высокий вклад";
    case "medium":
      return "Средний вклад";
    default:
      return "Низкий вклад";
  }
}
