package specialityexamination

type FinalDiagnosis struct {
	DiagnosisPrimer   string `json:"diagnosis_primer" binding:"required" bson:"diagnosis_primer"`
	DiagnosisSekunder string `json:"diagnosis_sekunder" binding:"required" bson:"diagnosis_sekunder"`
}

type Diagnosis struct {
	DiagnosisAwal  string         `json:"diagnosis_awal" binding:"required" bson:"diagnosis_awal"`
	DiagnosisAkhir FinalDiagnosis `json:"diagnosis_akhir" binding:"required" bson:"diagnosis_akhir"`
}
