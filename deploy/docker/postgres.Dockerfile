# Use the official PostgreSQL image as the base image
FROM postgres:13

# Set environment variables (these can be overridden at runtime)
ENV POSTGRES_USER=modula_u
ENV POSTGRES_PASSWORD=modula_pass
ENV POSTGRES_DB=modula_db

# Copy initialization script(s) to be executed during container startup.
# Any .sql or .sh files in this directory will be automatically executed.
#COPY ./init.sql /docker-entrypoint-initdb.d/

# Expose PostgreSQL port (optional, useful for documentation)
EXPOSE 5432

