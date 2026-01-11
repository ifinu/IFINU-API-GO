package enums

type StatusCobranca string

const (
	StatusCobrancaPendente  StatusCobranca = "PENDENTE"
	StatusCobrancaPago      StatusCobranca = "PAGO"
	StatusCobrancaVencido   StatusCobranca = "VENCIDO"
	StatusCobrancaCancelado StatusCobranca = "CANCELADO"
)

func (s StatusCobranca) String() string {
	return string(s)
}

func (s StatusCobranca) Valido() bool {
	switch s {
	case StatusCobrancaPendente, StatusCobrancaPago, StatusCobrancaVencido, StatusCobrancaCancelado:
		return true
	}
	return false
}
