# XSS Payload & Validation List

This document provides a list of Cross-Site Scripting (XSS) payloads for testing and validating the security of the TDS Suite application. The primary goal is to ensure that all user inputs are properly sanitized and encoded to prevent XSS vulnerabilities, in accordance with the guidelines in `05_SECURITY_PERMISSIONS.md`.

## XSS Handling and Prevention

As per the project documentation, the primary defense against XSS is:
1.  **Output Encoding**: All user-generated content displayed in the UI must be encoded. The standard method for this in ExtJS is `Ext.String.htmlEncode(userInput)`.
2.  **Content Security Policy (CSP)**: A strict CSP should be implemented to mitigate the impact of any potential injection.
3.  **Input Validation**: Server-side validation should be in place to reject malicious input patterns.

This payload list should be used to verify that these defenses are working correctly.

## Project-Specific XSS Handling

The TDS Suite uses a set of custom utility functions to handle XSS prevention. These functions are located in `source/ui/app/utility.js` and are used throughout the application.

### `util.escapeHtmlXSS(string)`
This is the primary function used to sanitize data before it is rendered in the UI. It should be used on any data that is displayed to the user and may contain untrusted content.

### `util.unEscapeHtmlXSS(string)`
This function is used to decode HTML entities. It should be used with caution and only on data that is known to be safe or has been previously sanitized.

### `util.escapeJsonArray(array)` and `util.escapeJsonObject(object)`
These functions are used to sanitize JSON data before it is processed or rendered in the UI.

### NOTES: Escape javscript injection
```javascript
   util.escapeJsonArray
   util.escapeJsonObject(object)
   util.escapeHtmlXSS(object[key]);
```
