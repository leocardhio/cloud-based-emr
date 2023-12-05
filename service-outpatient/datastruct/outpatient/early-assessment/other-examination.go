package earlyassessment

type OtherExamination struct {
	StatusPsikologis string `json:"status_psikologis" binding:"required" bson:"status_psikologis"`
	SosialEkonomi    string `json:"sosial_ekonomi" binding:"required" bson:"sosial_ekonomi"`
	Spiritual        string `json:"spiritual" binding:"required" bson:"spiritual"`
}
