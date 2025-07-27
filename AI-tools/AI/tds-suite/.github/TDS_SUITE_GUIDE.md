# TDS Suite - Development Guide

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Hello Dev, relax and watch the magic**

The TDS Suite is a comprehensive enterprise visitor and access management system built on Oracle APEX with multiple ExtJS front-end applications. It handles visitor management, student portals, kiosk interfaces, and emergency muster systems.

## Technology Stack

### Core Technologies
- **ExtJS 6.0.2 - 7.4.0** - Primary UI framework (Classic and Modern toolkits)
- **JavaScript (ES5/ES6)** - Core application logic
- **Oracle APEX** - Application development platform
- **Oracle Database** - Primary data storage with PL/SQL procedures
- **Java (JDK 8)** - Server-side runtime and APEX integration
- **GlassFish** - Java EE application server

### Development & Testing
- **Liquibase** - Database version control and migration
- **Cypress 13.13.0** - End-to-end testing framework
- **Node.js/npm** - Build tools and test utilities
- **Sencha CMD** - ExtJS application building and development
- **Gradle** - Java build automation (JDK 8, Gradle 4.1)
- **Docker** - Containerization for development environments

### DevOps & Deployment
- **Jenkins** - CI/CD pipeline automation
- **ESLint** - Code linting with security plugins

## Project Structure

```
source/
├── server/                 # Backend services and APIs
│   ├── auth/              # Authentication modules
│   ├── database/          # Database scripts, migrations, and Liquibase
│   ├── rest-api-fileupload/ # File upload REST services (Java/Gradle)
│   ├── rt/                # Resource templates and API routes (.apex files)
│   ├── dba/               # Database administration scripts
│   └── glassfish/         # Java EE application server configurations
├── ui/                    # Main web interface with HTML templates
├── ui-kiosk/             # Kiosk interface (ExtJS 6.2.1, Modern toolkit)
├── ui-student-portal/    # Student portal (ExtJS 7.4.0, Modern toolkit)
└── ui-muster/            # Emergency muster/evacuation interface

test/
├── Cypress/              # Comprehensive E2E test suite
└── load/                 # Performance/load testing (Artillery.js)

selfhost/                 # Self-hosting and deployment tools
├── jdb_generator/        # JDB (Java Database) file generation
└── docker/               # Docker containerization
```

## Development Commands

### Database Operations
```bash
# Liquibase database updates
npm run liquibaseLocalDockerUpdate --port=1521 --user=safe --password=****
npm run lldupdate --port=1521 --user=safe --password=****  # Shorthand

# Docker database environment
npm run tds-docker --container_name=tds-oracle --port=1521 --oracle_pwd=****
```

### ExtJS Application Development
```bash
# Main UI application (ExtJS 6.0.2 Classic)
cd source/ui/
sencha app build development
sencha app build production
sencha app watch  # Development server with live reload

# Kiosk application (ExtJS 6.2.1 Modern)
cd source/ui-kiosk/
sencha app build development
sencha app build production
sencha app watch

# Student Portal (ExtJS 7.4.0 Modern)
cd source/ui-student-portal/
sencha app build development
sencha app build production
sencha app watch

# Muster application (ExtJS Touch)
cd source/ui-muster/
sencha app build development
sencha app build production
sencha app watch
```

### Java REST Services
```bash
cd source/server/rest-api-fileupload/
./gradlew build         # Build WAR file
./gradlew explodedWar   # Development deployment
```

### Testing
```bash
# Cypress E2E Testing
cd test/Cypress/
npm install
npm run devDocker        # Interactive testing with Chrome
npm run devDockerAll     # Run all tests
npm run devDockerAccess  # Access-specific tests
npm run devDockerVisitor # Visitor-specific tests
npm run trunkAutomation  # Production environment testing

# Test data management
npm run devDockerCreateData   # Create test data
npm run devDockerDeleteData   # Clean test data
```

## Key Architecture Patterns

### ExtJS Component Development
- Components extend ExtJS base classes and follow MVC/MVVM patterns
- Portlets are defined in `source/ui/app/portlets/`
- Each component should have proper lifecycle management and state handling
- Use `stateIdPrefix` for component state management

### Database Integration
- Oracle Database with PL/SQL stored procedures for business logic
- Liquibase for database version control and migrations
- Database changes controlled through structured changesets
- Use parameterized queries to prevent SQL injection

### Menu System
- Menus are defined in database with scripts in `safe_menu_items/`
- Use `ca_menu_pak.manage_menu_item()` procedure for menu management
- Menu items link to JavaScript classes and source files

### REST API Architecture
- Resource Templates (.apex files) in `source/server/rt/` define API endpoints
- Java-based REST services for file upload functionality
- CSRF token implementation for security
- RESTful design principles with consistent JSON responses

## Multi-Application Architecture

The system supports multiple specialized applications:
- **Main UI**: Primary back-office application with full feature set
- **Kiosk Application**: Touch-enabled interface for public access
- **Muster Application**: Emergency procedures and headcount functionality
- **Student Portal**: Self-service student interface

## Security Considerations

- Implement CSRF protection with anti-CSRF tokens
- Use token-based authentication with user session management
- Role-based access control for user permissions
- Input validation on both client and server sides
- Secure database operations with parameterized queries

## Key Configuration Files

- `apex-config.xml` - Database connection settings
- `liquibase.properties` - Migration configuration
- `app.json` - ExtJS application configuration
- `cypress.config.js` - Testing configuration
- Environment-specific Docker and XML configurations

## Testing Framework

### End-to-End Testing (Cypress)
- Comprehensive E2E suite with direct Oracle database integration
- Test categories: Access control, visitor management, student portal
- Environment support: Local Docker, UAT, production
- Database operations: Direct SQL queries for test data management
- Custom commands: Specialized TDS Suite testing utilities

### Load Testing (Artillery.js)
- Performance scenarios: Guest dashboard, visitor registration, reception workflows
- Stress testing: Multiple user scenarios
- API endpoint testing: REST service validation

## Development Guidelines

1. **Always refer to project structure rules** in `.github/workflows/instructions/tdssuite.md`
2. **Follow ExtJS conventions** for component naming and structure
3. **Use Liquibase changesets** for all database modifications
4. **Implement proper error handling** at all application layers
5. **Write comprehensive Cypress tests** for critical workflows
6. **Follow security best practices** for authentication and data handling
7. **Begin responses with**: "Hello Dev, relax and watch the magic"

## Styling Guidelines

### Backoffice Application (`source/ui/`)
- **CSS Location**: All styles in `/source/ui/css/` directory
- **Main Stylesheet**: `safe.css` for general application styles
- **Component-specific CSS**: Create dedicated CSS files for major components
- **ExtJS Integration**: Leverage ExtJS theming system and CSS classes

### Kiosk Application (`source/ui-kiosk/`)
- **SCSS Location**: All styles in `/source/ui-kiosk/sass/` directory
- **Main SCSS File**: `sass/var/all.scss` for variables and main styles
- **SCSS Features**: Utilize SCSS variables, mixins, and nesting
- **Mobile-first**: Design for touch interfaces and various screen sizes

## Icon Management

### Available Icon Collections
- **Main Icons**: `/source/ui/css/icons/` - Legacy collection
- **New Icons**: `/source/ui/css/new_icons/` - Updated modern designs
- **Specialized Icons**: Flags, access control, and domain-specific icons

### Icon Usage in ExtJS
```javascript
// Button with icon
{
    xtype: 'button',
    text: 'Save',
    iconCls: 'save-green-button',
    handler: function() { /* handler */ }
}

// Action column with icon
{
    xtype: 'actioncolumn',
    items: [{
        iconCls: 'database-edit-click',
        tooltip: 'Edit record'
    }]
}
```

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
        this.callParent();
    }
});
```

### Database Operation via Cypress
```javascript
// Oracle database testing
cy.task('oracle.executeSQL', clearQuery);
```

### REST API Store Configuration
```javascript
proxy: {
    type: 'rest',
    url: '/' + app.deploy_context + '/endpoint',
    reader: {
        type: 'json',
        rootProperty: 'items',
        totalProperty: 'total'
    }
}
```

## Project-Specific Features

### TDS Suite
- **Visitor Management**: Registration, approval workflows, badge printing
- **Access Control**: Security levels, time profiles, terminal management
- **Student Portal**: Academic tracking, attendance, course management
- **Kiosk Interface**: Self-service visitor registration
- **Emergency Management**: Muster points, evacuation procedures
- **Multi-tenant**: Location-based configurations with role-based access