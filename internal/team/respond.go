package team

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/csnewman/team-cli/internal/gql"
)

const respondQuery = `mutation UpdateRequests(
    $input: UpdateRequestsInput!
    $condition: ModelRequestsConditionInput
  ) {
    updateRequests(input: $input, condition: $condition) {
      id
      email
      accountId
      accountName
      role
      roleId
      startTime
      duration
      justification
      status
      comment
      username
      approver
      approverId
      approvers
      approver_ids
      revoker
      revokerId
      endTime
      ticketNo
      revokeComment
      session_duration
      createdAt
      updatedAt
      owner
      __typename
    }
}`

type AccessResponse struct {
	ID      string
	Status  string
	Comment string
}

func Respond(ctx context.Context, remote *RemoteConfig, token *AuthToken, accResp *AccessResponse) error {
	slog.Info("Responding to request")

	resp, err := gql.Execute(ctx, remote.GraphQLEndpoint, token.AccessToken, &gql.Request{
		Query: respondQuery,
		Variables: map[string]any{
			"input": map[string]any{
				"id":      accResp.ID,
				"status":  accResp.Status,
				"comment": accResp.Comment,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to execute: %w", err)
	}

	if len(resp.Errors) > 0 {
		for _, err := range resp.Errors {
			slog.Error("Received error from server", "error", err)
		}

		return fmt.Errorf("%w: server returned an error", ErrUnexpected)
	}

	return nil
}
