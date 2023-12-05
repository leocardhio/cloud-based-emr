package datastruct

type SexType uint8
type ExaminationPriority uint8
type SendingMethod uint8
type AbnormalitiesEnum uint8
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
	DOKTER       RoleType = "Dokter"
	APOTEK       RoleType = "Apotek"
	LABORATORIUM RoleType = "Laboratorium"
	RADIOLOGI    RoleType = "Radiologi"
)

const (
	OPTIN  PatientConsent = true
	OPTOUT PatientConsent = false
)
