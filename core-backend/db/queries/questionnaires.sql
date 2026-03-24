-- name: CreateQuestionnaire :one
INSERT INTO questionnaires (
    title,
    description,
    is_active
) VALUES (
    $1, $2, $3
)
RETURNING id, title, description, is_active, created_at, updated_at;

-- name: UpdateQuestionnaire :one
UPDATE questionnaires
SET
    title = $2,
    description = $3,
    is_active = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING id, title, description, is_active, created_at, updated_at;

-- name: GetQuestionnaireByID :one
SELECT id, title, description, is_active, created_at, updated_at
FROM questionnaires
WHERE id = $1;

-- name: ListQuestionnaireDetails :many
SELECT
    q.id,
    q.title,
    q.description,
    q.is_active,
    q.created_at,
    q.updated_at,
    qq.question_id,
    ques.text AS question_text,
    qq.position
FROM questionnaires q
LEFT JOIN questionnaire_questions qq ON qq.questionnaire_id = q.id
LEFT JOIN questions ques ON ques.id = qq.question_id
ORDER BY q.id DESC, qq.position ASC;

-- name: GetQuestionnaireDetailsByID :many
SELECT
    q.id,
    q.title,
    q.description,
    q.is_active,
    q.created_at,
    q.updated_at,
    qq.question_id,
    ques.text AS question_text,
    qq.position
FROM questionnaires q
LEFT JOIN questionnaire_questions qq ON qq.questionnaire_id = q.id
LEFT JOIN questions ques ON ques.id = qq.question_id
WHERE q.id = $1
ORDER BY qq.position ASC;

-- name: CreateQuestion :one
INSERT INTO questions (
    text
) VALUES (
    $1
)
RETURNING id, text, created_at, updated_at;

-- name: DeleteQuestionnaireQuestions :exec
DELETE FROM questionnaire_questions
WHERE questionnaire_id = $1;

-- name: AddQuestionToQuestionnaire :exec
INSERT INTO questionnaire_questions (
    questionnaire_id,
    question_id,
    position
) VALUES (
    $1, $2, $3
);

-- name: DeleteOrphanQuestions :exec
DELETE FROM questions q
WHERE NOT EXISTS (
    SELECT 1
    FROM questionnaire_questions qq
    WHERE qq.question_id = q.id
);
