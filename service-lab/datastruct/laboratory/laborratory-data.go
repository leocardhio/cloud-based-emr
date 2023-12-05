package laboratory

import (
	"service-lab/datastruct"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LabExaminationResult struct {
	NilaiHasil   string                       `json:"nilai_hasil" binding:"required" bson:"nilai_hasil"`
	NilaiNormal  datastruct.AbnormalitiesEnum `json:"nilai_normal" binding:"required" bson:"nilai_normal"`
	NilaiRujukan string                       `json:"nilai_rujukan" binding:"required" bson:"nilai_rujukan"`
	NilaiKritis  string                       `json:"nilai_kritis" binding:"required" bson:"nilai_kritis"`
}

type ConfidentialLabData struct {
	TanggalLahir    time.Time           `json:"tanggal_lahir" binding:"required" bson:"tanggal_lahir"`
	JenisKelamin    *datastruct.SexType `json:"jenis_kelamin" binding:"required" bson:"jenis_kelamin"`
	WaktuPermintaan time.Time           `json:"waktu_permintaan" binding:"required" bson:"waktu_permintaan"`

	DokterPengirim                  string                         `json:"dokter_pengirim" bson:"dokter_pengirim"`
	NoTelpDokterPengirim            string                         `json:"no_telp_dokter_pengirim" bson:"no_telp_dokter_pengirim"`
	NamaFasyankesPengirimPermintaan string                         `json:"nama_fasyankes_pengirim_permintaan" bson:"nama_fasyankes_pengirim_permintaan"`
	UnitPengirimPermintaan          string                         `json:"unit_pengirim_permintaan" bson:"unit_pengirim_permintaan"`
	PrioritasPemeriksaan            datastruct.ExaminationPriority `json:"prioritas_pemeriksaan" bson:"prioritas_pemeriksaan"`
	Diagnosis                       string                         `json:"diagnosis" bson:"diagnosis"`
	CatatanPermintaan               string                         `json:"catatan_permintaan" bson:"catatan_permintaan"`

	MetodePengiriman          datastruct.SendingMethod `json:"metode_pengiriman" binding:"required" bson:"metode_pengiriman"`
	SumberSpesimen            string                   `json:"sumber_spesimen" binding:"required" bson:"sumber_spesimen"`
	LokasiPengambilanSpesimen string                   `json:"lokasi_pengambilan_spesimen" binding:"required" bson:"lokasi_pengambilan_spesimen"`
	JumlahSpesimen            uint8                    `json:"jumlah_spesimen" binding:"required" bson:"jumlah_spesimen"`
	VolumeSpesimen            uint16                   `json:"volume_spesimen" binding:"required" bson:"volume_spesimen"`
	MetodePengambilanSpesimen string                   `json:"metode_pengambilan_spesimen" binding:"required" bson:"metode_pengambilan_spesimen"`
	WaktuPengambilanSpesimen  time.Time                `json:"waktu_pengambilan_spesimen" binding:"required" bson:"waktu_pengambilan_spesimen"`
	KondisiSpesimen           string                   `json:"kondisi_spesimen" binding:"required" bson:"kondisi_spesimen"`

	WaktuFiksasi  time.Time `json:"waktu_fiksasi" binding:"required" bson:"waktu_fiksasi"`
	CairanFiksasi string    `json:"cairan_fiksasi" binding:"required" bson:"cairan_fiksasi"`
	VolumeCairan  uint16    `json:"volume_cairan" binding:"required" bson:"volume_cairan"`

	PetugasPengambilSpesimen    string `json:"petugas_pengambil_spesimen" binding:"required" bson:"petugas_pengambil_spesimen"`
	PetugasPengantarSpesimen    string `json:"petugas_pengantar_spesimen" binding:"required" bson:"petugas_pengantar_spesimen"`
	PetugasPenerimaSpesimen     string `json:"petugas_penerima_spesimen" binding:"required" bson:"petugas_penerima_spesimen"`
	PetugasPenganalisisSpesimen string `json:"petugas_penganalisis_spesimen" binding:"required" bson:"petugas_penganalisis_spesimen"`

	WaktuPengolahanSpesimen time.Time            `json:"waktu_pengolahan_spesimen" binding:"required" bson:"waktu_pengolahan_spesimen"`
	HasilPemeriksaan        LabExaminationResult `json:"hasil_pemeriksaan" binding:"required" bson:"hasil_pemeriksaan"`
	InterpretasiHasil       string               `json:"interpretasi_hasil" binding:"required" bson:"interpretasi_hasil"`

	DokterValidatorPemeriksaan        string `json:"dokter_validator_pemeriksaan" binding:"required" bson:"dokter_validator_pemeriksaan"`
	DokterPenginterpretasiPemeriksaan string `json:"dokter_penginterpretasi_pemeriksaan" binding:"required" bson:"dokter_penginterpretasi_pemeriksaan"`

	WaktuHasilKeluarLab            time.Time `json:"waktu_hasil_keluar_lab" binding:"required" bson:"waktu_hasil_keluar_lab"`
	WaktuHasilDiterimaUnitPengirim time.Time `json:"waktu_hasil_diterima_unit_pengirim" binding:"required" bson:"waktu_hasil_diterima_unit_pengirim"`
}

type LaboratoryData struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	ClientID        string  `json:"client_id" bson:"client_id"`
	Signature       *string `json:"signature" bson:"signature"`
	NoRegistrasiLab string  `json:"no_registrasi_lab" binding:"required" bson:"no_registrasi_lab"`

	NIK          *uint64           `json:"nik" binding:"required" bson:"nik,omitempty"`
	NIKEncrypted *primitive.Binary `json:"encrypted_nik,omitempty" bson:"encrypted_nik"`

	// IDPelanggan string `json:"id_pelanggan" binding:"required" bson:"id_pelanggan"`
	// NoPermintaan           *string `json:"no_permintaan" binding:"required" bson:"no_permintaan"`
	NamaPemeriksaan        string `json:"nama_pemeriksaan" binding:"required" bson:"nama_pemeriksaan"`
	NoIHS                  string `json:"no_ihs" binding:"required" bson:"no_ihs"`
	NamaFasyankesPemeriksa string `json:"nama_fasyankes_pemeriksa" binding:"required" bson:"nama_fasyankes_pemeriksa"`

	ConfidentialData      *ConfidentialLabData `json:"confidential_data" binding:"required" bson:"confidential_data,omitempty"`
	ConfidentialEncrypted *primitive.Binary    `json:"encrypted_confidential,omitempty" bson:"encrypted_confidential"`

	CreatedAt *time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"-" bson:"deleted_at"`
}

func (laboratoryData *LaboratoryData) PriorityString() string {
	switch laboratoryData.ConfidentialData.PrioritasPemeriksaan {
	case datastruct.CITO:
		return "CITO"
	case datastruct.NON_CITO:
		return "Non CITO"
	default:
		return ""
	}
}

func (laboratoryData *LaboratoryData) SendingMethodString() string {
	switch laboratoryData.ConfidentialData.MetodePengiriman {
	case datastruct.PENYERAHAN_LANGSUNG:
		return "Penyerahan langsung"
	case datastruct.VIA_SUREL:
		return "Dikirim via surel"
	default:
		return ""
	}
}

func (examinationResult *LabExaminationResult) NormalResultString() string {
	switch examinationResult.NilaiNormal {
	case datastruct.NORMAL:
		return "Normal"
	case datastruct.TIDAK_NORMAL:
		return "Tidak Normal"
	default:
		return ""
	}
}
