ALTER TABLE loan_agreement
ADD COLUMN cumulative_disbursement NUMERIC(20, 2) NOT NULL DEFAULT 0;
