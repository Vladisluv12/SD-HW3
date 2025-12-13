package models

func MapFileIDToSimilarWorks(originalId string, fileIDs string, reportID string) *SimilarWork {
	return &SimilarWork{
		OriginalWorkID:       originalId,
		SimilarWorkID:        fileIDs,
		ReportID:             reportID,
		SimilarityPercentage: 100.0,
	}
}
