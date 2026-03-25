-- name: CreateQuestionnaire :one
INSERT INTO questionnaires (
    title,
    description,
    is_active,
    last_edited_by_user_id,
    last_edited_at
) VALUES (
    $1, $2, $3, $4, NOW()
)
RETURNING id, title, description, is_active, last_edited_by_user_id, last_edited_at, created_at, updated_at;

-- name: UpdateQuestionnaire :one
UPDATE questionnaires
SET
    title = $2,
    description = $3,
    is_active = $4,
    last_edited_by_user_id = $5,
    last_edited_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING id, title, description, is_active, last_edited_by_user_id, last_edited_at, created_at, updated_at;

-- name: GetQuestionnaireByID :one
SELECT id, title, description, is_active, last_edited_by_user_id, last_edited_at, created_at, updated_at
FROM questionnaires
WHERE id = $1;

-- name: ListQuestionnaireDetails :many
SELECT
    q.id,
    q.title,
    q.description,
    q.is_active,
    usage_stats.usage_count,
    COALESCE(usage_stats.last_used_at, 'epoch'::timestamptz) AS last_used_at,
    q.last_edited_at,
    editor.id AS last_editor_user_id,
    editor.login AS last_editor_login,
    q.created_at,
    q.updated_at,
    qq.question_id,
    ques.text AS question_text,
    qq.position
FROM questionnaires q
LEFT JOIN (
    SELECT
        e.questionnaire_id,
        COUNT(*)::BIGINT AS usage_count,
        MAX(COALESCE(e.finished_at, e.started_at, e.created_at)) AS last_used_at
    FROM examinations e
    WHERE e.questionnaire_id IS NOT NULL
    GROUP BY e.questionnaire_id
) usage_stats ON usage_stats.questionnaire_id = q.id
LEFT JOIN users editor ON editor.id = q.last_edited_by_user_id
LEFT JOIN questionnaire_questions qq ON qq.questionnaire_id = q.id
LEFT JOIN questions ques ON ques.id = qq.question_id
ORDER BY q.id DESC, qq.position ASC;

-- name: GetQuestionnaireDetailsByID :many
SELECT
    q.id,
    q.title,
    q.description,
    q.is_active,
    usage_stats.usage_count,
    COALESCE(usage_stats.last_used_at, 'epoch'::timestamptz) AS last_used_at,
    q.last_edited_at,
    editor.id AS last_editor_user_id,
    editor.login AS last_editor_login,
    q.created_at,
    q.updated_at,
    qq.question_id,
    ques.text AS question_text,
    qq.position
FROM questionnaires q
LEFT JOIN (
    SELECT
        e.questionnaire_id,
        COUNT(*)::BIGINT AS usage_count,
        MAX(COALESCE(e.finished_at, e.started_at, e.created_at)) AS last_used_at
    FROM examinations e
    WHERE e.questionnaire_id IS NOT NULL
    GROUP BY e.questionnaire_id
) usage_stats ON usage_stats.questionnaire_id = q.id
LEFT JOIN users editor ON editor.id = q.last_edited_by_user_id
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
