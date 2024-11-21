# syntax=docker/dockerfile:1
FROM scratch

# Arguments (can be passed during build)
# Database Driver
ARG database="sqlite3"
# Database Name
ARG databaseName="modula.db"
# Database User
ARG databaseUserName
# Database Password
ARG databasePassword
# Database ConnectionString
ARG databaseConnectionString
# Bucket Storage
ARG bucket="local"
# Bucket Access Key
ARG bucketAccess=""
# Server port
ARG port=3055

# Metadata
LABEL project="ModulaCMS"

# Environment variables
ENV env="production"

# Set working directory
WORKDIR /

# Copy the executable binary (ensure `modula` exists and is executable)
COPY modulacms /


# Run the executable
CMD ["./modulacms"]

# Expose port
EXPOSE ${port}

# Stop signal
STOPSIGNAL SIGTERM

