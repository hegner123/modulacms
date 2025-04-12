CREATE OR REPLACE FUNCTION prevent_role_modification()
RETURNS trigger AS $$
BEGIN
    IF OLD.id = 1 THEN
        RAISE EXCEPTION 'Modification of constant row is not allowed';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_update
BEFORE UPDATE ON roles
FOR EACH ROW
EXECUTE FUNCTION prevent_role_modification();

CREATE TRIGGER prevent_delete
BEFORE DELETE ON roles
FOR EACH ROW
EXECUTE FUNCTION prevent_role_modification();


CREATE OR REPLACE FUNCTION prevent_user_modification()
RETURNS trigger AS $$
BEGIN
    IF OLD.id = 1 THEN
        RAISE EXCEPTION 'Modification of system_admin row is not allowed';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_update
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION prevent_user_modification();

CREATE TRIGGER prevent_delete
BEFORE DELETE ON users
FOR EACH ROW
EXECUTE FUNCTION prevent_user_modification();

CREATE OR REPLACE FUNCTION prevent_permission_modification()
RETURNS trigger AS $$
BEGIN
    IF OLD.id = 1 THEN
        RAISE EXCEPTION 'Modification of basic_permission row is not allowed';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_update
BEFORE UPDATE ON permissions
FOR EACH ROW
EXECUTE FUNCTION prevent_permission_modification();

CREATE TRIGGER prevent_delete
BEFORE DELETE ON permissions
FOR EACH ROW
EXECUTE FUNCTION prevent_permission_modification();
