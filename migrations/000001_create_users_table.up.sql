CREATE TABLE IF NOT EXISTS public.users (
    user_id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    "name" varchar NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT NOW()
);