package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/device"
	"backend/internal/generated/ent/order"
	"backend/internal/generated/ent/orderline"
	"backend/internal/repository"

	"github.com/google/uuid"
)

const dashboardTopProductWindow = time.Hour

type DashboardStatusLevel string

const (
	DashboardStatusGreen  DashboardStatusLevel = "green"
	DashboardStatusYellow DashboardStatusLevel = "yellow"
	DashboardStatusRed    DashboardStatusLevel = "red"
)

type SystemStatusChip struct {
	Key     string               `json:"key"`
	Label   string               `json:"label"`
	Status  DashboardStatusLevel `json:"status"`
	Summary string               `json:"summary"`
}

type StationOverviewCard struct {
	ID                      string               `json:"id"`
	Name                    string               `json:"name"`
	Status                  DashboardStatusLevel `json:"status"`
	OpenOrders              int                  `json:"openOrders"`
	Backlog                 int                  `json:"backlog"`
	MedianThroughputMinutes int                  `json:"medianThroughputMinutes"`
	RecentProductTitle      *string              `json:"recentProductTitle,omitempty"`
	RecentProductQuantity   int                  `json:"recentProductQuantity"`
}

type AdminOpsOverview struct {
	OverallStatus DashboardStatusLevel  `json:"overallStatus"`
	StatusChips   []SystemStatusChip    `json:"statusChips"`
	Stations      []StationOverviewCard `json:"stations"`
}

type DashboardTopProductWindowItem struct {
	Title        string `json:"title"`
	Quantity     int    `json:"quantity"`
	RevenueCents int64  `json:"revenueCents"`
}

type DashboardSeriesPoint struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type StationQueueMetric struct {
	OrderID         string   `json:"orderId"`
	CreatedAt       string   `json:"createdAt"`
	AgeMinutes      int      `json:"ageMinutes"`
	PendingItems    int      `json:"pendingItems"`
	PendingQuantity int      `json:"pendingQuantity"`
	Titles          []string `json:"titles"`
}

type StationDetailSummary struct {
	Station          StationOverviewCard             `json:"station"`
	Queue            []StationQueueMetric            `json:"queue"`
	TopProducts      []DashboardTopProductWindowItem `json:"topProducts"`
	ThroughputByHour []DashboardSeriesPoint          `json:"throughputByHour"`
}

type DashboardService interface {
	GetOpsOverview(ctx context.Context) (*AdminOpsOverview, error)
	GetStationDetail(ctx context.Context, stationID uuid.UUID) (*StationDetailSummary, error)
}

type dashboardService struct {
	client   *ent.Client
	devices  repository.DeviceRepository
	settings SettingsService
	payments PaymentService
	stations StationService
}

func NewDashboardService(
	client *ent.Client,
	devices repository.DeviceRepository,
	settings SettingsService,
	payments PaymentService,
	stations StationService,
) DashboardService {
	return &dashboardService{
		client:   client,
		devices:  devices,
		settings: settings,
		payments: payments,
		stations: stations,
	}
}

func (s *dashboardService) GetOpsOverview(ctx context.Context) (*AdminOpsOverview, error) {
	windowStart, windowEnd := currentOpsWindow()

	windowOrders, err := s.loadOrders(ctx, windowStart, windowEnd)
	if err != nil {
		return nil, err
	}

	stations, err := s.stations.ListStations(ctx, nil)
	if err != nil {
		return nil, err
	}

	stationCards := buildStationCards(stations, windowOrders, windowEnd)
	statusChips, overallStatus, err := s.buildStatusChips(ctx, stationCards)
	if err != nil {
		return nil, err
	}

	return &AdminOpsOverview{
		OverallStatus: overallStatus,
		StatusChips:   statusChips,
		Stations:      stationCards,
	}, nil
}

func (s *dashboardService) GetStationDetail(ctx context.Context, stationID uuid.UUID) (*StationDetailSummary, error) {
	windowStart, windowEnd := currentOpsWindow()

	windowOrders, err := s.loadOrders(ctx, windowStart, windowEnd)
	if err != nil {
		return nil, err
	}

	station, err := s.stations.GetStationByID(ctx, stationID)
	if err != nil {
		return nil, err
	}

	productIDs := stationProductIDs(station)
	card, queue, throughput := buildSingleStationMetrics(station, productIDs, windowOrders, windowEnd)

	return &StationDetailSummary{
		Station:          card,
		Queue:            queue,
		TopProducts:      buildStationTopProducts(productIDs, windowOrders, windowEnd),
		ThroughputByHour: throughput,
	}, nil
}

func currentOpsWindow() (time.Time, time.Time) {
	loc := swissLocation()
	now := time.Now().In(loc)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	return start, now
}

func (s *dashboardService) loadOrders(ctx context.Context, from, to time.Time) ([]*ent.Order, error) {
	return s.client.Order.Query().
		Where(
			order.CreatedAtGTE(from),
			order.CreatedAtLT(to),
		).
		WithPayments().
		WithLines(func(q *ent.OrderLineQuery) {
			q.WithRedemption()
		}).
		Order(order.ByCreatedAt()).
		All(ctx)
}

func (s *dashboardService) buildStatusChips(ctx context.Context, stationCards []StationOverviewCard) ([]SystemStatusChip, DashboardStatusLevel, error) {
	systemEnabled, err := s.settings.IsSystemEnabled(ctx)
	if err != nil {
		return nil, "", err
	}

	posDevices, err := s.devices.GetByType(ctx, device.TypePOS)
	if err != nil {
		return nil, "", err
	}
	stationDevices, err := s.devices.GetByType(ctx, device.TypeSTATION)
	if err != nil {
		return nil, "", err
	}

	chips := buildOpsStatusChips(systemEnabled, s.payments.IsPayrexxEnabled(), posDevices, stationDevices, stationCards)
	return chips, overallStatusOf(chips), nil
}

func buildOpsStatusChips(
	systemEnabled bool,
	payrexxEnabled bool,
	posDevices []*ent.Device,
	stationDevices []*ent.Device,
	stationCards []StationOverviewCard,
) []SystemStatusChip {
	posApproved, posPending := countDevices(posDevices)
	stationApproved, stationPending := countDevices(stationDevices)
	redStations, yellowStations, _ := stationPressure(stationCards)

	systemStatus := DashboardStatusGreen
	systemSummary := "System geöffnet"
	if !systemEnabled {
		systemStatus = DashboardStatusRed
		systemSummary = "System geschlossen"
	}

	paymentStatus := DashboardStatusGreen
	paymentSummary := "TWINT/Webhook bereit"
	if !payrexxEnabled {
		paymentStatus = DashboardStatusYellow
		paymentSummary = "TWINT nicht konfiguriert"
	}

	posStatus := DashboardStatusGreen
	posSummary := fmt.Sprintf("%d POS aktiv", posApproved)
	if posApproved == 0 {
		posStatus = DashboardStatusRed
		posSummary = "Kein POS freigegeben"
	} else if posPending > 0 {
		posStatus = DashboardStatusYellow
		posSummary = fmt.Sprintf("%d POS aktiv, %d ausstehend", posApproved, posPending)
	}

	stationStatus := DashboardStatusGreen
	stationSummary := fmt.Sprintf("%d Stationen aktiv", stationApproved)
	if stationApproved == 0 {
		stationStatus = DashboardStatusRed
		stationSummary = "Keine Station freigegeben"
	} else if redStations > 0 {
		stationStatus = DashboardStatusRed
		stationSummary = fmt.Sprintf("%d Stationen kritisch", redStations)
	} else if yellowStations > 0 {
		stationStatus = DashboardStatusYellow
		stationSummary = fmt.Sprintf("%d Stationen beobachten", yellowStations)
	} else if stationPending > 0 {
		stationStatus = DashboardStatusYellow
		stationSummary = fmt.Sprintf("%d Stationen aktiv, %d ausstehend", stationApproved, stationPending)
	}

	return []SystemStatusChip{
		{Key: "system", Label: "System", Status: systemStatus, Summary: systemSummary},
		{Key: "payments", Label: "Payments", Status: paymentStatus, Summary: paymentSummary},
		{Key: "pos", Label: "POS", Status: posStatus, Summary: posSummary},
		{Key: "stations", Label: "Stationen", Status: stationStatus, Summary: stationSummary},
	}
}

func overallStatusOf(chips []SystemStatusChip) DashboardStatusLevel {
	overall := DashboardStatusGreen
	for _, chip := range chips {
		if chip.Status == DashboardStatusRed {
			return DashboardStatusRed
		}
		if chip.Status == DashboardStatusYellow {
			overall = DashboardStatusYellow
		}
	}
	return overall
}

func stationPressure(cards []StationOverviewCard) (red int, yellow int, openOrders int) {
	for _, card := range cards {
		openOrders += card.OpenOrders
		switch card.Status {
		case DashboardStatusRed:
			red++
		case DashboardStatusYellow:
			yellow++
		}
	}
	return red, yellow, openOrders
}

func buildStationCards(stations []*ent.Device, orders []*ent.Order, windowEnd time.Time) []StationOverviewCard {
	cards := make([]StationOverviewCard, 0, len(stations))
	for _, station := range stations {
		if station.Status != device.StatusApproved {
			continue
		}
		card, _, _ := buildSingleStationMetrics(station, stationProductIDs(station), orders, windowEnd)
		cards = append(cards, card)
	}

	sort.Slice(cards, func(i, j int) bool {
		if cards[i].Status != cards[j].Status {
			return severityRank(cards[i].Status) > severityRank(cards[j].Status)
		}
		if cards[i].OpenOrders != cards[j].OpenOrders {
			return cards[i].OpenOrders > cards[j].OpenOrders
		}
		return cards[i].Name < cards[j].Name
	})

	return cards
}

func buildSingleStationMetrics(station *ent.Device, productIDs map[uuid.UUID]struct{}, orders []*ent.Order, windowEnd time.Time) (StationOverviewCard, []StationQueueMetric, []DashboardSeriesPoint) {
	queueByOrder := map[uuid.UUID]*StationQueueMetric{}
	var throughputMinutes []int
	recentProducts := map[string]int{}
	recentStart := windowEnd.Add(-dashboardTopProductWindow)
	redeemedByHour := map[string]int64{}
	loc := swissLocation()

	for _, ord := range orders {
		if ord.Status != order.StatusPaid {
			continue
		}
		for _, line := range flattenOrderLines(ord.Edges.Lines) {
			if _, ok := productIDs[line.ProductID]; !ok {
				continue
			}

			if ord.CreatedAt.After(recentStart) {
				recentProducts[line.Title] += line.Quantity
			}

			if line.Edges.Redemption == nil {
				entry := queueByOrder[ord.ID]
				if entry == nil {
					entry = &StationQueueMetric{
						OrderID:    ord.ID.String(),
						CreatedAt:  ord.CreatedAt.Format(time.RFC3339),
						AgeMinutes: maxInt(0, int(windowEnd.Sub(ord.CreatedAt).Minutes())),
					}
					queueByOrder[ord.ID] = entry
				}
				entry.PendingItems++
				entry.PendingQuantity += line.Quantity
				entry.Titles = append(entry.Titles, line.Title)
				continue
			}

			minutes := int(line.Edges.Redemption.RedeemedAt.Sub(ord.CreatedAt).Minutes())
			if minutes >= 0 {
				throughputMinutes = append(throughputMinutes, minutes)
			}
			hourLabel := line.Edges.Redemption.RedeemedAt.In(loc).Format("15:00")
			redeemedByHour[hourLabel] += int64(line.Quantity)
		}
	}

	queue := make([]StationQueueMetric, 0, len(queueByOrder))
	for _, entry := range queueByOrder {
		queue = append(queue, *entry)
	}
	sort.Slice(queue, func(i, j int) bool {
		if queue[i].AgeMinutes != queue[j].AgeMinutes {
			return queue[i].AgeMinutes > queue[j].AgeMinutes
		}
		return queue[i].OrderID < queue[j].OrderID
	})

	var topTitle *string
	topQty := 0
	for title, qty := range recentProducts {
		if qty > topQty {
			copyTitle := title
			topTitle = &copyTitle
			topQty = qty
		}
	}

	median := medianInt(throughputMinutes)
	openOrders := len(queue)
	status := stationSeverity(openOrders, median)

	points := make([]DashboardSeriesPoint, 0, len(redeemedByHour))
	for label, value := range redeemedByHour {
		points = append(points, DashboardSeriesPoint{Label: label, Value: value})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Label < points[j].Label })

	return StationOverviewCard{
		ID:                      station.ID.String(),
		Name:                    station.Name,
		Status:                  status,
		OpenOrders:              openOrders,
		Backlog:                 openOrders,
		MedianThroughputMinutes: median,
		RecentProductTitle:      topTitle,
		RecentProductQuantity:   topQty,
	}, queue, points
}

func buildStationTopProducts(productIDs map[uuid.UUID]struct{}, orders []*ent.Order, windowEnd time.Time) []DashboardTopProductWindowItem {
	recentStart := windowEnd.Add(-dashboardTopProductWindow)
	type agg struct {
		qty     int
		revenue int64
	}
	byTitle := map[string]*agg{}

	for _, ord := range orders {
		if ord.Status != order.StatusPaid || ord.CreatedAt.Before(recentStart) {
			continue
		}
		for _, line := range flattenOrderLines(ord.Edges.Lines) {
			if _, ok := productIDs[line.ProductID]; !ok {
				continue
			}
			entry := byTitle[line.Title]
			if entry == nil {
				entry = &agg{}
				byTitle[line.Title] = entry
			}
			entry.qty += line.Quantity
			entry.revenue += line.UnitPriceCents * int64(line.Quantity)
		}
	}

	items := make([]DashboardTopProductWindowItem, 0, len(byTitle))
	for title, entry := range byTitle {
		items = append(items, DashboardTopProductWindowItem{
			Title:        title,
			Quantity:     entry.qty,
			RevenueCents: entry.revenue,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Quantity != items[j].Quantity {
			return items[i].Quantity > items[j].Quantity
		}
		return items[i].RevenueCents > items[j].RevenueCents
	})
	if len(items) > 6 {
		items = items[:6]
	}
	return items
}

func stationProductIDs(station *ent.Device) map[uuid.UUID]struct{} {
	productIDs := make(map[uuid.UUID]struct{})
	for _, dp := range station.Edges.DeviceProducts {
		productIDs[dp.ProductID] = struct{}{}
	}
	return productIDs
}

func flattenOrderLines(lines []*ent.OrderLine) []*ent.OrderLine {
	out := make([]*ent.OrderLine, 0, len(lines))
	for _, line := range lines {
		if line.LineType != orderline.LineTypeBundle {
			out = append(out, line)
		}
	}
	return out
}

func countDevices(devices []*ent.Device) (approved int, pending int) {
	for _, dev := range devices {
		switch dev.Status {
		case device.StatusApproved:
			approved++
		case device.StatusPending:
			pending++
		}
	}
	return approved, pending
}

func stationSeverity(openOrders, medianMinutes int) DashboardStatusLevel {
	if openOrders >= 8 || medianMinutes >= 25 {
		return DashboardStatusRed
	}
	if openOrders >= 3 || medianMinutes >= 12 {
		return DashboardStatusYellow
	}
	return DashboardStatusGreen
}

func severityRank(level DashboardStatusLevel) int {
	switch level {
	case DashboardStatusRed:
		return 3
	case DashboardStatusYellow:
		return 2
	default:
		return 1
	}
}

func medianInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	cp := append([]int(nil), values...)
	sort.Ints(cp)
	mid := len(cp) / 2
	if len(cp)%2 == 1 {
		return cp[mid]
	}
	return (cp[mid-1] + cp[mid]) / 2
}

func swissLocation() *time.Location {
	loc, err := time.LoadLocation("Europe/Zurich")
	if err != nil {
		return time.Local
	}
	return loc
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
