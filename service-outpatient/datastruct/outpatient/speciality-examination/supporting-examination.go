package specialityexamination

import (
	"service-outpatient/datastruct"
	"time"
)

type SupportingExamination struct {
	NoRekamMedis     string              `json:"no_rekam_medis" binding:"required" bson:"no_rekam_medis"`
	NamaPasien       string              `json:"nama_pasien" binding:"required" bson:"nama_pasien"`
	NIK              uint64              `json:"nik" binding:"required" bson:"nik"`
	TanggalLahir     time.Time           `json:"tanggal_lahir" binding:"required" bson:"tanggal_lahir"`
	JenisKelamin     *datastruct.SexType `json:"jenis_kelamin" binding:"required" bson:"jenis_kelamin"`
	WaktuPemeriksaan time.Time           `json:"waktu_pemeriksaan" binding:"required" bson:"waktu_pemeriksaan"`
	StatusPuasa      bool                `json:"status_puasa" bson:"status_puasa"`

	LabResultRefId        *string            `json:"lab_result_ref_id" bson:"lab_result_ref_id,omitempty"` //ketika dokter make request, isi sebagian data di lab collection
	LabHTTPResponseStatus *string            `json:"lab_http_response_status" bson:"lab_http_response_status"`
	Laboratorium          *LaboratoryRequest `json:"laboratorium" bson:"laboratorium,omitempty"` //dokter create request ke LAB

	RadiologiResultRefId        *string           `json:"radiologi_result_ref_id" bson:"radiologi_result_ref_id,omitempty"` //ketika dokter make request, isi sebagian data di radiologi collection
	RadiologiHTTPResponseStatus *string           `json:"radiologi_http_response_status" bson:"radiologi_http_response_status"`
	Radiologi                   *RadiologyRequest `json:"radiologi" bson:"radiologi,omitempty"` //dokter create request ke RADIOLOGI
}
