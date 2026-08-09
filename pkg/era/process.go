// Package era provides processing for X12 835 ERA (Electronic Remittance Advice)
// files, matching claims to procedures and auto-posting payments to the ledger.
package era

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/pkg/x12/835"
)

// ProcessResult summarizes the results of processing an ERA.
type ProcessResult struct {
	ClaimsProcessed int
	PaymentsPosted  int
	Adjustments     int
	Errors          []string
}

// Process835ERA parses a parsed ERA835 and posts payments/adjustments to the ledger.
//   - db: active database/sql connection pool
//   - queries: sqlc-generated Queries instance
//   - era: the parsed 835 ERA
//   - userID: the user ID to attribute payment records to
func Process835ERA(db *sql.DB, queries *dbgen.Queries, era *x12_835.ERA835, userID int64) (*ProcessResult, error) {
	result := &ProcessResult{}
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("era: begin tx: %w", err)
	}
	defer tx.Rollback()

	// Use a tx-bound Queries for atomic operations
	q := queries.WithTx(tx)

	for _, claim := range era.Claims {
		// Step 1: Match claim to procedure by voucher (PatientControlNumber = procvoucher)
		proc, err := q.FindProcByVoucher(ctx, sql.NullString{
			String: claim.PatientControlNumber,
			Valid:  claim.PatientControlNumber != "",
		})
		if err != nil {
			if err == sql.ErrNoRows {
				result.Errors = append(result.Errors,
					fmt.Sprintf("claim %s: no procedure found for voucher %q", claim.PayerClaimID, claim.PatientControlNumber))
				continue
			}
			result.Errors = append(result.Errors,
				fmt.Sprintf("claim %s: DB error matching voucher: %v", claim.PayerClaimID, err))
			continue
		}

		// Step 2: Post claim payment to payrec
		if claim.ClaimPayment > 0 {
			desc := fmt.Sprintf("ERA payment from %s for claim %s", era.PayerName, claim.PayerClaimID)
			_, err := q.CreatePaymentRecord(ctx, dbgen.CreatePaymentRecordParams{
				PatientID:       proc.PatientID,
				ProcedureID:     proc.ID,
				Payrectype:      0, // 0 = insurance payment
				Amount:          claim.ClaimPayment,
				Description:     desc,
				ReferenceNumber: sql.NullString{String: era.TraceNumber, Valid: era.TraceNumber != ""},
				UserID:          userID,
			})
			if err != nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("claim %s: failed to post payment: %v", claim.PayerClaimID, err))
				continue
			}
			result.PaymentsPosted++
		}

		// Step 3: Post CAS adjustments as separate payrec entries
		for _, cas := range claim.CASAdjustments {
			desc := fmt.Sprintf("ERA adjustment %s:%s from %s for claim %s",
				cas.GroupCode, cas.ReasonCode, era.PayerName, claim.PayerClaimID)
			// CAS amounts are typically negative (reductions) but can be positive
			amount := -cas.Amount // CAS amounts are reductions, so negate for ledger
			if cas.GroupCode == "PI" { // Payer Initiated — positive adjustment
				amount = cas.Amount
			}

			_, err := q.CreatePaymentRecord(ctx, dbgen.CreatePaymentRecordParams{
				PatientID:       proc.PatientID,
				ProcedureID:     proc.ID,
				Payrectype:      0, // 0 = insurance payment
				Amount:          amount,
				Description:     desc,
				ReferenceNumber: sql.NullString{String: fmt.Sprintf("%s-%s", era.TraceNumber, cas.ReasonCode), Valid: true},
				UserID:          userID,
			})
			if err != nil {
				result.Errors = append(result.Errors,
					fmt.Sprintf("claim %s: failed to post CAS adjustment %s:%s: %v",
						claim.PayerClaimID, cas.GroupCode, cas.ReasonCode, err))
				continue
			}
			result.Adjustments++
		}

		// Step 4: Post service-line CAS adjustments
		for _, line := range claim.ServiceLines {
			for _, cas := range line.CASAdjustments {
				desc := fmt.Sprintf("ERA line adj %s:%s from %s for claim %s (CPT %s)",
					cas.GroupCode, cas.ReasonCode, era.PayerName, claim.PayerClaimID, line.ProcedureCode)
				amount := -cas.Amount
				if cas.GroupCode == "PI" {
					amount = cas.Amount
				}

				_, err := q.CreatePaymentRecord(ctx, dbgen.CreatePaymentRecordParams{
					PatientID:       proc.PatientID,
					ProcedureID:     proc.ID,
					Payrectype:      0,
					Amount:          amount,
					Description:     desc,
					ReferenceNumber: sql.NullString{String: fmt.Sprintf("%s-%s", era.TraceNumber, cas.ReasonCode), Valid: true},
					UserID:          userID,
				})
				if err != nil {
					result.Errors = append(result.Errors,
						fmt.Sprintf("claim %s line %s: failed to post CAS: %v",
							claim.PayerClaimID, line.ProcedureCode, err))
					continue
				}
				result.Adjustments++
			}
		}

		result.ClaimsProcessed++
	}

	// Step 5: Post provider-level adjustments
	for _, plb := range era.ProviderAdjustments {
		desc := fmt.Sprintf("ERA provider adjustment %s from %s (trace %s)",
			plb.AdjustmentReason, era.PayerName, era.TraceNumber)
		// PLB amounts can be positive (payments to provider) or negative (recoupments)
		_, err := q.CreatePaymentRecord(ctx, dbgen.CreatePaymentRecordParams{
			PatientID:       0, // Provider-level, not tied to a patient
			ProcedureID:     0,
			Payrectype:      0,
			Amount:          plb.Amount,
			Description:     desc,
			ReferenceNumber: sql.NullString{String: fmt.Sprintf("%s-PLB-%s", era.TraceNumber, plb.AdjustmentReason), Valid: true},
			UserID:          userID,
		})
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("PLB %s: failed to post: %v", plb.AdjustmentReason, err))
			continue
		}
		result.Adjustments++
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("era: commit tx: %w", err)
	}

	log.Printf("ERA processed: %d claims, %d payments, %d adjustments, %d errors",
		result.ClaimsProcessed, result.PaymentsPosted, result.Adjustments, len(result.Errors))

	return result, nil
}
