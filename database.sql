--

-- PostgreSQL database dump

--

SET statement_timeout = 0;

SET lock_timeout = 0;

SET client_encoding = 'UTF8';

SET standard_conforming_strings = on;

SET check_function_bodies = false;

SET client_min_messages = warning;

--

-- Name: golang_gin_db; Type: DATABASE; Schema: -; Owner: postgres

--

--DROP DATABASE golang_gin_db;

CREATE DATABASE golang_gin_db
WITH
    TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'en_US.UTF-8' LC_CTYPE = 'en_US.UTF-8';

ALTER DATABASE golang_gin_db OWNER TO postgres;

\connect golang_gin_db;

SET statement_timeout = 0;

SET lock_timeout = 0;

SET client_encoding = 'UTF8';

SET standard_conforming_strings = on;

SET check_function_bodies = false;

SET client_min_messages = warning;

--

-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner:

--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;

--

-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner:

--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';

CREATE FUNCTION CREATED_AT_COLUMN() RETURNS TRIGGER 
LANGUAGE PLPGSQL AS 
	$$ BEGIN NEW.created_at = EXTRACT(EPOCH FROM NOW());
	RETURN NEW;
END; 

$$;

ALTER FUNCTION public.created_at_column() OWNER TO postgres;

--

-- TOC entry 190 (class 1255 OID 36646)

-- Name: update_at_column(); Type: FUNCTION; Schema: public; Owner: postgres

--

CREATE FUNCTION UPDATE_AT_COLUMN() RETURNS TRIGGER 
LANGUAGE PLPGSQL AS 
	$$ BEGIN NEW.updated_at = EXTRACT(EPOCH FROM NOW());
	RETURN NEW;
END; 

$$;

ALTER FUNCTION public.update_at_column() OWNER TO postgres;

SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;