WITH deleted_questionnaires AS (
    DELETE FROM questionnaires
    WHERE title = 'Базовый опрос'
    RETURNING id
)
DELETE FROM questions q
WHERE q.text IN (
    'Расскажите кратко о своем текущем состоянии.',
    'Как вы спали и отдыхали в последние сутки?',
    'Есть ли что-то, что мешает вам сосредоточиться на работе?'
)
AND NOT EXISTS (
    SELECT 1
    FROM questionnaire_questions qq
    WHERE qq.question_id = q.id
);
