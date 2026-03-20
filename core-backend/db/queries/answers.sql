-- name: NextAnswerID :one
SELECT nextval('answers_id_seq');

-- name: GetExaminationQuestionByID :one
SELECT id, examination_id, specialist_id, questionnaire_id, source_question_id, position, question_text
FROM examination_questions
WHERE id = $1;

-- name: CreateAnswer :one
INSERT INTO answers (
    id,
    examination_id,
    examination_question_id,
    specialist_id,
    created_by_user_id,
    answer_text,
    audio_s3_key
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, examination_id, examination_question_id, specialist_id, created_by_user_id, answer_text, audio_s3_key, created_at;
