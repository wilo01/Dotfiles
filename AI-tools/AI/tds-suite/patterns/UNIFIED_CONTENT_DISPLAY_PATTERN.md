# Unified Content Display Pattern

## Overview
The Unified Content Display Pattern provides a standardized approach for components that need to display content based on multiple selection criteria (e.g., grid selection + language selection, filter + search, etc.).

## Problem Statement
Many components in the TDS Suite need to display content that depends on multiple user selections:
- **NDA Archive**: Grid selection + Language selection
- **Document Viewer**: Category selection + Document selection
- **Report Viewer**: Date range + Report type + Filters

Without a unified approach, each component implements its own content loading logic, leading to:
- **Inconsistent state management**
- **Duplicate loading logic**
- **Complex debugging**
- **Maintenance overhead**

## Solution: Unified Content Display Method

### Core Pattern Structure
```javascript
displayContent: function(primarySelection, secondarySelection, additionalParams) {
    var me = this;
    
    // 1. Update component state based on all provided parameters
    me.updateComponentState(primarySelection, secondarySelection, additionalParams);
    
    // 2. Show loading state with translatable text
    me.showLoadingState();
    
    // 3. Make unified API call with all current state
    me.callContentAPI();
}
```

### Real-World Example: NDA Archive Widget
```javascript
displayContent: function(versionId, languageData) {
    var me = this;
    
    // Update state based on provided parameters
    if (versionId) {
        me.ndaData.versionId = versionId;
    }
    if (languageData) {
        me.ndaData.selectedLanguageData = languageData;
        me.ndaData.languageCode = languageData.language_code;
        me.ndaData.languageDescription = languageData.language_description;
    }
    
    // Show loading with translatable text
    me.showLoadingOnPreviewPanel(app.text.loadingText());
    
    // Make unified API call
    me.loadNdaContent();
}
```

## Implementation Guidelines

### 1. Single Entry Point
Create one method that handles all content display scenarios:

```javascript
// ✅ GOOD: Single unified method
displayContent: function(gridSelection, filterSelection, searchParams) {
    // Handle all scenarios in one place
}

// ❌ AVOID: Multiple specialized methods
loadContentFromGrid: function(selection) { /* ... */ }
loadContentFromFilter: function(filter) { /* ... */ }
loadContentFromSearch: function(search) { /* ... */ }
```

### 2. State Management
Update component state consistently across all selection methods:

```javascript
displayContent: function(primaryParam, secondaryParam) {
    var me = this;
    
    // Update state based on what's provided
    if (primaryParam) {
        me.componentData.primarySelection = primaryParam;
    }
    if (secondaryParam) {
        me.componentData.secondarySelection = secondaryParam;
    }
    
    // State is now consistent regardless of how we got here
    me.loadContent();
}
```

### 3. Loading State Management
Use consistent loading indicators with translatable text:

```javascript
showLoadingState: function() {
    var me = this;
    
    // Use translatable loading text
    me.setLoading(app.text.loadingText());
    
    // Or component-specific loading
    me.showLoadingOnPreviewPanel(app.text.componentName.loadingContentText());
}
```

### 4. Error Handling
Centralize error handling for all content loading scenarios:

```javascript
handleContentLoadError: function(error) {
    var me = this;
    
    me.hideLoading();
    
    Ext.Msg.alert(
        app.text.errorDialogTitle(),
        app.text.componentName.loadContentErrorText()
    );
}
```

## Usage Scenarios

### Scenario 1: Grid Selection Only
```javascript
// User selects item from grid
onGridSelect: function(selection) {
    var me = this;
    
    // Call unified method with grid selection only
    me.displayContent(selection.get('id'), null);
}
```

### Scenario 2: Language Selection Only
```javascript
// User selects language from dropdown
onLanguageSelect: function(languageData) {
    var me = this;
    
    // Call unified method with language selection only
    me.displayContent(null, languageData);
}
```

### Scenario 3: Combined Selection
```javascript
// User selects both grid item and language
onCombinedSelection: function(gridSelection, languageData) {
    var me = this;
    
    // Call unified method with both selections
    me.displayContent(gridSelection.get('id'), languageData);
}
```

### Scenario 4: Refresh Current Content
```javascript
// Refresh with current state
refreshContent: function() {
    var me = this;
    
    // Call unified method with current state (no new selections)
    me.displayContent(null, null);
}
```

## Benefits

### 1. Consistency
- **Same loading behavior** across all selection methods
- **Consistent state management** regardless of entry point
- **Unified error handling** for all scenarios

### 2. Maintainability
- **Single place** to modify content loading logic
- **Easier debugging** - one method to trace
- **Reduced code duplication** across selection handlers

### 3. Testability
- **One method** to test for all scenarios
- **Predictable behavior** across different inputs
- **Easier mocking** for unit tests

### 4. User Experience
- **Consistent loading indicators** across the component
- **Predictable behavior** regardless of how content is selected
- **Seamless transitions** between different selection methods

## Advanced Pattern: State Preservation

For components with complex state, preserve selections across different operations:

```javascript
displayContent: function(newPrimary, newSecondary, preserveState) {
    var me = this;
    
    // Preserve existing state unless explicitly overridden
    var primary = newPrimary || me.componentData.currentPrimary;
    var secondary = newSecondary || me.componentData.currentSecondary;
    
    // Only proceed if we have necessary data
    if (!primary && !secondary) {
        me.showEmptyState();
        return;
    }
    
    // Update state
    me.componentData.currentPrimary = primary;
    me.componentData.currentSecondary = secondary;
    
    // Load content
    me.loadContent();
}
```

## Integration with Other Patterns

### With Text Translation
```javascript
displayContent: function(selection, filter) {
    var me = this;
    
    // Show loading with translatable text
    me.setLoading(app.text.componentName.loadingContentText());
    
    // Update state and load
    me.updateState(selection, filter);
    me.loadContent();
}
```

### With Error Handling
```javascript
loadContent: function() {
    var me = this;
    
    Ext.Ajax.request({
        url: me.getContentUrl(),
        success: function(response) {
            me.handleContentSuccess(response);
        },
        failure: function(response) {
            me.handleContentLoadError(response);
        }
    });
}
```

## Migration from Multiple Methods

### Before: Multiple specialized methods
```javascript
// Multiple methods handling different scenarios
loadFromGrid: function(selection) {
    me.currentSelection = selection;
    me.setLoading('Loading...');
    me.loadData();
},

loadFromLanguage: function(language) {
    me.currentLanguage = language;
    me.setLoading('Loading...');
    me.loadData();
},

loadFromBoth: function(selection, language) {
    me.currentSelection = selection;
    me.currentLanguage = language;
    me.setLoading('Loading...');
    me.loadData();
}
```

### After: Unified method
```javascript
// Single method handling all scenarios
displayContent: function(selection, language) {
    var me = this;
    
    // Update state based on what's provided
    if (selection) me.componentData.currentSelection = selection;
    if (language) me.componentData.currentLanguage = language;
    
    // Consistent loading and data loading
    me.setLoading(app.text.loadingText());
    me.loadData();
}
```

## Best Practices

### 1. Parameter Design
- Use **optional parameters** for flexibility
- **Preserve existing state** when parameters are null/undefined
- **Validate required combinations** before proceeding

### 2. State Management
- **Centralize state updates** in the unified method
- **Use consistent property names** across the component
- **Document state dependencies** clearly

### 3. Loading States
- **Always show loading** for async operations
- **Use translatable text** for loading messages
- **Clear loading states** on both success and failure

### 4. Error Recovery
- **Preserve valid state** when errors occur
- **Provide meaningful error messages** using translatable text
- **Allow user to retry** failed operations

This pattern ensures consistent, maintainable, and user-friendly content display across all TDS Suite components.