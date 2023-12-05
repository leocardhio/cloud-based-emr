package pharmacy

import (
	"service-pharmacy/datastruct"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HowToUse struct {
	Metode         string `json:"metode" binding:"required" bson:"metode"`
	DosisPakai     string `json:"dosis_pakai" binding:"required" bson:"dosis_pakai"`
	SatuanDosis    string `json:"satuan_dosis" binding:"required" bson:"satuan_dosis"`
	IntervalPakai  string `json:"interval_pakai" binding:"required" bson:"interval_pakai"`
	AturanTambahan string `json:"aturan_tambahan" binding:"required" bson:"aturan_tambahan"`
}

type UsageHistory struct {
	NamaObat    string `json:"nama_obat" binding:"required" bson:"nama_obat"`
	Bentuk      string `json:"bentuk" binding:"required" bson:"bentuk"`
	DosisPakai  string `json:"dosis_pakai" binding:"required" bson:"dosis_pakai"`
	AturanPakai string `json:"aturan_pakai" binding:"required" bson:"aturan_pakai"`
}

type AdministrationRegulation struct {
	IdentitasPasien string    `json:"identitas_pasien" binding:"required" bson:"identitas_pasien"`
	IdentitasDokter string    `json:"identitas_dokter" binding:"required" bson:"identitas_dokter"`
	TanggalResep    time.Time `json:"tanggal_resep" binding:"required" bson:"tanggal_resep"`
	UnitAsalResep   string    `json:"unit_asal_resep" binding:"required" bson:"unit_asal_resep"`
}

type PharmaceuticalRegulation struct {
	IdentitasObat  string `json:"identitas_obat" binding:"required" bson:"identitas_obat"`
	DosisObat      string `json:"dosis_obat" binding:"required" bson:"dosis_obat"`
	Stabilitas     string `json:"stabilitas" binding:"required" bson:"stabilitas"`
	AturanPakai    string `json:"aturan_pakai" binding:"required" bson:"aturan_pakai"`
	Kompatibilitas string `json:"kompatibilitas" binding:"required" bson:"kompatibilitas"`
}

type ClinicalRegulation struct {
	IndikasiDosisPenggunaan string `json:"indikasi_dosis_penggunaan" binding:"required" bson:"indikasi_dosis_penggunaan"`
	AturanCaraPenggunaan    string `json:"aturan_cara_penggunaan" binding:"required" bson:"aturan_cara_penggunaan"`
	DuplikasiPengobatan     string `json:"duplikasi_pengobatan" binding:"required" bson:"duplikasi_pengobatan"`
	AlergiDanROTD           string `json:"alergi_dan_rotd" binding:"required" bson:"alergi_dan_rotd"`
	Kontraindikasi          string `json:"kontraindikasi" binding:"required" bson:"kontraindikasi"`
	InteraksiObat           string `json:"interaksi_obat" binding:"required" bson:"interaksi_obat"`
}

type RecipeAssessment struct {
	Administrasi AdministrationRegulation `json:"administrasi" binding:"required" bson:"administrasi"`
	Farmasetik   PharmaceuticalRegulation `json:"farmasetik" binding:"required" bson:"farmasetik"`
	Klinis       ClinicalRegulation       `json:"klinis" binding:"required" bson:"klinis"`
}

type ConfidentialPharmacyData struct {
	NamaLengkap  string    `json:"nama_lengkap" binding:"required" bson:"nama_lengkap"`
	TanggalLahir time.Time `json:"tanggal_lahir" binding:"required" bson:"tanggal_lahir"`

	TinggiBadan            uint16 `json:"tinggi_badan" binding:"required" bson:"tinggi_badan"`
	BeratBadan             uint16 `json:"berat_badan" binding:"required" bson:"berat_badan"`
	LuasPermukaanTubuhAnak uint16 `json:"luas_permukaan_tubuh_anak" bson:"luas_permukaan_tubuh_anak"`

	NamaObat     string   `json:"nama_obat" binding:"required" bson:"nama_obat"`
	Bentuk       string   `json:"bentuk" binding:"required" bson:"bentuk"`
	JumlahObat   uint     `json:"jumlah_obat" binding:"required" bson:"jumlah_obat"`
	AturanPakai  HowToUse `json:"aturan_pakai" binding:"required" bson:"aturan_pakai"`
	CatatanResep string   `json:"catatan_resep" binding:"required" bson:"catatan_resep"`

	RiwayatPenggunaan UsageHistory `json:"riwayat_penggunaan" binding:"required" bson:"riwayat_penggunaan"`
	RiwayatAlergi     bool         `json:"riwayat_alergi" binding:"required" bson:"riwayat_alergi"`
	JenisAlergi       string       `json:"jenis_alergi" binding:"required" bson:"jenis_alergi"`

	NamaFasyankesPengirim string `json:"nama_fasyankes_pengirim" binding:"required" bson:"nama_fasyankes_pengirim"`
	UnitPengirim          string `json:"unit_pengirim" binding:"required" bson:"unit_pengirim"`
	DokterPenulis         string `json:"dokter_penulis" binding:"required" bson:"dokter_penulis"`
	SIPDokterPenulis      string `json:"sip_dokter_penulis" binding:"required" bson:"sip_dokter_penulis"`
	AlamatDokterPenulis   string `json:"alamat_dokter_penulis" binding:"required" bson:"alamat_dokter_penulis"`
	NoTelpSelularDokter   string `json:"no_telp_selular_dokter" binding:"required" bson:"no_telp_selular_dokter"`

	WaktuPenulisan    time.Time `json:"waktu_penulisan" binding:"required" bson:"waktu_penulisan"`
	TandaTanganDokter string    `json:"tanda_tangan_dokter" binding:"required" bson:"tanda_tangan_dokter"`

	StatusResep     *datastruct.RecipeStatus `json:"status_resep" binding:"required" bson:"status_resep"`
	PengkajianResep RecipeAssessment         `json:"pengkajian_resep" binding:"required" bson:"pengkajian_resep"`
}

type DrugRecipe struct {
	NoRekamMedis string `json:"no_rekam_medis" binding:"required" bson:"no_rekam_medis"`
	IDPelanggan  string `json:"id_pelanggan" binding:"required" bson:"id_pelanggan"`
	IDObat       string `json:"id_obat" binding:"required" bson:"id_obat"`
	NoIHS        string `json:"no_ihs" binding:"required" bson:"no_ihs"`

	NIK          *uint64           `json:"nik" binding:"required" bson:"nik,omitempty"`
	NIKEncrypted *primitive.Binary `json:"encrypted_nik" bson:"encrypted_nik"`

	ConfidentialData      *ConfidentialPharmacyData `json:"confidential_data" binding:"required" bson:"confidential_data,omitempty"`
	ConfidentialEncrypted *primitive.Binary         `json:"encrypted_confidential" bson:"encrypted_confidential"`
}

func (drugRecipe *DrugRecipe) StatusString() string {
	switch *drugRecipe.ConfidentialData.StatusResep {
	case datastruct.PENDING:
		return "pending"
	case datastruct.SUDAH_DIBERIKAN:
		return "sudah diberikan"
	default:
		return ""
	}
}
