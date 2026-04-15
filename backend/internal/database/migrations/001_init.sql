CREATE TABLE source_strings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE translations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id       UUID NOT NULL REFERENCES source_strings(id) ON DELETE CASCADE,
    locale          VARCHAR(10) NOT NULL,
    translated_text TEXT NOT NULL,
    llm_provider    VARCHAR(20) NOT NULL,
    translated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source_id, locale, llm_provider)
);

CREATE INDEX idx_translations_source_id ON translations(source_id);
CREATE INDEX idx_translations_locale ON translations(locale);
