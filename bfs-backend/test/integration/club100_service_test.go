package integration

import (
	"context"
	"sync/atomic"
	"testing"

	"backend/internal/generated/ent"
	entOrder "backend/internal/generated/ent/order"
	"backend/internal/repository"
	"backend/internal/service"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/require"
)

type countingDriver struct {
	dialect.Driver
	queries int64
	execs   int64
}

func (d *countingDriver) Query(ctx context.Context, query string, args, v any) error {
	atomic.AddInt64(&d.queries, 1)
	return d.Driver.Query(ctx, query, args, v)
}

func (d *countingDriver) Exec(ctx context.Context, query string, args, v any) error {
	atomic.AddInt64(&d.execs, 1)
	return d.Driver.Exec(ctx, query, args, v)
}

func TestClub100Service_GetPeopleWithRedemptions_NoNPlusOne(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()
	tdb.Cleanup(t)

	ctx := context.Background()

	repos := NewRepositories(tdb.Client)
	fixtures := NewFixtures(repos)
	require.NoError(t, repos.Settings.UpdateClub100Settings(ctx, nil, 2))

	people := []service.ElvantoPerson{
		{ID: "p1", FirstName: "Alice", LastName: "A"},
		{ID: "p2", FirstName: "Bob", LastName: "B"},
		{ID: "p3", FirstName: "Carol", LastName: "C"},
		{ID: "p4", FirstName: "Dave", LastName: "D"},
		{ID: "p5", FirstName: "Eve", LastName: "E"},
	}

	// Redemptions reference real orders (order_id is a NOT NULL FK to "order").
	for _, p := range people {
		ord := fixtures.CreateOrder(0, entOrder.StatusPaid, entOrder.OriginShop)
		_, err := repos.Club100Redemption.Create(ctx, p.ID, p.FirstName+" "+p.LastName, ord.ID, 1)
		require.NoError(t, err)
	}
	dupOrder := fixtures.CreateOrder(0, entOrder.StatusPaid, entOrder.OriginShop)
	_, err := repos.Club100Redemption.Create(ctx, "p1", "Alice A", dupOrder.ID, 1)
	require.NoError(t, err)

	counter := &countingDriver{Driver: entsql.OpenDB(dialect.Postgres, tdb.DB)}
	countedClient := ent.NewClient(ent.Driver(counter))

	countedRedemptions := repository.NewClub100RedemptionRepository(countedClient)
	countedSettings := repository.NewSettingsRepository(countedClient)
	countedOrderLines := repository.NewOrderLineRepository(countedClient)

	elvanto := &MockElvantoService{Configured: true, People: people}
	svc := service.NewClub100Service(elvanto, countedRedemptions, countedSettings, countedOrderLines)

	atomic.StoreInt64(&counter.queries, 0)
	atomic.StoreInt64(&counter.execs, 0)

	result, err := svc.GetPeopleWithRedemptions(ctx)
	require.NoError(t, err)
	require.Len(t, result, len(people))

	byID := make(map[string]service.Club100Person, len(result))
	for _, r := range result {
		byID[r.ID] = r
	}
	require.Equal(t, 2, byID["p1"].TotalRedemptions)
	require.Equal(t, 0, byID["p1"].Remaining)
	require.Equal(t, 1, byID["p2"].TotalRedemptions)
	require.Equal(t, 1, byID["p2"].Remaining)
	require.Equal(t, 2, byID["p1"].Max)

	queries := atomic.LoadInt64(&counter.queries)
	require.LessOrEqualf(t, queries, int64(3),
		"expected at most 3 SELECT queries (settings + batch redemptions), got %d — N+1 regression in /v1/club100/people",
		queries)
}
