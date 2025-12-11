CREATE TABLE IF NOT EXISTS similar_works (
    similar_id VARCHAR(255) PRIMARY KEY,
    report_id VARCHAR(255) NOT NULL,
    original_work_id VARCHAR(255) NOT NULL,
    similar_work_id VARCHAR(255) NOT NULL,
    similarity_percentage DECIMAL(5,2) DEFAULT 0.00,
    FOREIGN KEY (report_id) REFERENCES reports(report_id) ON DELETE CASCADE
);