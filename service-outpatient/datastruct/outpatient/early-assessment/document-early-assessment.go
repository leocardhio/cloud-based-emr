package earlyassessment

type EarlyAssessment struct {
	Anamnesis          Anamnesis           `json:"anamnesis" binding:"required" bson:"anamnesis"`
	PemeriksaanFisik   PhysicalExamination `json:"pemeriksaan_fisik" binding:"required" bson:"pemeriksaan_fisik"`
	PemeriksaanLainnya OtherExamination    `json:"pemeriksaan_lainnya" binding:"required" bson:"pemeriksaan_lainnya"`
}
