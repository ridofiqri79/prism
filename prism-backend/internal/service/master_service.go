package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/middleware"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type MasterService struct {
	db      *pgxpool.Pool
	queries *queries.Queries
}

var institutionLevels = []string{
	"Kementerian/Badan/Lembaga",
	"Eselon I",
	"Eselon II",
	"BUMN",
	"Pemerintah Daerah Tk. I",
	"Pemerintah Daerah Tk. II",
	"BUMD",
	"Lainya",
}

func NewMasterService(db *pgxpool.Pool, queries *queries.Queries) *MasterService {
	return &MasterService{db: db, queries: queries}
}

func (s *MasterService) ListCountries(ctx context.Context, params model.PaginationParams) (*model.ListResponse[model.CountryResponse], error) {
	page, limit, offset, search, sortField, sortOrder := normalizeMasterList(params, "name", "asc", "code", "name")
	rows, err := s.queries.ListCountries(ctx, queries.ListCountriesParams{
		Limit:     int32(limit),
		Offset:    int32(offset),
		Search:    search,
		SortField: sortField,
		SortOrder: sortOrder,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar country")
	}
	total, err := s.queries.CountCountries(ctx, search)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung country")
	}
	data := make([]model.CountryResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toCountryResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *MasterService) GetCountry(ctx context.Context, id pgtype.UUID) (*model.CountryResponse, error) {
	row, err := s.queries.GetCountry(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Country tidak ditemukan")
	}
	res := toCountryResponse(row)
	return &res, nil
}

func (s *MasterService) CreateCountry(ctx context.Context, req model.CountryRequest) (*model.CountryResponse, error) {
	if err := validateCountry(req); err != nil {
		return nil, err
	}
	var created queries.Country
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.CreateCountry(ctx, queries.CreateCountryParams{Name: req.Name, Code: strings.ToUpper(req.Code)})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	res := toCountryResponse(created)
	return &res, nil
}

func (s *MasterService) UpdateCountry(ctx context.Context, id pgtype.UUID, req model.CountryRequest) (*model.CountryResponse, error) {
	if err := validateCountry(req); err != nil {
		return nil, err
	}
	var updated queries.Country
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdateCountry(ctx, queries.UpdateCountryParams{ID: id, Name: req.Name, Code: strings.ToUpper(req.Code)})
		if err != nil {
			return mapNotFound(err, "Country tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	res := toCountryResponse(updated)
	return &res, nil
}

func (s *MasterService) DeleteCountry(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return fromPg(qtx.DeleteCountry(ctx, id))
	})
}

func (s *MasterService) ListCurrencies(ctx context.Context, params model.PaginationParams, active string) (*model.ListResponse[model.CurrencyResponse], error) {
	page, limit, offset, search, sortField, sortOrder := normalizeMasterList(params, "sort_order", "asc", "code", "name", "sort_order", "is_active")
	activeFilter, err := nullableBool(active)
	if err != nil {
		return nil, validation("active", "harus true atau false")
	}
	rows, err := s.queries.ListCurrencies(ctx, queries.ListCurrenciesParams{
		Limit:        int32(limit),
		Offset:       int32(offset),
		ActiveFilter: activeFilter,
		Search:       search,
		SortField:    sortField,
		SortOrder:    sortOrder,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar currency")
	}
	total, err := s.queries.CountCurrencies(ctx, queries.CountCurrenciesParams{ActiveFilter: activeFilter, Search: search})
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung currency")
	}
	data := make([]model.CurrencyResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toCurrencyResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *MasterService) GetCurrency(ctx context.Context, id pgtype.UUID) (*model.CurrencyResponse, error) {
	row, err := s.queries.GetCurrency(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Currency tidak ditemukan")
	}
	res := toCurrencyResponse(row)
	return &res, nil
}

func (s *MasterService) CreateCurrency(ctx context.Context, req model.CurrencyRequest) (*model.CurrencyResponse, error) {
	if err := validateCurrency(req); err != nil {
		return nil, err
	}
	var created queries.Currency
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.CreateCurrency(ctx, queries.CreateCurrencyParams{
			Code:      strings.ToUpper(strings.TrimSpace(req.Code)),
			Name:      req.Name,
			Symbol:    nullableTextPtr(req.Symbol),
			IsActive:  req.IsActive,
			SortOrder: req.SortOrder,
		})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	res := toCurrencyResponse(created)
	return &res, nil
}

func (s *MasterService) UpdateCurrency(ctx context.Context, id pgtype.UUID, req model.CurrencyRequest) (*model.CurrencyResponse, error) {
	if err := validateCurrency(req); err != nil {
		return nil, err
	}
	var updated queries.Currency
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdateCurrency(ctx, queries.UpdateCurrencyParams{
			ID:        id,
			Code:      strings.ToUpper(strings.TrimSpace(req.Code)),
			Name:      req.Name,
			Symbol:    nullableTextPtr(req.Symbol),
			IsActive:  req.IsActive,
			SortOrder: req.SortOrder,
		})
		if err != nil {
			return mapNotFound(err, "Currency tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	res := toCurrencyResponse(updated)
	return &res, nil
}

func (s *MasterService) DeleteCurrency(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return fromPg(qtx.DeleteCurrency(ctx, id))
	})
}

func (s *MasterService) ListLenders(ctx context.Context, params model.PaginationParams, lenderTypes []string) (*model.ListResponse[model.LenderResponse], error) {
	page, limit, offset, search, sortField, sortOrder := normalizeMasterList(params, "name", "asc", "name", "short_name", "type", "country")
	rows, err := s.queries.ListLenders(ctx, queries.ListLendersParams{
		Limit:       int32(limit),
		Offset:      int32(offset),
		TypeFilters: lenderTypes,
		Search:      search,
		SortField:   sortField,
		SortOrder:   sortOrder,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar lender")
	}
	total, err := s.queries.CountLenders(ctx, queries.CountLendersParams{TypeFilters: lenderTypes, Search: search})
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung lender")
	}
	data := make([]model.LenderResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toLenderListResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *MasterService) GetLender(ctx context.Context, id pgtype.UUID) (*model.LenderResponse, error) {
	row, err := s.queries.GetLender(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Lender tidak ditemukan")
	}
	res := toLenderResponse(row)
	return &res, nil
}

func (s *MasterService) CreateLender(ctx context.Context, req model.CreateLenderRequest) (*model.LenderResponse, error) {
	if err := validateLender(req); err != nil {
		return nil, err
	}
	countryID, err := nullableUUID(req.CountryID)
	if err != nil {
		return nil, validation("country_id", "UUID tidak valid")
	}
	shortName := nullableTextPtr(req.ShortName)
	var created queries.Lender
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.CreateLender(ctx, queries.CreateLenderParams{CountryID: countryID, Name: req.Name, ShortName: shortName, Type: req.Type})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	return s.GetLender(ctx, created.ID)
}

func (s *MasterService) UpdateLender(ctx context.Context, id pgtype.UUID, req model.UpdateLenderRequest) (*model.LenderResponse, error) {
	if err := validateLender(req); err != nil {
		return nil, err
	}
	countryID, err := nullableUUID(req.CountryID)
	if err != nil {
		return nil, validation("country_id", "UUID tidak valid")
	}
	shortName := nullableTextPtr(req.ShortName)
	var updated queries.Lender
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdateLender(ctx, queries.UpdateLenderParams{ID: id, CountryID: countryID, Name: req.Name, ShortName: shortName, Type: req.Type})
		if err != nil {
			return mapNotFound(err, "Lender tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	return s.GetLender(ctx, updated.ID)
}

func (s *MasterService) DeleteLender(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return fromPg(qtx.DeleteLender(ctx, id))
	})
}

func (s *MasterService) ListInstitutions(ctx context.Context, params model.PaginationParams, levels []string, parentID *string) (*model.ListResponse[model.InstitutionResponse], error) {
	page, limit, offset, search, sortField, sortOrder := normalizeMasterList(params, "level", "asc", "name", "short_name", "level")
	parentFilter, err := nullableUUID(parentID)
	if err != nil {
		return nil, validation("parent_id", "UUID tidak valid")
	}
	arg := queries.ListInstitutionsParams{
		Limit:          int32(limit),
		Offset:         int32(offset),
		LevelFilters:   levels,
		ParentIDFilter: parentFilter,
		Search:         search,
		SortField:      sortField,
		SortOrder:      sortOrder,
	}
	rows, err := s.queries.ListInstitutions(ctx, arg)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar institution")
	}
	total, err := s.queries.CountInstitutions(ctx, queries.CountInstitutionsParams{LevelFilters: arg.LevelFilters, ParentIDFilter: arg.ParentIDFilter, Search: search})
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung institution")
	}
	data := make([]model.InstitutionResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toInstitutionResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *MasterService) GetInstitution(ctx context.Context, id pgtype.UUID) (*model.InstitutionResponse, error) {
	row, err := s.queries.GetInstitution(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Institution tidak ditemukan")
	}
	res := toInstitutionDetailResponse(row)
	return &res, nil
}

func (s *MasterService) CreateInstitution(ctx context.Context, req model.InstitutionRequest) (*model.InstitutionResponse, error) {
	if err := validateInstitution(req); err != nil {
		return nil, err
	}
	parentID, err := nullableUUID(req.ParentID)
	if err != nil {
		return nil, validation("parent_id", "UUID tidak valid")
	}
	shortName := nullableTextPtr(req.ShortName)
	var created queries.Institution
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.CreateInstitution(ctx, queries.CreateInstitutionParams{ParentID: parentID, Name: req.Name, ShortName: shortName, Level: req.Level})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	res := toInstitutionResponse(created)
	return &res, nil
}

func (s *MasterService) UpdateInstitution(ctx context.Context, id pgtype.UUID, req model.InstitutionRequest) (*model.InstitutionResponse, error) {
	if err := validateInstitution(req); err != nil {
		return nil, err
	}
	parentID, err := nullableUUID(req.ParentID)
	if err != nil {
		return nil, validation("parent_id", "UUID tidak valid")
	}
	shortName := nullableTextPtr(req.ShortName)
	var updated queries.Institution
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdateInstitution(ctx, queries.UpdateInstitutionParams{ID: id, ParentID: parentID, Name: req.Name, ShortName: shortName, Level: req.Level})
		if err != nil {
			return mapNotFound(err, "Institution tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	res := toInstitutionResponse(updated)
	return &res, nil
}

func (s *MasterService) DeleteInstitution(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return fromPg(qtx.DeleteInstitution(ctx, id))
	})
}

func (s *MasterService) ListRegions(ctx context.Context, params model.PaginationParams, regionTypes []string, parentCode string) (*model.ListResponse[model.RegionResponse], error) {
	page, limit, offset, search, sortField, sortOrder := normalizeMasterList(params, "type", "asc", "code", "name", "type")
	arg := queries.ListRegionsParams{
		Limit:            int32(limit),
		Offset:           int32(offset),
		TypeFilters:      regionTypes,
		ParentCodeFilter: nullableText(parentCode),
		Search:           search,
		SortField:        sortField,
		SortOrder:        sortOrder,
	}
	rows, err := s.queries.ListRegions(ctx, arg)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar region")
	}
	total, err := s.queries.CountRegions(ctx, queries.CountRegionsParams{TypeFilters: arg.TypeFilters, ParentCodeFilter: arg.ParentCodeFilter, Search: search})
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung region")
	}
	data := make([]model.RegionResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toRegionResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *MasterService) GetRegion(ctx context.Context, id pgtype.UUID) (*model.RegionResponse, error) {
	row, err := s.queries.GetRegion(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Region tidak ditemukan")
	}
	res := toRegionResponse(row)
	return &res, nil
}

func (s *MasterService) CreateRegion(ctx context.Context, req model.RegionRequest) (*model.RegionResponse, error) {
	if err := validateRegion(req); err != nil {
		return nil, err
	}
	var created queries.Region
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.CreateRegion(ctx, queries.CreateRegionParams{Code: req.Code, Name: req.Name, Type: req.Type, ParentCode: nullableTextPtr(req.ParentCode)})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	res := toRegionResponse(created)
	return &res, nil
}

func (s *MasterService) UpdateRegion(ctx context.Context, id pgtype.UUID, req model.RegionRequest) (*model.RegionResponse, error) {
	if err := validateRegion(req); err != nil {
		return nil, err
	}
	var updated queries.Region
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdateRegion(ctx, queries.UpdateRegionParams{ID: id, Code: req.Code, Name: req.Name, Type: req.Type, ParentCode: nullableTextPtr(req.ParentCode)})
		if err != nil {
			return mapNotFound(err, "Region tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	res := toRegionResponse(updated)
	return &res, nil
}

func (s *MasterService) DeleteRegion(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return fromPg(qtx.DeleteRegion(ctx, id))
	})
}

func (s *MasterService) ListProgramTitles(ctx context.Context, params model.PaginationParams) (*model.ListResponse[model.ProgramTitleResponse], error) {
	page, limit, offset, search, sortField, sortOrder := normalizeMasterList(params, "title", "asc", "title")
	rows, err := s.queries.ListProgramTitles(ctx, queries.ListProgramTitlesParams{
		Limit:     int32(limit),
		Offset:    int32(offset),
		Search:    search,
		SortField: sortField,
		SortOrder: sortOrder,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar program title")
	}
	total, err := s.queries.CountProgramTitles(ctx, search)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung program title")
	}
	data := make([]model.ProgramTitleResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toProgramTitleResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *MasterService) GetProgramTitle(ctx context.Context, id pgtype.UUID) (*model.ProgramTitleResponse, error) {
	row, err := s.queries.GetProgramTitle(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Program title tidak ditemukan")
	}
	res := toProgramTitleResponse(row)
	return &res, nil
}

func (s *MasterService) CreateProgramTitle(ctx context.Context, req model.ProgramTitleRequest) (*model.ProgramTitleResponse, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, validation("title", "wajib diisi")
	}
	parentID, err := nullableUUID(req.ParentID)
	if err != nil {
		return nil, validation("parent_id", "UUID tidak valid")
	}
	var created queries.ProgramTitle
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.CreateProgramTitle(ctx, queries.CreateProgramTitleParams{ParentID: parentID, Title: req.Title})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	res := toProgramTitleResponse(created)
	return &res, nil
}

func (s *MasterService) UpdateProgramTitle(ctx context.Context, id pgtype.UUID, req model.ProgramTitleRequest) (*model.ProgramTitleResponse, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, validation("title", "wajib diisi")
	}
	parentID, err := nullableUUID(req.ParentID)
	if err != nil {
		return nil, validation("parent_id", "UUID tidak valid")
	}
	var updated queries.ProgramTitle
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdateProgramTitle(ctx, queries.UpdateProgramTitleParams{ID: id, ParentID: parentID, Title: req.Title})
		if err != nil {
			return mapNotFound(err, "Program title tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	res := toProgramTitleResponse(updated)
	return &res, nil
}

func (s *MasterService) DeleteProgramTitle(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return fromPg(qtx.DeleteProgramTitle(ctx, id))
	})
}

func (s *MasterService) ListBappenasPartners(ctx context.Context, params model.PaginationParams, levels []string) (*model.ListResponse[model.BappenasPartnerResponse], error) {
	page, limit, offset, search, sortField, sortOrder := normalizeMasterList(params, "level", "asc", "name", "level")
	rows, err := s.queries.ListBappenasPartners(ctx, queries.ListBappenasPartnersParams{
		Limit:        int32(limit),
		Offset:       int32(offset),
		LevelFilters: levels,
		Search:       search,
		SortField:    sortField,
		SortOrder:    sortOrder,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar bappenas partner")
	}
	total, err := s.queries.CountBappenasPartners(ctx, queries.CountBappenasPartnersParams{LevelFilters: levels, Search: search})
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung bappenas partner")
	}
	data := make([]model.BappenasPartnerResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toBappenasPartnerResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *MasterService) GetBappenasPartner(ctx context.Context, id pgtype.UUID) (*model.BappenasPartnerResponse, error) {
	row, err := s.queries.GetBappenasPartner(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Bappenas partner tidak ditemukan")
	}
	res := toBappenasPartnerResponse(row)
	return &res, nil
}

func (s *MasterService) CreateBappenasPartner(ctx context.Context, req model.BappenasPartnerRequest) (*model.BappenasPartnerResponse, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Level) == "" {
		return nil, validation("name", "name dan level wajib diisi")
	}
	parentID, err := nullableUUID(req.ParentID)
	if err != nil {
		return nil, validation("parent_id", "UUID tidak valid")
	}
	var created queries.BappenasPartner
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.CreateBappenasPartner(ctx, queries.CreateBappenasPartnerParams{ParentID: parentID, Name: req.Name, Level: req.Level})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	res := toBappenasPartnerResponse(created)
	return &res, nil
}

func (s *MasterService) UpdateBappenasPartner(ctx context.Context, id pgtype.UUID, req model.BappenasPartnerRequest) (*model.BappenasPartnerResponse, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Level) == "" {
		return nil, validation("name", "name dan level wajib diisi")
	}
	parentID, err := nullableUUID(req.ParentID)
	if err != nil {
		return nil, validation("parent_id", "UUID tidak valid")
	}
	var updated queries.BappenasPartner
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdateBappenasPartner(ctx, queries.UpdateBappenasPartnerParams{ID: id, ParentID: parentID, Name: req.Name, Level: req.Level})
		if err != nil {
			return mapNotFound(err, "Bappenas partner tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	res := toBappenasPartnerResponse(updated)
	return &res, nil
}

func (s *MasterService) DeleteBappenasPartner(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return fromPg(qtx.DeleteBappenasPartner(ctx, id))
	})
}

func (s *MasterService) ListPeriods(ctx context.Context, params model.PaginationParams) (*model.ListResponse[model.PeriodResponse], error) {
	page, limit, offset, _, sortField, sortOrder := normalizeMasterList(params, "year_start", "desc", "name", "year_start", "year_end")
	rows, err := s.queries.ListPeriods(ctx, queries.ListPeriodsParams{
		Limit:     int32(limit),
		Offset:    int32(offset),
		SortField: sortField,
		SortOrder: sortOrder,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar period")
	}
	total, err := s.queries.CountPeriods(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung period")
	}
	data := make([]model.PeriodResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toPeriodResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *MasterService) GetPeriod(ctx context.Context, id pgtype.UUID) (*model.PeriodResponse, error) {
	row, err := s.queries.GetPeriod(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Period tidak ditemukan")
	}
	res := toPeriodResponse(row)
	return &res, nil
}

func (s *MasterService) CreatePeriod(ctx context.Context, req model.PeriodRequest) (*model.PeriodResponse, error) {
	if err := validatePeriod(req); err != nil {
		return nil, err
	}
	var created queries.Period
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.CreatePeriod(ctx, queries.CreatePeriodParams{Name: req.Name, YearStart: req.YearStart, YearEnd: req.YearEnd})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	res := toPeriodResponse(created)
	return &res, nil
}

func (s *MasterService) UpdatePeriod(ctx context.Context, id pgtype.UUID, req model.PeriodRequest) (*model.PeriodResponse, error) {
	if err := validatePeriod(req); err != nil {
		return nil, err
	}
	var updated queries.Period
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdatePeriod(ctx, queries.UpdatePeriodParams{ID: id, Name: req.Name, YearStart: req.YearStart, YearEnd: req.YearEnd})
		if err != nil {
			return mapNotFound(err, "Period tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	res := toPeriodResponse(updated)
	return &res, nil
}

func (s *MasterService) DeletePeriod(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return fromPg(qtx.DeletePeriod(ctx, id))
	})
}

func (s *MasterService) ListNationalPriorities(ctx context.Context, params model.PaginationParams, periodIDs []string) (*model.ListResponse[model.NationalPriorityResponse], error) {
	page, limit, offset, search, sortField, sortOrder := normalizeMasterList(params, "title", "asc", "title", "period")
	periodFilters, err := uuidFilters(periodIDs, "period_id")
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListNationalPriorities(ctx, queries.ListNationalPrioritiesParams{
		Limit:           int32(limit),
		Offset:          int32(offset),
		PeriodIDFilters: periodFilters,
		Search:          search,
		SortField:       sortField,
		SortOrder:       sortOrder,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar national priority")
	}
	total, err := s.queries.CountNationalPriorities(ctx, queries.CountNationalPrioritiesParams{PeriodIDFilters: periodFilters, Search: search})
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung national priority")
	}
	data := make([]model.NationalPriorityResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toNationalPriorityListResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *MasterService) GetNationalPriority(ctx context.Context, id pgtype.UUID) (*model.NationalPriorityResponse, error) {
	row, err := s.queries.GetNationalPriority(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "National priority tidak ditemukan")
	}
	res := toNationalPriorityResponse(row)
	return &res, nil
}

func (s *MasterService) CreateNationalPriority(ctx context.Context, req model.NationalPriorityRequest) (*model.NationalPriorityResponse, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, validation("title", "wajib diisi")
	}
	periodID, err := model.ParseUUID(req.PeriodID)
	if err != nil {
		return nil, validation("period_id", "UUID tidak valid")
	}
	var created queries.NationalPriority
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.CreateNationalPriority(ctx, queries.CreateNationalPriorityParams{PeriodID: periodID, Title: req.Title})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	return s.GetNationalPriority(ctx, created.ID)
}

func (s *MasterService) UpdateNationalPriority(ctx context.Context, id pgtype.UUID, req model.NationalPriorityRequest) (*model.NationalPriorityResponse, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, validation("title", "wajib diisi")
	}
	periodID, err := model.ParseUUID(req.PeriodID)
	if err != nil {
		return nil, validation("period_id", "UUID tidak valid")
	}
	var updated queries.NationalPriority
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdateNationalPriority(ctx, queries.UpdateNationalPriorityParams{ID: id, PeriodID: periodID, Title: req.Title})
		if err != nil {
			return mapNotFound(err, "National priority tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	return s.GetNationalPriority(ctx, updated.ID)
}

func (s *MasterService) DeleteNationalPriority(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return fromPg(qtx.DeleteNationalPriority(ctx, id))
	})
}

func (s *MasterService) withTx(ctx context.Context, fn func(*queries.Queries) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return apperrors.Internal("Gagal memulai transaksi")
	}
	defer tx.Rollback(ctx)
	if err := middleware.ApplyAuditUser(ctx, tx); err != nil {
		return apperrors.Internal("Gagal menyiapkan audit user")
	}
	if err := fn(s.queries.WithTx(tx)); err != nil {
		if _, ok := err.(*apperrors.AppError); ok {
			return err
		}
		return apperrors.FromPgError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return apperrors.Internal("Gagal menyimpan data")
	}
	return nil
}

func normalizeList(params model.PaginationParams) (int, int, int) {
	page, limit := normalizePagination(params.Page, params.Limit)
	return page, limit, (page - 1) * limit
}

func normalizeMasterList(params model.PaginationParams, defaultSort, defaultOrder string, allowedSorts ...string) (int, int, int, pgtype.Text, string, string) {
	page, limit, offset := normalizeList(params)
	sortField := defaultSort
	for _, allowed := range allowedSorts {
		if params.Sort == allowed {
			sortField = params.Sort
			break
		}
	}

	sortOrder := strings.ToLower(params.Order)
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = defaultOrder
	}

	return page, limit, offset, nullableText(params.Search), sortField, sortOrder
}

func listResponse[T any](data []T, page, limit int, total int64) *model.ListResponse[T] {
	return &model.ListResponse[T]{
		Data: data,
		Meta: model.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}
}

func validateCountry(req model.CountryRequest) error {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Code) == "" {
		return validation("name", "name dan code wajib diisi")
	}
	if len(req.Code) != 3 {
		return validation("code", "harus 3 karakter")
	}
	return nil
}

func validateCurrency(req model.CurrencyRequest) error {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Code) == "" {
		return validation("name", "name dan code wajib diisi")
	}
	if len(strings.TrimSpace(req.Code)) != 3 {
		return validation("code", "harus 3 karakter")
	}
	for _, char := range strings.ToUpper(strings.TrimSpace(req.Code)) {
		if char < 'A' || char > 'Z' {
			return validation("code", "harus kode ISO 4217")
		}
	}
	return nil
}

func validateLender(req model.CreateLenderRequest) error {
	if req.Type != "Bilateral" && req.Type != "Multilateral" && req.Type != "KSA" {
		return validation("type", "harus Bilateral, Multilateral, atau KSA")
	}
	if req.Type != "Multilateral" && (req.CountryID == nil || strings.TrimSpace(*req.CountryID) == "") {
		return validation("country_id", "Wajib diisi untuk Bilateral dan KSA")
	}
	if req.Type == "Multilateral" && req.CountryID != nil && strings.TrimSpace(*req.CountryID) != "" {
		return validation("country_id", "Harus kosong untuk Multilateral")
	}
	if strings.TrimSpace(req.Name) == "" {
		return validation("name", "wajib diisi")
	}
	return nil
}

func validateInstitution(req model.InstitutionRequest) error {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Level) == "" {
		return validation("name", "name dan level wajib diisi")
	}
	if !isAllowedValue(req.Level, institutionLevels) {
		return validation("level", "level institution tidak valid")
	}
	return nil
}

func isAllowedValue(value string, allowed []string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}

func validateRegion(req model.RegionRequest) error {
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Type) == "" {
		return validation("code", "code, name, dan type wajib diisi")
	}
	return nil
}

func validatePeriod(req model.PeriodRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return validation("name", "wajib diisi")
	}
	if req.YearEnd <= req.YearStart {
		return validation("year_end", "harus lebih besar dari year_start")
	}
	return nil
}

func mapNotFound(err error, msg string) error {
	if err == pgx.ErrNoRows {
		return apperrors.NotFound(msg)
	}
	return apperrors.Internal("Gagal mengambil data")
}

func validation(field, msg string) error {
	return apperrors.Validation(apperrors.FieldError{Field: field, Message: msg})
}

func fromPg(err error) error {
	if err == nil {
		return nil
	}
	return apperrors.FromPgError(err)
}

func nullableUUID(value *string) (pgtype.UUID, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.UUID{}, nil
	}
	return model.ParseUUID(*value)
}

func uuidFilters(values []string, field string) ([]pgtype.UUID, error) {
	filters := make([]pgtype.UUID, 0, len(values))

	for _, value := range values {
		parsed, err := model.ParseUUID(value)
		if err != nil {
			return nil, validation(field, "UUID tidak valid")
		}
		filters = append(filters, parsed)
	}

	return filters, nil
}

func nullableText(value string) pgtype.Text {
	if strings.TrimSpace(value) == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func nullableTextPtr(value *string) pgtype.Text {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func nullableBool(value string) (pgtype.Bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return pgtype.Bool{}, nil
	case "true":
		return pgtype.Bool{Bool: true, Valid: true}, nil
	case "false":
		return pgtype.Bool{Bool: false, Valid: true}, nil
	default:
		return pgtype.Bool{}, pgx.ErrNoRows
	}
}

func stringPtrFromText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func stringPtrFromUUID(value pgtype.UUID) *string {
	if !value.Valid {
		return nil
	}
	str := model.UUIDToString(value)
	return &str
}

func formatMasterTime(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}

func toCountryResponse(row queries.Country) model.CountryResponse {
	return model.CountryResponse{ID: model.UUIDToString(row.ID), Name: row.Name, Code: row.Code, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func toCurrencyResponse(row queries.Currency) model.CurrencyResponse {
	return model.CurrencyResponse{ID: model.UUIDToString(row.ID), Code: row.Code, Name: row.Name, Symbol: stringPtrFromText(row.Symbol), IsActive: row.IsActive, SortOrder: row.SortOrder, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func toLenderListResponse(row queries.ListLendersRow) model.LenderResponse {
	res := model.LenderResponse{ID: model.UUIDToString(row.ID), Name: row.Name, ShortName: stringPtrFromText(row.ShortName), Type: row.Type, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
	if row.CountryID.Valid {
		res.Country = &model.CountryInfo{ID: model.UUIDToString(row.CountryID), Name: row.CountryName.String, Code: row.CountryCode.String}
	}
	return res
}

func toLenderResponse(row queries.GetLenderRow) model.LenderResponse {
	res := model.LenderResponse{ID: model.UUIDToString(row.ID), Name: row.Name, ShortName: stringPtrFromText(row.ShortName), Type: row.Type, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
	if row.CountryID.Valid {
		res.Country = &model.CountryInfo{ID: model.UUIDToString(row.CountryID), Name: row.CountryName.String, Code: row.CountryCode.String}
	}
	return res
}

func toInstitutionResponse(row queries.Institution) model.InstitutionResponse {
	return model.InstitutionResponse{ID: model.UUIDToString(row.ID), ParentID: stringPtrFromUUID(row.ParentID), Name: row.Name, ShortName: stringPtrFromText(row.ShortName), Level: row.Level, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func toInstitutionDetailResponse(row queries.GetInstitutionRow) model.InstitutionResponse {
	return model.InstitutionResponse{ID: model.UUIDToString(row.ID), ParentID: stringPtrFromUUID(row.ParentID), ParentName: stringPtrFromText(row.ParentName), Name: row.Name, ShortName: stringPtrFromText(row.ShortName), Level: row.Level, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func toRegionResponse(row queries.Region) model.RegionResponse {
	return model.RegionResponse{ID: model.UUIDToString(row.ID), Code: row.Code, Name: row.Name, Type: row.Type, ParentCode: stringPtrFromText(row.ParentCode), CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func toProgramTitleResponse(row queries.ProgramTitle) model.ProgramTitleResponse {
	return model.ProgramTitleResponse{ID: model.UUIDToString(row.ID), ParentID: stringPtrFromUUID(row.ParentID), Title: row.Title, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func toBappenasPartnerResponse(row queries.BappenasPartner) model.BappenasPartnerResponse {
	return model.BappenasPartnerResponse{ID: model.UUIDToString(row.ID), ParentID: stringPtrFromUUID(row.ParentID), Name: row.Name, Level: row.Level, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func toPeriodResponse(row queries.Period) model.PeriodResponse {
	return model.PeriodResponse{ID: model.UUIDToString(row.ID), Name: row.Name, YearStart: row.YearStart, YearEnd: row.YearEnd, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func toNationalPriorityListResponse(row queries.ListNationalPrioritiesRow) model.NationalPriorityResponse {
	return model.NationalPriorityResponse{ID: model.UUIDToString(row.ID), PeriodID: model.UUIDToString(row.PeriodID), PeriodName: row.PeriodName, Title: row.Title, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func toNationalPriorityResponse(row queries.GetNationalPriorityRow) model.NationalPriorityResponse {
	return model.NationalPriorityResponse{ID: model.UUIDToString(row.ID), PeriodID: model.UUIDToString(row.PeriodID), PeriodName: row.PeriodName, Title: row.Title, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}
