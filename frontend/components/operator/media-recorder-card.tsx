"use client";

import { CheckCircle2, FileAudio, Mic, Pause, RotateCcw, StopCircle, Upload } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient, ApiError } from "@/lib/api/client";
import { appendDraftAnswer } from "@/lib/examination-drafts";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";
import type { Answer } from "@/lib/api/types";

type RecorderState = "idle" | "recording" | "recorded";

export function MediaRecorderCard({
  examinationId,
  examinationQuestionId,
  specialistId,
  answerIndex,
  totalQuestions,
  onUploaded,
}: {
  examinationId: number;
  examinationQuestionId?: number;
  specialistId?: number;
  answerIndex: number;
  totalQuestions: number;
  onUploaded: (answer: Answer) => void;
}) {
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const [recorderState, setRecorderState] = useState<RecorderState>("idle");
  const [audioBlob, setAudioBlob] = useState<Blob | File | null>(null);
  const [audioUrl, setAudioUrl] = useState<string | null>(null);
  const [audioSourceLabel, setAudioSourceLabel] = useState<string>("Микрофон готов");
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
      if (!examinationQuestionId || !specialistId) {
        throw new Error("Для этого обследования ещё не определён текущий вопрос.");
      }

      const file = new File([audioBlob], "answer.webm", {
        type: audioBlob.type || "audio/webm",
      });
      const uploadFile =
        audioBlob instanceof File
          ? audioBlob
          : file;

      return apiClient.uploadAnswer({
        examination_id: examinationId,
        examination_question_id: examinationQuestionId,
        specialist_id: specialistId,
        audio: uploadFile,
      });
    },
    onSuccess: (answer) => {
      appendDraftAnswer(answer);
      onUploaded(answer);
      toast.success("Ответ сохранён", {
        description: "Аудиозапись добавлена в текущее обследование.",
      });
      setAudioBlob(null);
      if (audioUrl) {
        URL.revokeObjectURL(audioUrl);
      }
      setAudioUrl(null);
      setAudioSourceLabel("Микрофон готов");
      setRecorderState("idle");
    },
  });

  const durationHint = useMemo(() => {
    if (recorderState === "recording") {
      return "Идёт запись";
    }

    if (audioBlob) {
      return audioBlob instanceof File ? "Файл готов" : "Запись готова";
    }

    return audioSourceLabel;
  }, [audioBlob, audioSourceLabel, recorderState]);

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
        setAudioSourceLabel("Запись готова");
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
    mediaRecorderRef.current?.stream?.getTracks().forEach((track) => track.stop());
    setAudioBlob(null);
    setRecorderState("idle");
    if (audioUrl) {
      URL.revokeObjectURL(audioUrl);
    }
    setAudioUrl(null);
    setAudioSourceLabel("Микрофон готов");
    setRecorderError(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }

  function openFilePicker() {
    fileInputRef.current?.click();
  }

  function handleFileSelected(event: React.ChangeEvent<HTMLInputElement>) {
    const nextFile = event.target.files?.[0];
    if (!nextFile) {
      return;
    }
    if (!nextFile.type.startsWith("audio/")) {
      setRecorderError("Выберите аудиофайл для загрузки.");
      event.target.value = "";
      return;
    }

    mediaRecorderRef.current?.stream?.getTracks().forEach((track) => track.stop());
    setRecorderError(null);
    setAudioBlob(nextFile);
    setRecorderState("recorded");
    setAudioSourceLabel("Файл готов");
    if (audioUrl) {
      URL.revokeObjectURL(audioUrl);
    }
    setAudioUrl(URL.createObjectURL(nextFile));
  }

  return (
    <Card>
      <CardContent className="space-y-5 p-5">
        <div className="rounded-[24px] border border-border/70 bg-secondary/15 p-5">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                <Mic className="h-5 w-5" />
              </div>
              <div>
                <p className="text-sm font-semibold">Ответ {totalQuestions > 0 ? `${answerIndex} / ${totalQuestions}` : answerIndex}</p>
                <p className="text-sm text-muted-foreground">
                  {recorderState === "recording"
                    ? "Идёт запись"
                    : recorderState === "recorded"
                      ? audioBlob instanceof File
                        ? "Файл готов"
                        : "Запись готова"
                      : "Микрофон готов"}
                </p>
              </div>
            </div>
            {recorderState === "recording" ? (
              <span className="inline-flex items-center gap-2 rounded-full bg-danger/10 px-3 py-2 text-sm text-danger">
                <Pause className="h-4 w-4" />
                {durationHint}
              </span>
            ) : recorderState === "recorded" ? (
              <span className="inline-flex items-center gap-2 rounded-full bg-emerald-500/10 px-3 py-2 text-sm text-emerald-700">
                <CheckCircle2 className="h-4 w-4" />
                {durationHint}
              </span>
            ) : (
              <span className="inline-flex items-center gap-2 rounded-full bg-background px-3 py-2 text-sm text-muted-foreground">
                <Mic className="h-4 w-4" />
                {durationHint}
              </span>
            )}
          </div>

          <div className="mt-5 grid grid-cols-8 gap-2">
            {Array.from({ length: 8 }).map((_, index) => {
              const active = recorderState === "recording" && index % 2 === 0;
              const height = recorderState === "recorded" ? ((index % 4) + 2) * 12 : ((index % 3) + 1) * 10;

              return (
                <div
                  key={index}
                  className="flex h-20 items-end rounded-2xl bg-background/70 px-1.5 py-2"
                >
                  <div
                    className={`w-full rounded-full transition-all ${active ? "animate-pulse bg-primary" : "bg-primary/35"}`}
                    style={{ height }}
                  />
                </div>
              );
            })}
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <input
            ref={fileInputRef}
            type="file"
            accept="audio/*,.wav,.mp3,.m4a,.ogg,.webm"
            className="hidden"
            onChange={handleFileSelected}
          />
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
          <Button
            type="button"
            variant="outline"
            className="h-10 w-10 rounded-full border-border/70 p-0 text-muted-foreground"
            onClick={openFilePicker}
            disabled={uploadMutation.isPending || recorderState === "recording"}
            title="Загрузить аудиофайл"
            aria-label="Загрузить аудиофайл"
          >
            <FileAudio className="h-4 w-4" />
          </Button>
          <Button type="button" variant="outline" onClick={resetRecording} disabled={!audioBlob && recorderState !== "recording"}>
            <RotateCcw className="mr-2 h-4 w-4" />
            Перезаписать
          </Button>
        </div>

        {audioUrl ? (
          <div className="rounded-2xl border border-border/70 bg-background p-4">
            <p className="text-sm font-medium">Предпрослушивание</p>
            <audio className="mt-4 w-full" controls src={audioUrl} />
          </div>
        ) : null}

        {recorderError ? <Alert variant="danger">{recorderError}</Alert> : null}
        {uploadMutation.isError ? (
          <Alert variant="danger">{(uploadMutation.error as ApiError).message}</Alert>
        ) : null}

        <Button
          type="button"
          className="w-full"
          disabled={uploadMutation.isPending || !audioBlob || !examinationQuestionId || !specialistId}
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
              Сохранить аудиоответ
            </>
          )}
        </Button>
      </CardContent>
    </Card>
  );
}
