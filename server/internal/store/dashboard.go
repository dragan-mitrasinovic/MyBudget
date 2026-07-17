package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const (
	DashboardLayoutVersion                 = 8
	DashboardWidgetTypeCategoryExpenses    = "category-expenses"
	DashboardWidgetTypeCategoryBreakdown   = "category-breakdown"
	DashboardExpenseDisplayNumberOnly      = "number-only"
	DashboardExpenseDisplayAmountOnly      = "amount-only"
	DashboardExpenseDisplayTitleAmount     = "title-amount"
	DashboardExpenseDisplayTitleAmountDate = "title-amount-date"
	DashboardVisualizationPie              = "pie"
	DashboardVisualizationDonut            = "donut"
	DashboardVisualizationBar              = "bar"
	DashboardVisualizationTreemap          = "treemap"
	dashboardOriginalLayoutVersion         = 1
	dashboardFineLayoutVersion             = 2
	dashboardCategoryOnlyLayoutVersion     = 3
	dashboardChartLayoutVersion            = 4
	dashboardBidirectionalLayoutVersion    = 5
	dashboardVisualizationLayoutVersion    = 6
	dashboardExpenseDisplayLayoutVersion   = 7
	dashboardInitialGridColumns            = 36
	dashboardMaximumGridColumns            = 20000
	dashboardDefaultWidgetColumns          = 12
	dashboardDefaultWidgetRows             = 10
	dashboardMinimumWidgetColumns          = 9
	dashboardMaximumWidgetColumns          = 36
	dashboardMinimumWidgetRows             = 6
	dashboardMaximumWidgetRows             = 24
	dashboardMaximumGridRows               = 20000
)

var dashboardCategoryBreakdownWidgetIDPattern = regexp.MustCompile(
	`^category-breakdown:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
)

// DashboardLayout is the versioned arrangement of widgets for one budget.
type DashboardLayout struct {
	Version int               `json:"version"`
	Widgets []DashboardWidget `json:"widgets"`
}

// DashboardWidget describes the identity, type, position, and size of one dashboard widget.
type DashboardWidget struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	CategoryID     int64  `json:"categoryId,omitempty"`
	ExpenseDisplay string `json:"expenseDisplay,omitempty"`
	Visualization  string `json:"visualization,omitempty"`
	X              int    `json:"x"`
	Y              int    `json:"y"`
	Columns        int    `json:"cols"`
	Rows           int    `json:"rows"`
}

// GetDashboardLayout returns the saved layout reconciled with the budget's current
// categories. It returns ErrNotFound when the budget is not owned by the user.
func (s *Store) GetDashboardLayout(ctx context.Context, userID, budgetID int64) (*DashboardLayout, error) {
	owned, err := s.isBudgetOwnedByUser(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}

	if !owned {
		return nil, ErrNotFound
	}

	categoryIDs, err := s.listDashboardCategoryIDs(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}

	layout := DashboardLayout{Version: DashboardLayoutVersion, Widgets: []DashboardWidget{}}
	var layoutJSON string
	const query = `SELECT layout_json FROM dashboard_layouts WHERE budget_id = ?`
	err = s.db.QueryRowContext(ctx, query, budgetID).Scan(&layoutJSON)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == nil {
		if err := json.Unmarshal([]byte(layoutJSON), &layout); err != nil {
			return nil, fmt.Errorf("decode dashboard layout: %w", err)
		}
	}

	return reconcileAndValidateDashboardLayout(layout, categoryIDs)
}

// SaveDashboardLayout validates and replaces the layout for a budget owned by the user.
// Stale category widgets are removed and widgets for new categories are appended.
func (s *Store) SaveDashboardLayout(ctx context.Context, userID, budgetID int64, layout DashboardLayout) (*DashboardLayout, error) {
	owned, err := s.isBudgetOwnedByUser(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}

	if !owned {
		return nil, ErrNotFound
	}

	categoryIDs, err := s.listDashboardCategoryIDs(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}

	normalized, err := reconcileAndValidateDashboardLayout(layout, categoryIDs)
	if err != nil {
		return nil, err
	}

	layoutJSON, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO dashboard_layouts (budget_id, version, layout_json, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(budget_id) DO UPDATE SET
			version = excluded.version,
			layout_json = excluded.layout_json,
			updated_at = excluded.updated_at`

	if _, err := s.db.ExecContext(
		ctx,
		query,
		budgetID,
		normalized.Version,
		string(layoutJSON),
		time.Now().UTC().Format(time.RFC3339),
	); err != nil {
		return nil, err
	}

	return normalized, nil
}

func (s *Store) listDashboardCategoryIDs(ctx context.Context, userID, budgetID int64) ([]int64, error) {
	const query = `
		SELECT c.id
		FROM categories c
		JOIN budgets b ON b.id = c.budget_id
		WHERE c.budget_id = ? AND b.user_id = ?
		ORDER BY c.name COLLATE NOCASE ASC, c.id ASC`

	rows, err := s.db.QueryContext(ctx, query, budgetID, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	categoryIDs := []int64{}
	for rows.Next() {
		var categoryID int64
		if err := rows.Scan(&categoryID); err != nil {
			return nil, err
		}

		categoryIDs = append(categoryIDs, categoryID)
	}

	return categoryIDs, rows.Err()
}

func reconcileAndValidateDashboardLayout(layout DashboardLayout, categoryIDs []int64) (*DashboardLayout, error) {
	if layout.Version != DashboardLayoutVersion &&
		layout.Version != dashboardOriginalLayoutVersion &&
		layout.Version != dashboardFineLayoutVersion &&
		layout.Version != dashboardCategoryOnlyLayoutVersion &&
		layout.Version != dashboardChartLayoutVersion &&
		layout.Version != dashboardBidirectionalLayoutVersion &&
		layout.Version != dashboardVisualizationLayoutVersion &&
		layout.Version != dashboardExpenseDisplayLayoutVersion {
		return nil, fmt.Errorf("%w: version must be %d", ErrInvalidDashboardLayout, DashboardLayoutVersion)
	}

	validCategoryIDs := make(map[int64]bool, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		validCategoryIDs[categoryID] = true
	}

	seenCategoryIDs := make(map[int64]bool, len(categoryIDs))
	seenWidgetIDs := make(map[string]bool, len(layout.Widgets))
	widgets := make([]DashboardWidget, 0, len(layout.Widgets)+len(categoryIDs))
	for _, widget := range layout.Widgets {
		switch widget.Type {
		case DashboardWidgetTypeCategoryExpenses:
			if widget.CategoryID <= 0 ||
				widget.ID != dashboardCategoryWidgetID(widget.CategoryID) ||
				widget.Visualization != "" {
				return nil, fmt.Errorf("%w: invalid category widget", ErrInvalidDashboardLayout)
			}

			if layout.Version < dashboardExpenseDisplayLayoutVersion && widget.ExpenseDisplay == "" {
				widget.ExpenseDisplay = DashboardExpenseDisplayTitleAmount
			}

			if !isDashboardExpenseDisplaySupported(widget.ExpenseDisplay) {
				return nil, fmt.Errorf("%w: invalid category expense display", ErrInvalidDashboardLayout)
			}

			if !validCategoryIDs[widget.CategoryID] {
				continue
			}

			if seenCategoryIDs[widget.CategoryID] {
				return nil, fmt.Errorf("%w: duplicate category widget", ErrInvalidDashboardLayout)
			}
		case DashboardWidgetTypeCategoryBreakdown:
			if layout.Version < dashboardChartLayoutVersion ||
				widget.CategoryID != 0 ||
				widget.ExpenseDisplay != "" ||
				!dashboardCategoryBreakdownWidgetIDPattern.MatchString(widget.ID) {
				return nil, fmt.Errorf("%w: invalid category breakdown widget", ErrInvalidDashboardLayout)
			}

			if (layout.Version == dashboardChartLayoutVersion ||
				layout.Version == dashboardBidirectionalLayoutVersion) && widget.Visualization == "" {
				widget.Visualization = DashboardVisualizationPie
			}

			if !isDashboardVisualizationSupported(widget.Visualization) {
				return nil, fmt.Errorf("%w: invalid category breakdown visualization", ErrInvalidDashboardLayout)
			}
		default:
			return nil, fmt.Errorf("%w: unknown widget type", ErrInvalidDashboardLayout)
		}

		switch layout.Version {
		case dashboardOriginalLayoutVersion:
			if err := validateDashboardWidgetGeometryWithin(widget, 3, 12, 3, 12, 12, 10000); err != nil {
				return nil, err
			}

			widget = scaleDashboardWidget(widget, 3, 2)
		case dashboardFineLayoutVersion:
			if err := validateDashboardWidgetGeometryWithin(widget, 18, 72, 12, 48, 72, 40000); err != nil {
				return nil, err
			}

			widget = halveDashboardWidget(widget)
		case dashboardCategoryOnlyLayoutVersion, dashboardChartLayoutVersion:
			if err := validateDashboardWidgetGeometryWithin(
				widget,
				dashboardMinimumWidgetColumns,
				dashboardMaximumWidgetColumns,
				dashboardMinimumWidgetRows,
				dashboardMaximumWidgetRows,
				dashboardInitialGridColumns,
				dashboardMaximumGridRows,
			); err != nil {
				return nil, err
			}
		}

		if seenWidgetIDs[widget.ID] {
			return nil, fmt.Errorf("%w: duplicate widget ID", ErrInvalidDashboardLayout)
		}

		if err := validateDashboardWidgetGeometry(widget); err != nil {
			return nil, err
		}

		for _, existing := range widgets {
			if dashboardWidgetsOverlap(widget, existing) {
				return nil, fmt.Errorf("%w: widgets overlap", ErrInvalidDashboardLayout)
			}
		}

		seenWidgetIDs[widget.ID] = true
		if widget.Type == DashboardWidgetTypeCategoryExpenses {
			seenCategoryIDs[widget.CategoryID] = true
		}
		widgets = append(widgets, widget)
	}

	for _, categoryID := range categoryIDs {
		if seenCategoryIDs[categoryID] {
			continue
		}

		widget, err := nextDashboardWidgetPosition(categoryID, widgets)
		if err != nil {
			return nil, err
		}

		widgets = append(widgets, widget)
	}

	return &DashboardLayout{Version: DashboardLayoutVersion, Widgets: widgets}, nil
}

func isDashboardExpenseDisplaySupported(expenseDisplay string) bool {
	switch expenseDisplay {
	case DashboardExpenseDisplayNumberOnly,
		DashboardExpenseDisplayAmountOnly,
		DashboardExpenseDisplayTitleAmount,
		DashboardExpenseDisplayTitleAmountDate:
		return true
	default:
		return false
	}
}

func isDashboardVisualizationSupported(visualization string) bool {
	switch visualization {
	case DashboardVisualizationPie,
		DashboardVisualizationDonut,
		DashboardVisualizationBar,
		DashboardVisualizationTreemap:
		return true
	default:
		return false
	}
}

func validateDashboardWidgetGeometry(widget DashboardWidget) error {
	return validateDashboardWidgetGeometryWithin(
		widget,
		dashboardMinimumWidgetColumns,
		dashboardMaximumWidgetColumns,
		dashboardMinimumWidgetRows,
		dashboardMaximumWidgetRows,
		dashboardMaximumGridColumns,
		dashboardMaximumGridRows,
	)
}

func validateDashboardWidgetGeometryWithin(widget DashboardWidget, minColumns, maxColumns, minRows, maxRows, gridColumns, gridRows int) error {
	if widget.Columns < minColumns || widget.Columns > maxColumns ||
		widget.Rows < minRows || widget.Rows > maxRows ||
		widget.X < 0 || widget.X+widget.Columns > gridColumns ||
		widget.Y < 0 || widget.Y+widget.Rows > gridRows {
		return fmt.Errorf("%w: widget geometry is outside the grid", ErrInvalidDashboardLayout)
	}

	return nil
}

func scaleDashboardWidget(widget DashboardWidget, columnScale, rowScale int) DashboardWidget {
	widget.X *= columnScale
	widget.Y *= rowScale
	widget.Columns *= columnScale
	widget.Rows *= rowScale
	return widget
}

func halveDashboardWidget(widget DashboardWidget) DashboardWidget {
	right := roundFineGridCoordinate(widget.X + widget.Columns)
	bottom := roundFineGridCoordinate(widget.Y + widget.Rows)
	widget.X = roundFineGridCoordinate(widget.X)
	widget.Y = roundFineGridCoordinate(widget.Y)
	widget.Columns = right - widget.X
	widget.Rows = bottom - widget.Y
	return widget
}

func roundFineGridCoordinate(value int) int {
	return (value + 1) / 2
}

func dashboardWidgetsOverlap(first, second DashboardWidget) bool {
	return first.X < second.X+second.Columns &&
		first.X+first.Columns > second.X &&
		first.Y < second.Y+second.Rows &&
		first.Y+first.Rows > second.Y
}

func nextDashboardWidgetPosition(categoryID int64, widgets []DashboardWidget) (DashboardWidget, error) {
	widget := DashboardWidget{
		ID:             dashboardCategoryWidgetID(categoryID),
		Type:           DashboardWidgetTypeCategoryExpenses,
		CategoryID:     categoryID,
		ExpenseDisplay: DashboardExpenseDisplayTitleAmount,
		Columns:        dashboardDefaultWidgetColumns,
		Rows:           dashboardDefaultWidgetRows,
	}

	for y := 0; y <= dashboardMaximumGridRows-dashboardDefaultWidgetRows; y++ {
		for x := 0; x <= dashboardInitialGridColumns-dashboardDefaultWidgetColumns; x++ {
			widget.X = x
			widget.Y = y
			isAvailable := true
			for _, existing := range widgets {
				if dashboardWidgetsOverlap(widget, existing) {
					isAvailable = false
					break
				}
			}

			if isAvailable {
				return widget, nil
			}
		}
	}

	return DashboardWidget{}, fmt.Errorf("%w: no grid space available", ErrInvalidDashboardLayout)
}

func dashboardCategoryWidgetID(categoryID int64) string {
	return DashboardWidgetTypeCategoryExpenses + ":" + strconv.FormatInt(categoryID, 10)
}
