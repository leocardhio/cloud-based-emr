package pharmacy

import (
	"service-pharmacy/datastruct"
	"time"
)

type Etiquette struct {
	NamaPasien            string    `json:"nama_pasien" binding:"required" bson:"nama_pasien"`
	TanggalLahir          time.Time `json:"tanggal_lahir" binding:"required" bson:"tanggal_lahir"`
	NamaObat              string    `json:"nama_obat" binding:"required" bson:"nama_obat"`
	Bentuk                string    `json:"bentuk" binding:"required" bson:"bentuk"`
	JumlahObat            uint      `json:"jumlah_obat" binding:"required" bson:"jumlah_obat"`
	AturanPakai           HowToUse  `json:"aturan_pakai" binding:"required" bson:"aturan_pakai"`
	CatatanResep          string    `json:"catatan_resep" binding:"required" bson:"catatan_resep"`
	TanggalObatDiserahkan time.Time `json:"tanggal_obat_diserahkan" binding:"required" bson:"tanggal_obat_diserahkan"`
}

type Dispensing struct {
	StatusResep           *datastruct.RecipeStatus `json:"status_resep" binding:"required" bson:"status_resep"`
	WaktuPenyerahan       time.Time                `json:"waktu_penyerahan" binding:"required" bson:"waktu_penyerahan"`
	NamaPetugasDispensing string                   `json:"nama_petugas_dispensing" binding:"required" bson:"nama_petugas_dispensing"`
	Etiket                Etiquette                `json:"etiket" binding:"required" bson:"etiket"`
}

func (dispensing *Dispensing) StatusString() string {
	switch *dispensing.StatusResep {
	case datastruct.PENDING:
		return "pending"
	case datastruct.SUDAH_DIBERIKAN:
		return "sudah diberikan"
	default:
		return ""
	}
}
