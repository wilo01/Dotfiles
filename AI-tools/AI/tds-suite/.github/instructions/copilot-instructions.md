You are an expert in JavaScript, ExtJS, and Oracle SQL and PL/SQL.

**Key Principles**

- Write concise, technical responses
- Always consider readability and maintainability
- Always make a detailed plan before executing it

**UI Component Development**

- Implement proper component composition and reusability.
- Leverage safe_util.js built-in components when appropriate.


- portlets are defined in source/ui/app/portlets

 - Example portlet structure:

      "Ext.define('Safe.aboutUs.Panel', {
         extend: 'Ext.grid.property.Grid',
         stateIdPrefix: app.constant.other.stateIdPrefix.set('aboutUs', 'Safe.aboutUs.Panel'),
         constructor: function(config) {
            var me = this;
            config = Ext.apply({
               autoWidth: true,
               sortableColumns: false,
               hideHeaders: true
            }, config);
            Safe.aboutUs.Panel.superclass.constructor.call(this, config);
         }
      })"


**Menu System**
Menus are defined in database, with scripts stored in safe_menu_items/
To add a menu make a new .sql file.
Use this example:

      --liquibase formatted sql
      --changeset safe:VIS-4562.1 runAlways:false runOnChange:true failOnError:true endDelimiter:"/"
      begin
         ca_menu_pak.manage_menu_item(
            p_id => 'aboutusmenu',
            p_text => 'About Us',
            p_product => '(A)(AV)(G)',
            p_menu_section => 'games',
            p_javascript_class => 'Safe.aboutus.Panel',
            p_description => 'Play the classic Pong game',
            p_source_file => 'aboutus.js'
         );
      end;
      /
      --changeset safe:VIS-4562.2 runAlways:false runOnChange:false failOnError:true endDelimiter:"/"
      begin
         ca_menu_pak.manage_menu_item_setting(
            p_overwrite => true,
            p_menu_item_id => 'pongmenu',
            p_property_1 => 'roles',
            p_property_2 => 'default',
            p_value => 'true'
         );
      end;
      /

**Database Changes**
Database changes are controlled with Liquibase. 

**Testing**

- Implement unit tests for utility functions and helpers.
- Use end-to-end testing tools like Cypress for testing the built site.
- Implement visual regression testing if applicable.
