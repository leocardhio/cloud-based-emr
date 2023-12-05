package specialityexamination

type SpecialityExamination struct {
	RiwayatPenggunaanObat []DrugHistory         `json:"riwayat_penggunaan_obat" binding:"required" bson:"riwayat_penggunaan_obat"`
	RencanaRawat          string                `json:"rencana_rawat" binding:"required" bson:"rencana_rawat"`
	InstruksiMedik        string                `json:"instruksi_medik" binding:"required" bson:"instruksi_medik"`
	PemeriksaanPenunjang  SupportingExamination `json:"pemeriksaan_penunjang" binding:"required" bson:"pemeriksaan_penunjang"`
	Diagnosis             Diagnosis             `json:"diagnosis" binding:"required" bson:"diagnosis"`
	PersetujuanTindakan   InformedConsent       `json:"persetujuan_tindakan" binding:"required" bson:"persetujuan_tindakan"`
	Terapi                Therapy               `json:"terapi" binding:"required" bson:"terapi"`
}
