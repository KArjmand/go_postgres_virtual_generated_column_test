-- Table WITH virtual generated column (PostgreSQL 12+)
-- total_cents is computed by the database automatically
CREATE TABLE IF NOT EXISTS invoices_with_virtual (
    id           BIGSERIAL PRIMARY KEY,
    customer_id  BIGINT NOT NULL,
    amount_cents BIGINT NOT NULL,
    tax_rate     NUMERIC(4,2) NOT NULL,
    total_cents  BIGINT GENERATED ALWAYS AS (
        ROUND(amount_cents * (1 + tax_rate))
    ) STORED
);

-- Table WITHOUT virtual generated column
-- total_cents must be calculated in the application
CREATE TABLE IF NOT EXISTS invoices_without_virtual (
    id           BIGSERIAL PRIMARY KEY,
    customer_id  BIGINT NOT NULL,
    amount_cents BIGINT NOT NULL,
    tax_rate     NUMERIC(4,2) NOT NULL
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_invoices_with_virtual_customer ON invoices_with_virtual(customer_id);
CREATE INDEX IF NOT EXISTS idx_invoices_without_virtual_customer ON invoices_without_virtual(customer_id);
