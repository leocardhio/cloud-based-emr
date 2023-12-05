package outpatient

import (
	"service-outpatient/datastruct/outpatient/consent"
	earlyassessment "service-outpatient/datastruct/outpatient/early-assessment"
	specialityexamination "service-outpatient/datastruct/outpatient/speciality-examination"
	"time"
)

type OutpatientAdult struct {
	CaraPembayaran          string                                      `json:"cara_pembayaran" binding:"required" bson:"cara_pembayaran"`
	PersetujuanUmum         consent.GeneralConsent                      `json:"persetujuan_umum" binding:"required" bson:"persetujuan_umum"`
	AsesmenAwal             earlyassessment.EarlyAssessment             `json:"asesmen_awal" binding:"required" bson:"asesmen_awal"`
	PemeriksaanSpesialistik specialityexamination.SpecialityExamination `json:"pemeriksaan_spesialistik" binding:"required" bson:"pemeriksaan_spesialistik"`

	CreatedAt *time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"-" bson:"deleted_at"`
}
