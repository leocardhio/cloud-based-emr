package consent

import (
	"service-outpatient/datastruct"
	"time"
)

type InformationSecrecy struct {
	DiberikanKePenjamin bool `json:"diberikan_ke_penjamin" bson:"diberikan_ke_penjamin"`
	DiaksesPesertaDidik bool `json:"diakses_peserta_didik" bson:"diakses_peserta_didik"`
	DiberikanKeKeluarga bool `json:"diberikan_ke_keluarga" bson:"diberikan_ke_keluarga"`
	UntukRujukan        bool `json:"untuk_rujukan" bson:"untuk_rujukan"`
}

type PatientConsent struct {
	KetentuanPembayaran bool               `json:"ketentuan_pembayaran" bson:"ketentuan_pembayaran"`
	HakKewajiban        bool               `json:"hak_kewajiban" bson:"hak_kewajiban"`
	TataTertibRS        bool               `json:"tata_tertib_rs" bson:"tata_tertib_rs"`
	KebutuhanPenerjemah bool               `json:"kebutuhan_penerjemah" bson:"kebutuhan_penerjemah"`
	KebutuhanRohaniawan bool               `json:"kebutuhan_rohaniawan" bson:"kebutuhan_rohaniawan"`
	PelepasanInformasi  InformationSecrecy `json:"pelepasan_informasi" binding:"required" bson:"pelepasan_informasi"`
}

type GeneralConsent struct {
	NamaLengkap       string              `json:"nama_lengkap" binding:"required" bson:"nama_lengkap"`
	TanggalLahir      time.Time           `json:"tanggal_lahir" binding:"required" bson:"tanggal_lahir"`
	JenisKelamin      *datastruct.SexType `json:"jenis_kelamin" binding:"required" bson:"jenis_kelamin"`
	NoRekamMedis      string              `json:"no_rekam_medis" binding:"required" bson:"no_rekam_medis"`
	PersetujuanPasien PatientConsent      `json:"persetujuan_pasien" binding:"required" bson:"persetujuan_pasien"`

	PenanggungJawab string `json:"penanggung_jawab" binding:"required" bson:"penanggung_jawab"`
	PetugasConsent  string `json:"petugas_consent" binding:"required" bson:"petugas_consent"`
}
