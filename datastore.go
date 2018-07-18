package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"net/http"

	"google.golang.org/appengine/datastore"
)

type Package struct {
	GoGetCount int64
	GoDocCount int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func incrementPackageCounts(ctx context.Context, r *http.Request) error {
	updateGoGetCount := strings.HasPrefix(r.UserAgent(), "Go-http-client")
	updateGoDocCount := strings.HasPrefix(r.UserAgent(), "GoDocBot")

	if !updateGoGetCount && !updateGoDocCount {
		return nil
	}

	path := strings.Split(r.URL.Path, "/")
	if len(path) < 2 || path[1] == "" {
		return nil
	}

	now := time.Now()
	date := now.Format("2006-01-02")
	parentPkg := fmt.Sprintf("%s/%s", r.URL.Host, path[1])
	pkg := fmt.Sprintf("%s%s", r.URL.Host, r.URL.Path)

	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		dateKey := datastore.NewKey(ctx, "Date", date, 0, nil)
		libraryKey := datastore.NewKey(ctx, "Library", parentPkg, 0, dateKey)
		key := datastore.NewKey(ctx, "Package", pkg, 0, libraryKey)

		var x Package
		err := datastore.Get(ctx, key, &x)
		if err == datastore.ErrNoSuchEntity {
			x.CreatedAt = now
		} else if err != nil {
			return err
		}

		if updateGoGetCount {
			x.GoGetCount++
		}

		if updateGoDocCount {
			x.GoDocCount++
		}

		x.UpdatedAt = now
		if _, err := datastore.Put(ctx, key, &x); err != nil {
			return err
		}
		return nil
	}, nil)

	return err
}
