import { Badge } from "@/components/ui/badge";
import type { ExaminationProcessingStatus } from "@/lib/api/types";
import { buildChannelStates, channelLabels, getChannelStatusBadge } from "./helpers";

export function StatusStrip({ processing }: { processing?: ExaminationProcessingStatus }) {
  const channelStates = buildChannelStates(processing);

  return (
    <div className="grid gap-3 md:grid-cols-3">
      {channelStates.map(({ channel, state }) => {
        const badge = getChannelStatusBadge(state);
        return (
          <div
            key={channel}
            className="flex items-center justify-between gap-3 rounded-2xl border border-border/70 bg-card px-4 py-3"
          >
            <div>
              <p className="text-sm font-medium text-foreground">{channelLabels[channel]}</p>
              <p className="text-xs text-muted-foreground">
                {state ? `Попытка ${state.attempt_count} из ${state.max_attempts}` : "Статус ещё не получен"}
              </p>
            </div>
            <Badge variant={badge.variant}>{badge.label}</Badge>
          </div>
        );
      })}
    </div>
  );
}
