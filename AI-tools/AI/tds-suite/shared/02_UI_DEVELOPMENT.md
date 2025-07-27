# UI Development Guidelines

## ExtJS Framework

### Core Principles
- Write concise, technical responses
- Always consider readability and maintainability
- Implement proper component composition and reusability
- Leverage safe_util.js built-in components when appropriate

### Component Development Patterns

#### ExtJS Component Definition
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

### ExtJS Application Architecture

#### Viewport Management
- **Main Application**: `source/ui/app/app.js` - Central application controller
- **Contractor Viewport**: `source/ui/app/contractorViewport.js` - Specialized contractor interface
- **Portlets**: Individual screen components located in `source/ui/app/portlets/`

#### Data Management
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

### Data Store Patterns

#### REST API Integration
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

### Portlets System
- Components extend ExtJS base classes and follow MVC/MVVM patterns
- Portlets are defined in `source/ui/app/portlets/`
- Each component should have proper lifecycle management and state handling
- Use `stateIdPrefix` for component state management

## Styling Architecture

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

## Icons and Assets

### Available Icon Collections
- **Main Icons**: `/source/ui/css/icons/` - Legacy icon collection
- **New Icons**: `/source/ui/css/new_icons/` - Updated icon collection with modern designs
- **Flags**: `/source/ui/css/icons/flags/` and `/source/ui/css/new_icons/flags/` - Country flag icons
- **Access Control**: `/source/ui/css/icons/ac/` and `/source/ui/css/new_icons/ac/` - Access control specific icons

### Common Icon Usage in ExtJS
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

## Multi-Application Architecture

### Main UI (`source/ui/`)
- Primary back-office application
- Full feature set for administrators
- ExtJS Classic toolkit

### Kiosk Application (`source/ui-kiosk/`)
- Touch-enabled kiosk interface
- Simplified UI for public access
- ExtJS Modern toolkit

### Muster Application (`source/ui-muster/`)
- Emergency muster functionality
- Specialized for headcount and safety procedures
- Touch/mobile optimized

### Student Portal (`source/ui-student-portal/`)
- Student-specific interface
- Self-service functionality
- Modern UI design

## JavaScript Coding Standards

### Variable Declarations
Always use `let` or `const` for variable declarations. Avoid using `var`.

```javascript
// ✅ DO: Use let or const
let userName = 'John Doe';
const API_KEY = 'your_api_key';

// ❌ DON'T: Use var
var oldVariable = 'deprecated';
```

### Performance Optimization
- Minimize use of client-side JavaScript; leverage static generation
- Implement proper lazy loading for images and other assets

## Development Best Practices

### Code Organization
1. **Component-based Architecture**: Each UI component should be self-contained
2. **Separation of Concerns**: Clear distinction between UI, business logic, and data layers
3. **ExtJS Conventions**: Follow ExtJS conventions for component naming and structure

### Component Lifecycle
- Check component lifecycle methods
- Verify proper store configuration
- Ensure correct proxy settings

### Text Translation System
All user-facing text must be translatable using the `app.text` functions from `@source/ui/app/text.js`.

#### Translation Guidelines
1. **Always use `app.text` functions** for any user-facing text
2. **Check existing functions first** - reuse when available instead of creating duplicates
3. **Create new functions** for text not already covered
4. **Function naming convention**: Use descriptive names like `loadingText()`, `saveButtonText()`

#### Text Function Examples
```javascript
// ✅ DO: Use app.text functions
text: app.text.loadingText(),           // Returns "Loading..."
title: app.text.errorDialogTitle(),     // Returns "Error"
tooltip: app.text.saveButtonTooltip()   // Returns "Save changes"

// ❌ DON'T: Hardcode text strings
text: 'Loading...',
title: 'Error',
tooltip: 'Save changes'
```

#### Creating New Text Functions
```javascript
// In @source/ui/app/text.js - add new functions as needed
loadingText: function() {
    return 'Loading...';
},

// Organized by component/feature
ndaArchive: {
    selectLanguageText: function() {
        return 'Select Language';
    },
    noLanguagesAvailableText: function() {
        return 'No languages available';
    }
}
```

#### Usage in ExtJS Components
```javascript
// In component definitions
{
    xtype: 'button',
    text: app.text.saveButtonText(),
    tooltip: app.text.saveButtonTooltip(),
    handler: function() {
        Ext.Msg.alert(app.text.successTitle(), app.text.saveSuccessMessage());
    }
}

// In loading masks
me.setLoading(app.text.loadingText());

// In grid columns
{
    text: app.text.nameColumnHeader(),
    dataIndex: 'name'
}
```

### Unified Content Display Pattern
For components that display content based on multiple selection criteria (e.g., grid + language selection):

#### Pattern Structure
```javascript
// Single unified method to handle all content display scenarios
displayContent: function(versionId, languageData) {
    var me = this;
    
    // Update state based on provided parameters
    if (versionId) {
        me.ndaData.versionId = versionId;
    }
    if (languageData) {
        me.ndaData.selectedLanguageData = languageData;
        me.ndaData.languageCode = languageData.language_code;
    }
    
    // Show loading with translatable text
    me.showLoadingOnPreviewPanel(app.text.loadingText());
    
    // Make unified API call
    me.callUnifiedContentAPI();
}
```

#### Benefits
1. **Single source of truth** for content loading logic
2. **Consistent state management** across all selection methods
3. **Unified error handling** and loading states
4. **Easier testing and maintenance**

### Accessibility
- Ensure proper semantic HTML structure in components
- Implement ARIA attributes where necessary
- Ensure keyboard navigation support for interactive elements