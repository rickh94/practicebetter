package librarypages

import "database/sql"

func SpotMeasuresOrEmpty(measures sql.NullString) string {
	if measures.Valid {
		return measures.String
	}
	return ""
}
