CREATE OR REPLACE FUNCTION audit_trigger_by_column_fn()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_user_id UUID;
    v_old_data JSONB;
    v_new_data JSONB;
    v_record_data JSONB;
    v_record_id UUID;
BEGIN
    BEGIN
        v_user_id := current_setting('app.current_user_id', true)::UUID;
    EXCEPTION WHEN OTHERS THEN
        v_user_id := NULL;
    END;

    v_old_data := CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE to_jsonb(OLD) END;
    v_new_data := CASE WHEN TG_OP = 'DELETE' THEN NULL ELSE to_jsonb(NEW) END;
    v_record_data := COALESCE(v_new_data, v_old_data);
    v_record_id := (v_record_data ->> TG_ARGV[0])::UUID;

    INSERT INTO audit_log (table_name, record_id, action, old_data, new_data, changed_by)
    VALUES (TG_TABLE_NAME, v_record_id, TG_OP, v_old_data, v_new_data, v_user_id);

    RETURN COALESCE(NEW, OLD);
END;
$$;

DROP TRIGGER IF EXISTS trg_audit_bb_project_institution ON bb_project_institution;
DROP TRIGGER IF EXISTS trg_audit_bb_project_bappenas_partner ON bb_project_bappenas_partner;
DROP TRIGGER IF EXISTS trg_audit_bb_project_location ON bb_project_location;
DROP TRIGGER IF EXISTS trg_audit_bb_project_national_priority ON bb_project_national_priority;
DROP TRIGGER IF EXISTS trg_audit_gb_project_bb_project ON gb_project_bb_project;
DROP TRIGGER IF EXISTS trg_audit_gb_project_bappenas_partner ON gb_project_bappenas_partner;
DROP TRIGGER IF EXISTS trg_audit_gb_project_institution ON gb_project_institution;
DROP TRIGGER IF EXISTS trg_audit_gb_project_location ON gb_project_location;
DROP TRIGGER IF EXISTS trg_audit_dk_project_gb_project ON dk_project_gb_project;
DROP TRIGGER IF EXISTS trg_audit_dk_project_bappenas_partner ON dk_project_bappenas_partner;
DROP TRIGGER IF EXISTS trg_audit_dk_project_location ON dk_project_location;

CREATE TRIGGER trg_audit_bb_project_institution
    AFTER INSERT OR UPDATE OR DELETE ON bb_project_institution
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_by_column_fn('bb_project_id');

CREATE TRIGGER trg_audit_bb_project_bappenas_partner
    AFTER INSERT OR UPDATE OR DELETE ON bb_project_bappenas_partner
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_by_column_fn('bb_project_id');

CREATE TRIGGER trg_audit_bb_project_location
    AFTER INSERT OR UPDATE OR DELETE ON bb_project_location
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_by_column_fn('bb_project_id');

CREATE TRIGGER trg_audit_bb_project_national_priority
    AFTER INSERT OR UPDATE OR DELETE ON bb_project_national_priority
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_by_column_fn('bb_project_id');

CREATE TRIGGER trg_audit_gb_project_bb_project
    AFTER INSERT OR UPDATE OR DELETE ON gb_project_bb_project
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_by_column_fn('gb_project_id');

CREATE TRIGGER trg_audit_gb_project_bappenas_partner
    AFTER INSERT OR UPDATE OR DELETE ON gb_project_bappenas_partner
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_by_column_fn('gb_project_id');

CREATE TRIGGER trg_audit_gb_project_institution
    AFTER INSERT OR UPDATE OR DELETE ON gb_project_institution
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_by_column_fn('gb_project_id');

CREATE TRIGGER trg_audit_gb_project_location
    AFTER INSERT OR UPDATE OR DELETE ON gb_project_location
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_by_column_fn('gb_project_id');

CREATE TRIGGER trg_audit_dk_project_gb_project
    AFTER INSERT OR UPDATE OR DELETE ON dk_project_gb_project
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_by_column_fn('dk_project_id');

CREATE TRIGGER trg_audit_dk_project_bappenas_partner
    AFTER INSERT OR UPDATE OR DELETE ON dk_project_bappenas_partner
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_by_column_fn('dk_project_id');

CREATE TRIGGER trg_audit_dk_project_location
    AFTER INSERT OR UPDATE OR DELETE ON dk_project_location
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_by_column_fn('dk_project_id');
