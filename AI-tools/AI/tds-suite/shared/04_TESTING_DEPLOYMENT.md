# Testing & Deployment Guidelines

## Cypress E2E Testing

### Test Location and Structure
- **Test Location**: `test/Cypress/cypress/e2e/`
- **Oracle Integration**: Direct database testing capabilities
- **Multi-environment Support**: Docker and local configurations

### Cypress Oracle Integration
```javascript
// Cypress Oracle Integration Example
cy.task('oracle.executeSQL', clearQuery);

// Database operation via Cypress
const executeQuery = async (sql) => {
   return cy.task('oracle.executeSQL', [sql]);
};
```

### Testing Framework Configuration
- **Configuration Files**:
  - `cypress.config.js` - Main testing configuration
  - `cypress.config.local.js` - Local development settings
  - `cypress.config.docker.js` - Docker environment settings

### Test Development Best Practices
1. **Database Testing**: Write tests that verify database operations
2. **UI Testing**: Test ExtJS components and user interactions
3. **API Testing**: Verify REST API endpoints and responses
4. **Cross-browser Testing**: Ensure compatibility across different browsers

## Load Testing

### K6 Framework
- **Purpose**: Performance testing scripts
- **Location**: `test/load/`
- **Metrics Collection**: Response time and throughput analysis

### Performance Testing Guidelines
1. **Baseline Metrics**: Establish performance baselines
2. **Load Scenarios**: Test various load conditions
3. **Bottleneck Identification**: Identify performance bottlenecks
4. **Scalability Testing**: Test system scalability limits

## Build & Deployment

### Build Commands

#### Development Commands
```bash
# ExtJS Development Build
sencha app build development

# ExtJS Production Build
sencha app build production

# Development Server with Live Reload
sencha app watch

# Webpack Development Server (port 1962)
webpack-dev-server --env.environment=development
```

#### Backend Build Commands
```bash
# Build WAR file
./gradlew war

# Create Exploded WAR
./gradlew explodedWar
```

#### Database Operations
```bash
# Database updates via Docker
npm run liquibaseLocalDockerUpdate

# Shorthand for database updates
npm run lldupdate

# Direct Liquibase updates
make db-update

# Database rollbacks
make db-rollback

# PL/SQL testing
make test-plsql
```

#### Testing Commands
```bash
# Interactive Cypress testing
npm run devDocker

# Run all Cypress tests
npm run devDockerAll

# Access-specific tests
npm run devDockerAccess

# Visitor-specific tests
npm run devDockerVisitor

# Production testing
npm run trunkAutomation
```

### JDB Generation

#### Purpose and Process
- **Purpose**: Creates Java Database files for APEX deployment
- **Process**: Automated ZIP creation and APEX listener interaction
- **Location**: `selfhost/jdb_generator/`

### Docker Integration

#### Development Environment
- **Containerized Development**: Docker setup for consistent development environment
- **Configuration**: Environment-specific Docker configurations
- **Database Integration**: Oracle database containers

## Configuration Files

### Key Configuration Files

#### Database Configuration
- `apex-config.xml` - Database connection settings
- `liquibase.properties` - Migration configuration
- `changelog.xml` - Database schema versioning

#### Application Configuration
- `app.json` - ExtJS application configuration
- `package.json` - Node.js dependencies and scripts
- `cypress.config.js` - Testing configuration

#### Environment-specific Settings
- `cypress.config.local.js` - Local development settings
- `cypress.config.docker.js` - Docker environment settings
- Various environment-specific XML configurations

## CI/CD Pipeline

### Jenkins Integration
- **Pipeline Automation**: Jenkins CI/CD pipeline automation
- **Build Triggers**: Automated builds on code changes
- **Testing Integration**: Automated test execution
- **Deployment Process**: Automated deployment to different environments

### AWS Integration
- **Cloud Infrastructure**: AWS S3, RDS integration
- **Deployment Targets**: Cloud-based deployment environments
- **Scaling**: Auto-scaling capabilities

## Testing Best Practices

### Unit Testing
1. **Component Testing**: Test individual ExtJS components
2. **Function Testing**: Test utility functions and helpers
3. **Model Testing**: Test data models and validation
4. **Store Testing**: Test ExtJS stores and data operations

### Integration Testing
1. **API Integration**: Test REST API endpoints
2. **Database Integration**: Test database operations and procedures
3. **Component Integration**: Test component interactions
4. **User Workflow Testing**: Test complete user workflows

### End-to-End Testing
1. **Critical Path Testing**: Test critical business workflows
2. **Cross-browser Testing**: Test on multiple browsers
3. **Performance Testing**: Test application performance
4. **Security Testing**: Test security features and vulnerabilities

### Testing Guidelines
1. **Test Coverage**: Aim for high test coverage
2. **Test Documentation**: Document test cases and scenarios
3. **Test Data Management**: Manage test data effectively
4. **Continuous Testing**: Integrate testing into development workflow

## Deployment Process

### Environment Management
1. **Development Environment**: Local development setup
2. **Testing Environment**: Dedicated testing environment
3. **Staging Environment**: Pre-production environment
4. **Production Environment**: Live production environment

### Deployment Best Practices
1. **Version Control**: Use proper version control for deployments
2. **Rollback Strategy**: Have rollback procedures in place
3. **Health Checks**: Implement deployment health checks
4. **Monitoring**: Monitor application performance post-deployment

### Build Optimization
1. **Asset Optimization**: Optimize CSS, JS, and image assets
2. **Minification**: Minify production assets
3. **Caching**: Implement proper caching strategies
4. **Performance Monitoring**: Monitor build and deployment performance

## Troubleshooting

### Build Issues
- Check Java version compatibility
- Validate Docker container configurations
- Ensure proper environment variable settings

### Test Issues
- Verify test database configuration
- Check Cypress installation and setup
- Validate test data and fixtures

### Deployment Issues
- Check deployment target configuration
- Verify network connectivity
- Validate environment-specific settings
- Review deployment logs for errors