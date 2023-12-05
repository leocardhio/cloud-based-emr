package earlyassessment

type Anamnesis struct {
	KeluhanUtama      string   `json:"keluhan_utama" binding:"required" bson:"keluhan_utama"`
	RiwayatPenyakit   []string `json:"riwayat_penyakit" binding:"required" bson:"riwayat_penyakit"`
	RiwayatAlergi     []string `json:"riwayat_alergi" binding:"required" bson:"riwayat_alergi"`
	RiwayatPengobatan []string `json:"riwayat_pengobatan" binding:"required" bson:"riwayat_pengobatan"`
}
