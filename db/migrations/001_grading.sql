-- Phase 1: Grading, Deadlines, Quiz Enhancements

-- Assignment: add max_grade field
ALTER TABLE assignment ADD COLUMN IF NOT EXISTS max_grade DECIMAL(5,2) NOT NULL DEFAULT 20;

-- Submission: add grade columns
ALTER TABLE submission ADD COLUMN IF NOT EXISTS grade DECIMAL(5,2);
ALTER TABLE submission ADD COLUMN IF NOT EXISTS max_grade DECIMAL(5,2);
ALTER TABLE submission ADD COLUMN IF NOT EXISTS graded_at TIMESTAMPTZ;

-- Quiz: add deadline, time limit, max attempts
ALTER TABLE quiz ADD COLUMN IF NOT EXISTS deadline TIMESTAMPTZ;
ALTER TABLE quiz ADD COLUMN IF NOT EXISTS time_limit_minutes INT NOT NULL DEFAULT 0;
ALTER TABLE quiz ADD COLUMN IF NOT EXISTS max_attempts INT NOT NULL DEFAULT 1;

-- Quiz questions: expand allowed types to include file_upload
ALTER TABLE quiz_question DROP CONSTRAINT IF EXISTS quiz_question_question_type_check;
ALTER TABLE quiz_question ADD CONSTRAINT quiz_question_question_type_check
    CHECK (question_type IN ('mcq', 'true_false', 'fill_blank', 'open_ended', 'file_upload'));

-- Quiz attempt: add file_answers for file_upload questions
ALTER TABLE quiz_attempt ADD COLUMN IF NOT EXISTS file_answers JSONB NOT NULL DEFAULT '{}';
