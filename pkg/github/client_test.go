package github

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchProjectState(t *testing.T) {
	tests := []struct {
		name       string
		responses  []string
		startField string
		endField   string
		wantDates  bool
		wantStart  time.Time
		wantEnd    time.Time
	}{
		{
			name: "with start and end fields",
			responses: []string{
				`{
					"data": {
						"viewer": {
							"projectV2": {
								"id": "PVT_123"
							}
						}
					}
				}`,
				`{
					"data": {
						"node": {
							"__typename": "ProjectV2",
							"items": {
								"pageInfo": { "hasNextPage": false },
								"nodes": [{
									"id": "item1",
									"fieldValues": {
										"nodes": [
											{
												"__typename": "ProjectV2ItemFieldDateValue",
												"field": { "name": "Start Date" },
												"date": "2024-01-01"
											},
											{
												"__typename": "ProjectV2ItemFieldDateValue",
												"field": { "name": "Due Date" },
												"date": "2024-01-10"
											}
										]
									},
									"content": {
										"__typename": "Issue",
										"title": "Test Issue",
										"createdAt": "2024-01-01T00:00:00Z",
										"updatedAt": "2024-01-01T00:00:00Z"
									}
								}]
							}
						}
					}
				}`,
			},
			startField: "Start Date",
			endField:   "Due Date",
			wantDates:  true,
			wantStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:    time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "with date fields but not marked as start/end",
			responses: []string{
				`{
					"data": {
						"viewer": {
							"projectV2": {
								"id": "PVT_123"
							}
						}
					}
				}`,
				`{
					"data": {
						"node": {
							"__typename": "ProjectV2",
							"items": {
								"pageInfo": { "hasNextPage": false },
								"nodes": [{
									"id": "item1",
									"fieldValues": {
										"nodes": [
											{
												"__typename": "ProjectV2ItemFieldDateValue",
												"field": { "name": "Start Date" },
												"date": "2024-01-01"
											},
											{
												"__typename": "ProjectV2ItemFieldDateValue",
												"field": { "name": "Due Date" },
												"date": "2024-01-10"
											}
										]
									},
									"content": {
										"__typename": "Issue",
										"title": "Test Issue",
										"createdAt": "2024-01-01T00:00:00Z",
										"updatedAt": "2024-01-01T00:00:00Z"
									}
								}]
							}
						}
					}
				}`,
			},
			startField: "Other Field",
			endField:   "Another Field",
			wantDates:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			responseIndex := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responses[responseIndex]))
				responseIndex++
			}))
			defer server.Close()

			// Create client with test server URL
			httpClient := &http.Client{
				Transport: &http.Transport{
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(server.URL)
					},
				},
			}
			client := NewClientWithBaseURL(httpClient, server.URL, false)

			// Fetch state
			state, err := client.FetchProjectState(123, "", tt.startField, tt.endField)
			assert.NoError(t, err)
			assert.NotNil(t, state)
			assert.Len(t, state.Items, 1)

			item := state.Items[0]
			if tt.wantDates {
				// Verify dates are in DateSpan
				assert.Equal(t, tt.wantStart, item.DateSpan.Start)
				assert.Equal(t, tt.wantEnd, item.DateSpan.End)
				// Verify dates are not in Attributes
				_, hasStart := item.Attributes[tt.startField]
				_, hasEnd := item.Attributes[tt.endField]
				assert.False(t, hasStart, "start date should not be in Attributes")
				assert.False(t, hasEnd, "end date should not be in Attributes")
			} else {
				// Verify dates are in Attributes
				assert.Equal(t, "2024-01-01", item.Attributes["Start Date"])
				assert.Equal(t, "2024-01-10", item.Attributes["Due Date"])
				// Verify DateSpan is empty
				assert.True(t, item.DateSpan.Start.IsZero())
				assert.True(t, item.DateSpan.End.IsZero())
			}
		})
	}
}

func TestFetchProjectStateErrors(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantErrMsg string
	}{
		{
			name:       "invalid json response",
			response:   "invalid json",
			statusCode: 200,
			wantErrMsg: "GraphQL query failed",
		},
		{
			name:       "server error",
			response:   `{"errors":[{"message":"Server Error"}]}`,
			statusCode: 500,
			wantErrMsg: "GraphQL query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			assert.NoError(t, err)

			httpClient := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(serverURL),
				},
			}
			client := NewClientWithBaseURL(httpClient, server.URL, false)

			_, err = client.FetchProjectState(123, "", "Timeline", "Due Date")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}

func TestLookupProjectNodeID(t *testing.T) {
	tests := []struct {
		name         string
		response     string
		projectNum   int
		organization string
		wantID       string
		wantErr      string
	}{
		{
			name: "user project found",
			response: `{
				"data": {
					"viewer": {
						"projectV2": {
							"id": "PVT_123"
						}
					}
				}
			}`,
			projectNum: 123,
			wantID:     "PVT_123",
		},
		{
			name: "org project found",
			response: `{
				"data": {
					"organization": {
						"projectV2": {
							"id": "PVT_456"
						}
					}
				}
			}`,
			projectNum:   456,
			organization: "testorg",
			wantID:       "PVT_456",
		},
		{
			name: "project not found in org",
			response: `{
				"data": {
					"organization": {
						"projectV2": {
							"id": ""
						}
					}
				}
			}`,
			projectNum:   789,
			organization: "testorg",
			wantErr:      "project 789 not found in organization testorg",
		},
		{
			name: "project not found for user",
			response: `{
				"data": {
					"viewer": {
						"projectV2": {
							"id": ""
						}
					}
				}
			}`,
			projectNum: 999,
			wantErr:    "project 999 not found",
		},
		{
			name: "graphql error",
			response: `{
				"errors": [
					{
						"message": "Something went wrong"
					}
				]
			}`,
			projectNum: 123,
			wantErr:    "GraphQL query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			httpClient := &http.Client{
				Transport: &http.Transport{
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(server.URL)
					},
				},
			}
			client := NewClientWithBaseURL(httpClient, server.URL, false)

			gotID, err := client.LookupProjectNodeID(tt.projectNum, tt.organization)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantID, gotID)
		})
	}
}
