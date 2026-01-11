package enums

type TipoRecorrencia string

const (
	TipoRecorrenciaUnica       TipoRecorrencia = "UNICA"
	TipoRecorrenciaMensal      TipoRecorrencia = "MENSAL"
	TipoRecorrenciaTrimestral  TipoRecorrencia = "TRIMESTRAL"
	TipoRecorrenciaSemestral   TipoRecorrencia = "SEMESTRAL"
	TipoRecorrenciaAnual       TipoRecorrencia = "ANUAL"
	TipoRecorrenciaPersonalizado TipoRecorrencia = "PERSONALIZADO"
)

func (t TipoRecorrencia) String() string {
	return string(t)
}

func (t TipoRecorrencia) Valido() bool {
	switch t {
	case TipoRecorrenciaUnica, TipoRecorrenciaMensal, TipoRecorrenciaTrimestral,
		TipoRecorrenciaSemestral, TipoRecorrenciaAnual, TipoRecorrenciaPersonalizado:
		return true
	}
	return false
}
