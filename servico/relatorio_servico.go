package servico

import (
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/enums"
	"github.com/ifinu/ifinu-api-go/repositorio"
)

type RelatorioServico struct {
	clienteRepo  *repositorio.ClienteRepositorio
	cobrancaRepo *repositorio.CobrancaRepositorio
}

func NovoRelatorioServico(clienteRepo *repositorio.ClienteRepositorio, cobrancaRepo *repositorio.CobrancaRepositorio) *RelatorioServico {
	return &RelatorioServico{
		clienteRepo:  clienteRepo,
		cobrancaRepo: cobrancaRepo,
	}
}

// ObterHistoricoPagamentos retorna histórico completo de pagamentos com paginação
func (s *RelatorioServico) ObterHistoricoPagamentos(usuarioID uuid.UUID, pagina, tamanhoPagina int) (map[string]interface{}, error) {
	if pagina < 1 {
		pagina = 1
	}
	if tamanhoPagina < 1 {
		tamanhoPagina = 20
	}

	cobrancas, err := s.cobrancaRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		return nil, err
	}

	var pagamentosList []map[string]interface{}
	var valorTotalPago float64

	for _, cobranca := range cobrancas {
		if cobranca.Status == enums.StatusCobrancaPago && cobranca.DataPagamento != nil {
			pagamentosList = append(pagamentosList, map[string]interface{}{
				"cobrancaId":     cobranca.ID,
				"clienteNome":    cobranca.Cliente.Nome,
				"clienteId":      cobranca.ClienteID,
				"valor":          cobranca.Valor,
				"dataPagamento":  cobranca.DataPagamento,
				"dataVencimento": cobranca.DataVencimento,
				"descricao":      cobranca.Descricao,
				"tipoRecorrencia": cobranca.TipoRecorrencia,
			})
			valorTotalPago += cobranca.Valor
		}
	}

	sort.Slice(pagamentosList, func(i, j int) bool {
		dataI := pagamentosList[i]["dataPagamento"].(*time.Time)
		dataJ := pagamentosList[j]["dataPagamento"].(*time.Time)
		return dataI.After(*dataJ)
	})

	total := len(pagamentosList)
	inicio := (pagina - 1) * tamanhoPagina
	fim := inicio + tamanhoPagina

	if inicio > total {
		inicio = total
	}
	if fim > total {
		fim = total
	}

	pagamentosPaginados := []map[string]interface{}{}
	if inicio < total {
		pagamentosPaginados = pagamentosList[inicio:fim]
	}

	totalPaginas := int(math.Ceil(float64(total) / float64(tamanhoPagina)))

	return map[string]interface{}{
		"pagamentos":     pagamentosPaginados,
		"total":          total,
		"pagina":         pagina - 1,
		"tamanhoPagina":  tamanhoPagina,
		"totalPaginas":   totalPaginas,
		"valorTotalPago": valorTotalPago,
	}, nil
}

// ObterDashboard retorna estatísticas completas do dashboard
func (s *RelatorioServico) ObterDashboard(usuarioID uuid.UUID) (map[string]interface{}, error) {
	cobrancas, err := s.cobrancaRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		return nil, err
	}

	totalCobrancas := int64(len(cobrancas))

	var valorTotal, valorPago, valorPendente, valorVencido, valorPagoMes float64
	var totalPagas, totalPendentes, totalVencidas int64

	agora := time.Now()
	inicioMes := time.Date(agora.Year(), agora.Month(), 1, 0, 0, 0, 0, agora.Location())

	clientesMap := make(map[uuid.UUID]struct {
		nome          string
		totalCobrancas int64
		totalPago     float64
		totalPendente float64
	})

	evolucaoMap := make(map[string]struct {
		mes           string
		ano           int
		receita       float64
		qtdCobrancas  int64
		qtdPagas      int64
	})

	type UltimoPagamento struct {
		cobrancaID     uuid.UUID
		clienteNome    string
		valor          float64
		dataPagamento  time.Time
		descricao      string
	}
	var ultimosPagamentos []UltimoPagamento

	for _, cobranca := range cobrancas {
		valorTotal += cobranca.Valor

		if cobranca.Status == enums.StatusCobrancaPago {
			valorPago += cobranca.Valor
			totalPagas++

			if cobranca.DataPagamento != nil && cobranca.DataPagamento.After(inicioMes) {
				valorPagoMes += cobranca.Valor
			}

			if cobranca.DataPagamento != nil {
				ultimosPagamentos = append(ultimosPagamentos, UltimoPagamento{
					cobrancaID:    cobranca.ID,
					clienteNome:   cobranca.Cliente.Nome,
					valor:         cobranca.Valor,
					dataPagamento: *cobranca.DataPagamento,
					descricao:     cobranca.Descricao,
				})
			}

			mesAno := cobranca.DataVencimento.Format("January")
			key := mesAno + "-" + string(rune(cobranca.DataVencimento.Year()))
			evo := evolucaoMap[key]
			evo.mes = mesAno
			evo.ano = cobranca.DataVencimento.Year()
			evo.receita += cobranca.Valor
			evo.qtdCobrancas++
			evo.qtdPagas++
			evolucaoMap[key] = evo
		} else if cobranca.Status == enums.StatusCobrancaPendente {
			valorPendente += cobranca.Valor
			totalPendentes++
		} else if cobranca.Status == enums.StatusCobrancaVencido {
			valorVencido += cobranca.Valor
			totalVencidas++
		}

		cliente := clientesMap[cobranca.ClienteID]
		cliente.nome = cobranca.Cliente.Nome
		cliente.totalCobrancas++
		if cobranca.Status == enums.StatusCobrancaPago {
			cliente.totalPago += cobranca.Valor
		} else {
			cliente.totalPendente += cobranca.Valor
		}
		clientesMap[cobranca.ClienteID] = cliente
	}

	taxaConversao := 0.0
	if totalCobrancas > 0 {
		taxaConversao = (float64(totalPagas) / float64(totalCobrancas)) * 100
	}

	taxaInadimplencia := 0.0
	if totalCobrancas > 0 {
		taxaInadimplencia = (float64(totalVencidas) / float64(totalCobrancas)) * 100
	}

	sort.Slice(ultimosPagamentos, func(i, j int) bool {
		return ultimosPagamentos[i].dataPagamento.After(ultimosPagamentos[j].dataPagamento)
	})

	ultimosPagamentosSlice := make([]map[string]interface{}, 0)
	for i, p := range ultimosPagamentos {
		if i >= 5 {
			break
		}
		ultimosPagamentosSlice = append(ultimosPagamentosSlice, map[string]interface{}{
			"cobrancaId":     p.cobrancaID,
			"clienteNome":    p.clienteNome,
			"valor":          p.valor,
			"dataPagamento":  p.dataPagamento,
			"descricao":      p.descricao,
		})
	}

	type TopCliente struct {
		clienteID      uuid.UUID
		clienteNome    string
		totalCobrancas int64
		totalPago      float64
		totalPendente  float64
	}
	var topClientesSlice []TopCliente
	for id, c := range clientesMap {
		topClientesSlice = append(topClientesSlice, TopCliente{
			clienteID:      id,
			clienteNome:    c.nome,
			totalCobrancas: c.totalCobrancas,
			totalPago:      c.totalPago,
			totalPendente:  c.totalPendente,
		})
	}
	sort.Slice(topClientesSlice, func(i, j int) bool {
		return (topClientesSlice[i].totalPago + topClientesSlice[i].totalPendente) >
			(topClientesSlice[j].totalPago + topClientesSlice[j].totalPendente)
	})

	topClientes := make([]map[string]interface{}, 0)
	for i, c := range topClientesSlice {
		if i >= 5 {
			break
		}
		topClientes = append(topClientes, map[string]interface{}{
			"clienteId":      c.clienteID,
			"clienteNome":    c.clienteNome,
			"totalCobrancas": c.totalCobrancas,
			"totalPago":      c.totalPago,
			"totalPendente":  c.totalPendente,
		})
	}

	evolucaoMensal := make([]map[string]interface{}, 0)
	for _, evo := range evolucaoMap {
		taxaConversaoMes := 0.0
		if evo.qtdCobrancas > 0 {
			taxaConversaoMes = (float64(evo.qtdPagas) / float64(evo.qtdCobrancas)) * 100
		}
		evolucaoMensal = append(evolucaoMensal, map[string]interface{}{
			"mes":            evo.mes,
			"ano":            evo.ano,
			"receita":        evo.receita,
			"qtdCobrancas":   evo.qtdCobrancas,
			"qtdPagas":       evo.qtdPagas,
			"taxaConversao":  taxaConversaoMes,
		})
	}

	return map[string]interface{}{
		"metricas": map[string]interface{}{
			"totalPendentes":     totalPendentes,
			"totalPagas":         totalPagas,
			"totalVencidas":      totalVencidas,
			"totalGeral":         totalCobrancas,
			"valorPendente":      valorPendente,
			"valorPago":          valorPago,
			"valorPagoMes":       valorPagoMes,
			"valorTotal":         valorTotal,
			"taxaInadimplencia":  taxaInadimplencia,
			"taxaConversao":      taxaConversao,
		},
		"evolucaoMensal": evolucaoMensal,
		"topClientes":    topClientes,
		"distribuicaoStatus": map[string]interface{}{
			"pendentes": totalPendentes,
			"pagas":     totalPagas,
			"vencidas":  totalVencidas,
		},
		"resumoPeriodo": map[string]interface{}{
			"mesAtual":       agora.Format("January"),
			"ano":            agora.Year(),
			"cobrancasDoMes": len(cobrancas),
			"receitaDoMes":   valorPagoMes,
		},
		"ultimosPagamentos": ultimosPagamentosSlice,
	}, nil
}
