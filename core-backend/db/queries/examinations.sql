-- name: CreateExamination :one
INSERT INTO examinations (
    specialist_id,
    created_by_user_id,
    questionnaire_id,
    status
) VALUES (
    $1, $2, $3, 'created'
)
RETURNING id, specialist_id, created_by_user_id, questionnaire_id, status, created_at, started_at, finished_at, updated_at;

-- name: CreateExaminationQuestionSnapshots :many
INSERT INTO examination_questions (
    examination_id,
    specialist_id,
    questionnaire_id,
    source_question_id,
    position,
    question_text
)
SELECT
    $1,
    $2,
    $3,
    qq.question_id,
    qq.position,
    q.text
FROM questionnaire_questions qq
JOIN questions q ON q.id = qq.question_id
WHERE qq.questionnaire_id = $3
ORDER BY qq.position
RETURNING id, examination_id, specialist_id, questionnaire_id, source_question_id, position, question_text;

-- name: GetExaminationByID :one
SELECT id, specialist_id, created_by_user_id, questionnaire_id, status, created_at, started_at, finished_at, updated_at
FROM examinations
WHERE id = $1;

-- name: ListExaminations :many
SELECT id, specialist_id, created_by_user_id, questionnaire_id, status, created_at, started_at, finished_at, updated_at
FROM examinations
ORDER BY created_at DESC, id DESC;

-- name: ListExaminationsBySpecialistID :many
SELECT id, specialist_id, created_by_user_id, questionnaire_id, status, created_at, started_at, finished_at, updated_at
FROM examinations
WHERE specialist_id = $1
ORDER BY created_at DESC, id DESC;

-- name: UpdateExaminationStatus :one
UPDATE examinations
SET
    status = $2,
    started_at = CASE
        WHEN $2 = 'collecting_answers' AND started_at IS NULL THEN NOW()
        ELSE started_at
    END,
    finished_at = CASE
        WHEN $2 = 'ready_for_processing' THEN NOW()
        ELSE finished_at
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING id, specialist_id, created_by_user_id, questionnaire_id, status, created_at, started_at, finished_at, updated_at;
