# TDS Suite Project Overview

## Project Description
The **TDS Suite** is a comprehensive enterprise visitor and access management system built on Oracle APEX with multiple ExtJS front-end applications. It handles visitor management, student portals, kiosk interfaces, and emergency muster systems.

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
- **Webpack** - Module bundling and development server
- **Docker** - Containerization for development environments

### DevOps & Deployment
- **Jenkins** - CI/CD pipeline automation
- **AWS** - Cloud infrastructure (S3, RDS)
- **ESLint** - Code linting with security plugins

## Project Structure

### Main Projects Location: `~/Dev/branch-opener/branches/`

### TDS Suite Structure (`~/Dev/branch-opener/branches/tds-suite/`)
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

## Key Architecture Patterns

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

## Project-Specific Features

### TDS Suite
- **Visitor Management**: Registration, approval workflows, badge printing
- **Access Control**: Security levels, time profiles, terminal management
- **Student Portal**: Academic tracking, attendance, course management
- **Kiosk Interface**: Self-service visitor registration
- **Emergency Management**: Muster points, evacuation procedures
- **Multi-tenant**: Location-based configurations with role-based access

## Development Guidelines

1. **Always refer to project structure rules** in `.github/workflows/instructions/tdssuite.md`
2. **Follow ExtJS conventions** for component naming and structure
3. **Use Liquibase changesets** for all database modifications
4. **Implement proper error handling** at all application layers
5. **Write comprehensive Cypress tests** for critical workflows
6. **Follow security best practices** for authentication and data handling