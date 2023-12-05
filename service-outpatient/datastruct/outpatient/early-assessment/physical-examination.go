package earlyassessment

import (
	"service-outpatient/datastruct"
)

type BloodPressure struct {
	Sistole  uint8 `json:"sistole" binding:"required" bson:"sistole"`
	Diastole uint8 `json:"diastole" binding:"required" bson:"diastole"`
}

type VitalSign struct {
	DenyutJantung string        `json:"denyut_jantung" binding:"required" bson:"denyut_jantung"`
	Pernapasan    string        `json:"pernapasan" binding:"required" bson:"pernapasan"`
	TekananDarah  BloodPressure `json:"tekanan_darah" binding:"required" bson:"tekanan_darah"`
	SuhuTubuh     uint8         `json:"suhu_tubuh" binding:"required" bson:"suhu_tubuh"`

	Kepala           string `json:"kepala" bson:"kepala"`
	Mata             string `json:"mata" bson:"mata"`
	Telinga          string `json:"telinga" bson:"telinga"`
	Hidung           string `json:"hidung" bson:"hidung"`
	Rambut           string `json:"rambut" bson:"rambut"`
	Bibir            string `json:"bibir" bson:"bibir"`
	GigiGeligi       string `json:"gigi_geligi" bson:"gigi_geligi"`
	Lidah            string `json:"lidah" bson:"lidah"`
	LangitLangit     string `json:"langit_langit" bson:"langit_langit"`
	Leher            string `json:"leher" bson:"leher"`
	Tenggorokan      string `json:"tenggorokan" bson:"tenggorokan"`
	Tonsil           string `json:"tonsil" bson:"tonsil"`
	Dada             string `json:"dada" bson:"dada"`
	Payudara         string `json:"payudara" bson:"payudara"`
	Punggung         string `json:"punggung" bson:"punggung"`
	Perut            string `json:"perut" bson:"perut"`
	Genital          string `json:"genital" bson:"genital"`
	Anus             string `json:"anus" bson:"anus"`
	LenganAtas       string `json:"lengan_atas" bson:"lengan_atas"`
	LenganBawah      string `json:"lengan_bawah" bson:"lengan_bawah"`
	JariTangan       string `json:"jari_tangan" bson:"jari_tangan"`
	KukuTangan       string `json:"kuku_tangan" bson:"kuku_tangan"`
	PersendianTangan string `json:"persendian_tangan" bson:"persendian_tangan"`
	TungkaiAtas      string `json:"tungkai_atas" bson:"tungkai_atas"`
	TungkaiBawah     string `json:"tungkai_bawah" bson:"tungkai_bawah"`
	JariKaki         string `json:"jari_kaki" bson:"jari_kaki"`
	KukuKaki         string `json:"kuku_kaki" bson:"kuku_kaki"`
	PersendianKaki   string `json:"persendian_kaki" bson:"persendian_kaki"`
}

type GeneralCondition struct {
	TingkatKesadaran *datastruct.ConsciousnessLevel `json:"tingkat_kesadaran" binding:"required" bson:"tingkat_kesadaran"`
	VitalSign        VitalSign                      `json:"vital_sign" binding:"required" bson:"vital_sign"`
}

type PhysicalExamination struct {
	URLAnatomiTubuh string           `json:"url_anatomi_tubuh" binding:"required" bson:"url_anatomi_tubuh"`
	KeadaanUmum     GeneralCondition `json:"keadaan_umum" binding:"required" bson:"keadaan_umum"`
}

func (generalCondition *GeneralCondition) ConsciousnessString() string {
	switch *generalCondition.TingkatKesadaran {
	case datastruct.ALERT:
		return "Alert"
	case datastruct.VOICE:
		return "Voice"
	case datastruct.PAIN:
		return "Pain"
	case datastruct.UNRESPONSIVE:
		return "Unresponsive"
	case datastruct.CONFUSE:
		return "Gelisah atau bingung"
	case datastruct.ACUTE_CONFUSION:
		return "Acute Confusional States"
	default:
		return ""
	}
}
