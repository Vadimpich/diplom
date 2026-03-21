package processing

import (
	"encoding/json"
	"time"
)

const (
	MessageVersionV1 = 1

	ChannelText           = "text"
	ChannelAcoustic       = "acoustic"
	ChannelParalinguistic = "paralinguistic"

	ResultStatusSucceeded      = "succeeded"
	ResultStatusTemporaryError = "temporary_error"
	ResultStatusFatalError     = "fatal_error"
)

var MandatoryChannels = []string{
	ChannelText,
	ChannelAcoustic,
	ChannelParalinguistic,
}

type CommandAnswerReference struct {
	AnswerID      int64  `json:"answer_id"`
	QuestionID    int64  `json:"question_id"`
	AudioS3Bucket string `json:"audio_s3_bucket"`
	AudioS3Key    string `json:"audio_s3_key"`
	AnswerText    string `json:"answer_text"`
}

type ProcessingCommandEnvelope struct {
	MessageVersion int                      `json:"message_version"`
	MessageID      string                   `json:"message_id"`
	CorrelationID  string                   `json:"correlation_id"`
	ExaminationID  int64                    `json:"examination_id"`
	SpecialistID   int64                    `json:"specialist_id"`
	Channel        string                   `json:"channel"`
	Attempt        int32                    `json:"attempt"`
	MaxAttempts    int32                    `json:"max_attempts"`
	RequestedAt    time.Time                `json:"requested_at"`
	Answers        []CommandAnswerReference `json:"answers"`
}

type ChannelResultEnvelope struct {
	MessageVersion int             `json:"message_version"`
	MessageID      string          `json:"message_id"`
	CorrelationID  string          `json:"correlation_id"`
	ExaminationID  int64           `json:"examination_id"`
	Channel        string          `json:"channel"`
	Attempt        int32           `json:"attempt"`
	Status         string          `json:"status"`
	CompletedAt    time.Time       `json:"completed_at"`
	ModelVersion   string          `json:"model_version"`
	ErrorCode      *string         `json:"error_code"`
	ErrorMessage   *string         `json:"error_message"`
	Payload        json.RawMessage `json:"payload"`
}

type ChannelStatusDTO struct {
	Channel             string     `json:"channel"`
	Status              string     `json:"status"`
	AttemptCount        int32      `json:"attempt_count"`
	MaxAttempts         int32      `json:"max_attempts"`
	MessageVersion      int        `json:"message_version"`
	QueuedAt            *time.Time `json:"queued_at"`
	StartedAt           *time.Time `json:"started_at"`
	FinishedAt          *time.Time `json:"finished_at"`
	LastErrorCode       *string    `json:"last_error_code"`
	LastErrorMessage    *string    `json:"last_error_message"`
	BrokerMessageID     *string    `json:"broker_message_id"`
	BrokerCorrelationID *string    `json:"broker_correlation_id"`
}

type ProcessingStatusResponse struct {
	ExaminationID    int64              `json:"examination_id"`
	Status           string             `json:"status"`
	MessageVersion   int                `json:"message_version"`
	ChannelsTotal    int                `json:"channels_total"`
	ChannelsComplete int                `json:"channels_completed"`
	Terminal         bool               `json:"terminal"`
	StartedAt        *time.Time         `json:"started_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	FinishedAt       *time.Time         `json:"finished_at"`
	FailedAt         *time.Time         `json:"failed_at"`
	Channels         []ChannelStatusDTO `json:"channels"`
}
