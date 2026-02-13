# Use the official MySQL image as the base image
FROM mysql:8.0

# Set environment variables (these can be overridden at runtime)
ENV MYSQL_ROOT_PASSWORD=root_root
ENV MYSQL_DATABASE=modula_db
ENV MYSQL_USER=modula_u
ENV MYSQL_PASSWORD=modula_pass

# Copy initialization script(s) into the proper directory.
# Any .sql or .sh files in this directory will be executed during container initialization.
#COPY ./init.sql /docker-entrypoint-initdb.d/

# Expose MySQL's default port (optional, for documentation)
EXPOSE 3306
