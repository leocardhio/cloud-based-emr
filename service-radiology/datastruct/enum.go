package datastruct

type SexType uint8
type EduEnum uint8
type MarriageEnum uint8
type RecipeStatus uint8
type ExaminationPriority uint8
type SendingMethod uint8
type AbnormalitiesEnum uint8
type ConsciousnessLevel uint8
type RadiologyExaminationType string
type TreatmentType uint8
type RoleType string
type PatientConsent bool

const (
	UNKNOWN SexType = iota
	MALE
	FEMALE
	UNDEFINED
	NOTFILLED
)

const (
	TIDAKSEKOLAH EduEnum = iota
	SD
	SLTP
	SLTA
	D1_D3
	D4
	S1
	S2
	S3
)

const (
	BELUMKAWIN MarriageEnum = iota + 1
	KAWIN
	CERAI_HIDUP
	CERAI_MATI
)

const (
	PENDING RecipeStatus = iota
	SUDAH_DIBERIKAN
)

const (
	CITO ExaminationPriority = iota + 1
	NON_CITO
)

const (
	PENYERAHAN_LANGSUNG SendingMethod = iota + 1
	VIA_SUREL
)

const (
	NORMAL AbnormalitiesEnum = iota + 1
	TIDAK_NORMAL
)

const (
	ALERT ConsciousnessLevel = iota
	VOICE
	PAIN
	UNRESPONSIVE
	CONFUSE
	ACUTE_CONFUSION
)

const (
	CRANIUM                 RadiologyExaminationType = "1. Cranium"
	GIGI_GELIGI             RadiologyExaminationType = "2. Gigi Geligi"
	VERTEBRA                RadiologyExaminationType = "3. Vertebra"
	BADAN                   RadiologyExaminationType = "4. Badan"
	EKSTREMITAS_ATAS        RadiologyExaminationType = "5. Extremitas Atas"
	EKSTREMITAS_BAWAH       RadiologyExaminationType = "6. Extremitas Bawah"
	KONTRAS_SALURAN_CERNA   RadiologyExaminationType = "7. Kontras Saluran Cerna"
	KONTRAS_SALURAN_KENCING RadiologyExaminationType = "8. Kontras Saluran Kencing"
)

const (
	IGD TreatmentType = iota
	RAWAT_INAP
	RAWAT_JALAN
)

const (
	DOKTER       RoleType = "Dokter"
	APOTEK       RoleType = "Apotek"
	LABORATORIUM RoleType = "Laboratorium"
	RADIOLOGI    RoleType = "Radiologi"
)

const (
	OPTIN  PatientConsent = true
	OPTOUT PatientConsent = false
)
