package specialityexamination

type DrugHistory struct {
	NamaObat        string `json:"nama_obat" binding:"required" bson:"nama_obat"`
	DosisPakai      string `json:"dosis_pakai" binding:"required" bson:"dosis_pakai"`
	WaktuPenggunaan string `json:"waktu_penggunaan" binding:"required" bson:"waktu_penggunaan"`
}
