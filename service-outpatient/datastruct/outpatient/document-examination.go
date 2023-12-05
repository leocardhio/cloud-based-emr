package outpatient

import (
	"service-outpatient/datastruct/outpatient/consent"
	earlyassessment "service-outpatient/datastruct/outpatient/early-assessment"
	specialityexamination "service-outpatient/datastruct/outpatient/speciality-examination"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConfidentialExaminationData struct {
	CaraPembayaran          string                                      `json:"cara_pembayaran" binding:"required" bson:"cara_pembayaran"`
	PersetujuanUmum         consent.GeneralConsent                      `json:"persetujuan_umum" binding:"required" bson:"persetujuan_umum"`
	AsesmenAwal             earlyassessment.EarlyAssessment             `json:"asesmen_awal" binding:"required" bson:"asesmen_awal"`
	PemeriksaanSpesialistik specialityexamination.SpecialityExamination `json:"pemeriksaan_spesialistik" binding:"required" bson:"pemeriksaan_spesialistik"`
}

type ExaminationDocument struct {
	ID primitive.ObjectID `json:"_id" bson:"_id,omitempty"`

	ClientID  string  `json:"client_id" bson:"client_id"`
	Signature *string `json:"signature" bson:"signature"`
	NoIHS     string  `json:"no_ihs" binding:"required" bson:"no_ihs"`

	ConfidentialData      *ConfidentialExaminationData `json:"confidential_data" binding:"required" bson:"confidential_data,omitempty"`
	ConfidentialEncrypted *primitive.Binary            `json:"encrypted_confidential" bson:"encrypted_confidential"`

	CreatedAt *time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"-" bson:"deleted_at"`
}
