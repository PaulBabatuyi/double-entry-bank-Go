package api

import "github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"

func toAccountResponse(acc sqlc.Account) AccountResponse {
	var ownerID *string
	if acc.OwnerID.Valid {
		// Convert nullable UUID into pointer so omitempty works in JSON output.
		s := acc.OwnerID.UUID.String()
		ownerID = &s
	}

	return AccountResponse{
		ID:        acc.ID.String(),
		OwnerID:   ownerID,
		Name:      acc.Name,
		Balance:   acc.Balance,
		Currency:  acc.Currency,
		IsSystem:  acc.IsSystem,
		CreatedAt: acc.CreatedAt.Time,
	}
}

func toEntryResponse(entry sqlc.Entry) EntryResponse {
	var description string
	if entry.Description.Valid {
		// Preserve optional descriptions only when present in DB rows.
		description = entry.Description.String
	}

	operationType := operationTypeToString(entry.OperationType)

	return EntryResponse{
		ID:            entry.ID.String(),
		AccountID:     entry.AccountID.String(),
		Debit:         entry.Debit,
		Credit:        entry.Credit,
		TransactionID: entry.TransactionID.String(),
		OperationType: operationType,
		Description:   description,
		CreatedAt:     entry.CreatedAt.Time,
	}
}

func operationTypeToString(v interface{}) string {
	// sqlc enum decoding can arrive as string or []byte depending on driver path.
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}
