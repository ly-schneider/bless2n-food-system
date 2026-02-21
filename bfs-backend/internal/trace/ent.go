package trace

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
)

func QueryInterceptor() ent.InterceptFunc {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, q ent.Query) (ent.Value, error) {
			entity := typeName(q)
			ctx, finish := StartSpan(ctx, "db.query", entity)
			defer finish()
			start := time.Now()
			v, err := next.Query(ctx, q)
			Data(ctx, "duration_ms", time.Since(start).Milliseconds())
			if err != nil {
				Tag(ctx, "error", "true")
			}
			return v, err
		})
	})
}

func MutationHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			desc := fmt.Sprintf("%s %s", m.Op(), m.Type())
			ctx, finish := StartSpan(ctx, "db.mutation", desc)
			defer finish()
			start := time.Now()
			v, err := next.Mutate(ctx, m)
			Data(ctx, "duration_ms", time.Since(start).Milliseconds())
			if err != nil {
				Tag(ctx, "error", "true")
			}
			return v, err
		})
	}
}

func typeName(v any) string {
	t := fmt.Sprintf("%T", v)
	if i := strings.LastIndex(t, "."); i >= 0 {
		t = t[i+1:]
	}
	return strings.TrimPrefix(t, "*")
}
