package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
)

func (r *Client) parsePostWithComments(ctx context.Context, rawResp []json.RawMessage, maxComments int, includeMore bool) (*Post, error) {
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
	count := 0
	for _, raw := range commentListing.Data.Children {
		if count >= maxComments {
			break
		}
		comment, cCount, err := r.parseCommentTree(ctx, raw, maxComments-count, includeMore)
		if err == nil && comment != nil {
			comments = append(comments, comment)
			count += cCount
		}
	}
	post.Comments = comments
	return &post, nil
}

func (r *Client) parseCommentTree(
	ctx context.Context,
	raw json.RawMessage,
	remaining int,
	includeMore bool,
) (*Comment, int, error) {
	if remaining <= 0 {
		return nil, 0, nil
	}

	var kindWrapper struct {
		Kind string          `json:"kind"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &kindWrapper); err != nil {
		return nil, 0, err
	}

	switch kindWrapper.Kind {
	case "t1": // Comment
		var c Comment
		if err := json.Unmarshal(kindWrapper.Data, &c); err != nil {
			return nil, 0, err
		}

		total := 0
		var replies []*Comment

		// Recurse into replies
		if rawReplies, ok := extractReplies(kindWrapper.Data); ok && remaining > 1 {
			var repliesWrapper struct {
				Data struct {
					Children []json.RawMessage `json:"children"`
				} `json:"data"`
			}
			if err := json.Unmarshal(rawReplies, &repliesWrapper); err == nil {
				for _, child := range repliesWrapper.Data.Children {
					if total >= remaining {
						break
					}
					childComment, used, err := r.parseCommentTree(ctx, child, remaining-total, includeMore)
					if err == nil && childComment != nil {
						replies = append(replies, childComment)
						total += used
					}
				}
			}
		}

		if c.ShouldInclude() {
			c.Comments = replies
			total++ // Count this comment
			return &c, total, nil
		} else {
			r.logger.Info("skipped comment",
				zap.Int("score", c.Score),
				zap.Int("body_length", len(c.Body)))
		}

		// If this comment isn't included but replies are, pass those up
		if len(replies) > 0 {
			// Return a dummy parent to hold children
			c.Comments = replies
			return &c, total, nil
		}

		// Fully filtered out
		return nil, total, nil

	case "more":
		if !includeMore {
			return nil, remaining, nil
		}
		var more struct {
			Children []string `json:"children"`
			ParentID string   `json:"parent_id"`
			ID       string   `json:"id"`
		}
		if err := json.Unmarshal(kindWrapper.Data, &more); err != nil {
			return nil, 0, err
		}
		moreComment, err := r.fetchMoreComments(ctx, more.Children, more.ParentID, remaining)
		if err != nil || moreComment == nil {
			return nil, remaining, err
		}
		return moreComment, len(moreComment.Comments), nil

	default:
		return nil, 0, nil
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

func (r *Client) fetchMoreComments(ctx context.Context, children []string, parentID string, remaining int) (*Comment, error) {
	if len(children) == 0 || remaining <= 0 {
		return nil, nil
	}

	reqURL := fmt.Sprintf(
		"%s/api/morechildren.json?api_type=json&raw_json=1&children=%s&link_id=%s",
		r.baseURL, strings.Join(children, ","), parentID,
	)

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.config.AccessToken))

	resp, err := r.DoWithRateLimit(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

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
	totalParsed := 0

	for _, thing := range response.Json.Data.Things {
		if totalParsed >= remaining {
			break
		}
		comment, used, err := r.parseCommentTree(ctx, thing, remaining-totalParsed, true)
		if err == nil && comment != nil {
			moreParent.Comments = append(moreParent.Comments, comment)
			totalParsed += used
		}
	}

	// If no comments were added, return nil instead of empty node
	if totalParsed == 0 {
		return nil, nil
	}

	return moreParent, nil
}
