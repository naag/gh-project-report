package github

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/naag/gh-project-report/pkg/types"
	"github.com/shurcooL/graphql"
)

// Client represents a GitHub client
type Client struct {
	graphql *graphql.Client
	verbose bool
}

// NewClient creates a new GitHub client
func NewClient(httpClient *http.Client, verbose bool) *Client {
	return NewClientWithBaseURL(httpClient, "https://api.github.com/graphql", verbose)
}

// NewClientWithBaseURL creates a new GitHub client with a custom base URL
func NewClientWithBaseURL(httpClient *http.Client, baseURL string, verbose bool) *Client {
	if verbose {
		// Wrap the transport with our logging transport
		transport := httpClient.Transport
		if transport == nil {
			transport = http.DefaultTransport
		}

		httpClient.Transport = &loggingTransport{
			transport: transport,
		}
	}

	client := graphql.NewClient(baseURL, httpClient)

	return &Client{
		graphql: client,
		verbose: verbose,
	}
}

// FetchProjectState fetches the current state of a project
func (c *Client) FetchProjectState(projectNumber int, organization, startField, endField string) (*types.ProjectState, error) {
	// First, lookup the project's node ID
	projectNodeID, err := c.LookupProjectNodeID(projectNumber, organization)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup project ID: %w", err)
	}

	// Common field types that will be embedded
	type ProjectV2FieldCommon struct {
		Name graphql.String
	}

	type ProjectV2Field struct {
		Common ProjectV2FieldCommon `graphql:"... on ProjectV2FieldCommon"`
	}

	// Field value types that will be embedded
	type TextFieldValue struct {
		Text  graphql.String
		Field ProjectV2Field
	}

	type NumberFieldValue struct {
		Number float64
		Field  ProjectV2Field
	}

	type DateFieldValue struct {
		Date  graphql.String
		Field ProjectV2Field
	}

	type SingleSelectFieldValue struct {
		Name  graphql.String
		Field ProjectV2Field
	}

	type RepositoryFieldValue struct {
		Repository struct {
			Name  graphql.String
			Owner struct {
				Login graphql.String
			}
		}
		Field ProjectV2Field
	}

	// Content types that will be embedded
	type IssueContent struct {
		Title     graphql.String
		CreatedAt graphql.String
		UpdatedAt graphql.String
	}

	type PullRequestContent struct {
		Title     graphql.String
		CreatedAt graphql.String
		UpdatedAt graphql.String
	}

	type DraftIssueContent struct {
		Title     graphql.String
		CreatedAt graphql.String
		UpdatedAt graphql.String
	}

	var query struct {
		Node struct {
			TypeName  graphql.String `graphql:"__typename"`
			ProjectV2 struct {
				Title graphql.String
				Items struct {
					PageInfo struct {
						HasNextPage graphql.Boolean
						EndCursor   graphql.String
					}
					Nodes []struct {
						ID          graphql.String
						FieldValues struct {
							Nodes []struct {
								TypeName     graphql.String         `graphql:"__typename"`
								TextValue    TextFieldValue         `graphql:"... on ProjectV2ItemFieldTextValue"`
								NumberValue  NumberFieldValue       `graphql:"... on ProjectV2ItemFieldNumberValue"`
								DateValue    DateFieldValue         `graphql:"... on ProjectV2ItemFieldDateValue"`
								SingleSelect SingleSelectFieldValue `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
								Repository   RepositoryFieldValue   `graphql:"... on ProjectV2ItemFieldRepositoryValue"`
							}
						} `graphql:"fieldValues(first: 100)"`
						Content struct {
							TypeName    graphql.String     `graphql:"__typename"`
							Issue       IssueContent       `graphql:"... on Issue"`
							PullRequest PullRequestContent `graphql:"... on PullRequest"`
							DraftIssue  DraftIssueContent  `graphql:"... on DraftIssue"`
						}
					}
				} `graphql:"items(first: 100, after: $cursor)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $id)"`
	}

	// Initialize state
	state := &types.ProjectState{
		Timestamp:     time.Now(),
		ProjectNumber: projectNumber,
		ProjectID:     projectNodeID,
		Organization:  organization,
		Items:         make([]types.Item, 0),
	}

	var cursor *graphql.String
	for {
		variables := map[string]interface{}{
			"id":     graphql.ID(projectNodeID),
			"cursor": cursor,
		}

		err = c.graphql.Query(context.Background(), &query, variables)
		if err != nil {
			return nil, fmt.Errorf("GraphQL query failed: %w", err)
		}

		// Process items from current page
		for _, item := range query.Node.ProjectV2.Items.Nodes {
			// Get title and timestamps based on content type
			var (
				title     string
				createdAt time.Time
				updatedAt time.Time
			)

			switch item.Content.TypeName {
			case "Issue":
				title = string(item.Content.Issue.Title)
				createdAt, _ = time.Parse(time.RFC3339, string(item.Content.Issue.CreatedAt))
				updatedAt, _ = time.Parse(time.RFC3339, string(item.Content.Issue.UpdatedAt))
			case "PullRequest":
				title = string(item.Content.PullRequest.Title)
				createdAt, _ = time.Parse(time.RFC3339, string(item.Content.PullRequest.CreatedAt))
				updatedAt, _ = time.Parse(time.RFC3339, string(item.Content.PullRequest.UpdatedAt))
			case "DraftIssue":
				title = string(item.Content.DraftIssue.Title)
				createdAt, _ = time.Parse(time.RFC3339, string(item.Content.DraftIssue.CreatedAt))
				updatedAt, _ = time.Parse(time.RFC3339, string(item.Content.DraftIssue.UpdatedAt))
			}

			if title == "" {
				title = fmt.Sprintf("Unknown type: %s", item.Content.TypeName)
			}

			projectItem := types.Item{
				ID: string(item.ID),
				Attributes: map[string]interface{}{
					"Title":      title,
					"created_at": createdAt,
					"updated_at": updatedAt,
				},
			}

			// Process field values
			for _, fieldValue := range item.FieldValues.Nodes {
				switch fieldValue.TypeName {
				case "ProjectV2ItemFieldTextValue":
					name := string(fieldValue.TextValue.Field.Common.Name)
					if name == "Title" {
						continue
					}
					projectItem.Attributes[name] = string(fieldValue.TextValue.Text)
				case "ProjectV2ItemFieldNumberValue":
					name := string(fieldValue.NumberValue.Field.Common.Name)
					projectItem.Attributes[name] = fieldValue.NumberValue.Number
				case "ProjectV2ItemFieldDateValue":
					name := string(fieldValue.DateValue.Field.Common.Name)
					dateStr := string(fieldValue.DateValue.Date)

					if name == startField || name == endField {
						if date, err := time.Parse("2006-01-02", dateStr); err == nil {
							if name == startField {
								projectItem.DateSpan.Start = date
							} else {
								projectItem.DateSpan.End = date
							}
						}
					} else {
						projectItem.Attributes[name] = dateStr
					}
				case "ProjectV2ItemFieldSingleSelectValue":
					name := string(fieldValue.SingleSelect.Field.Common.Name)
					projectItem.Attributes[name] = string(fieldValue.SingleSelect.Name)
				case "ProjectV2ItemFieldRepositoryValue":
					name := string(fieldValue.Repository.Field.Common.Name)
					repoValue := fmt.Sprintf("%s/%s",
						fieldValue.Repository.Repository.Owner.Login,
						fieldValue.Repository.Repository.Name)
					projectItem.Attributes[name] = repoValue
				}
			}

			state.Items = append(state.Items, projectItem)
		}

		// Check if there are more pages
		if !query.Node.ProjectV2.Items.PageInfo.HasNextPage {
			break
		}

		// Update cursor for next page
		endCursor := graphql.String(query.Node.ProjectV2.Items.PageInfo.EndCursor)
		cursor = &endCursor
	}

	return state, nil
}

// LookupProjectNodeID looks up the node ID for a project based on its number and optional organization
func (c *Client) LookupProjectNodeID(projectNumber int, organization string) (string, error) {
	if organization != "" {
		// Try organization project first
		var orgQuery struct {
			Organization struct {
				ProjectV2 struct {
					ID graphql.String
				} `graphql:"projectV2(number: $number)"`
			} `graphql:"organization(login: $login)"`
		}

		variables := map[string]interface{}{
			"number": graphql.Int(projectNumber),
			"login":  graphql.String(organization),
		}

		err := c.graphql.Query(context.Background(), &orgQuery, variables)
		if err != nil {
			return "", fmt.Errorf("GraphQL query failed: %w", err)
		}

		if id := string(orgQuery.Organization.ProjectV2.ID); id != "" {
			return id, nil
		}
		return "", fmt.Errorf("project %d not found in organization %s", projectNumber, organization)
	}

	// Fall back to viewer's project
	var viewerQuery struct {
		Viewer struct {
			ProjectV2 struct {
				ID graphql.String
			} `graphql:"projectV2(number: $number)"`
		}
	}

	variables := map[string]interface{}{
		"number": graphql.Int(projectNumber),
	}

	err := c.graphql.Query(context.Background(), &viewerQuery, variables)
	if err != nil {
		return "", fmt.Errorf("GraphQL query failed: %w", err)
	}

	if id := string(viewerQuery.Viewer.ProjectV2.ID); id != "" {
		return id, nil
	}

	return "", fmt.Errorf("project %d not found", projectNumber)
}

type loggingTransport struct {
	transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		// Log the request
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		// Restore the body for the actual request
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		fmt.Printf("\nGraphQL Request:\n%s\n", string(body))
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Log the response
	if resp.Body != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		// Restore the body for the actual response
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		fmt.Printf("\nGraphQL Response:\n%s\n", string(body))
	}

	return resp, nil
}
