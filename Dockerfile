# syntax=docker/dockerfile:1
FROM scratch

# Arguments (can be passed during build)
ARG database="sqlite3"
ARG databaseName="modula.db"
ARG bucket="local"
ARG bucketAccess=""
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

