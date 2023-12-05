package identity

import (
	"service-outpatient/datastruct"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Address struct {
	Alamat        string `json:"alamat" binding:"required" bson:"alamat"`
	RT            string `json:"rt" binding:"required" bson:"rt"`
	RW            string `json:"rw" binding:"required" bson:"rw"`
	KelurahanDesa string `json:"kelurahan_desa" binding:"required" bson:"kelurahan_desa"`
	Kecamatan     string `json:"kecamatan" binding:"required" bson:"kecamatan"`
	KotaKabupaten string `json:"kota_kabupaten" binding:"required" bson:"kota_kabupaten"`
	KodePos       uint32 `json:"kode_pos" binding:"required" bson:"kode_pos"`
	Provinsi      uint8  `json:"provinsi" binding:"required" bson:"provinsi"`
	Negara        string `json:"negara" binding:"required" bson:"negara"`
}

type ConfidentialIdentityData struct {
	NamaIbu      string              `json:"nama_ibu" binding:"required" bson:"nama_ibu"`
	TempatLahir  string              `json:"tempat_lahir" binding:"required" bson:"tempat_lahir"`
	TanggalLahir time.Time           `json:"tanggal_lahir" binding:"required" bson:"tanggal_lahir"`
	JenisKelamin *datastruct.SexType `json:"jenis_kelamin" binding:"required" bson:"jenis_kelamin"`
	Agama        string              `json:"agama" binding:"required" bson:"agama"`
	Suku         string              `json:"suku" binding:"required" bson:"suku"`

	AlamatIdentitas Address `json:"alamat_identitas" binding:"required" bson:"alamat_identitas"`
	AlamatDomisili  Address `json:"alamat_domisili" binding:"required" bson:"alamat_domisili"`

	NoTelpRumah   string `json:"no_telp_rumah" binding:"required" bson:"no_telp_rumah"`
	NoTelpSelular string `json:"no_telp_selular" binding:"required" bson:"no_telp_selular"`

	Pendidikan *datastruct.EduEnum `json:"pendidikan" binding:"required" bson:"pendidikan"`
	Pekerjaan  string              `json:"pekerjaan" binding:"required" bson:"pekerjaan"`

	StatusPernikahan datastruct.MarriageEnum `json:"status_pernikahan" binding:"required" bson:"status_pernikahan"`
	BahasaDikuasai   string                  `json:"bahasa_dikuasai" binding:"required" bson:"bahasa_dikuasai"`
}

type AdultPatient struct {
	NoIHS         string            `json:"no_ihs" binding:"required" bson:"no_ihs"`
	NamaLengkap   *string           `json:"nama_lengkap" binding:"required" bson:"nama_lengkap,omitempty"`
	NamaEncrypted *primitive.Binary `json:"encrypted_nama" bson:"encrypted_nama,omitempty"`

	NoRekamMedis string            `json:"no_rekam_medis" binding:"required" bson:"no_rekam_medis"`
	NIK          *uint64           `json:"nik" binding:"required" bson:"nik,omitempty"`
	NIKEncrypted *primitive.Binary `json:"encrypted_nik" bson:"encrypted_nik"`

	IdentitasLain          *string           `json:"identitas_lain" binding:"required" bson:"identitas_lain,omitempty"`
	IdentitasLainEncrypted *primitive.Binary `json:"encrypted_identitas_lain" bson:"encrypted_identitas_lain"`

	ConfidentialData      *ConfidentialIdentityData `json:"confidential_data" binding:"required" bson:"confidential_data,omitempty"`
	ConfidentialEncrypted *primitive.Binary         `json:"encrypted_confidential" bson:"encrypted_confidential"`

	CreatedAt *time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"-" bson:"deleted_at"`
}

func (adultPatient *AdultPatient) EduString() string {
	switch *adultPatient.ConfidentialData.Pendidikan {
	case datastruct.TIDAKSEKOLAH:
		return "Tidak sekolah"
	case datastruct.SD:
		return "SD"
	case datastruct.SLTP:
		return "SLTP sederajat"
	case datastruct.SLTA:
		return "SLTA sederajat"
	case datastruct.D1_D3:
		return "D1-D3 sederajat"
	case datastruct.D4:
		return "D4"
	case datastruct.S1:
		return "S1"
	case datastruct.S2:
		return "S2"
	case datastruct.S3:
		return "S3"
	default:
		return ""
	}
}
