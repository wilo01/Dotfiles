# TDS Suite Database Connection Guide

## Overview

The TDS Suite uses Oracle Database running in Docker container 'trunk' for development and testing.

## Connection Details

### Container Information
- **Container Name**: `trunk`
- **Database**: `xepdb1`
- **Port**: Configured via docker-env.sh (default: 6021)

### User Credentials

#### Primary Application User
- **User**: `coreaccess`
- **Password**: `xy*0m9`
- **Purpose**: Application database operations, testing queries, development work

#### System Administrator
- **User**: `sys`
- **Password**: `TDSSuite1`
- **Role**: `sysdba`
- **Purpose**: Administrative tasks, schema management, system-level operations

## Quick Start Commands

### Basic Connection Test
```bash
# Test application user connection
docker exec -i trunk bash -c "sqlplus -s 'coreaccess/xy*0m9@xepdb1'" <<EOF
SELECT 'Connection successful' FROM dual;
EXIT;
EOF

# Test system admin connection
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT 'Admin connection successful' FROM dual;
EXIT;
EOF
```

### Common Development Tasks

#### 1. Check Application Tables
```bash
docker exec -i trunk bash -c "sqlplus -s 'coreaccess/xy*0m9@xepdb1'" <<EOF
SELECT table_name FROM user_tables WHERE table_name LIKE 'CA_%' ORDER BY table_name;
EXIT;
EOF
```

#### 2. Inspect Table Structure
```bash
docker exec -i trunk bash -c "sqlplus -s 'coreaccess/xy*0m9@xepdb1'" <<EOF
DESC ca_nda_mapping;
EXIT;
EOF
```

#### 3. Query NDA Location Data
```bash
docker exec -i trunk bash -c "sqlplus -s 'coreaccess/xy*0m9@xepdb1'" <<EOF
SELECT DISTINCT
    CASE
        WHEN cnm.cnpt_code IS NOT NULL AND ccp.collection_point_description IS NOT NULL
        THEN ccp.collection_point_description
        WHEN cnm.zone_code IS NOT NULL AND cpv.zone_description IS NOT NULL
        THEN cpv.zone_description
        ELSE 'Unknown Location'
    END as location_name
FROM ca_nda_mapping cnm
LEFT JOIN ca_collection_point ccp ON ccp.collection_point_code = cnm.cnpt_code
LEFT JOIN safe_cp_location_view cpv ON cpv.collection_point_code = cnm.cnpt_code
WHERE cnm.nda_code = 'test_nda_code'
AND ROWNUM <= 10;
EXIT;
EOF
```

#### 4. Check Table Row Counts
```bash
docker exec -i trunk bash -c "sqlplus -s 'coreaccess/xy*0m9@xepdb1'" <<EOF
SELECT 'CA_NDA_MAPPING' as table_name, COUNT(*) as row_count FROM ca_nda_mapping
UNION ALL
SELECT 'CA_COLLECTION_POINT', COUNT(*) FROM ca_collection_point
UNION ALL
SELECT 'SAFE_CP_LOCATION_VIEW', COUNT(*) FROM safe_cp_location_view;
EXIT;
EOF
```

## Environment Setup

### Docker Container Management
```bash
# Check if trunk container is running
docker ps | grep trunk

# Start trunk container if stopped
docker start trunk

# Check container logs
docker logs trunk --tail 50
```

### Database Scripts Location
Database setup and migration scripts are located in:
- `source/server/database/sql/dev-scripts/docker-env.sh`
- `source/server/database/liquibase.properties`
- `source/server/database/sql/safe/`

## Integration with Development

### For API Development
When developing or testing REST endpoints (like `/getNdaLocations`), use the `coreaccess` user to:
1. Validate query logic
2. Test data retrieval
3. Verify permissions and access

### For Schema Changes
When making database schema modifications:
1. Use `sys` user for administrative operations
2. Apply changes through Liquibase migrations
3. Test with `coreaccess` user for application compatibility

## Best Practices

1. **Use Application User First**: Always try `coreaccess` user for application-related queries
2. **Silent Mode**: Use `-s` flag for cleaner output: `sqlplus -s`
3. **Limit Results**: Use `ROWNUM <= N` for large result sets
4. **Proper Exit**: Always end SQL sessions with `EXIT;`
5. **Error Handling**: Check command exit codes for scripting

## Troubleshooting

### Connection Issues
```bash
# Check container status
docker ps -a | grep trunk

# Verify database is ready
docker exec trunk sqlplus -v

# Check database logs for errors
docker logs trunk | grep -i error
```

### Permission Problems
```bash
# List user privileges
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT * FROM dba_sys_privs WHERE grantee = 'COREACCESS';
EXIT;
EOF
```

### Performance Issues
```bash
# Check active sessions
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT username, status, COUNT(*)
FROM v\$session
WHERE username IS NOT NULL
GROUP BY username, status;
EXIT;
EOF
```

## Security Notes

- The `coreaccess` user has application-level permissions
- The `sys` user has full database administrator privileges
- Passwords are stored in `docker-env.sh` for development environment only
- Production environments should use secure credential management

## Related Documentation

- [Docker Environment Script](../source/server/database/sql/dev-scripts/docker-env.sh)
- [Liquibase Configuration](../source/server/database/liquibase.properties)
- [Copilot Instructions](./instructions/copilot-instructions.md)
- [TDS Suite Guide](./TDS_SUITE_GUIDE.md)
