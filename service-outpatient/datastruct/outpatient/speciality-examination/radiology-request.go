package specialityexamination

import (
	"service-outpatient/datastruct"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RadiologyExaminationResult struct {
	URLFotoHasilPemeriksaan           string `json:"url_foto_hasil_pemeriksaan" bson:"url_foto_hasil_pemeriksaan"`
	DokterPenginterpretasiPemeriksaan string `json:"dokter_penginterpretasi_pemeriksaan" bson:"dokter_penginterpretasi_pemeriksaan"`
	InterpretasiRadiologi             string `json:"interpretasi_radiologi" bson:"interpretasi_radiologi"`
}

type ConfidentialRadiologyRequestData struct {
	WaktuPermintaan time.Time `json:"waktu_permintaan" binding:"required" bson:"waktu_permintaan"`

	DokterPengirim                  string                         `json:"dokter_pengirim" binding:"required" bson:"dokter_pengirim"`
	NoTelpDokterPengirim            string                         `json:"no_telp_dokter_pengirim" binding:"required" bson:"no_telp_dokter_pengirim"`
	NamaFasyankesPengirimPermintaan string                         `json:"nama_fasyankes_pengirim_permintaan" binding:"required" bson:"nama_fasyankes_pengirim_permintaan"`
	UnitPengirimPermintaan          string                         `json:"unit_pengirim_permintaan" binding:"required" bson:"unit_pengirim_permintaan"`
	PrioritasPemeriksaan            datastruct.ExaminationPriority `json:"prioritas_pemeriksaan" binding:"required" bson:"prioritas_pemeriksaan"`
	Diagnosis                       string                         `json:"diagnosis" binding:"required" bson:"diagnosis"`
	CatatanPermintaan               string                         `json:"catatan_permintaan" binding:"required" bson:"catatan_permintaan"`

	MetodePengiriman  datastruct.SendingMethod   `json:"metode_pengiriman" bson:"metode_pengiriman"`
	StatusAlergi      bool                       `json:"status_alergi" bson:"status_alergi"`
	StatusKehamilan   bool                       `json:"status_kehamilan" bson:"status_kehamilan"`
	WaktuPemeriksaan  time.Time                  `json:"waktu_pemeriksaan" bson:"waktu_pemeriksaan"`
	JenisBahanKontras string                     `json:"jenis_bahan_kontras" bson:"jenis_bahan_kontras"`
	HasilPemeriksaan  RadiologyExaminationResult `json:"hasil_pemeriksaan" bson:"hasil_pemeriksaan"`
}

type RadiologyRequest struct {
	Signature        *string                             `json:"signature" bson:"signature"`
	NoIHS            string                              `json:"no_ihs" binding:"required" bson:"no_ihs"`
	NamaPemeriksaan  string                              `json:"nama_pemeriksaan" binding:"required" bson:"nama_pemeriksaan"`
	JenisPemeriksaan datastruct.RadiologyExaminationType `json:"jenis_pemeriksaan" binding:"required" bson:"jenis_pemeriksaan"`
	// NoPermintaan     string                              `json:"no_permintaan" binding:"required" bson:"no_permintaan"`

	ConfidentialData      *ConfidentialRadiologyRequestData `json:"confidential_data" binding:"required" bson:"confidential_data,omitempty"`
	ConfidentialEncrypted *primitive.Binary                 `json:"encrypted_confidential" bson:"encrypted_confidential"`

	CreatedAt *time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"-" bson:"deleted_at"`
}
