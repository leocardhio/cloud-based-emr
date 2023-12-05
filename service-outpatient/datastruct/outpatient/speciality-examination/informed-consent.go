package specialityexamination

import "time"

type StatementSigner struct {
	DokterPenjelas     string `json:"dokter_penjelas" binding:"required" bson:"dokter_penjelas"`
	PenerimaPenjelasan string `json:"penerima_penjelasan" binding:"required" bson:"penerima_penjelasan"`
	Saksi1             string `json:"saksi1" binding:"required" bson:"saksi1"`
	Saksi2             string `json:"saksi2" binding:"required" bson:"saksi2"`
}

type InformedConsent struct {
	NamaPasien          string          `json:"nama_pasien" binding:"required" bson:"nama_pasien"`
	DokterPenjelas      string          `json:"dokter_penjelas" binding:"required" bson:"dokter_penjelas"`
	PetugasPendamping   string          `json:"petugas_pendamping" binding:"required" bson:"petugas_pendamping"`
	NamaKeluarga        string          `json:"nama_keluarga" binding:"required" bson:"nama_keluarga"`
	Tindakan            string          `json:"tindakan" binding:"required" bson:"tindakan"`
	KonsekuensiTindakan string          `json:"konsekuensi_tindakan" binding:"required" bson:"konsekuensi_tindakan"`
	Persetujuan         bool            `json:"persetujuan" binding:"required" bson:"persetujuan"`
	WaktuMenjelaskan    time.Time       `json:"waktu_menjelaskan" binding:"required" bson:"waktu_menjelaskan"`
	PembuatPernyataan   StatementSigner `json:"pembuat_pernyataan" binding:"required" bson:"pembuat_pernyataan"`
}
