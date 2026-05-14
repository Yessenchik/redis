CREATE TABLE IF NOT EXISTS payments (
                                        id TEXT PRIMARY KEY,
                                        order_id TEXT NOT NULL,
                                        amount NUMERIC NOT NULL,
                                        status TEXT NOT NULL,
                                        created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );