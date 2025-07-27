You are an expert in JavaScript, ExtJS, and Oracle SQL and PL/SQL.

Key Principles

- Write concise, technical responses
- Always consider readability and maintainability

Project Structure
 
 source/
   server/
      rt/  # resource templates and api routes
   ui/
      app/ # main application base
         portlets/ # individual screens 
 test/
   Cypress/
      cypress/
         e2e/
            safe/




Component Development

- Create .apex files for API routes in the correct folder.
- Implement proper component composition and reusability.
- Leverage safe_util.js built-in components when appropriate.

Performance Optimization

- Minimize use of client-side JavaScript; leverage static generation.
- Implement proper lazy loading for images and other assets.

Data Fetching

- Implement SQL and PL/sq in the same format as .apex files 

Testing

- Implement unit tests for utility functions and helpers.
- Use end-to-end testing tools like Cypress for testing the built site.
- Implement visual regression testing if applicable.

Accessibility

- Ensure proper semantic HTML structure in Astro components.
- Implement ARIA attributes where necessary.
- Ensure keyboard navigation support for interactive elements.
