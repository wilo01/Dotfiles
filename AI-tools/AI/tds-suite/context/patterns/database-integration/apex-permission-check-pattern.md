# APEX Permission Check Pattern

## Overview
Critical security pattern that must be implemented in every Oracle APEX Resource Template (.apex file) to ensure proper authorization before data access.

## Context
Every REST API endpoint in the TDS Suite must verify that the requesting user has permission to access the specific resource. This is enforced through the `safe_request_permissions_pak.check_permission_sql` function.

## Implementation

### Standard Pattern
```sql
where safe_request_permissions_pak.check_permission_sql(
   :Uidtoken, 
   '(portletName)', 
   'resourceName', 
   'HTTP_METHOD'
) = 'Y'
```

### Real Example
```xml
<?xml version="1.0" encoding="UTF-8"?>
<template xmlns="http://xmlns.oracle.com/apex/resource-template" 
          pattern="getNdaTranslationLanguages?nda_code={nda_code}">
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
  and safe_request_permissions_pak.check_permission_sql(
      :Uidtoken, '(ndaDashboard)', 'getNdaTranslationLanguages', 'GET'
  ) = 'Y'
order by
    case when cnv.end_date is null then 0 else 1 end,
    skl.language_description asc
]]></content>
<parameter name="Uidtoken" source="header" aliasing="Uidtoken"/>
</handler>
</template>
```

## Key Points
- **Always required**: Never create an .apex file without permission checks
- **Portlet name format**: Use parentheses around portlet name: `'(portletName)'`
- **Resource name**: Should match the file name (without .apex extension)
- **HTTP method**: Must match the actual HTTP method being used
- **Uidtoken parameter**: Always include as a header parameter
- **Database security**: This prevents unauthorized data access at the database level

## Parameter Details

### Uidtoken
- **Source**: HTTP header
- **Purpose**: User identification token
- **Required**: Always include `<parameter name="Uidtoken" source="header" aliasing="Uidtoken"/>`

### Portlet Name
- **Format**: Wrapped in parentheses `'(portletName)'`
- **Purpose**: Associates the permission with a specific UI component
- **Examples**: `'(ndaDashboard)'`, `'(visitorList)'`, `'(accessControl)'`

### Resource Name
- **Format**: String matching the API endpoint name
- **Purpose**: Identifies the specific resource being accessed
- **Convention**: Usually matches the .apex filename

### HTTP Method
- **Values**: `'GET'`, `'POST'`, `'PUT'`, `'DELETE'`
- **Purpose**: Method-specific permission checking
- **Requirement**: Must match the actual HTTP method used

## Common Mistakes
- ❌ Forgetting to include permission check entirely
- ❌ Wrong portlet name format (missing parentheses)
- ❌ Mismatched HTTP method
- ❌ Missing Uidtoken parameter definition
- ❌ Placing permission check in wrong part of WHERE clause

## Related Patterns
- [CSRF Protection Pattern](../security-implementations/csrf-protection-pattern.md)
- [XSS Protection Pattern](../security-implementations/xss-protection-pattern.md)
- [User Session Management](../../project-knowledge/business-logic/user-session-management.md)

## Examples
- All files in `source/server/rt/` directory
- Specifically: `getNdaTranslationLanguages.apex`, `getNdaArchiveList.apex`

## Generated On
2024-07-23 - Extracted from codebase analysis of security patterns