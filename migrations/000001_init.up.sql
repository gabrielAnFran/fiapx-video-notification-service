CREATE TABLE notification_log (
  id UUID PRIMARY KEY,
  video_id UUID NOT NULL,
  user_id UUID NOT NULL,
  recipient TEXT NOT NULL,
  channel TEXT NOT NULL DEFAULT 'email',
  type TEXT NOT NULL CHECK (type IN ('COMPLETED','FAILED')),
  status TEXT NOT NULL CHECK (status IN ('SENT','FAILED')),
  error_message TEXT,
  sent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE processed_events (
  event_id UUID PRIMARY KEY,
  processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
