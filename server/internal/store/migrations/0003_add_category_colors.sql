ALTER TABLE categories
ADD COLUMN color_key TEXT NOT NULL DEFAULT 'chart-1';

ALTER TABLE categories
ADD COLUMN color_hex TEXT;

UPDATE categories AS category
SET color_key = printf(
    'chart-%d',
    ((
        SELECT COUNT(*)
        FROM categories AS earlier
        WHERE earlier.budget_id = category.budget_id
          AND (
              earlier.name COLLATE NOCASE < category.name COLLATE NOCASE
              OR (
                  earlier.name COLLATE NOCASE = category.name COLLATE NOCASE
                  AND earlier.id <= category.id
              )
          )
    ) - 1) % 12 + 1
);