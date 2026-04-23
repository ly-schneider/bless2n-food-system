CREATE TYPE volunteer_campaign_status AS ENUM ('draft', 'active', 'ended');

CREATE TABLE volunteer_campaign (
    id             UUID PRIMARY KEY DEFAULT uuidv7(),
    claim_token    UUID NOT NULL DEFAULT uuidv7(),
    name           VARCHAR(100) NOT NULL,
    access_code    VARCHAR(4) NOT NULL,
    valid_from     TIMESTAMPTZ NULL,
    valid_until    TIMESTAMPTZ NULL,
    status         volunteer_campaign_status NOT NULL DEFAULT 'active',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_volunteer_campaign_claim_token ON volunteer_campaign (claim_token);
CREATE INDEX idx_volunteer_campaign_status ON volunteer_campaign (status);

CREATE TABLE volunteer_campaign_product (
    campaign_id  UUID NOT NULL REFERENCES volunteer_campaign (id) ON DELETE CASCADE,
    product_id   UUID NOT NULL REFERENCES product (id) ON DELETE RESTRICT,
    quantity     INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (campaign_id, product_id)
);

CREATE TABLE volunteer_slot (
    id                   UUID PRIMARY KEY DEFAULT uuidv7(),
    campaign_id          UUID NOT NULL REFERENCES volunteer_campaign (id) ON DELETE CASCADE,
    order_id             UUID NOT NULL UNIQUE REFERENCES "order" (id) ON DELETE RESTRICT,
    reserved_by_session  VARCHAR(64) NULL,
    reserved_at          TIMESTAMPTZ NULL,
    reserved_until       TIMESTAMPTZ NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_volunteer_slot_campaign_id_reserved_until ON volunteer_slot (campaign_id, reserved_until);
CREATE INDEX idx_volunteer_slot_reserved_by_session ON volunteer_slot (reserved_by_session);
