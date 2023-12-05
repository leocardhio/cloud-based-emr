package datastruct

type RecipeStatus uint8
type RoleType string
type PatientConsent bool

const (
	PENDING RecipeStatus = iota
	SUDAH_DIBERIKAN
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
