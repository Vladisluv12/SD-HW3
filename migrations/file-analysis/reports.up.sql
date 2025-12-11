CREATE TABLE IF NOT EXISTS reports (
    report_id VARCHAR(255) PRIMARY KEY,
    work_id VARCHAR(255) NOT NULL,
    student_id VARCHAR(255) NOT NULL,
    assignment_id VARCHAR(255) NOT NULL,
    plagiarism_score DECIMAL(5,2) DEFAULT 0.00,
    is_plagiarism BOOLEAN DEFAULT FALSE,
    word_count INT DEFAULT 0,
    analysis_duration_ms INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'completed',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);