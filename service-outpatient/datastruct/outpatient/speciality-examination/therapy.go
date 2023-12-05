package specialityexamination

import (
	"time"
)

type Action struct {
	NamaTindakan       string    `json:"nama_tindakan" binding:"required" bson:"nama_tindakan"`
	PetugasPelaksana   string    `json:"petugas_pelaksana" binding:"required" bson:"petugas_pelaksana"`
	TanggalPelaksanaan time.Time `json:"tanggal_pelaksanaan" binding:"required" bson:"tanggal_pelaksanaan"`
	WaktuMulai         time.Time `json:"waktu_mulai" binding:"required" bson:"waktu_mulai"`
	WaktuSelesai       time.Time `json:"waktu_selesai" binding:"required" bson:"waktu_selesai"`
	AlatMedisDigunakan string    `json:"alat_medis_digunakan" binding:"required" bson:"alat_medis_digunakan"`
	BMHP               string    `json:"bmhp" binding:"required" bson:"bmhp"`
}

type Therapy struct {
	Tindakan           Action             `json:"tindakan" binding:"required" bson:"tindakan"`
	ResepObatRefId     *string            `json:"resep_obat_ref_id" bson:"resep_obat_ref_id"`
	HTTPResponseStatus *string            `json:"http_response_status" bson:"http_response_status"`
	ResepObat          *DrugRecipeRequest `json:"resep_obat" bson:"resep_obat,omitempty"` //bisa empty on create, harus ada ketika get (beda collection)
}
