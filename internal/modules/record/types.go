package record

type RecordType string

const (
	TypeLogin  RecordType = "login"
	TypeText   RecordType = "text"
	TypeBinary RecordType = "binary"
	TypeCard   RecordType = "card"
)

type LoginData struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
	URL      string `json:"url,omitempty"`
}

type TextData struct {
	Text string `json:"text" validate:"required"`
}

type BinaryData struct {
	Blob string `json:"blob" validate:"required,base64"`
}

type CardData struct {
	Number string `json:"number" validate:"required,len=16"`
	Expiry string `json:"expiry" validate:"required,expiry"`
	CVV    string `json:"cvv" validate:"required,len=3,numeric"`
}
