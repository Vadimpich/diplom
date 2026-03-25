"use client";

import { Mic, Pause, RotateCcw, StopCircle, Upload } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient, ApiError } from "@/lib/api/client";
import { appendDraftAnswer } from "@/lib/examination-drafts";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import type { Answer } from "@/lib/api/types";

type RecorderState = "idle" | "recording" | "recorded";

export function MediaRecorderCard({
  examinationId,
  questionLabel,
  answerIndex,
  onUploaded,
}: {
  examinationId: number;
  questionLabel?: string;
  answerIndex: number;
  onUploaded: (answer: Answer) => void;
}) {
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const [recorderState, setRecorderState] = useState<RecorderState>("idle");
  const [audioBlob, setAudioBlob] = useState<Blob | null>(null);
  const [audioUrl, setAudioUrl] = useState<string | null>(null);
  const [transcript, setTranscript] = useState("");
  const [recorderError, setRecorderError] = useState<string | null>(null);

  useEffect(() => {
    return () => {
      if (audioUrl) {
        URL.revokeObjectURL(audioUrl);
      }
    };
  }, [audioUrl]);

  const uploadMutation = useMutation({
    mutationFn: async () => {
      if (!audioBlob) {
        throw new Error("Сначала запишите ответ.");
      }

      const file = new File([audioBlob], "answer.webm", {
        type: audioBlob.type || "audio/webm",
      });

      return apiClient.uploadAnswer({
        examination_id: examinationId,
        text: transcript,
        audio: file,
      });
    },
    onSuccess: (answer) => {
      appendDraftAnswer(answer);
      onUploaded(answer);
      toast.success("Ответ сохранён", {
        description: "Запись и текст ответа добавлены в текущее обследование.",
      });
      setTranscript("");
      setAudioBlob(null);
      if (audioUrl) {
        URL.revokeObjectURL(audioUrl);
      }
      setAudioUrl(null);
      setRecorderState("idle");
    },
  });

  const durationHint = useMemo(() => {
    if (recorderState === "recording") {
      return "Идёт запись. Завершите ответ после окончания реплики.";
    }

    if (audioBlob) {
      return "Запись готова к отправке. При необходимости перезапишите ответ.";
    }

    return "Для работы требуется разрешение браузера на микрофон.";
  }, [audioBlob, recorderState]);

  async function startRecording() {
    try {
      setRecorderError(null);
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const mediaRecorder = new MediaRecorder(stream);
      chunksRef.current = [];
      mediaRecorderRef.current = mediaRecorder;

      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          chunksRef.current.push(event.data);
        }
      };

      mediaRecorder.onstop = () => {
        const blob = new Blob(chunksRef.current, { type: mediaRecorder.mimeType || "audio/webm" });
        setAudioBlob(blob);
        setRecorderState("recorded");
        if (audioUrl) {
          URL.revokeObjectURL(audioUrl);
        }
        setAudioUrl(URL.createObjectURL(blob));
        stream.getTracks().forEach((track) => track.stop());
      };

      mediaRecorder.start();
      setRecorderState("recording");
    } catch {
      setRecorderError("Не удалось получить доступ к микрофону. Проверьте разрешения браузера.");
    }
  }

  function stopRecording() {
    mediaRecorderRef.current?.stop();
  }

  function resetRecording() {
    setAudioBlob(null);
    setRecorderState("idle");
    if (audioUrl) {
      URL.revokeObjectURL(audioUrl);
    }
    setAudioUrl(null);
  }

  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle>Запись ответа</CardTitle>
        <CardDescription>{durationHint}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="rounded-2xl border border-border/70 bg-secondary/20 p-4">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Текущий ответ</p>
          <p className="mt-2 text-base font-semibold">Ответ {answerIndex}</p>
          <p className="mt-2 text-sm text-muted-foreground">
            {questionLabel ?? "Запишите реплику обследуемого и затем сохраните её в карточке сеанса."}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          {recorderState !== "recording" ? (
            <Button type="button" onClick={startRecording}>
              <Mic className="mr-2 h-4 w-4" />
              Начать запись
            </Button>
          ) : (
            <Button type="button" variant="danger" onClick={stopRecording}>
              <StopCircle className="mr-2 h-4 w-4" />
              Остановить
            </Button>
          )}
          <Button type="button" variant="outline" onClick={resetRecording} disabled={!audioBlob && recorderState !== "recording"}>
            <RotateCcw className="mr-2 h-4 w-4" />
            Перезаписать
          </Button>
          {recorderState === "recording" ? (
            <span className="inline-flex items-center gap-2 rounded-full bg-danger/10 px-3 py-2 text-sm text-danger">
              <Pause className="h-4 w-4" />
              Идёт запись
            </span>
          ) : null}
        </div>

        {audioUrl ? <audio className="w-full" controls src={audioUrl} /> : null}

        <div className="space-y-2">
          <Label htmlFor="answer-text">Текст ответа</Label>
          <Textarea
            id="answer-text"
            placeholder="Зафиксируйте содержание ответа после записи"
            value={transcript}
            onChange={(event) => setTranscript(event.target.value)}
          />
          <p className="text-sm text-muted-foreground">
            Сначала остановите запись, затем сохраните ответ. После сохранения он сразу появится в журнале сеанса справа.
          </p>
        </div>

        {recorderError ? <Alert variant="danger">{recorderError}</Alert> : null}
        {uploadMutation.isError ? (
          <Alert variant="danger">{(uploadMutation.error as ApiError).message}</Alert>
        ) : null}

        <Button
          type="button"
          className="w-full"
          disabled={uploadMutation.isPending || !audioBlob || transcript.trim().length === 0}
          onClick={() => uploadMutation.mutate()}
        >
          {uploadMutation.isPending ? (
            <>
              <Spinner />
              <span className="ml-2">Отправка ответа...</span>
            </>
          ) : (
            <>
              <Upload className="mr-2 h-4 w-4" />
              Сохранить ответ в сеансе
            </>
          )}
        </Button>
      </CardContent>
    </Card>
  );
}
