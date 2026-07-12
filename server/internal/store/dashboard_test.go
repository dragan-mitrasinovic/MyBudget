package store

import (
	"context"
	"encoding/json"
	"errors"
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

	if len(layout.Widgets) != 3 {
		t.Fatalf("widget count = %d, want 3", len(layout.Widgets))
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

	if len(saved.Widgets) != 2 {
		t.Fatalf("saved widget count = %d, want 2", len(saved.Widgets))
	}

	utilities, err := st.CreateCategory(ctx, user.ID, budget.ID, "Utilities")
	if err != nil {
		t.Fatalf("create Utilities category: %v", err)
	}

	if err := st.DeleteCategory(ctx, user.ID, food.ID); err != nil {
		t.Fatalf("delete Food category: %v", err)
	}

	reconciled, err := st.GetDashboardLayout(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("get reconciled layout: %v", err)
	}

	if len(reconciled.Widgets) != 2 {
		t.Fatalf("reconciled widget count = %d, want 2", len(reconciled.Widgets))
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

func TestGetDashboardLayoutUpgradesVersionOneGeometry(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "dashboard-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	categories, err := st.ListCategories(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	legacyLayout := DashboardLayout{
		Version: dashboardOriginalLayoutVersion,
		Widgets: []DashboardWidget{
			{
				ID:         dashboardCategoryWidgetID(categories[0].ID),
				Type:       DashboardWidgetTypeCategoryExpenses,
				CategoryID: categories[0].ID,
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

	categories, err := st.ListCategories(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	fineLayout := DashboardLayout{
		Version: dashboardFineLayoutVersion,
		Widgets: []DashboardWidget{
			{
				ID:         dashboardCategoryWidgetID(categories[0].ID),
				Type:       DashboardWidgetTypeCategoryExpenses,
				CategoryID: categories[0].ID,
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

func TestSaveDashboardLayoutValidatesOwnershipAndGeometry(t *testing.T) {
	st := openTestStore(t)
	owner := createStoreTestUser(t, st, "dashboard-owner")
	other := createStoreTestUser(t, st, "other-user")
	budget := createStoreTestBudget(t, st, owner.ID, "Household")
	ctx := context.Background()

	if _, err := st.GetDashboardLayout(ctx, other.ID, budget.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("other user's get error = %v, want ErrNotFound", err)
	}

	layout, err := st.GetDashboardLayout(ctx, owner.ID, budget.ID)
	if err != nil {
		t.Fatalf("get owner layout: %v", err)
	}

	layout.Widgets[0].Columns = dashboardGridColumns + 1
	if _, err := st.SaveDashboardLayout(ctx, owner.ID, budget.ID, *layout); !errors.Is(err, ErrInvalidDashboardLayout) {
		t.Errorf("invalid geometry error = %v, want ErrInvalidDashboardLayout", err)
	}

	if _, err := st.SaveDashboardLayout(ctx, other.ID, budget.ID, *layout); !errors.Is(err, ErrNotFound) {
		t.Errorf("other user's save error = %v, want ErrNotFound", err)
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
