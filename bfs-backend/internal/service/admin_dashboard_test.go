package service

import (
	"testing"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/device"
	"backend/internal/generated/ent/order"

	"github.com/google/uuid"
)

func TestBuildOpsStatusChipsSystemClosedRed(t *testing.T) {
	chips := buildOpsStatusChips(
		false,
		true,
		[]*ent.Device{{Status: device.StatusApproved, Type: device.TypePOS}},
		[]*ent.Device{{Status: device.StatusApproved, Type: device.TypeSTATION}},
		nil,
	)

	if chip := chipByKey(chips, "system"); chip == nil || chip.Status != DashboardStatusRed {
		t.Fatalf("expected system chip to be red, got %#v", chip)
	}
	if overall := overallStatusOf(chips); overall != DashboardStatusRed {
		t.Fatalf("expected overall status red, got %q", overall)
	}
}

func TestBuildOpsStatusChipsPayrexxDisabledYellow(t *testing.T) {
	chips := buildOpsStatusChips(
		true,
		false,
		[]*ent.Device{{Status: device.StatusApproved, Type: device.TypePOS}},
		[]*ent.Device{{Status: device.StatusApproved, Type: device.TypeSTATION}},
		nil,
	)

	if chip := chipByKey(chips, "payments"); chip == nil || chip.Status != DashboardStatusYellow {
		t.Fatalf("expected payments chip to be yellow, got %#v", chip)
	}
	if overall := overallStatusOf(chips); overall != DashboardStatusYellow {
		t.Fatalf("expected overall status yellow, got %q", overall)
	}
}

func TestBuildOpsStatusChipsNoApprovedPOSRed(t *testing.T) {
	chips := buildOpsStatusChips(
		true,
		true,
		nil,
		[]*ent.Device{{Status: device.StatusApproved, Type: device.TypeSTATION}},
		nil,
	)

	if chip := chipByKey(chips, "pos"); chip == nil || chip.Status != DashboardStatusRed {
		t.Fatalf("expected pos chip to be red, got %#v", chip)
	}
}

func TestBuildOpsStatusChipsStationPressureRed(t *testing.T) {
	chips := buildOpsStatusChips(
		true,
		true,
		[]*ent.Device{{Status: device.StatusApproved, Type: device.TypePOS}},
		[]*ent.Device{{Status: device.StatusApproved, Type: device.TypeSTATION}},
		[]StationOverviewCard{{Status: DashboardStatusRed, OpenOrders: 9}},
	)

	if chip := chipByKey(chips, "stations"); chip == nil || chip.Status != DashboardStatusRed {
		t.Fatalf("expected stations chip to be red, got %#v", chip)
	}
}

func TestBuildSingleStationMetricsCountsQueueAndThroughput(t *testing.T) {
	loc := swissLocation()
	windowEnd := time.Date(2026, 4, 27, 11, 0, 0, 0, loc)
	stationID := uuid.MustParse("01900000-0000-7000-8000-000000000001")
	productID := uuid.MustParse("01900000-0000-7000-8000-000000000101")
	otherProductID := uuid.MustParse("01900000-0000-7000-8000-000000000102")

	orderOneID := uuid.MustParse("01900000-0000-7000-8000-000000000201")
	orderTwoID := uuid.MustParse("01900000-0000-7000-8000-000000000202")

	queueLineID := uuid.MustParse("01900000-0000-7000-8000-000000000301")
	redeemedLineOneID := uuid.MustParse("01900000-0000-7000-8000-000000000302")
	redeemedLineTwoID := uuid.MustParse("01900000-0000-7000-8000-000000000303")
	ignoredLineID := uuid.MustParse("01900000-0000-7000-8000-000000000304")

	orders := []*ent.Order{
		{
			ID:        orderOneID,
			Status:    order.StatusPaid,
			CreatedAt: time.Date(2026, 4, 27, 10, 15, 0, 0, loc),
			Edges: ent.OrderEdges{Lines: []*ent.OrderLine{
				{
					ID:             queueLineID,
					OrderID:        orderOneID,
					ProductID:      productID,
					Title:          "Fries",
					Quantity:       2,
					UnitPriceCents: 900,
				},
			}},
		},
		{
			ID:        orderTwoID,
			Status:    order.StatusPaid,
			CreatedAt: time.Date(2026, 4, 27, 9, 30, 0, 0, loc),
			Edges: ent.OrderEdges{Lines: []*ent.OrderLine{
				{
					ID:             redeemedLineOneID,
					OrderID:        orderTwoID,
					ProductID:      productID,
					Title:          "Fries",
					Quantity:       1,
					UnitPriceCents: 450,
					Edges: ent.OrderLineEdges{Redemption: &ent.OrderLineRedemption{
						OrderLineID: redeemedLineOneID,
						RedeemedAt:  time.Date(2026, 4, 27, 9, 42, 0, 0, loc),
					}},
				},
				{
					ID:             redeemedLineTwoID,
					OrderID:        orderTwoID,
					ProductID:      productID,
					Title:          "Fries",
					Quantity:       3,
					UnitPriceCents: 1350,
					Edges: ent.OrderLineEdges{Redemption: &ent.OrderLineRedemption{
						OrderLineID: redeemedLineTwoID,
						RedeemedAt:  time.Date(2026, 4, 27, 9, 50, 0, 0, loc),
					}},
				},
				{
					ID:             ignoredLineID,
					OrderID:        orderTwoID,
					ProductID:      otherProductID,
					Title:          "Salad",
					Quantity:       5,
					UnitPriceCents: 700,
				},
			}},
		},
	}

	card, queue, throughput := buildSingleStationMetrics(
		&ent.Device{ID: stationID, Name: "Fryer", Status: device.StatusApproved},
		map[uuid.UUID]struct{}{productID: {}},
		orders,
		windowEnd,
	)

	if card.OpenOrders != 1 || card.Backlog != 1 {
		t.Fatalf("expected 1 open order/backlog, got %d/%d", card.OpenOrders, card.Backlog)
	}
	if card.MedianThroughputMinutes != 16 {
		t.Fatalf("expected median throughput 16, got %d", card.MedianThroughputMinutes)
	}
	if card.Status != DashboardStatusYellow {
		t.Fatalf("expected yellow station status, got %q", card.Status)
	}
	if card.RecentProductTitle == nil || *card.RecentProductTitle != "Fries" || card.RecentProductQuantity != 2 {
		t.Fatalf("expected recent product Fries x2, got %#v / %d", card.RecentProductTitle, card.RecentProductQuantity)
	}

	if len(queue) != 1 {
		t.Fatalf("expected one queued order, got %d", len(queue))
	}
	if queue[0].OrderID != orderOneID.String() || queue[0].PendingItems != 1 || queue[0].PendingQuantity != 2 {
		t.Fatalf("unexpected queue metrics: %#v", queue[0])
	}

	if len(throughput) != 1 || throughput[0].Label != "09:00" || throughput[0].Value != 4 {
		t.Fatalf("unexpected throughput series: %#v", throughput)
	}
}

func TestBuildStationTopProductsUsesLastHourOnly(t *testing.T) {
	loc := swissLocation()
	windowEnd := time.Date(2026, 4, 27, 11, 0, 0, 0, loc)
	productID := uuid.MustParse("01900000-0000-7000-8000-000000000101")
	orderRecentID := uuid.MustParse("01900000-0000-7000-8000-000000000201")
	orderOldID := uuid.MustParse("01900000-0000-7000-8000-000000000202")

	orders := []*ent.Order{
		{
			ID:        orderRecentID,
			Status:    order.StatusPaid,
			CreatedAt: time.Date(2026, 4, 27, 10, 30, 0, 0, loc),
			Edges: ent.OrderEdges{Lines: []*ent.OrderLine{
				{
					OrderID:        orderRecentID,
					ProductID:      productID,
					Title:          "Burger",
					Quantity:       2,
					UnitPriceCents: 1200,
				},
			}},
		},
		{
			ID:        orderOldID,
			Status:    order.StatusPaid,
			CreatedAt: time.Date(2026, 4, 27, 8, 30, 0, 0, loc),
			Edges: ent.OrderEdges{Lines: []*ent.OrderLine{
				{
					OrderID:        orderOldID,
					ProductID:      productID,
					Title:          "Burger",
					Quantity:       7,
					UnitPriceCents: 1200,
				},
			}},
		},
	}

	items := buildStationTopProducts(map[uuid.UUID]struct{}{productID: {}}, orders, windowEnd)

	if len(items) != 1 {
		t.Fatalf("expected one top-product entry, got %d", len(items))
	}
	if items[0].Title != "Burger" || items[0].Quantity != 2 || items[0].RevenueCents != 2400 {
		t.Fatalf("unexpected top-product entry: %#v", items[0])
	}
}

func chipByKey(chips []SystemStatusChip, key string) *SystemStatusChip {
	for _, chip := range chips {
		if chip.Key == key {
			copyChip := chip
			return &copyChip
		}
	}
	return nil
}
