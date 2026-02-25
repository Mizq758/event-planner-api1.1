CREATE TABLE IF NOT EXISTS public.resume (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    name        TEXT NOT NULL,
    city        TEXT NOT NULL,
    job_title   TEXT NOT NULL,
    email       TEXT NOT NULL,
    phone       TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
); 

CREATE OR REPLACE VIEW public.notes AS
SELECT
    id,
    user_id,
    name,
    city,
    job_title,
    email,
    phone,
    created_at
FROM public.resume;

CREATE INDEX IF NOT EXISTS idx_resume_user_id ON public.resume(user_id);
