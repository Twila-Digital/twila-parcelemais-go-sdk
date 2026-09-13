package parcelemais

// PagedResult é o retorno de qualquer listagem paginada. Sem auto-paginação — o
// consumidor avança de página explicitamente via HasNext/PageNumber.
type PagedResult[T any] struct {
	Items       []T
	HasNext     bool
	HasPrevious bool
	PageNumber  int64
	PageSize    int64
	TotalCount  int64
}
