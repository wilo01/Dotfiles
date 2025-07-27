# TDS Suite - GitHub Copilot Instructions

Hello Dev, relax and watch the magic

## AI Role & Instructions
See detailed role configuration in: `AI_INSTRUCTIONS/agents/COPILOT_CONFIG.md`

## Shared Knowledge Base
- **Project Overview**: `AI_INSTRUCTIONS/shared/00_PROJECT_OVERVIEW.md` (comprehensive TDS Suite details)
- **Task Management**: `AI_INSTRUCTIONS/shared/01_TASK_MANAGEMENT.md`
- **UI Development**: `AI_INSTRUCTIONS/shared/02_UI_DEVELOPMENT.md` (ExtJS patterns, styling)
- **Backend Development**: `AI_INSTRUCTIONS/shared/03_BACKEND_DEVELOPMENT.md` (Oracle, APEX)
- **Testing & Deployment**: `AI_INSTRUCTIONS/shared/04_TESTING_DEPLOYMENT.md`
- **Security & Permissions**: `AI_INSTRUCTIONS/shared/05_SECURITY_PERMISSIONS.md`

## Multi-Agent Coordination
- **Coordination Protocols**: `AI_INSTRUCTIONS/protocols/MULTI_AGENT_COORDINATION.md`
- **File Conflict Prevention**: `AI_INSTRUCTIONS/protocols/FILE_CONFLICT_PREVENTION.md`

## Project Overview Summary

The **TDS Suite** is a comprehensive enterprise web application built using **ExtJS**, **JavaScript**, **Oracle Database**, and **PL/SQL**. This is a monolithic full-stack application with multiple UI components, server-side APIs, database management, and testing infrastructure.

## Technology Stack

### Frontend Technologies
- **ExtJS 6.0.2 / 6.2.0** - Primary UI framework for rich web applications
- **Modern JavaScript (ES5/ES6)** - Core application logic
- **SCSS/CSS** - Styling and theming
- **HTML5** - Markup structure

### Backend Technologies
- **Oracle Database** - Primary data storage with PL/SQL procedures
- **Java (JDK)** - Server-side runtime and APEX integration
- **Liquibase** - Database version control and migration
- **REST APIs** - Communication layer between UI and server

### Development & Testing
- **Cypress** - End-to-end testing framework
- **ESLint** - Code quality and security linting
- **Node.js** - Build tools and test utilities
- **Docker** - Containerization for development environments

## Architecture Overview

### Project Structure

```
source/
├── server/                 # Backend services and APIs
│   ├── auth/              # Authentication modules
│   ├── database/          # Database scripts, migrations, and Liquibase
│   ├── rest-api-fileupload/ # File upload REST services
│   ├── rt/                # Resource templates and API routes
│   └── rtapi/             # Runtime API implementations
├── ui/                    # Main web application (Backoffice)
│   ├── app/               # Core application logic
│   │   └── portlets/      # Individual screen components
│   ├── lib/               # Third-party libraries (ExtJS, plugins)
│   ├── css/               # Backoffice styling files
│   └── api/               # Client-side API wrappers
├── ui-kiosk/             # Kiosk application variant
│   ├── app/              # Kiosk application logic
│   ├── sass/             # Kiosk SCSS styling files
│   └── resources/        # Kiosk resources and assets
├── ui-muster/            # Muster application variant
└── ui-student-portal/    # Student portal application variant

test/
├── Cypress/              # E2E test suites
└── load/                 # Performance testing

selfhost/                 # Self-hosting and deployment tools
├── jdb_generator/        # JDB (Java Database) file generation
└── docker/               # Docker containerization
```

## Database Access & Docker Integration

### **Docker Database Connection**
The TDS Suite uses Oracle Database running in Docker container named **'trunk'**.

**Connection Details:**
- **Container Name**: `trunk`
- **Database User**: `sys`
- **Password**: `TDSSuite1`
- **Database Service**: `xepdb1`
- **Connection Role**: `sysdba`
- **SQL*Plus Mode**: Silent mode (`-s`) for clean output

### **Standard Docker Database Query Pattern**
```bash
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
[SQL COMMANDS]
EXIT;
EOF
```

### **Common Database Operations**

#### **1. Check Table Existence**
```bash
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT owner, table_name FROM all_tables WHERE table_name = 'CA_NDA_VERSION';
EXIT;
EOF
```

#### **2. Describe Table Structure**
```bash
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
DESC ca_nda_mapping;
EXIT;
EOF
```

#### **3. Query Table Data with Row Limit**
```bash
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT * FROM ca_nda_mapping WHERE ROWNUM <= 5;
EXIT;
EOF
```

#### **4. Find Tables by Pattern**
```bash
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT table_name FROM all_tables WHERE table_name LIKE 'CA_%' ORDER BY table_name;
EXIT;
EOF
```

#### **5. Check Table Constraints**
```bash
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT constraint_name, constraint_type, search_condition 
FROM all_constraints 
WHERE table_name = 'CA_NDA_MAPPING';
EXIT;
EOF
```

#### **6. View Table Columns and Data Types**
```bash
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT column_name, data_type, nullable, data_default 
FROM all_tab_columns 
WHERE table_name = 'CA_NDA_MAPPING' 
ORDER BY column_id;
EXIT;
EOF
```

### **Database Schema Investigation Scripts**

#### **Create Reusable Database Inspection Scripts**

```bash
#!/bin/bash
# File: .github/scripts/inspect-table.sh
# Usage: ./inspect-table.sh TABLE_NAME

TABLE_NAME=${1:-CA_NDA_MAPPING}

echo "=== TDS Suite Database Table Inspection: $TABLE_NAME ==="

echo "1. Checking table existence:"
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT owner, table_name FROM all_tables WHERE table_name = '$TABLE_NAME';
EXIT;
EOF

echo "2. Table structure:"
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
DESC $TABLE_NAME;
EXIT;
EOF

echo "3. Sample data (first 3 rows):"
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT * FROM $TABLE_NAME WHERE ROWNUM <= 3;
EXIT;
EOF

echo "4. Row count:"
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT COUNT(*) as total_rows FROM $TABLE_NAME;
EXIT;
EOF
```

#### **Schema Discovery Script**
```bash
#!/bin/bash
# File: .github/scripts/discover-schema.sh

echo "=== TDS Suite Schema Discovery ==="

echo "1. All CA_ tables:"
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT table_name FROM all_tables WHERE table_name LIKE 'CA_%' ORDER BY table_name;
EXIT;
EOF

echo "2. All users/schemas:"
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT username FROM all_users ORDER BY username;
EXIT;
EOF

echo "3. Database version:"
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT banner FROM v\$version WHERE ROWNUM = 1;
EXIT;
EOF
```

### **Database Development Best Practices**

#### **1. Always Use Silent Mode**
- Use `-s` flag in sqlplus for clean output without SQL*Plus banners
- Reduces noise in scripts and automated processes

#### **2. Proper Session Management**
- Always end SQL sessions with `EXIT;`
- Use heredoc (`<<EOF ... EOF`) for multi-line SQL blocks
- Container automatically cleans up sessions

#### **3. Safe Query Practices**
- Use `ROWNUM <= N` for limiting large result sets
- Test queries with small datasets first
- Always verify table existence before complex operations

#### **4. Error Handling**
```bash
# Check if Docker container is running
if ! docker ps | grep -q "trunk"; then
    echo "Error: Docker container 'trunk' is not running"
    exit 1
fi

# Execute SQL with error checking
RESULT=$(docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT COUNT(*) FROM ca_nda_mapping;
EXIT;
EOF)

if [[ $RESULT == *"ORA-"* ]]; then
    echo "SQL Error: $RESULT"
    exit 1
fi
```

### **Troubleshooting Database Connectivity**

#### **Common Issues and Solutions**

**1. Container Not Running**
```bash
# Check container status
docker ps | grep trunk

# Start container if stopped
docker start trunk
```

**2. Database Service Not Ready**
```bash
# Check database availability
docker exec trunk sqlplus -v

# Wait for database to be ready
docker exec trunk bash -c "
until sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba' <<< 'SELECT 1 FROM dual; EXIT;' >/dev/null 2>&1; do
    echo 'Waiting for database...'
    sleep 2
done
echo 'Database is ready!'
"
```

**3. Permission Issues**
```bash
# Verify connection credentials
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT user FROM dual;
EXIT;
EOF
```

**4. Table/View Does Not Exist (ORA-00942)**
```bash
# First, verify table exists
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT owner, table_name FROM all_tables WHERE table_name = 'YOUR_TABLE_NAME';
EXIT;
EOF

# Check for case sensitivity
docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT table_name FROM all_tables WHERE UPPER(table_name) = UPPER('your_table_name');
EXIT;
EOF
```

### **Integration with Development Workflow**

#### **Pre-commit Database Checks**
```bash
# File: .github/scripts/pre-commit-db-check.sh
#!/bin/bash

echo "Running pre-commit database checks..."

# Check critical tables exist
TABLES=("CA_NDA_VERSION" "CA_NDA_MAPPING" "CA_USER")

for table in "${TABLES[@]}"; do
    RESULT=$(docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
SELECT COUNT(*) FROM all_tables WHERE table_name = '$table';
EXIT;
EOF)
    
    if [[ "$RESULT" -eq "0" ]]; then
        echo "ERROR: Required table $table does not exist"
        exit 1
    fi
    echo "✓ Table $table exists"
done

echo "All database checks passed!"
```

## Styling and CSS Architecture

### Backoffice Application (`/source/ui`)
- **Styling Location**: `/source/ui/css/` directory
- **Main CSS File**: `safe.css` - Primary stylesheet for backoffice functionality
- **Specialized CSS Files**: 
  - `accessControl.css` - Access control specific styles
  - `dashboard.css` - Dashboard component styles
  - `studentPortal.css` - Student portal specific styles
  - `virtualReceptionist.css` - Virtual receptionist styles
- **Styling Approach**: Primarily uses ExtJS theming and CSS classes
- **Framework**: ExtJS Classic toolkit with traditional CSS

### Kiosk Application (`/source/ui-kiosk`)
- **Styling Location**: `/source/ui-kiosk/sass/` directory
- **Main SCSS Files**: 
  - `sass/var/all.scss` - Main SCSS variables and styles
  - `sass/etc/all.scss` - Additional SCSS imports
- **Styling Approach**: Uses SCSS (Sass) preprocessing
- **Framework**: ExtJS Modern toolkit with SCSS compilation
- **Build Process**: SCSS files are compiled to CSS during build

## 🏠 Icons and Assets

### Available Icon Collections
- **Main Icons**: `/source/ui/css/icons/` - Legacy icon collection
- **New Icons**: `/source/ui/css/new_icons/` - Updated icon collection with modern designs
- **Flags**: `/source/ui/css/icons/flags/` and `/source/ui/css/new_icons/flags/` - Country flag icons
- **Access Control**: `/source/ui/css/icons/ac/` and `/source/ui/css/new_icons/ac/` - Access control specific icons

### 📋 Common Icon Usage in ExtJS
```javascript
// Using iconCls in buttons
{
    xtype: 'button',
    text: 'Save',
    iconCls: 'save-green-button',
    handler: function() { /* handler code */ }
}

// Using icons in grid action columns
{
    xtype: 'actioncolumn',
    items: [{
        iconCls: 'database-edit-click',
        tooltip: 'Edit record'
    }]
}

// Using icons in menu items
{
    xtype: 'menuitem',
    text: 'Export',
    iconCls: 'export-excel-icon',
    handler: function() { /* handler code */ }
}

// Using icons in toolbar buttons
{
    xtype: 'toolbar',
    items: [{
        iconCls: 'add-green-button',
        text: 'Add New',
        tooltip: 'Add new record'
    }]
}
```

### Icon Naming Conventions
- **Action Icons**: Use descriptive names like `save-green-button`, `delete-red-button`
- **Status Icons**: Include status in name like `status-active`, `status-inactive`
- **File Type Icons**: Use format like `file-pdf`, `file-excel`, `export-csv`
- **Navigation Icons**: Use direction like `arrow-left`, `arrow-right`, `arrow-up`, `arrow-down`

### Best Practices for Icons
1. **Consistency**: Use the same icon for the same action across the application
2. **Accessibility**: Always provide tooltips for icon-only buttons
3. **Size Compatibility**: Ensure icons work well at different sizes (16px, 24px, 32px)
4. **Color Coding**: Use consistent color schemes (green for success/add, red for delete/error)
5. **Fallback**: Always provide text labels alongside icons for better UX

## Key Components & Logic

### 1. ExtJS Application Architecture

The application follows ExtJS MVC/MVVM patterns:

#### **Viewport Management**
- **Main Application**: `source/ui/app/app.js` - Central application controller
- **Contractor Viewport**: `source/ui/app/contractorViewport.js` - Specialized contractor interface
- **Portlets**: Individual screen components located in `source/ui/app/portlets/`

#### **Data Management**
- **Stores**: ExtJS data stores for client-side data management
- **Models**: Data models defining structure and validation
- **Proxies**: REST and AJAX proxies for server communication

```javascript
// Example Store Configuration
Ext.create('Ext.data.Store', {
   model: 'formItemsModel',
   restful: true,
   proxy: {
      type: 'rest',
      url: '/' + app.deploy_context + '/form',
      reader: {
         type: 'json',
         rootProperty: 'items',
         totalProperty: 'total'
      }
   }
});
```

### 2. Database Layer

#### **Oracle Database Integration**
- **Connection Management**: Database connections configured via XML configuration files
- **PL/SQL Procedures**: Stored procedures for business logic
- **Liquibase Migrations**: Version-controlled database schema changes

```xml
<!-- Database Configuration Example -->
<entry key="apex.db.hostname">hostname</entry>
<entry key="apex.db.port">1521</entry>
<entry key="apex.db.sid">database_sid</entry>
```

#### **Database Migration Process**
- **Liquibase Scripts**: Located in `source/server/database/`
- **Changelog Management**: Structured changelog system for tracking database changes
- **Environment-specific Configurations**: Separate configs for dev, test, and production

### 3. REST API Architecture

#### **Resource Templates (.apex files)**
- **Location**: `source/server/rt/` directory
- **Purpose**: Define API endpoints and business logic
- **Format**: Custom .apex file format containing SQL and PL/SQL procedures

#### **File Upload API**
- **Location**: `source/server/rest-api-fileupload/`
- **Technology**: Java-based REST service
- **Security**: Encrypted password authentication

### 4. Multi-Application Architecture

The system supports multiple specialized applications:

#### **Main UI** (`source/ui/`)
- Primary back-office application
- Full feature set for administrators
- ExtJS Classic toolkit

#### **Kiosk Application** (`source/ui-kiosk/`)
- Touch-enabled kiosk interface
- Simplified UI for public access
- ExtJS Modern toolkit

#### **Muster Application** (`source/ui-muster/`)
- Emergency muster functionality
- Specialized for headcount and safety procedures
- Touch/mobile optimized

#### **Student Portal** (`source/ui-student-portal/`)
- Student-specific interface
- Self-service functionality
- Modern UI design

### 5. Security & Authentication

#### **User Authentication**
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

### 6. Testing Framework

#### **Cypress E2E Testing**
- **Test Location**: `test/Cypress/cypress/e2e/`
- **Oracle Integration**: Direct database testing capabilities
- **Multi-environment Support**: Docker and local configurations

```javascript
// Cypress Oracle Integration Example
cy.task('oracle.executeSQL', clearQuery);
```

#### **Load Testing**
- **K6 Framework**: Performance testing scripts
- **Location**: `test/load/`
- **Metrics Collection**: Response time and throughput analysis

### 7. Build & Deployment

#### **JDB Generation**
- **Purpose**: Creates Java Database files for APEX deployment
- **Process**: Automated ZIP creation and APEX listener interaction
- **Location**: `selfhost/jdb_generator/`

#### **Docker Integration**
- **Development Environment**: Containerized development setup
- **Configuration**: Environment-specific Docker configurations
- **Database Integration**: Oracle database containers

## Development Guidelines

### Code Organization

1. **Component-based Architecture**: Each UI component should be self-contained
2. **Separation of Concerns**: Clear distinction between UI, business logic, and data layers
3. **RESTful API Design**: Consistent API endpoints following REST principles
4. **Database Best Practices**: Use stored procedures for complex business logic

### JavaScript Coding Style

- **Variable Declarations**: Always use `let` or `const` for variable declarations. Avoid using `var`.
  ```javascript
  // ✅ DO: Use let or const
  let userName = 'John Doe';
  const API_KEY = 'your_api_key';

  // ❌ DON'T: Use var
  var oldVariable = 'deprecated';
  ```

### Styling Guidelines

#### Backoffice Application (`/source/ui`)
1. **CSS Location**: All styles should be placed in `/source/ui/css/` directory
2. **Main Stylesheet**: Use `safe.css` for general application styles
3. **Component-specific CSS**: Create dedicated CSS files for major components
4. **ExtJS Integration**: Leverage ExtJS theming system and CSS classes
5. **Naming Convention**: Use descriptive class names with component prefixes
6. **CSS Specificity**: Use `!important` sparingly, prefer specific selectors

#### Kiosk Application (`/source/ui-kiosk`)
1. **SCSS Location**: All styles should be placed in `/source/ui-kiosk/sass/` directory
2. **Main SCSS File**: Use `sass/var/all.scss` for variables and main styles
3. **SCSS Features**: Utilize SCSS variables, mixins, and nesting
4. **Build Process**: Ensure SCSS compilation works properly
5. **Mobile-first**: Design for touch interfaces and various screen sizes
6. **Performance**: Optimize CSS output for mobile devices

## Key Configuration Files

### Database Configuration
- `apex-config.xml` - Database connection settings
- `liquibase.properties` - Migration configuration
- `changelog.xml` - Database schema versioning

### Application Configuration
- `app.json` - ExtJS application configuration
- `package.json` - Node.js dependencies and scripts
- `cypress.config.js` - Testing configuration

### Environment-specific Settings
- `cypress.config.local.js` - Local development settings
- `cypress.config.docker.js` - Docker environment settings
- Various environment-specific XML configurations

## Common Development Patterns

### ExtJS Component Definition
```javascript
Ext.define('MyApp.view.MyComponent', {
   extend: 'Ext.panel.Panel',
   xtype: 'mycomponent',
   
   requires: [
      'MyApp.model.MyModel',
      'MyApp.store.MyStore'
   ],
   
   initComponent: function() {
      // Component initialization logic
      this.callParent();
   }
});
```

### Database Operation Pattern
```javascript
// Oracle database operation via Cypress
const executeQuery = async (sql) => {
   return cy.task('oracle.executeSQL', [sql]);
};
```

### REST API Integration
```javascript
// ExtJS Store with REST proxy
proxy: {
   type: 'rest',
   url: '/api/endpoint',
   reader: {
      type: 'json',
      rootProperty: 'data'
   },
   extraParams: {
      // Additional parameters
   }
}
```

## Troubleshooting Common Issues

### ExtJS Component Issues
- Check component lifecycle methods
- Verify proper store configuration
- Ensure correct proxy settings

### Database Connection Issues
- Validate Oracle connection strings
- Check user permissions and roles
- Verify Liquibase migration status

### Build and Deployment Issues
- Check Java version compatibility
- Validate Docker container configurations
- Ensure proper environment variable settings

## Best Practices Summary

1. **Always refer to the project structure rules** in `.github/workflows/instructions/tdssuite.md`
2. **Follow ExtJS conventions** for component naming and structure
3. **Use correct styling paths**:
   - **Backoffice UI**: Place CSS in `/source/ui/css/` directory
   - **Kiosk UI**: Place SCSS in `/source/ui-kiosk/sass/` directory
4. **Use parameterized queries** for all database operations
5. **Implement proper error handling** at all application layers
6. **Write comprehensive tests** for critical business logic
7. **Document API endpoints** and database schema changes
8. **Follow security best practices** for authentication and authorization
9. **Optimize performance** through proper caching and lazy loading strategies
10. **Use Docker for consistent database access** following the established pattern:
    ```bash
    docker exec -i trunk bash -c "sqlplus -s 'sys/TDSSuite1@xepdb1 as sysdba'" <<EOF
    [SQL COMMANDS]
    EXIT;
    EOF
    ```

---

## Important Note

**This file contains historical comprehensive documentation.** For the latest modular organization and multi-agent coordination:

- **Complete guidelines**: Reference `AI_INSTRUCTIONS/` directory
- **Agent-specific rules**: See `AI_INSTRUCTIONS/agents/COPILOT_CONFIG.md`
- **Coordination protocols**: See `AI_INSTRUCTIONS/protocols/`
- **Shared knowledge**: See `AI_INSTRUCTIONS/shared/` modules

*Always ensure code changes align with the established patterns, security requirements, and multi-agent coordination protocols.*
