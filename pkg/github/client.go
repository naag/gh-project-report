package github

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/naag/gh-project-report/pkg/types"
	"github.com/shurcooL/graphql"
)

//go:embed queries/get_viewer_project_node_id.graphql
var getViewerProjectNodeIDQuery string

//go:embed queries/get_org_project_node_id.graphql
var getOrgProjectNodeIDQuery string

//go:embed queries/get_project_state.graphql
var getProjectStateQuery string

// Client represents a GitHub client
type Client struct {
	httpClient *http.Client
	baseURL    string
	verbose    bool
}

// graphQLRequest represents a GraphQL request
type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// loadQuery loads a GraphQL query from embedded files
func (c *Client) loadQuery(name string) (string, error) {
	switch name {
	case "get_viewer_project_node_id":
		return getViewerProjectNodeIDQuery, nil
	case "get_org_project_node_id":
		return getOrgProjectNodeIDQuery, nil
	case "get_project_state":
		return getProjectStateQuery, nil
	default:
		return "", fmt.Errorf("unknown query: %s", name)
	}
}

// executeQuery executes a GraphQL query and unmarshals the response into the result
func (c *Client) executeQuery(ctx context.Context, queryName string, variables map[string]interface{}, result interface{}) error {
	queryStr, err := c.loadQuery(queryName)
	if err != nil {
		return err
	}

	reqBody := graphQLRequest{
		Query:     queryStr,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.verbose {
		fmt.Printf("\nGraphQL Request:\n%s\n", string(jsonBody))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if c.verbose {
		fmt.Printf("\nGraphQL Response:\n%s\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data   json.RawMessage  `json:"data"`
		Errors []map[string]any `json:"errors,omitempty"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Errors) > 0 {
		return fmt.Errorf("GraphQL query failed: %v", response.Errors)
	}

	if err := json.Unmarshal(response.Data, result); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// NewClient creates a new GitHub client
func NewClient(httpClient *http.Client, verbose bool) *Client {
	return NewClientWithBaseURL(httpClient, "https://api.github.com/graphql", verbose)
}

// NewClientWithBaseURL creates a new GitHub client with a custom base URL
func NewClientWithBaseURL(httpClient *http.Client, baseURL string, verbose bool) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		verbose:    verbose,
	}
}

// GetProjectNodeId fetches the node ID for a project, either from an organization or personal project
func (c *Client) GetProjectNodeId(projectNumber int, org string) (string, error) {
	var result struct {
		Organization *struct {
			ProjectV2 *struct {
				ID string
			}
		}
		Viewer *struct {
			ProjectV2 *struct {
				ID string
			}
		}
	}

	variables := map[string]interface{}{
		"number": projectNumber,
	}

	queryName := "get_viewer_project_node_id"
	if org != "" {
		queryName = "get_org_project_node_id"
		variables["login"] = org
	}

	if err := c.executeQuery(context.Background(), queryName, variables, &result); err != nil {
		return "", err
	}

	if org != "" {
		if result.Organization == nil || result.Organization.ProjectV2 == nil {
			return "", fmt.Errorf("GraphQL query failed for org project: project not found")
		}
		return result.Organization.ProjectV2.ID, nil
	}

	if result.Viewer == nil || result.Viewer.ProjectV2 == nil {
		return "", fmt.Errorf("GraphQL query failed for viewer project: project not found")
	}
	return result.Viewer.ProjectV2.ID, nil
}

// FetchProjectState fetches the current state of a project
func (c *Client) FetchProjectState(projectNumber int, startField, endField string) (*types.ProjectState, error) {
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
		Viewer struct {
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
				} `graphql:"items(first: 100)"`
			} `graphql:"projectV2(number: $number)"`
		}
	}

	variables := map[string]interface{}{
		"number": projectNumber,
	}

	if err := c.executeQuery(context.Background(), "get_project_state", variables, &query); err != nil {
		return nil, fmt.Errorf("failed to fetch project state: %w", err)
	}

	// Convert the GraphQL response to our ProjectState type
	state := &types.ProjectState{
		Timestamp:     time.Now(),
		ProjectNumber: projectNumber,

		Items: make([]types.Item, 0),
	}

	for _, item := range query.Viewer.ProjectV2.Items.Nodes {
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

	return state, nil
}
