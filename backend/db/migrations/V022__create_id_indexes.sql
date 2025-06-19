DO $$
DECLARE
    tbl text;
BEGIN
FOR tbl IN
    SELECT table_name FROM information_schema.tables
    WHERE table_schema = 'public'
      AND table_type   = 'BASE TABLE'
      AND EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = tbl AND column_name = 'id')
LOOP
    EXECUTE format('CREATE INDEX IF NOT EXISTS %I_id_hash_idx
                    ON %I USING hash (left(id,8));', tbl, tbl);
END LOOP;
END$$;