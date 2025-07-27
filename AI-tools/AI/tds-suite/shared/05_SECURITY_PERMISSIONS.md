# Security & Permissions Guidelines

## APEX File Permissions

### Permission System Overview
- Always get permissions for UI backoffice app for .apex files
- Use `safe_request_permissions_pak.check_permission_sql` function to check permissions
- Portlet name is crucial for permission checks
- Permissions apply specifically to files in:
  - `@source/ui/`
  - `@source/server/rt/`

### Permission Check Implementation

#### Standard Permission Check Pattern
```sql
safe_request_permissions_pak.check_permission_sql (:Uidtoken, '(ndaDashboard)', 'getNdaContent', 'GET') = 'Y'
```

#### Parameters Explanation
- `:Uidtoken` - User ID token for authentication
- `'(ndaDashboard)'` - Portlet name (critical for proper permission mapping)
- `'getNdaContent'` - Resource/function name
- `'GET'` - HTTP method being accessed

#### Permission Verification Process
1. Always verify the specific portlet name when extending permissions
2. Test permissions with different user roles
3. Validate permission inheritance and delegation
4. Check for proper error handling when permissions are denied

### Example Permission Implementation in .apex Files

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
    case when cnv.end_date is null then 0 else 1 end,
    skl.language_description asc
]]></content>
<parameter name="Uidtoken" source="header" aliasing="Uidtoken"/>
</handler>
</template>
```

## Authentication & Authorization

### Token-Based Security

#### CSRF Token Implementation
```javascript
// Security Headers Configuration
Ext.Ajax.defaultHeaders = {
   'Uidtoken': app.user.id,
   'X-Tds-Sec': app.user.antiCSRFToken
};
```

#### User Session Management
- Server-side session handling
- Token validation on each request
- Session timeout management
- Secure token generation and storage

### Role-Based Access Control

#### User Permissions System
- Role-based access control for user permissions
- Hierarchical permission structure
- Permission inheritance
- Dynamic permission evaluation

#### Access Control Implementation
```javascript
// Permission check before UI action
if (app.user.hasPermission('nda.manage')) {
    // Allow action
} else {
    // Deny action and show appropriate message
}
```

## Input Validation & Security

### SQL Injection Prevention

#### Parameterized Queries
- Always use parameterized queries for database operations
- Never concatenate user input directly into SQL strings
- Use proper SQL parameter binding

```sql
-- ✅ CORRECT: Using parameterized query
SELECT * FROM users WHERE user_id = :user_id

-- ❌ INCORRECT: String concatenation (vulnerable to SQL injection)
SELECT * FROM users WHERE user_id = '" + userInput + "'
```

#### Input Sanitization
- Validate all input data at multiple layers
- Sanitize user input before processing
- Use whitelist validation where possible
- Implement proper encoding for output

### XSS Prevention

#### Output Encoding
- Encode all user-generated content before display
- Use proper HTML encoding functions
- Validate and sanitize rich text content
- Implement Content Security Policy (CSP)

#### Client-Side Security
```javascript
// Proper way to handle user input in ExtJS
{
    xtype: 'displayfield',
    value: Ext.String.htmlEncode(userInput) // Always encode user input
}
```

## Database Security

### Connection Security

#### Secure Database Connections
- Use encrypted connections where possible
- Implement proper connection pooling
- Limit database user privileges
- Regular credential rotation

#### Database User Management
```sql
-- Example of proper user privilege management
GRANT SELECT, INSERT, UPDATE ON ca_nda_mapping TO app_user;
REVOKE DELETE ON ca_nda_mapping FROM app_user;
```

### Stored Procedure Security

#### Secure PL/SQL Development
- Use DEFINER'S RIGHTS for security procedures
- Implement proper exception handling
- Validate all input parameters
- Log security-relevant operations

```sql
CREATE OR REPLACE PROCEDURE secure_procedure(
    p_user_id IN NUMBER,
    p_action IN VARCHAR2
)
AUTHID DEFINER
AS
BEGIN
    -- Validate input parameters
    IF p_user_id IS NULL OR p_action IS NULL THEN
        RAISE_APPLICATION_ERROR(-20001, 'Invalid parameters');
    END IF;

    -- Implement security logic
    IF NOT check_user_permissions(p_user_id, p_action) THEN
        RAISE_APPLICATION_ERROR(-20002, 'Access denied');
    END IF;

    -- Continue with procedure logic
END;
```

## Data Protection

### Sensitive Data Handling

#### Data Encryption
- Encrypt sensitive data at rest
- Use proper encryption algorithms
- Implement secure key management
- Regular encryption key rotation

#### Data Masking
- Mask sensitive data in non-production environments
- Implement data anonymization for testing
- Use proper data classification
- Follow data retention policies

### Audit Trail

#### Security Logging
- Log all security-relevant events
- Implement comprehensive audit trails
- Monitor for suspicious activities
- Regular log review and analysis

```sql
-- Example audit trail implementation
INSERT INTO security_audit_log (
    user_id,
    action,
    resource,
    timestamp,
    ip_address,
    result
) VALUES (
    :user_id,
    :action,
    :resource,
    SYSTIMESTAMP,
    :ip_address,
    :result
);
```

## Security Configuration

### Environment-Specific Security

#### Development Environment
- Use separate security configurations for development
- Implement proper secret management
- Regular security testing
- Secure development practices

#### Production Environment
- Implement strict security controls
- Regular security assessments
- Incident response procedures
- Security monitoring and alerting

### Security Headers

#### HTTP Security Headers
```javascript
// Example security headers configuration
{
    'X-Content-Type-Options': 'nosniff',
    'X-Frame-Options': 'DENY',
    'X-XSS-Protection': '1; mode=block',
    'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
    'Content-Security-Policy': "default-src 'self'"
}
```

## Security Best Practices

### General Security Guidelines

1. **Principle of Least Privilege**: Grant minimal necessary permissions
2. **Defense in Depth**: Implement multiple layers of security
3. **Security by Design**: Build security into the application from the start
4. **Regular Updates**: Keep all components updated with security patches
5. **Security Testing**: Regular penetration testing and vulnerability assessments

### Code Security Practices

1. **Input Validation**: Validate all input at multiple points
2. **Output Encoding**: Properly encode all output
3. **Error Handling**: Don't expose sensitive information in error messages
4. **Secure Communications**: Use HTTPS for all communications
5. **Session Management**: Implement secure session handling

### Database Security Practices

1. **Access Control**: Implement proper database access controls
2. **Encryption**: Encrypt sensitive data in the database
3. **Backup Security**: Secure database backups
4. **Monitoring**: Monitor database access and operations
5. **Regular Audits**: Conduct regular security audits

## Incident Response

### Security Incident Handling

#### Incident Detection
- Implement security monitoring
- Set up alerting for suspicious activities
- Regular log analysis
- User reporting mechanisms

#### Incident Response Process
1. **Detection**: Identify security incidents
2. **Analysis**: Analyze the scope and impact
3. **Containment**: Contain the incident
4. **Eradication**: Remove the threat
5. **Recovery**: Restore normal operations
6. **Lessons Learned**: Document and improve processes

### Security Monitoring

#### Continuous Monitoring
- Real-time security monitoring
- Automated threat detection
- Regular security assessments
- Performance impact monitoring

#### Compliance Requirements
- Meet regulatory compliance requirements
- Regular compliance audits
- Documentation and reporting
- Staff training and awareness
