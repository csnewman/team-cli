package main

import (
	"fmt"
	"time"

	"github.com/csnewman/team-cli/internal/team"
	"github.com/spf13/cobra"
)

func approveCmdRun(cmd *cobra.Command, args []string) error {
	cfg, err := readConfigReAuth(cmd.Context())
	if err != nil {
		return fmt.Errorf("could not read config and authenticate: %w", err)
	}

	requests, err := team.ListRequests(cmd.Context(), cfg.ServerConfig, cfg.AuthToken, team.ListRequestsFilterRequiresMyApproval)
	if err != nil {
		return fmt.Errorf("could not fetch requests: %w", err)
	}

	fmt.Println()

	if len(requests) == 0 {
		fmt.Println("There are no requests to approve")

		return nil
	}

	fmt.Println("Please select the request:")
	for i, req := range requests {
		fmt.Printf(
			"  [%d] requester=%q account=%q role=%q\n",
			i+1,
			req.Email,
			req.AccountName,
			req.Role,
		)
		fmt.Printf(
			"\taccount_id=%q requested=%q start_time=%q duration=%q \n",
			req.AccountID, fmtDate(req.CreatedAt), fmtDate(req.StartTime), req.Duration+" hours",
		)
		fmt.Printf(
			"\tticket=%q justification=%q\n",
			req.TicketNo,
			req.Justification,
		)
	}

	fmt.Println()

	idx, err := promptSelection("Request option? ", 1, len(requests))
	if err != nil {
		return fmt.Errorf("could not select request: %w", err)
	}

	selectedRequest := requests[idx-1]

	fmt.Println()
	fmt.Println("Please select the response:")
	fmt.Println("  [1] Approve")
	fmt.Println("  [2] Approve without comment")
	fmt.Println("  [3] Reject")
	fmt.Println("  [4] Reject without comment")
	fmt.Println()

	idx, err = promptSelection("Response option? ", 1, 4)
	if err != nil {
		return fmt.Errorf("could not select request: %w", err)
	}

	comment := "No comment."
	approve := idx < 3

	if idx == 1 || idx == 3 {
		comment, err = promptString("Comment? ")
		if err != nil {
			return fmt.Errorf("could not read comment: %w", err)
		}
	}

	accResp := &team.AccessResponse{
		ID:      selectedRequest.ID,
		Comment: comment,
	}

	fmt.Println("")
	fmt.Println("Details:")
	fmt.Printf("  ID: %q\n", selectedRequest.ID)
	fmt.Printf("  Requester: email=%q\n", selectedRequest.Email)
	fmt.Printf("  Account: id=%q name=%q\n", selectedRequest.AccountID, selectedRequest.AccountName)
	fmt.Printf("  Role: name=%q\n", selectedRequest.Role)
	fmt.Printf("  Created: %q\n", fmtDate(selectedRequest.CreatedAt))
	fmt.Printf("  Start: %q\n", fmtDate(selectedRequest.StartTime))
	fmt.Printf("  Duration: %q\n", selectedRequest.Duration+" Hours")
	fmt.Printf("  Ticket: %q\n", selectedRequest.TicketNo)
	fmt.Printf("  Justification: %q\n", selectedRequest.Justification)

	if approve {
		fmt.Print("  Response Action: Approve\n")
		accResp.Status = "approved"
	} else {
		fmt.Print("  Response Action: Reject\n")
		accResp.Status = "rejected"
	}

	fmt.Printf("  Response Comment: %q\n", comment)

	fmt.Println()

	cont, err := promptBool("Confirm (y/n)? ")
	if err != nil {
		return fmt.Errorf("could not select confirmation: %w", err)
	}

	if !cont {
		return fmt.Errorf("%w: confirmation rejected", ErrInvalid)
	}

	if err := team.Respond(cmd.Context(), cfg.ServerConfig, cfg.AuthToken, accResp); err != nil {
		return fmt.Errorf("could not respond to request: %w", err)
	}

	fmt.Println("Responded")

	return nil
}

func fmtDate(t time.Time) string {
	return t.Local().Format(time.UnixDate)
}
