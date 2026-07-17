package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestGetDashboardLayoutCreatesWidgetForEveryCategory(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	food, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create Food category: %v", err)
	}

	transport, err := st.CreateCategory(ctx, user.ID, budget.ID, "Transport")
	if err != nil {
		t.Fatalf("create Transport category: %v", err)
	}

	layout, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get dashboard layout: %v", err)
	}

	if layout.Version != DashboardLayoutVersion {
		t.Errorf("layout version = %d, want %d", layout.Version, DashboardLayoutVersion)
	}

	if len(layout.Widgets) != 2 {
		t.Fatalf("widget count = %d, want 2", len(layout.Widgets))
	}

	widgetCategoryIDs := map[int64]bool{}
	for _, widget := range layout.Widgets {
		widgetCategoryIDs[widget.CategoryID] = true
		if widget.ID != dashboardCategoryWidgetID(widget.CategoryID) {
			t.Errorf("widget ID = %q, want %q", widget.ID, dashboardCategoryWidgetID(widget.CategoryID))
		}
	}

	for _, categoryID := range []int64{food.ID, transport.ID} {
		if !widgetCategoryIDs[categoryID] {
			t.Errorf("category %d has no widget", categoryID)
		}
	}
}

func TestGetDashboardLayoutReturnsEmptyLayoutForOwnedBudgetWithoutCategories(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")

	layout, err := st.GetDashboardLayout(context.Background(), user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get dashboard layout: %v", err)
	}

	if layout.Version != DashboardLayoutVersion || len(layout.Widgets) != 0 {
		t.Fatalf("empty dashboard layout = %#v, want version %d with no widgets", layout, DashboardLayoutVersion)
	}
}

func TestDashboardLayoutPersistsAndReconcilesCategoryChanges(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	food, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create Food category: %v", err)
	}

	layout, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get initial layout: %v", err)
	}

	for index := range layout.Widgets {
		if layout.Widgets[index].CategoryID == food.ID {
			layout.Widgets[index].X = 18
			layout.Widgets[index].Y = 14
			layout.Widgets[index].Columns = 18
			layout.Widgets[index].Rows = 8
		}
	}

	saved, err := st.SaveDashboardLayout(ctx, user.ID, budget.ID, *layout)
	if err != nil {
		t.Fatalf("save layout: %v", err)
	}

	if len(saved.Widgets) != 1 {
		t.Fatalf("saved widget count = %d, want 1", len(saved.Widgets))
	}

	utilities, err := st.CreateCategory(ctx, user.ID, budget.ID, "Utilities")
	if err != nil {
		t.Fatalf("create Utilities category: %v", err)
	}

	if err := st.DeleteCategory(ctx, user.ID, food.ID, nil); err != nil {
		t.Fatalf("delete Food category: %v", err)
	}

	reconciled, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get reconciled layout: %v", err)
	}

	if len(reconciled.Widgets) != 1 {
		t.Fatalf("reconciled widget count = %d, want 1", len(reconciled.Widgets))
	}

	for _, widget := range reconciled.Widgets {
		if widget.CategoryID == food.ID {
			t.Error("deleted category widget is still present")
		}
	}

	if !hasDashboardWidgetForCategory(reconciled.Widgets, utilities.ID) {
		t.Error("new category widget was not added")
	}
}

func TestDashboardLayoutPersistsExpenseDisplay(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()
	createDashboardTestCategory(t, st, user.ID, budget.ID)

	layout, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get dashboard layout: %v", err)
	}

	if layout.Widgets[0].ExpenseDisplay != DashboardExpenseDisplayTitleAmount {
		t.Fatalf("default expense display = %q, want %q", layout.Widgets[0].ExpenseDisplay, DashboardExpenseDisplayTitleAmount)
	}

	layout.Widgets[0].ExpenseDisplay = DashboardExpenseDisplayNumberOnly
	if _, err := st.SaveDashboardLayout(ctx, user.ID, budget.ID, *layout); err != nil {
		t.Fatalf("save dashboard layout: %v", err)
	}

	persisted, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get persisted dashboard layout: %v", err)
	}

	if persisted.Widgets[0].ExpenseDisplay != DashboardExpenseDisplayNumberOnly {
		t.Errorf("persisted expense display = %q, want %q", persisted.Widgets[0].ExpenseDisplay, DashboardExpenseDisplayNumberOnly)
	}
}

func TestDashboardLayoutPersistsMultipleCategoryBreakdownWidgets(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	layout, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get initial layout: %v", err)
	}

	charts := []DashboardWidget{
		{
			ID:            "category-breakdown:11111111-1111-4111-8111-111111111111",
			Type:          DashboardWidgetTypeCategoryBreakdown,
			Visualization: DashboardVisualizationDonut,
			X:             0,
			Y:             10,
			Columns:       18,
			Rows:          12,
		},
		{
			ID:            "category-breakdown:22222222-2222-4222-8222-222222222222",
			Type:          DashboardWidgetTypeCategoryBreakdown,
			Visualization: DashboardVisualizationTreemap,
			X:             18,
			Y:             10,
			Columns:       18,
			Rows:          12,
		},
	}
	layout.Widgets = append(layout.Widgets, charts...)

	if _, err := st.SaveDashboardLayout(ctx, user.ID, budget.ID, *layout); err != nil {
		t.Fatalf("save layout with charts: %v", err)
	}

	if _, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food"); err != nil {
		t.Fatalf("create category: %v", err)
	}

	reconciled, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get reconciled layout: %v", err)
	}

	chartCount := 0
	for _, widget := range reconciled.Widgets {
		if widget.Type == DashboardWidgetTypeCategoryBreakdown {
			chartCount++
		}
	}

	if chartCount != 2 {
		t.Errorf("chart count = %d, want 2", chartCount)
	}
}

func TestDashboardLayoutAllowsWidgetsBeyondInitialColumns(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()
	createDashboardTestCategory(t, st, user.ID, budget.ID)

	layout, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get dashboard layout: %v", err)
	}

	layout.Widgets[0].X = dashboardInitialGridColumns
	saved, err := st.SaveDashboardLayout(ctx, user.ID, budget.ID, *layout)
	if err != nil {
		t.Fatalf("save horizontally extended layout: %v", err)
	}

	if saved.Widgets[0].X != dashboardInitialGridColumns {
		t.Errorf("saved widget x = %d, want %d", saved.Widgets[0].X, dashboardInitialGridColumns)
	}
}

func TestDashboardLayoutAcceptsAllSupportedVisualizations(t *testing.T) {
	visualizations := []string{
		DashboardVisualizationPie,
		DashboardVisualizationDonut,
		DashboardVisualizationBar,
		DashboardVisualizationTreemap,
	}

	for index, visualization := range visualizations {
		layout := DashboardLayout{
			Version: DashboardLayoutVersion,
			Widgets: []DashboardWidget{
				{
					ID:             dashboardCategoryWidgetID(1),
					Type:           DashboardWidgetTypeCategoryExpenses,
					CategoryID:     1,
					ExpenseDisplay: DashboardExpenseDisplayTitleAmount,
					Columns:        12,
					Rows:           10,
				},
				{
					ID:            fmt.Sprintf("category-breakdown:%08d-1111-4111-8111-111111111111", index+1),
					Type:          DashboardWidgetTypeCategoryBreakdown,
					Visualization: visualization,
					Y:             10,
					Columns:       18,
					Rows:          12,
				},
			},
		}

		normalized, err := reconcileAndValidateDashboardLayout(layout, []int64{1})
		if err != nil {
			t.Fatalf("normalize %q layout: %v", visualization, err)
		}

		if normalized.Widgets[1].Visualization != visualization {
			t.Errorf("normalized visualization = %q, want %q", normalized.Widgets[1].Visualization, visualization)
		}
	}
}

func TestGetDashboardLayoutUpgradesVersionOneGeometry(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	category := createDashboardTestCategory(t, st, user.ID, budget.ID)

	legacyLayout := DashboardLayout{
		Version: dashboardOriginalLayoutVersion,
		Widgets: []DashboardWidget{
			{
				ID:         dashboardCategoryWidgetID(category.ID),
				Type:       DashboardWidgetTypeCategoryExpenses,
				CategoryID: category.ID,
				X:          2,
				Y:          3,
				Columns:    4,
				Rows:       5,
			},
		},
	}
	layoutJSON, err := json.Marshal(legacyLayout)
	if err != nil {
		t.Fatalf("encode legacy layout: %v", err)
	}

	const insert = `
		INSERT INTO dashboard_layouts (budget_id, version, layout_json, updated_at)
		VALUES (?, ?, ?, '2026-07-12T00:00:00Z')`
	if _, err := st.db.ExecContext(ctx, insert, budget.ID, legacyLayout.Version, string(layoutJSON)); err != nil {
		t.Fatalf("insert legacy layout: %v", err)
	}

	upgraded, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get upgraded layout: %v", err)
	}

	if upgraded.Version != DashboardLayoutVersion {
		t.Errorf("upgraded version = %d, want %d", upgraded.Version, DashboardLayoutVersion)
	}

	widget := upgraded.Widgets[0]
	if widget.X != 6 || widget.Y != 6 || widget.Columns != 12 || widget.Rows != 10 {
		t.Errorf("upgraded geometry = (%d, %d, %d, %d), want (6, 6, 12, 10)", widget.X, widget.Y, widget.Columns, widget.Rows)
	}
}

func TestGetDashboardLayoutUpgradesVersionTwoGeometry(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	category := createDashboardTestCategory(t, st, user.ID, budget.ID)

	fineLayout := DashboardLayout{
		Version: dashboardFineLayoutVersion,
		Widgets: []DashboardWidget{
			{
				ID:         dashboardCategoryWidgetID(category.ID),
				Type:       DashboardWidgetTypeCategoryExpenses,
				CategoryID: category.ID,
				X:          1,
				Y:          1,
				Columns:    25,
				Rows:       21,
			},
		},
	}
	layoutJSON, err := json.Marshal(fineLayout)
	if err != nil {
		t.Fatalf("encode version two layout: %v", err)
	}

	const insert = `
		INSERT INTO dashboard_layouts (budget_id, version, layout_json, updated_at)
		VALUES (?, ?, ?, '2026-07-12T00:00:00Z')`
	if _, err := st.db.ExecContext(ctx, insert, budget.ID, fineLayout.Version, string(layoutJSON)); err != nil {
		t.Fatalf("insert version two layout: %v", err)
	}

	upgraded, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get upgraded layout: %v", err)
	}

	widget := upgraded.Widgets[0]
	if widget.X != 1 || widget.Y != 1 || widget.Columns != 12 || widget.Rows != 10 {
		t.Errorf("upgraded geometry = (%d, %d, %d, %d), want (1, 1, 12, 10)", widget.X, widget.Y, widget.Columns, widget.Rows)
	}
}

func TestGetDashboardLayoutUpgradesVersionThreeWithoutChangingGeometry(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	category := createDashboardTestCategory(t, st, user.ID, budget.ID)

	legacyLayout := DashboardLayout{
		Version: dashboardCategoryOnlyLayoutVersion,
		Widgets: []DashboardWidget{
			{
				ID:         dashboardCategoryWidgetID(category.ID),
				Type:       DashboardWidgetTypeCategoryExpenses,
				CategoryID: category.ID,
				X:          6,
				Y:          7,
				Columns:    18,
				Rows:       11,
			},
		},
	}
	layoutJSON, err := json.Marshal(legacyLayout)
	if err != nil {
		t.Fatalf("encode version three layout: %v", err)
	}

	const insert = `
		INSERT INTO dashboard_layouts (budget_id, version, layout_json, updated_at)
		VALUES (?, ?, ?, '2026-07-12T00:00:00Z')`
	if _, err := st.db.ExecContext(ctx, insert, budget.ID, legacyLayout.Version, string(layoutJSON)); err != nil {
		t.Fatalf("insert version three layout: %v", err)
	}

	upgraded, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get upgraded layout: %v", err)
	}

	widget := upgraded.Widgets[0]
	if upgraded.Version != DashboardLayoutVersion ||
		widget.X != 6 || widget.Y != 7 || widget.Columns != 18 || widget.Rows != 11 {
		t.Errorf("upgraded layout = %#v, want current version with unchanged geometry", upgraded)
	}
}

func TestGetDashboardLayoutUpgradesVersionFourWithChart(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	category := createDashboardTestCategory(t, st, user.ID, budget.ID)

	legacyLayout := DashboardLayout{
		Version: dashboardChartLayoutVersion,
		Widgets: []DashboardWidget{
			{
				ID:         dashboardCategoryWidgetID(category.ID),
				Type:       DashboardWidgetTypeCategoryExpenses,
				CategoryID: category.ID,
				Columns:    12,
				Rows:       10,
			},
			{
				ID:      "category-breakdown:11111111-1111-4111-8111-111111111111",
				Type:    DashboardWidgetTypeCategoryBreakdown,
				X:       12,
				Columns: 18,
				Rows:    12,
			},
		},
	}
	layoutJSON, err := json.Marshal(legacyLayout)
	if err != nil {
		t.Fatalf("encode version four layout: %v", err)
	}

	const insert = `
		INSERT INTO dashboard_layouts (budget_id, version, layout_json, updated_at)
		VALUES (?, ?, ?, '2026-07-12T00:00:00Z')`
	if _, err := st.db.ExecContext(ctx, insert, budget.ID, legacyLayout.Version, string(layoutJSON)); err != nil {
		t.Fatalf("insert version four layout: %v", err)
	}

	upgraded, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get upgraded layout: %v", err)
	}

	if upgraded.Version != DashboardLayoutVersion || len(upgraded.Widgets) != 2 {
		t.Fatalf("upgraded layout = %#v, want current version with two widgets", upgraded)
	}

	chart := upgraded.Widgets[1]
	if chart.Visualization != DashboardVisualizationPie ||
		chart.X != 12 || chart.Columns != 18 || chart.Rows != 12 {
		t.Errorf("upgraded chart = %#v, want Pie with unchanged geometry", chart)
	}
}

func TestGetDashboardLayoutUpgradesVersionFiveWithExplicitVisualization(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	category := createDashboardTestCategory(t, st, user.ID, budget.ID)

	legacyLayout := DashboardLayout{
		Version: dashboardBidirectionalLayoutVersion,
		Widgets: []DashboardWidget{
			{
				ID:         dashboardCategoryWidgetID(category.ID),
				Type:       DashboardWidgetTypeCategoryExpenses,
				CategoryID: category.ID,
				Columns:    12,
				Rows:       10,
			},
			{
				ID:            "category-breakdown:11111111-1111-4111-8111-111111111111",
				Type:          DashboardWidgetTypeCategoryBreakdown,
				Visualization: DashboardVisualizationBar,
				X:             48,
				Columns:       18,
				Rows:          12,
			},
		},
	}
	layoutJSON, err := json.Marshal(legacyLayout)
	if err != nil {
		t.Fatalf("encode version five layout: %v", err)
	}

	const insert = `
		INSERT INTO dashboard_layouts (budget_id, version, layout_json, updated_at)
		VALUES (?, ?, ?, '2026-07-12T00:00:00Z')`
	if _, err := st.db.ExecContext(ctx, insert, budget.ID, legacyLayout.Version, string(layoutJSON)); err != nil {
		t.Fatalf("insert version five layout: %v", err)
	}

	upgraded, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get upgraded layout: %v", err)
	}

	chart := upgraded.Widgets[1]
	if upgraded.Version != DashboardLayoutVersion ||
		chart.Visualization != DashboardVisualizationBar || chart.X != 48 {
		t.Errorf("upgraded layout = %#v, want Bar at x=48", upgraded)
	}
}

func TestGetDashboardLayoutUpgradesVersionSixWithDefaultExpenseDisplay(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()
	category := createDashboardTestCategory(t, st, user.ID, budget.ID)

	legacyLayout := DashboardLayout{
		Version: dashboardVisualizationLayoutVersion,
		Widgets: []DashboardWidget{
			{
				ID:         dashboardCategoryWidgetID(category.ID),
				Type:       DashboardWidgetTypeCategoryExpenses,
				CategoryID: category.ID,
				Columns:    12,
				Rows:       10,
			},
		},
	}
	layoutJSON, err := json.Marshal(legacyLayout)
	if err != nil {
		t.Fatalf("encode version six layout: %v", err)
	}

	const insert = `
		INSERT INTO dashboard_layouts (budget_id, version, layout_json, updated_at)
		VALUES (?, ?, ?, '2026-07-17T00:00:00Z')`
	if _, err := st.db.ExecContext(ctx, insert, budget.ID, legacyLayout.Version, string(layoutJSON)); err != nil {
		t.Fatalf("insert version six layout: %v", err)
	}

	upgraded, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get upgraded layout: %v", err)
	}

	if upgraded.Version != DashboardLayoutVersion ||
		upgraded.Widgets[0].ExpenseDisplay != DashboardExpenseDisplayTitleAmount {
		t.Errorf("upgraded layout = %#v, want current version with title and amount display", upgraded)
	}
}

func TestGetDashboardLayoutUpgradesVersionSevenWithExpenseDisplay(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()
	category := createDashboardTestCategory(t, st, user.ID, budget.ID)

	legacyLayout := DashboardLayout{
		Version: dashboardExpenseDisplayLayoutVersion,
		Widgets: []DashboardWidget{
			{
				ID:             dashboardCategoryWidgetID(category.ID),
				Type:           DashboardWidgetTypeCategoryExpenses,
				CategoryID:     category.ID,
				ExpenseDisplay: DashboardExpenseDisplayAmountOnly,
				Columns:        12,
				Rows:           10,
			},
		},
	}
	layoutJSON, err := json.Marshal(legacyLayout)
	if err != nil {
		t.Fatalf("encode version seven layout: %v", err)
	}

	const insert = `
		INSERT INTO dashboard_layouts (budget_id, version, layout_json, updated_at)
		VALUES (?, ?, ?, '2026-07-17T00:00:00Z')`
	if _, err := st.db.ExecContext(ctx, insert, budget.ID, legacyLayout.Version, string(layoutJSON)); err != nil {
		t.Fatalf("insert version seven layout: %v", err)
	}

	upgraded, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get upgraded layout: %v", err)
	}

	if upgraded.Version != DashboardLayoutVersion ||
		upgraded.Widgets[0].ExpenseDisplay != DashboardExpenseDisplayAmountOnly {
		t.Errorf("upgraded layout = %#v, want current version preserving amount-only display", upgraded)
	}
}

func TestSaveDashboardLayoutValidatesOwnershipAndGeometry(t *testing.T) {
	st := openTestStore(t)
	owner := createStoreTestUser(t, st, "dashboard-owner")
	other := createStoreTestUser(t, st, "other-user")
	budget := createStoreTestBudget(t, st, owner.ID, "Household")
	ctx := context.Background()
	createDashboardTestCategory(t, st, owner.ID, budget.ID)

	if _, err := st.GetDashboardLayout(ctx, other.ID, budget.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("other user's get error = %v, want ErrNotFound", err)
	}

	layout, err := st.GetDashboardLayout(ctx, owner.ID, budget.ID)
	if err != nil {
		t.Fatalf("get owner layout: %v", err)
	}

	layout.Widgets[0].Columns = dashboardMaximumWidgetColumns + 1
	if _, err := st.SaveDashboardLayout(ctx, owner.ID, budget.ID, *layout); !errors.Is(err, ErrInvalidDashboardLayout) {
		t.Errorf("invalid geometry error = %v, want ErrInvalidDashboardLayout", err)
	}

	if _, err := st.SaveDashboardLayout(ctx, other.ID, budget.ID, *layout); !errors.Is(err, ErrNotFound) {
		t.Errorf("other user's save error = %v, want ErrNotFound", err)
	}
}

func TestSaveDashboardLayoutRejectsInvalidCategoryBreakdownWidgets(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()
	category := createDashboardTestCategory(t, st, user.ID, budget.ID)

	layout, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get layout: %v", err)
	}

	invalidChart := DashboardWidget{
		ID:         "category-breakdown:not-a-uuid",
		Type:       DashboardWidgetTypeCategoryBreakdown,
		CategoryID: category.ID,
		X:          0,
		Y:          10,
		Columns:    18,
		Rows:       12,
	}
	layout.Widgets = append(layout.Widgets, invalidChart)

	if _, err := st.SaveDashboardLayout(ctx, user.ID, budget.ID, *layout); !errors.Is(err, ErrInvalidDashboardLayout) {
		t.Errorf("invalid chart error = %v, want ErrInvalidDashboardLayout", err)
	}
}

func TestSaveDashboardLayoutValidatesVisualizationFields(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()
	createDashboardTestCategory(t, st, user.ID, budget.ID)

	layout, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get layout: %v", err)
	}

	tests := []struct {
		name   string
		widget DashboardWidget
	}{
		{
			name: "chart missing visualization",
			widget: DashboardWidget{
				ID:      "category-breakdown:11111111-1111-4111-8111-111111111111",
				Type:    DashboardWidgetTypeCategoryBreakdown,
				Y:       10,
				Columns: 18,
				Rows:    12,
			},
		},
		{
			name: "chart unknown visualization",
			widget: DashboardWidget{
				ID:            "category-breakdown:11111111-1111-4111-8111-111111111111",
				Type:          DashboardWidgetTypeCategoryBreakdown,
				Visualization: "radar",
				Y:             10,
				Columns:       18,
				Rows:          12,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := *layout
			candidate.Widgets = append(append([]DashboardWidget{}, layout.Widgets...), test.widget)
			if _, err := st.SaveDashboardLayout(ctx, user.ID, budget.ID, candidate); !errors.Is(err, ErrInvalidDashboardLayout) {
				t.Errorf("save error = %v, want ErrInvalidDashboardLayout", err)
			}
		})
	}

	candidate := *layout
	candidate.Widgets = append([]DashboardWidget{}, layout.Widgets...)
	candidate.Widgets[0].Visualization = DashboardVisualizationPie
	if _, err := st.SaveDashboardLayout(ctx, user.ID, budget.ID, candidate); !errors.Is(err, ErrInvalidDashboardLayout) {
		t.Errorf("category visualization error = %v, want ErrInvalidDashboardLayout", err)
	}
}

func TestSaveDashboardLayoutValidatesExpenseDisplayFields(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()
	createDashboardTestCategory(t, st, user.ID, budget.ID)

	layout, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get layout: %v", err)
	}

	for _, expenseDisplay := range []string{
		DashboardExpenseDisplayNumberOnly,
		DashboardExpenseDisplayAmountOnly,
		DashboardExpenseDisplayTitleAmount,
		DashboardExpenseDisplayTitleAmountDate,
	} {
		candidate := *layout
		candidate.Widgets = append([]DashboardWidget{}, layout.Widgets...)
		candidate.Widgets[0].ExpenseDisplay = expenseDisplay
		if _, err := st.SaveDashboardLayout(ctx, user.ID, budget.ID, candidate); err != nil {
			t.Errorf("save %q expense display: %v", expenseDisplay, err)
		}
	}

	for _, expenseDisplay := range []string{"", "description-only"} {
		candidate := *layout
		candidate.Widgets = append([]DashboardWidget{}, layout.Widgets...)
		candidate.Widgets[0].ExpenseDisplay = expenseDisplay
		if _, err := st.SaveDashboardLayout(ctx, user.ID, budget.ID, candidate); !errors.Is(err, ErrInvalidDashboardLayout) {
			t.Errorf("save %q expense display error = %v, want ErrInvalidDashboardLayout", expenseDisplay, err)
		}
	}

	candidate := *layout
	candidate.Widgets = append([]DashboardWidget{}, layout.Widgets...)
	candidate.Widgets = append(candidate.Widgets, DashboardWidget{
		ID:             "category-breakdown:11111111-1111-4111-8111-111111111111",
		Type:           DashboardWidgetTypeCategoryBreakdown,
		ExpenseDisplay: DashboardExpenseDisplayAmountOnly,
		Visualization:  DashboardVisualizationPie,
		Y:              10,
		Columns:        18,
		Rows:           12,
	})
	if _, err := st.SaveDashboardLayout(ctx, user.ID, budget.ID, candidate); !errors.Is(err, ErrInvalidDashboardLayout) {
		t.Errorf("chart expense display error = %v, want ErrInvalidDashboardLayout", err)
	}
}

func hasDashboardWidgetForCategory(widgets []DashboardWidget, categoryID int64) bool {
	for _, widget := range widgets {
		if widget.CategoryID == categoryID {
			return true
		}
	}

	return false
}

func createDashboardTestCategory(t *testing.T, st *Store, userID, budgetID int64) *Category {
	t.Helper()

	category, err := st.CreateCategory(context.Background(), userID, budgetID, "Food")
	if err != nil {
		t.Fatalf("create dashboard test category: %v", err)
	}

	return category
}
