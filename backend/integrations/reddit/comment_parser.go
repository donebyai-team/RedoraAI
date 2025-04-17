package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
)

func (r *Client) parsePostWithComments(ctx context.Context, rawResp []json.RawMessage) (*Post, error) {
	// Parse post metadata
	var postListing struct {
		Data struct {
			Children []struct {
				Data Post `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rawResp[0], &postListing); err != nil {
		return nil, err
	}
	post := postListing.Data.Children[0].Data

	// Parse top-level comments
	var commentListing struct {
		Data struct {
			Children []json.RawMessage `json:"children"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rawResp[1], &commentListing); err != nil {
		return nil, err
	}

	var comments []*Comment
	for _, raw := range commentListing.Data.Children {
		comment, err := r.parseCommentTree(ctx, raw)
		if err == nil && comment != nil {
			comments = append(comments, comment)
		}
	}
	post.Comments = comments
	return &post, nil
}

func (r *Client) parseCommentTree(ctx context.Context, raw json.RawMessage) (*Comment, error) {
	var kindWrapper struct {
		Kind string          `json:"kind"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &kindWrapper); err != nil {
		return nil, err
	}

	switch kindWrapper.Kind {
	case "t1": // comment
		var c Comment
		if err := json.Unmarshal(kindWrapper.Data, &c); err != nil {
			return nil, err
		}
		// Handle replies recursively
		var repliesWrapper struct {
			Data struct {
				Children []json.RawMessage `json:"children"`
			} `json:"data"`
		}
		if rawReplies, ok := extractReplies(kindWrapper.Data); ok {
			if err := json.Unmarshal(rawReplies, &repliesWrapper); err == nil {
				for _, child := range repliesWrapper.Data.Children {
					childComment, err := r.parseCommentTree(ctx, child)
					if err == nil && childComment != nil {
						c.Comments = append(c.Comments, childComment)
					}
				}
			}
		}
		return &c, nil

	case "more":
		// Handle "more" children
		var more struct {
			Children []string `json:"children"`
			ParentID string   `json:"parent_id"`
			ID       string   `json:"id"`
		}
		if err := json.Unmarshal(kindWrapper.Data, &more); err != nil {
			return nil, err
		}
		return r.fetchMoreComments(ctx, more.Children, more.ParentID)
	default:
		return nil, nil
	}
}

func extractReplies(data json.RawMessage) (json.RawMessage, bool) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, false
	}
	if replies, ok := obj["replies"]; ok && string(replies) != `""` {
		return replies, true
	}
	return nil, false
}

func (r *Client) fetchMoreComments(ctx context.Context, children []string, parentID string) (*Comment, error) {
	if len(children) == 0 {
		return nil, nil
	}

	reqURL := fmt.Sprintf("%s/api/morechildren.json?api_type=json&raw_json=1&children=%s&link_id=%s",
		r.baseURL, strings.Join(children, ","), parentID)

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.config.AccessToken))

	resp, err := r.DoWithRateLimit(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var response struct {
		Json struct {
			Data struct {
				Things []json.RawMessage `json:"things"`
			} `json:"data"`
		} `json:"json"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	moreParent := &Comment{}
	for _, thing := range response.Json.Data.Things {
		comment, err := r.parseCommentTree(ctx, thing)
		if err == nil && comment != nil {
			moreParent.Comments = append(moreParent.Comments, comment)
		}
	}
	return moreParent, nil
}
