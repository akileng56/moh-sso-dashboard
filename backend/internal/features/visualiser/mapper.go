package visualiser

func toDatasetResponses(items []DatasetResponse) []DatasetResponse {
	if items == nil {
		return []DatasetResponse{}
	}
	out := make([]DatasetResponse, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func toDataElementResponses(items []DataElementResponse) []DataElementResponse {
	if items == nil {
		return []DataElementResponse{}
	}
	out := make([]DataElementResponse, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func toDataValuesResponse(item DataValuesResponse) DataValuesResponse {
	if item.Rows == nil {
		item.Rows = []DataValueRowResponse{}
	}
	return item
}
