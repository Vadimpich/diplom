WITH existing_questionnaire AS (
    SELECT id
    FROM questionnaires
    WHERE title = 'Базовый опрос'
    ORDER BY id
    LIMIT 1
),
created_questionnaire AS (
    INSERT INTO questionnaires (
        title,
        description,
        is_active,
        last_edited_by_user_id,
        last_edited_at
    )
    SELECT
        'Базовый опрос',
        'Стартовый опросник, автоматически создаваемый при инициализации системы.',
        TRUE,
        NULL,
        NOW()
    WHERE NOT EXISTS (SELECT 1 FROM existing_questionnaire)
    RETURNING id
),
target_questionnaire AS (
    SELECT id FROM existing_questionnaire
    UNION ALL
    SELECT id FROM created_questionnaire
),
inserted_questions AS (
    INSERT INTO questions (text)
    SELECT question_text
    FROM (
        VALUES
            (1, 'Расскажите кратко о своем текущем состоянии.'),
            (2, 'Как вы спали и отдыхали в последние сутки?'),
            (3, 'Есть ли что-то, что мешает вам сосредоточиться на работе?')
    ) AS seed(position, question_text)
    WHERE EXISTS (SELECT 1 FROM created_questionnaire)
    RETURNING id, text
)
INSERT INTO questionnaire_questions (questionnaire_id, question_id, position)
SELECT tq.id, iq.id, seed.position
FROM target_questionnaire tq
JOIN (
    VALUES
        (1, 'Расскажите кратко о своем текущем состоянии.'),
        (2, 'Как вы спали и отдыхали в последние сутки?'),
        (3, 'Есть ли что-то, что мешает вам сосредоточиться на работе?')
) AS seed(position, question_text) ON TRUE
JOIN inserted_questions iq ON iq.text = seed.question_text
WHERE EXISTS (SELECT 1 FROM created_questionnaire);
