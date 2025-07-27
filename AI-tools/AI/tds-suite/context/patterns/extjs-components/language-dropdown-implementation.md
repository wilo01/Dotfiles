# Language Dropdown Implementation Pattern

## Overview
Pattern for implementing clickable language dropdown components in ExtJS widgets that fetch available languages from REST endpoints and update content dynamically.

## Context
Used in NDA archive widgets and portlets where users need to switch between different language versions of content. Based on Figma design mockups with specific visual requirements.

## Implementation

### Component Structure
```javascript
// Three-component layout matching Figma design
{
   xtype: 'container',
   layout: { type: 'hbox', align: 'middle' },
   items: [
      {
         xtype: 'component',
         cls: 'language cursor-pointer',
         width: 20, height: 20,
         itemId: 'languageIcon',
         listeners: {
            afterrender: function(component) {
               component.getEl().on('click', function(e) {
                  me.showLanguageDropdown(component, e);
               });
               me.setupLanguageTooltip(component);
            }
         }
      },
      {
         xtype: 'component',
         itemId: 'languageText',
         cls: 'nda-language-text',
         listeners: {
            afterrender: function(component) {
               component.getEl().on('click', function(e) {
                  me.showLanguageDropdown(component, e);
               });
               component.getEl().setStyle('cursor', 'pointer');
            }
         }
      },
      {
         xtype: 'component',
         cls: 'language-caret',
         width: 12, height: 20,
         listeners: {
            afterrender: function(component) {
               component.getEl().on('click', function(e) {
                  me.showLanguageDropdown(component, e);
               });
               component.getEl().setStyle('cursor', 'pointer');
            }
         }
      }
   ]
}
```

### API Integration Pattern
```javascript
showLanguageDropdown: function(component, event) {
   const me = this;
   const requestUrl = '/' + app.deploy_context + '/getNdaTranslationLanguages';
   const numericNdaCode = Number(me.currentNdaCode) || null;

   Ext.Ajax.request({
      url: requestUrl,
      method: 'GET',
      params: { nda_code: numericNdaCode },
      success: function(response, opts) {
         try {
            const data = safe_util.getResponseObject(response);
            if (data && data.length > 0) {
               me.showLanguageMenu(component, event, data);
            } else {
               safe_util.toastNotify('Info', 'No language versions available', true);
            }
         } catch (e) {
            console.error('Error parsing language data:', e);
            safe_util.toastNotify('Error', 'Failed to load languages', false);
         }
      },
      failure: function(response, opts) {
         console.error('Failed to fetch languages:', response);
         safe_util.toastNotify('Error', 'Failed to load languages', false);
      }
   });
}
```

### Menu Creation Pattern
```javascript
showLanguageMenu: function(component, event, languages) {
   const menuItems = [];
   
   Ext.Array.each(languages, function(lang) {
      menuItems.push({
         text: safe_util.escapeHtmlXSS(lang.language_description),
         iconCls: 'language-active', // Green dot for all languages per Figma
         langData: lang,
         handler: function(item) {
            me.selectLanguageFromMenu(item.langData);
            me.currentLanguageMenu.hide();
         }
      });
   });

   me.currentLanguageMenu = Ext.create('Ext.menu.Menu', {
      plain: true, floating: true, shadow: true,
      cls: 'nda-language-dropdown-menu',
      items: menuItems
   });
}
```

## Key Points
- **Three-component structure**: Icon, text, and caret must be separate components for proper styling
- **All clickable**: Each component needs click listeners to trigger dropdown
- **XSS protection**: Always use `safe_util.escapeHtmlXSS()` for user-controlled content
- **Figma compliance**: All languages show with green dots, not just active ones
- **Error handling**: Comprehensive error handling with user-friendly messages
- **Menu cleanup**: Proper cleanup of menu instances to prevent memory leaks

## CSS Requirements
```css
.language { 
   background-image: url("new_icons/language.png");
   filter: brightness(0) saturate(100%) invert(20%) sepia(100%) saturate(5000%) hue-rotate(85deg) brightness(1.2);
}

.nda-language-text {
   color: #1A6A05; font-size: 14px; font-weight: bold;
}

.language-caret {
   background-image: url("data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTAiIGhlaWdodD0iNiIgdmlld0JveD0iMCAwIDEwIDYiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxwYXRoIGQ9Ik01IDZMMCAwSDE0TDUgNloiIGZpbGw9IiMxQTZBMDUiLz4KPC9zdmc+");
}

.language-active:before { content: "●"; color: #4CAF50; margin-right: 8px; }
```

## Related Patterns
- [APEX Permission Check Pattern](../database-integration/apex-permission-check-pattern.md)
- [XSS Protection Pattern](../security-implementations/xss-protection-pattern.md)
- [ExtJS Menu Component Pattern](./extjs-menu-component-pattern.md)

## Examples
- Implemented in `source/ui/app/widgets/visitor/ndaList.js` lines 500-577 and 1186-1414
- Working reference in `source/ui/app/widgets/visitor/ndaArchive.js` lines 972-1049

## Generated On
2024-07-23 - Created from VIS-5029 language dropdown implementation task