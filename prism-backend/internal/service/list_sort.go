package service

import "strings"

func normalizeListSort(sortField, sortOrder, defaultSort, defaultOrder string, allowedSorts map[string]struct{}) (string, string, error) {
	sortField = strings.TrimSpace(sortField)
	if sortField == "" {
		sortField = defaultSort
	}
	if _, ok := allowedSorts[sortField]; !ok {
		return "", "", validation("sort", "nilai tidak valid")
	}

	sortOrder = strings.ToLower(strings.TrimSpace(sortOrder))
	if sortOrder == "" {
		sortOrder = defaultOrder
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		return "", "", validation("order", "harus asc atau desc")
	}

	return sortField, sortOrder, nil
}
