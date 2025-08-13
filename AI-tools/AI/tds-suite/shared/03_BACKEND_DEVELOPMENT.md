# Backend Development Guidelines

## Oracle Database Integration

### Database Connection Details

The TDS Suite uses Oracle Database running in Docker container 'trunk'.

**Default Connection Details:**
- **Container Name**: `trunk`
- **Database User**: `some_user`
- **Password**: `some_pass`
- **Database Service**: `serv`
- **Connection Role**: Normal user

### Standard Docker Database Query Pattern

```bash
# Primary user access
docker exec -i --env-file .env trunk bash -c "sqlplus -s '\$DB_USER/\$DB_PASS@\$DB_SERVICE'" <<EOF
[SQL COMMANDS]
EXIT;
EOF

# System administrator access
docker exec -i --env-file .env trunk bash -c "sqlplus -s '\$DB_USER/\$DB_PASS@\$DB_SERVICE'" <<EOF
[SQL COMMANDS]
EXIT;
EOF
```

### Common Database Operations

#### 1. Check Table Existence
```bash
docker exec -i --env-file .env trunk bash -c "sqlplus -s '\$DB_USER/\$DB_PASS@\$DB_SERVICE'" <<EOF
SELECT owner, table_name FROM all_tables WHERE table_name = 'CA_NDA_VERSION';
EXIT;
EOF
```

#### 2. Describe Table Structure
```bash
docker exec -i --env-file .env trunk bash -c "sqlplus -s '\$DB_USER/\$DB_PASS@\$DB_SERVICE'" <<EOF
DESC ca_nda_mapping;
EXIT;
EOF
```

#### 3. Query Table Data with Row Limit
```bash
docker exec -i --env-file .env trunk bash -c "sqlplus -s '\$DB_USER/\$DB_PASS@\$DB_SERVICE'" <<EOF
SELECT * FROM ca_nda_mapping WHERE ROWNUM <= 5;
EXIT;
EOF
```

#### 4. Complex Query with Filters
```bash
docker exec -i --env-file .env trunk bash -c "sqlplus -s '\$DB_USER/\$DB_PASS@\$DB_SERVICE'" <<EOF
SELECT DISTINCT location_name FROM (
    SELECT CASE
        WHEN cnm.cnpt_code IS NOT NULL AND ccp.collection_point_description IS NOT NULL
        THEN ccp.collection_point_description
        WHEN cnm.zone_code IS NOT NULL AND cpv.zone_description IS NOT NULL
        THEN cpv.zone_description
        ELSE NULL
    END as location_name
    FROM ca_nda_mapping cnm
    LEFT JOIN ca_collection_point ccp ON ccp.collection_point_code = cnm.cnpt_code
    LEFT JOIN safe_cp_location_view cpv ON cpv.collection_point_code = cnm.cnpt_code
    WHERE cnm.nda_code = 'your_nda_code'
) WHERE location_name IS NOT NULL;
EXIT;
EOF
```

#### 5. Test Query Performance
```bash
docker exec -i --env-file .env trunk bash -c "sqlplus -s '\$DB_USER/\$DB_PASS@\$DB_SERVICE'" <<EOF
SET TIMING ON
SELECT COUNT(*) FROM ca_nda_mapping;
EXIT;
EOF
```

### Database Development Best Practices

#### 1. Always Use Silent Mode
- Use `-s` flag in sqlplus for clean output without SQL*Plus banners
- Reduces noise in scripts and automated processes

#### 2. Proper Session Management
- Always end SQL sessions with `EXIT;`
- Use heredoc (`<<EOF ... EOF`) for multi-line SQL blocks
- Container automatically cleans up sessions

#### 3. Safe Query Practices
- Use `ROWNUM <= N` for limiting large result sets
- Test queries with small datasets first
- Always verify table existence before complex operations

## REST API Architecture

### Resource Templates (.apex files)

#### Location and Purpose
- **Location**: `source/server/rt/` directory
- **Purpose**: Define API endpoints and business logic
- **Format**: Custom .apex file format containing SQL and PL/SQL procedures

#### Example .apex File Structure
```xml
<?xml version="1.0" encoding="UTF-8"?>
<template xmlns="http://xmlns.oracle.com/apex/resource-template" pattern="getNdaTranslationLanguages?nda_code={nda_code}">
<etag/>
<handler>
<content><![CDATA[
select distinct
    cnv.language_code,
    skl.language_description,
    cnv.version_id,
    cnv.end_date,
    case when cnv.end_date is null then 'Y' else 'N' end as is_active
from ca_nda_version cnv
left join safe_kiosk_languages_view skl on cnv.language_code = skl.language_code
where cnv.nda_code = trim(:nda_code)
  and cnv.language_code is not null
  and skl.language_description is not null
  and safe_request_permissions_pak.check_permission_sql (:Uidtoken, '(ndaDashboard)', 'getNdaTranslationLanguages', 'GET') = 'Y'
order by
    case when cnv.end_date is null then 0 else 1 end, -- Active versions first (no end_date)
    skl.language_description asc
]]></content>
<parameter name="Uidtoken" source="header" aliasing="Uidtoken"/>
</handler>
</template>
```

### File Upload API

#### Location and Technology
- **Location**: `source/server/rest-api-fileupload/`
- **Technology**: Java-based REST service
- **Security**: Encrypted password authentication

### API Development Patterns

#### RESTful Design Principles
- Consistent API endpoints following REST principles
- CSRF token implementation for security
- JSON responses with proper error handling

#### Data Fetching
- Implement SQL and PL/SQL in the same format as .apex files
- Use parameterized queries for all database operations

## Oracle APEX Development

### Database Layer Integration

#### Connection Management
- Database connections configured via XML configuration files
- PL/SQL procedures for business logic
- Liquibase migrations for version-controlled database schema changes

```xml
<!-- Database Configuration Example -->
<entry key="apex.db.hostname">hostname</entry>
<entry key="apex.db.port">1521</entry>
<entry key="apex.db.sid">database_sid</entry>
```

#### Database Migration Process
- **Liquibase Scripts**: Located in `source/server/database/`
- **Changelog Management**: Structured changelog system for tracking database changes
- **Environment-specific Configurations**: Separate configs for dev, test, and production

### APEX Resource Templates

#### Best Practices
1. Create .apex files for API routes in the correct folder
2. Implement proper SQL and PL/SQL procedures
3. Use parameterized queries to prevent SQL injection
4. Include proper error handling and validation

## Security Implementation

### Authentication & Authorization
- **Token-based Security**: CSRF token implementation
- **User Session Management**: Server-side session handling
- **Role-based Access Control**: User permissions and restrictions

```javascript
// Security Headers Example
Ext.Ajax.defaultHeaders = {
   'Uidtoken': app.user.id,
   'X-Tds-Sec': app.user.antiCSRFToken
};
```

### Database Security
- Use stored procedures for complex business logic
- Implement proper user permissions and roles
- Always use parameterized queries
- Validate all input data

## Development Best Practices

### Code Organization
1. **RESTful API Design**: Consistent API endpoints following REST principles
2. **Database Best Practices**: Use stored procedures for complex business logic
3. **Separation of Concerns**: Clear distinction between UI, business logic, and data layers

### Performance Optimization
- Optimize database queries for performance
- Use proper indexing strategies
- Implement caching where appropriate
- Monitor query execution times

### Error Handling
- Implement comprehensive error handling at all layers
- Provide meaningful error messages
- Log errors appropriately for debugging
- Handle database connection failures gracefully

## Troubleshooting

### Database Connection Issues
- Validate Oracle connection strings
- Check user permissions and roles
- Verify Liquibase migration status
- Ensure Docker container is running

### API Issues
- Check .apex file syntax and structure
- Verify permissions and security tokens
- Test SQL queries independently
- Validate input parameters
