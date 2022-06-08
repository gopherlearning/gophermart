package internal

import "context"

type contextKey string

// PathKey .
const PathKey contextKey = "path"

// GetPath .
func GetPath(ctx context.Context) string {
	v := ctx.Value(PathKey)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
