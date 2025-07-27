# Text Translation Patterns

## Overview
All user-facing text in the TDS Suite must use the centralized translation system located in `@source/ui/app/text.js`. This ensures consistent translations and eliminates duplicate text strings across the application.

## Core Principles

### 1. Always Use app.text Functions
Every piece of user-facing text must go through the `app.text` system:

```javascript
// ✅ CORRECT: Use app.text functions
text: app.text.loadingText()

// ❌ INCORRECT: Hardcoded strings
text: 'Loading...'
```

### 2. Check Existing Functions First
Before creating new text functions, always check `@source/ui/app/text.js` for existing functions:

```javascript
// Common existing functions to reuse:
app.text.loadingText()           // "Loading..."
app.text.saveButtonText()        // "Save"
app.text.cancelButtonText()      // "Cancel"
app.text.errorDialogTitle()      // "Error"
app.text.successDialogTitle()    // "Success"
```

### 3. Organized Function Structure
Text functions are organized by feature/component areas:

```javascript
// General functions (top level)
loadingText: function() { return 'Loading...'; },

// Component-specific functions (nested)
ndaArchive: {
    selectLanguageText: function() { return 'Select Language'; },
    noLanguagesText: function() { return 'No languages available'; }
},

visitor: {
    checkInButtonText: function() { return 'Check In'; },
    checkOutButtonText: function() { return 'Check Out'; }
}
```

## Implementation Patterns

### ExtJS Button Text
```javascript
{
    xtype: 'button',
    text: app.text.saveButtonText(),
    tooltip: app.text.saveButtonTooltip(),
    iconCls: 'save-green-button'
}
```

### ExtJS Grid Columns
```javascript
{
    text: app.text.nameColumnHeader(),
    dataIndex: 'name',
    flex: 1
}
```

### Loading Masks
```javascript
// Component loading
me.setLoading(app.text.loadingText());

// Panel loading with custom message
me.showLoadingOnPreviewPanel(app.text.ndaArchive.loadingContentText());
```

### Dialog Messages
```javascript
Ext.Msg.alert(
    app.text.successDialogTitle(), 
    app.text.ndaArchive.languageSelectedText()
);

Ext.Msg.confirm(
    app.text.confirmDialogTitle(),
    app.text.ndaArchive.confirmDeleteText(),
    function(btn) { /* handler */ }
);
```

### Menu Items
```javascript
{
    text: app.text.ndaArchive.selectLanguageText(),
    handler: function() { /* handler */ }
}
```

## Creating New Text Functions

### Function Naming Convention
- Use descriptive names ending with `Text()` for display text
- Use descriptive names ending with `Tooltip()` for tooltips
- Use descriptive names ending with `Title()` for dialog titles

### Adding New Functions
1. **Determine the appropriate section** in `text.js`
2. **Check for similar existing functions** to maintain consistency
3. **Follow the established naming pattern**
4. **Group related functions together**

```javascript
// Example: Adding new NDA Archive functions
ndaArchive: {
    // Existing functions...
    selectLanguageText: function() { return 'Select Language'; },
    
    // New functions - maintain consistency
    languageLoadingText: function() { return 'Loading languages...'; },
    switchLanguageTooltip: function() { return 'Switch to this language'; },
    noContentAvailableText: function() { return 'No content available for this language'; }
}
```

## Component-Specific Examples

### NDA Archive Widget
```javascript
// Language dropdown
{
    text: app.text.ndaArchive.selectLanguageText(),
    menu: {
        items: languageMenuItems.map(function(lang) {
            return {
                text: lang.language_description,
                handler: function() {
                    me.displayContent(null, lang);
                }
            };
        })
    }
}

// Loading states
me.showLoadingOnPreviewPanel(app.text.ndaArchive.loadingContentText());

// Error messages
Ext.Msg.alert(
    app.text.errorDialogTitle(), 
    app.text.ndaArchive.loadContentErrorText()
);
```

### Visitor Management
```javascript
// Check-in button
{
    text: app.text.visitor.checkInButtonText(),
    tooltip: app.text.visitor.checkInTooltip(),
    handler: me.performCheckIn
}

// Status updates
me.updateStatusMessage(app.text.visitor.checkInSuccessText());
```

## Best Practices

### 1. Consistency
- Use the same text function for the same concept across all components
- Maintain consistent naming patterns
- Group related functions logically

### 2. Reusability
- Create general functions for common concepts (Save, Cancel, Loading, etc.)
- Create specific functions for component-unique text
- Avoid duplicating similar text with slight variations

### 3. Maintainability
- Keep functions organized by feature/component
- Use descriptive function names
- Document complex or non-obvious text requirements

### 4. Performance
- Text functions are lightweight and cached
- No performance concerns with frequent usage
- Prefer function calls over string concatenation

## Migration Strategy

### For Existing Components
1. **Identify hardcoded strings** in component
2. **Check existing text functions** for matches
3. **Create new functions** if needed
4. **Replace strings** with function calls
5. **Test functionality** remains unchanged

### Example Migration
```javascript
// BEFORE: Hardcoded strings
{
    text: 'Select Language',
    tooltip: 'Choose a language from the dropdown'
}

// AFTER: Using app.text functions
{
    text: app.text.ndaArchive.selectLanguageText(),
    tooltip: app.text.ndaArchive.selectLanguageTooltip()
}
```

## Common Anti-Patterns to Avoid

### ❌ Don't Hardcode Strings
```javascript
// Wrong
text: 'Loading...'
title: 'Error'
```

### ❌ Don't Duplicate Functions
```javascript
// Wrong - creating duplicates
loadingDataText: function() { return 'Loading...'; }
loadingContentText: function() { return 'Loading...'; }

// Right - reuse existing
text: app.text.loadingText()  // Use existing general function
```

### ❌ Don't Concatenate When Possible
```javascript
// Less ideal
text: app.text.loadingText() + ' ' + objectType

// Better - create specific function
text: app.text.ndaArchive.loadingContentText()  // "Loading content..."
```

## Integration with Translation Framework

The `app.text` functions integrate with the automatic translation framework:

1. **Functions return base language strings** (typically English)
2. **Translation framework intercepts** function calls
3. **Automatic translation** based on user's language preference
4. **Fallback to base language** if translation not available

This system ensures:
- **Centralized text management**
- **Automatic language switching**
- **Consistent user experience**
- **Easy maintenance and updates**