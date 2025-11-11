package team

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/csnewman/team-cli/internal/gql"
)

const listQuery = `query ListRequests(
    $filter: ModelRequestsFilterInput
    $limit: Int
    $nextToken: String
  ) {
    listRequests(filter: $filter, limit: $limit, nextToken: $nextToken) {
      items {
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
      nextToken
      __typename
    }
}`

type PermissionRequest struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Status string `json:"status"`

	AccountID     string    `json:"accountId"`
	AccountName   string    `json:"accountName"`
	Role          string    `json:"role"`
	RoleID        string    `json:"roleId"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	Duration      string    `json:"duration"`
	TicketNo      string    `json:"ticketNo"`
	Justification string    `json:"justification"`

	Comment    string   `json:"comment"`
	Approver   string   `json:"approver"`
	ApproverID string   `json:"approverId"`
	Approvers  []string `json:"approvers"`

	Revoker   string `json:"revoker"`
	RevokerID string `json:"revokerId"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type rawListResponse struct {
	ListRequests struct {
		Items     []*PermissionRequest `json:"items"`
		NextToken any                  `json:"nextToken"`
	} `json:"listRequests"`
}

type ListRequestsFilter string

const (
	ListRequestsFilterAll                ListRequestsFilter = "all"
	ListRequestsFilterRequiresMyApproval ListRequestsFilter = "requires-my-approval"
)

func ListRequests(
	ctx context.Context,
	remote *RemoteConfig,
	token *AuthToken,
	filter ListRequestsFilter,
) ([]*PermissionRequest, error) {
	idTok, err := token.ParseIDToken()
	if err != nil {
		return nil, fmt.Errorf("failed to parse ID token: %w", err)
	}

	var filterBlob map[string]any

	switch filter {
	case ListRequestsFilterAll:
	// no filter
	case ListRequestsFilterRequiresMyApproval:
		filterBlob = map[string]any{
			"and": []map[string]any{
				{
					"email": map[string]any{
						"ne": idTok.Email,
					},
				},
				{
					"status": map[string]any{
						"eq": "pending",
					},
				},
				{
					"approvers": map[string]any{
						"contains": idTok.Email,
					},
				},
			},
		}
	default:
		panic("unknown filter")
	}

	resp, err := gql.Execute(ctx, remote.GraphQLEndpoint, token.AccessToken, &gql.Request{
		Query: listQuery,
		Variables: map[string]any{
			"filter":    filterBlob,
			"nextToken": nil,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute: %w", err)
	}

	if len(resp.Errors) > 0 {
		for _, err := range resp.Errors {
			slog.Error("Received error from server", "error", err)
		}

		return nil, fmt.Errorf("%w: server returned an error", ErrUnexpected)
	}

	var rawResult rawListResponse

	if err := resp.UnmarshalData(&rawResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return rawResult.ListRequests.Items, nil
}
